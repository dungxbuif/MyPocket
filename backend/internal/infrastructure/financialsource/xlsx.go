package financialsource

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	maxWorkbookBytes             = 20 << 20
	maxArchiveMembers            = 128
	maxArchiveUncompressed       = 32 << 20
	maxArchiveMemberBytes        = 8 << 20
	maxWorksheetRows             = 20_000
	maxWorksheetCells            = 200_000
	maxSharedStrings             = 100_000
	maxCellTextBytes             = 32 << 10
	maxCanonicalAmount     int64 = 9_007_199_254_740_991
)

var errUnsupportedLayout = errors.New("unsupported workbook layout")

type Document struct {
	Name string
	Rows []Row
}

type Row struct {
	SourceID       string
	SourceName     string
	SheetName      string
	RowNumber      int
	OccurredAt     time.Time
	Type           string
	Amount         int64
	SourceCurrency string
	Description    string
	Reference      string
}

type archiveIndex struct {
	files map[string]*zip.File
}

type relationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr"`
}

type relationshipsXML struct {
	Items []relationship `xml:"Relationship"`
}

type contentTypesXML struct {
	Defaults []struct {
		Extension   string `xml:"Extension,attr"`
		ContentType string `xml:"ContentType,attr"`
	} `xml:"Default"`
	Overrides []struct {
		PartName    string `xml:"PartName,attr"`
		ContentType string `xml:"ContentType,attr"`
	} `xml:"Override"`
}

type workbookXML struct {
	Properties struct {
		Date1904 bool `xml:"date1904,attr"`
	} `xml:"workbookPr"`
	Sheets []struct {
		Name  string `xml:"name,attr"`
		RelID string `xml:"id,attr"`
	} `xml:"sheets>sheet"`
}

type styleSheetXML struct {
	NumberFormats []struct {
		ID   int    `xml:"numFmtId,attr"`
		Code string `xml:"formatCode,attr"`
	} `xml:"numFmts>numFmt"`
	CellFormats []struct {
		NumberFormatID int `xml:"numFmtId,attr"`
	} `xml:"cellXfs>xf"`
}

type richText struct {
	Value string
}

func (r *richText) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var parts []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local != "t" {
				continue
			}
			var text string
			if err := decoder.DecodeElement(&text, &typed); err != nil {
				return err
			}
			parts = append(parts, text)
		case xml.EndElement:
			if typed.Name == start.Name {
				r.Value = strings.Join(parts, "")
				return nil
			}
		}
	}
}

type sharedStringsXML struct {
	Items []richText `xml:"si"`
}

type worksheetXML struct {
	Rows []worksheetRow `xml:"sheetData>row"`
}

type worksheetRow struct {
	Number int             `xml:"r,attr"`
	Cells  []worksheetCell `xml:"c"`
}

type formulaMarker struct{}

type worksheetCell struct {
	Reference string         `xml:"r,attr"`
	Type      string         `xml:"t,attr"`
	Style     string         `xml:"s,attr"`
	Value     string         `xml:"v"`
	Inline    richText       `xml:"is"`
	Formula   *formulaMarker `xml:"f"`
}

type cellValue struct {
	text    string
	numeric bool
	style   int
}

type parsedWorksheetRow struct {
	number int
	cells  map[int]cellValue
}

type headerColumns struct {
	date        int
	reference   int
	description int
	debit       int
	credit      int
	currency    int
}

func ParseWorkbook(name string, data []byte, timezone string) (Document, error) {
	if strings.TrimSpace(name) == "" {
		return Document{}, errors.New("workbook name is required")
	}
	location, err := time.LoadLocation(timezone)
	if err != nil || timezone == "Local" {
		return Document{}, errors.New("invalid workbook timezone")
	}
	archive, err := openArchive(data)
	if err != nil {
		return Document{}, err
	}
	if err := archive.rejectActiveRelationships(); err != nil {
		return Document{}, err
	}

	workbookPath, err := archive.validateContentTypes()
	if err != nil {
		return Document{}, err
	}
	workbookBytes, err := archive.read(workbookPath)
	if err != nil {
		return Document{}, err
	}
	var workbook workbookXML
	if err := decodeXML(workbookBytes, &workbook); err != nil || len(workbook.Sheets) == 0 {
		return Document{}, errors.New("invalid workbook metadata")
	}

	relsPath := path.Join(path.Dir(workbookPath), "_rels", path.Base(workbookPath)+".rels")
	workbookRels, err := archive.relationships(relsPath)
	if err != nil {
		return Document{}, err
	}
	relTargets := make(map[string]string, len(workbookRels))
	for _, rel := range workbookRels {
		resolved, resolveErr := resolveRelationshipTarget(workbookPath, rel.Target)
		if resolveErr != nil {
			return Document{}, resolveErr
		}
		relTargets[rel.ID] = resolved
	}

	sharedStrings, err := archive.sharedStrings(path.Dir(workbookPath), workbookRels)
	if err != nil {
		return Document{}, err
	}
	dateStyles, err := archive.dateStyles(path.Dir(workbookPath), workbookRels)
	if err != nil {
		return Document{}, err
	}

	document := Document{Name: name}
	foundLayout := false
	for _, sheet := range workbook.Sheets {
		sheetPath, ok := relTargets[sheet.RelID]
		if !ok || !strings.HasPrefix(sheetPath, path.Dir(workbookPath)+"/") {
			return Document{}, errors.New("invalid worksheet relationship")
		}
		sheetBytes, readErr := archive.read(sheetPath)
		if readErr != nil {
			return Document{}, readErr
		}
		rows, parseErr := parseWorksheet(sheetBytes, sharedStrings)
		if parseErr != nil {
			return Document{}, parseErr
		}
		headers, headerRow, ok, headerErr := findHeaders(rows)
		if headerErr != nil {
			return Document{}, headerErr
		}
		if !ok {
			continue
		}
		foundLayout = true
		parsedRows, parseErr := canonicalRows(name, sheet.Name, rows, headerRow, headers, location, workbook.Properties.Date1904, dateStyles)
		if parseErr != nil {
			return Document{}, parseErr
		}
		document.Rows = append(document.Rows, parsedRows...)
	}
	if !foundLayout {
		return Document{}, errUnsupportedLayout
	}
	return document, nil
}

func openArchive(data []byte) (*archiveIndex, error) {
	if len(data) < 4 || len(data) > maxWorkbookBytes || data[0] != 'P' || data[1] != 'K' {
		return nil, errors.New("invalid XLSX bytes")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, errors.New("invalid XLSX archive")
	}
	if len(reader.File) == 0 || len(reader.File) > maxArchiveMembers {
		return nil, errors.New("XLSX archive member limit exceeded")
	}
	index := &archiveIndex{files: make(map[string]*zip.File, len(reader.File))}
	var total uint64
	for _, file := range reader.File {
		isDirectory := strings.HasSuffix(file.Name, "/")
		archiveName := strings.TrimSuffix(file.Name, "/")
		clean := path.Clean(archiveName)
		if archiveName == "" || clean != archiveName || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") || strings.Contains(file.Name, "\\") {
			return nil, errors.New("invalid XLSX archive path")
		}
		if isDirectory {
			if file.UncompressedSize64 != 0 {
				return nil, errors.New("invalid XLSX directory entry")
			}
			continue
		}
		if _, exists := index.files[clean]; exists {
			return nil, errors.New("duplicate XLSX archive member")
		}
		if file.FileInfo().Mode()&(^file.FileInfo().Mode().Perm()) != 0 && !file.FileInfo().Mode().IsRegular() {
			return nil, errors.New("unsupported XLSX archive member")
		}
		if file.UncompressedSize64 > maxArchiveMemberBytes || total > maxArchiveUncompressed-file.UncompressedSize64 {
			return nil, errors.New("XLSX archive size limit exceeded")
		}
		total += file.UncompressedSize64
		lowerName := strings.ToLower(clean)
		if strings.Contains(lowerName, "vbaproject") || strings.HasSuffix(lowerName, ".xlsm") || strings.HasPrefix(lowerName, "xl/externallinks/") {
			return nil, errors.New("active workbook content is not supported")
		}
		if !allowedArchivePart(clean) {
			return nil, errors.New("unsupported XLSX package part")
		}
		index.files[clean] = file
	}
	return index, nil
}

func (archive *archiveIndex) read(name string) ([]byte, error) {
	file, ok := archive.files[name]
	if !ok {
		return nil, errors.New("required XLSX member is missing")
	}
	reader, err := file.Open()
	if err != nil {
		return nil, errors.New("cannot open XLSX member")
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, maxArchiveMemberBytes+1))
	if err != nil || len(data) > maxArchiveMemberBytes {
		return nil, errors.New("cannot read bounded XLSX member")
	}
	return data, nil
}

func (archive *archiveIndex) relationships(name string) ([]relationship, error) {
	data, err := archive.read(name)
	if err != nil {
		return nil, err
	}
	var parsed relationshipsXML
	if err := decodeXML(data, &parsed); err != nil {
		return nil, errors.New("invalid XLSX relationships")
	}
	return parsed.Items, nil
}

func (archive *archiveIndex) rejectActiveRelationships() error {
	for name := range archive.files {
		if !strings.HasSuffix(strings.ToLower(name), ".rels") {
			continue
		}
		source, ok := relationshipSource(name)
		if !ok {
			return errors.New("unsupported XLSX relationship part")
		}
		rels, err := archive.relationships(name)
		if err != nil {
			return err
		}
		seenIDs := make(map[string]struct{}, len(rels))
		for _, rel := range rels {
			if rel.ID == "" {
				return errors.New("invalid XLSX relationship")
			}
			if _, exists := seenIDs[rel.ID]; exists {
				return errors.New("duplicate XLSX relationship")
			}
			seenIDs[rel.ID] = struct{}{}
			if rel.TargetMode != "" {
				return errors.New("external workbook links are not supported")
			}
			kind, allowed := allowedRelationshipKind(name, rel.Type)
			if !allowed {
				return errors.New("unsupported workbook relationship")
			}
			target, resolveErr := resolveRelationshipTarget(source, rel.Target)
			if resolveErr != nil || !relationshipTargetMatches(kind, target) {
				return errors.New("invalid XLSX relationship target")
			}
		}
	}
	return nil
}

func (archive *archiveIndex) validateContentTypes() (string, error) {
	data, err := archive.read("[Content_Types].xml")
	if err != nil {
		return "", errors.New("invalid XLSX content type")
	}
	var contentTypes contentTypesXML
	if err := decodeXML(data, &contentTypes); err != nil {
		return "", errors.New("invalid XLSX content type")
	}
	const workbookMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
	workbookPath := ""
	for _, defaultType := range contentTypes.Defaults {
		extension := strings.ToLower(defaultType.Extension)
		contentType := strings.ToLower(defaultType.ContentType)
		allowed := extension == "rels" && contentType == "application/vnd.openxmlformats-package.relationships+xml" ||
			extension == "xml" && contentType == "application/xml" ||
			extension == "vml" && contentType == "application/vnd.openxmlformats-officedocument.vmldrawing"
		if !allowed {
			return "", errors.New("unsupported XLSX default content type")
		}
	}
	for _, override := range contentTypes.Overrides {
		partName, partErr := contentTypePartName(override.PartName)
		if partErr != nil || !allowedContentType(partName, override.ContentType) {
			return "", errors.New("unsupported XLSX content type")
		}
		if override.ContentType == workbookMIME {
			if workbookPath != "" {
				return "", errors.New("duplicate XLSX workbook content type")
			}
			workbookPath = partName
		}
	}
	if workbookPath == "" || workbookPath == "." {
		return "", errors.New("invalid XLSX content type")
	}
	rootRels, err := archive.relationships("_rels/.rels")
	if err != nil {
		return "", errors.New("invalid XLSX package relationships")
	}
	foundOfficeDocument := false
	for _, rel := range rootRels {
		if !strings.HasSuffix(strings.ToLower(rel.Type), "/officedocument") {
			continue
		}
		resolved, resolveErr := resolveRelationshipTarget("", rel.Target)
		if resolveErr != nil || resolved != workbookPath {
			return "", errors.New("invalid XLSX workbook relationship")
		}
		foundOfficeDocument = true
	}
	if !foundOfficeDocument {
		return "", errors.New("invalid XLSX workbook relationship")
	}
	return workbookPath, nil
}

func allowedArchivePart(name string) bool {
	switch name {
	case "[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/styles.xml", "xl/sharedStrings.xml", "docProps/core.xml", "docProps/app.xml", "docProps/custom.xml":
		return true
	}
	return numberedXMLPart(name, "xl/worksheets/sheet", ".xml") ||
		numberedXMLPart(name, "xl/theme/theme", ".xml") ||
		numberedXMLPart(name, "xl/worksheets/_rels/sheet", ".xml.rels")
}

func numberedXMLPart(name, prefix, suffix string) bool {
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
		return false
	}
	number := strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix)
	parsed, err := strconv.Atoi(number)
	return err == nil && parsed > 0
}

func relationshipSource(name string) (string, bool) {
	switch name {
	case "_rels/.rels":
		return "", true
	case "xl/_rels/workbook.xml.rels":
		return "xl/workbook.xml", true
	}
	const prefix = "xl/worksheets/_rels/"
	if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".rels") {
		source := "xl/worksheets/" + strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".rels")
		if numberedXMLPart(source, "xl/worksheets/sheet", ".xml") {
			return source, true
		}
	}
	return "", false
}

func allowedRelationshipKind(relationshipPart, relationshipType string) (string, bool) {
	kind := ""
	for _, base := range []string{
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/",
		"http://purl.oclc.org/ooxml/officeDocument/relationships/",
	} {
		if strings.HasPrefix(relationshipType, base) {
			kind = strings.TrimPrefix(relationshipType, base)
			break
		}
	}
	if relationshipType == "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" || relationshipType == "http://purl.oclc.org/ooxml/package/relationships/metadata/core-properties" {
		kind = "core-properties"
	}
	if relationshipPart == "_rels/.rels" {
		switch kind {
		case "officeDocument", "core-properties", "extended-properties", "custom-properties":
			return kind, true
		}
		return "", false
	}
	if relationshipPart == "xl/_rels/workbook.xml.rels" {
		switch kind {
		case "worksheet", "styles", "sharedStrings", "theme":
			return kind, true
		}
	}
	return "", false
}

func relationshipTargetMatches(kind, target string) bool {
	switch kind {
	case "officeDocument":
		return target == "xl/workbook.xml"
	case "core-properties":
		return target == "docProps/core.xml"
	case "extended-properties":
		return target == "docProps/app.xml"
	case "custom-properties":
		return target == "docProps/custom.xml"
	case "worksheet":
		return numberedXMLPart(target, "xl/worksheets/sheet", ".xml")
	case "styles":
		return target == "xl/styles.xml"
	case "sharedStrings":
		return target == "xl/sharedStrings.xml"
	case "theme":
		return numberedXMLPart(target, "xl/theme/theme", ".xml")
	default:
		return false
	}
}

func contentTypePartName(raw string) (string, error) {
	if !strings.HasPrefix(raw, "/") || strings.Contains(raw, "\\") {
		return "", errors.New("invalid content type part")
	}
	trimmed := strings.TrimPrefix(raw, "/")
	if trimmed == "" || path.Clean(trimmed) != trimmed || !allowedArchivePart(trimmed) {
		return "", errors.New("invalid content type part")
	}
	return trimmed, nil
}

func allowedContentType(partName, rawContentType string) bool {
	contentType := strings.ToLower(rawContentType)
	want := ""
	switch partName {
	case "xl/workbook.xml":
		want = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
	case "xl/styles.xml":
		want = "application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"
	case "xl/sharedStrings.xml":
		want = "application/vnd.openxmlformats-officedocument.spreadsheetml.sharedstrings+xml"
	case "docProps/core.xml":
		want = "application/vnd.openxmlformats-package.core-properties+xml"
	case "docProps/app.xml":
		want = "application/vnd.openxmlformats-officedocument.extended-properties+xml"
	case "docProps/custom.xml":
		want = "application/vnd.openxmlformats-officedocument.custom-properties+xml"
	default:
		switch {
		case numberedXMLPart(partName, "xl/worksheets/sheet", ".xml"):
			want = "application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"
		case numberedXMLPart(partName, "xl/theme/theme", ".xml"):
			want = "application/vnd.openxmlformats-officedocument.theme+xml"
		}
	}
	return contentType == strings.ToLower(want) && want != ""
}

func (archive *archiveIndex) sharedStrings(workbookDir string, rels []relationship) ([]string, error) {
	for _, rel := range rels {
		if !strings.HasSuffix(strings.ToLower(rel.Type), "/sharedstrings") {
			continue
		}
		name, err := resolveRelationshipTarget(path.Join(workbookDir, "workbook.xml"), rel.Target)
		if err != nil {
			return nil, err
		}
		data, err := archive.read(name)
		if err != nil {
			return nil, err
		}
		var parsed sharedStringsXML
		if err := decodeXML(data, &parsed); err != nil || len(parsed.Items) > maxSharedStrings {
			return nil, errors.New("invalid or oversized shared strings")
		}
		values := make([]string, len(parsed.Items))
		for i, item := range parsed.Items {
			if len(item.Value) > maxCellTextBytes {
				return nil, errors.New("shared string exceeds cell limit")
			}
			values[i] = item.Value
		}
		return values, nil
	}
	return nil, nil
}

func (archive *archiveIndex) dateStyles(workbookDir string, rels []relationship) (map[int]bool, error) {
	dateStyles := make(map[int]bool)
	for _, rel := range rels {
		if !strings.HasSuffix(strings.ToLower(rel.Type), "/styles") {
			continue
		}
		name, err := resolveRelationshipTarget(path.Join(workbookDir, "workbook.xml"), rel.Target)
		if err != nil {
			return nil, err
		}
		data, err := archive.read(name)
		if err != nil {
			return nil, err
		}
		var styles styleSheetXML
		if err := decodeXML(data, &styles); err != nil || len(styles.CellFormats) > 4096 || len(styles.NumberFormats) > 4096 {
			return nil, errors.New("invalid or oversized workbook styles")
		}
		custom := make(map[int]string, len(styles.NumberFormats))
		for _, format := range styles.NumberFormats {
			custom[format.ID] = format.Code
		}
		for index, format := range styles.CellFormats {
			dateStyles[index] = builtInDateFormat(format.NumberFormatID) || looksLikeDateFormat(custom[format.NumberFormatID])
		}
		return dateStyles, nil
	}
	return dateStyles, nil
}

func resolveRelationshipTarget(source, target string) (string, error) {
	if target == "" || strings.Contains(target, "\\") || strings.Contains(target, "://") || strings.ContainsAny(target, "?#%") {
		return "", errors.New("invalid XLSX relationship target")
	}
	for _, segment := range strings.Split(strings.TrimPrefix(target, "/"), "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", errors.New("invalid XLSX relationship target")
		}
	}
	var resolved string
	if strings.HasPrefix(target, "/") {
		resolved = path.Clean(strings.TrimPrefix(target, "/"))
	} else {
		resolved = path.Clean(path.Join(path.Dir(source), target))
	}
	if resolved == "." || resolved == ".." || strings.HasPrefix(resolved, "../") {
		return "", errors.New("invalid XLSX relationship target")
	}
	return resolved, nil
}

func decodeXML(data []byte, target any) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	return decoder.Decode(target)
}

func parseWorksheet(data []byte, sharedStrings []string) ([]parsedWorksheetRow, error) {
	if hasFormula(data) {
		return nil, errors.New("formula cells are not supported")
	}
	var sheet worksheetXML
	if err := decodeXML(data, &sheet); err != nil {
		return nil, errors.New("invalid worksheet XML")
	}
	if len(sheet.Rows) > maxWorksheetRows {
		return nil, errors.New("worksheet row limit exceeded")
	}
	rows := make([]parsedWorksheetRow, 0, len(sheet.Rows))
	totalCells := 0
	previousRow := 0
	for rowIndex, sourceRow := range sheet.Rows {
		rowNumber := sourceRow.Number
		if rowNumber == 0 {
			rowNumber = rowIndex + 1
		}
		if rowNumber <= previousRow || rowNumber > maxWorksheetRows {
			return nil, errors.New("invalid worksheet row numbering")
		}
		previousRow = rowNumber
		totalCells += len(sourceRow.Cells)
		if totalCells > maxWorksheetCells {
			return nil, errors.New("worksheet cell limit exceeded")
		}
		row := parsedWorksheetRow{number: rowNumber, cells: make(map[int]cellValue, len(sourceRow.Cells))}
		for _, cell := range sourceRow.Cells {
			column, err := cellColumn(cell.Reference)
			if err != nil || column >= 16_384 {
				return nil, errors.New("invalid worksheet cell reference")
			}
			if _, exists := row.cells[column]; exists {
				return nil, errors.New("duplicate worksheet cell reference")
			}
			value, err := decodeCell(cell, sharedStrings)
			if err != nil {
				return nil, err
			}
			row.cells[column] = value
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func hasFormula(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "f" {
			return true
		}
	}
}

func decodeCell(cell worksheetCell, sharedStrings []string) (cellValue, error) {
	if cell.Formula != nil {
		return cellValue{}, errors.New("formula cells are not supported")
	}
	style := 0
	if cell.Style != "" {
		parsed, err := strconv.Atoi(cell.Style)
		if err != nil || parsed < 0 || parsed > 4095 {
			return cellValue{}, errors.New("invalid cell style")
		}
		style = parsed
	}
	value := cellValue{style: style}
	switch cell.Type {
	case "inlineStr":
		value.text = cell.Inline.Value
	case "s":
		index, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return cellValue{}, errors.New("invalid shared string reference")
		}
		value.text = sharedStrings[index]
	case "", "n":
		value.text = strings.TrimSpace(cell.Value)
		value.numeric = value.text != ""
	case "str", "d":
		value.text = cell.Value
	case "b":
		value.text = cell.Value
	case "e":
		return cellValue{}, errors.New("error cells are not supported")
	default:
		return cellValue{}, errors.New("unsupported worksheet cell type")
	}
	if len(value.text) > maxCellTextBytes {
		return cellValue{}, errors.New("cell text exceeds limit")
	}
	return value, nil
}

func cellColumn(reference string) (int, error) {
	if reference == "" {
		return 0, errors.New("missing reference")
	}
	column := 0
	letters := 0
	for _, current := range reference {
		if current >= 'A' && current <= 'Z' {
			column = column*26 + int(current-'A'+1)
			letters++
			continue
		}
		if current >= 'a' && current <= 'z' {
			column = column*26 + int(current-'a'+1)
			letters++
			continue
		}
		break
	}
	if letters == 0 || letters == len(reference) {
		return 0, errors.New("invalid reference")
	}
	return column - 1, nil
}

func findHeaders(rows []parsedWorksheetRow) (headerColumns, int, bool, error) {
	limit := len(rows)
	if limit > 50 {
		limit = 50
	}
	for index := 0; index < limit; index++ {
		headers := headerColumns{date: -1, reference: -1, description: -1, debit: -1, credit: -1, currency: -1}
		recognized := make(map[string]int)
		for column, cell := range rows[index].cells {
			kind := headerKind(cell.text)
			if kind != "" {
				recognized[kind]++
			}
			switch kind {
			case "date":
				headers.date = column
			case "reference":
				headers.reference = column
			case "description":
				headers.description = column
			case "debit":
				headers.debit = column
			case "credit":
				headers.credit = column
			case "currency":
				headers.currency = column
			}
		}
		if headers.date >= 0 && headers.description >= 0 && headers.debit >= 0 && headers.credit >= 0 {
			for kind, count := range recognized {
				if count > 1 {
					return headerColumns{}, 0, false, fmt.Errorf("duplicate recognized workbook header: %s", kind)
				}
			}
			return headers, index, true, nil
		}
	}
	return headerColumns{}, 0, false, nil
}

func headerKind(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.Join(strings.FieldsFunc(normalized, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("_-:/()[]", r)
	}), " ")
	switch normalized {
	case "ngày", "ngay", "ngày giao dịch", "ngay giao dich", "ngày gd", "transaction date", "ngày hạch toán", "ngay hach toan":
		return "date"
	case "chi tiết", "chi tiet", "số tham chiếu", "so tham chieu", "số giao dịch", "so giao dich", "mã giao dịch", "ma giao dich", "reference", "reference number":
		return "reference"
	case "diễn giải", "dien giai", "mô tả", "mo ta", "nội dung giao dịch", "noi dung giao dich", "description", "transaction description":
		return "description"
	case "nợ", "no", "ghi nợ", "ghi no", "số tiền ghi nợ", "so tien ghi no", "debit", "debit amount":
		return "debit"
	case "có", "co", "ghi có", "ghi co", "số tiền ghi có", "so tien ghi co", "credit", "credit amount":
		return "credit"
	case "loại tiền", "loai tien", "currency", "currency code":
		return "currency"
	default:
		return ""
	}
}

func canonicalRows(name, sheetName string, rows []parsedWorksheetRow, headerRow int, headers headerColumns, location *time.Location, date1904 bool, dateStyles map[int]bool) ([]Row, error) {
	result := make([]Row, 0, len(rows)-headerRow-1)
	for _, sourceRow := range rows[headerRow+1:] {
		dateCell := sourceRow.cells[headers.date]
		if strings.TrimSpace(dateCell.text) == "" {
			if strings.TrimSpace(sourceRow.cells[headers.debit].text) != "" || strings.TrimSpace(sourceRow.cells[headers.credit].text) != "" {
				return nil, fmt.Errorf("missing transaction date at sheet %q row %d", sheetName, sourceRow.number)
			}
			continue
		}
		occurredAt, err := parseDate(dateCell, location, date1904, dateStyles)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction date at sheet %q row %d", sheetName, sourceRow.number)
		}
		debitCell := sourceRow.cells[headers.debit]
		creditCell := sourceRow.cells[headers.credit]
		debit, debitPresent, err := parseAmount(debitCell)
		if err != nil {
			return nil, fmt.Errorf("invalid debit amount at sheet %q row %d", sheetName, sourceRow.number)
		}
		credit, creditPresent, err := parseAmount(creditCell)
		if err != nil {
			return nil, fmt.Errorf("invalid credit amount at sheet %q row %d", sheetName, sourceRow.number)
		}
		if debitPresent && debit == 0 && creditPresent && credit == 0 {
			continue
		}
		if debitPresent && debit == 0 {
			debitPresent = false
		}
		if creditPresent && credit == 0 {
			creditPresent = false
		}
		if debitPresent == creditPresent {
			return nil, fmt.Errorf("ambiguous debit and credit at sheet %q row %d", sheetName, sourceRow.number)
		}

		currency := "VND"
		if headers.currency >= 0 {
			if raw := strings.TrimSpace(sourceRow.cells[headers.currency].text); raw != "" {
				currency = strings.ToUpper(raw)
			}
		}
		if currency != "VND" {
			return nil, fmt.Errorf("unsupported currency at sheet %q row %d", sheetName, sourceRow.number)
		}

		row := Row{
			SourceID:       sourceID(name, sheetName, sourceRow.number),
			SourceName:     name,
			SheetName:      sheetName,
			RowNumber:      sourceRow.number,
			OccurredAt:     occurredAt,
			SourceCurrency: currency,
			Description:    strings.TrimSpace(sourceRow.cells[headers.description].text),
		}
		if headers.reference >= 0 {
			row.Reference = strings.TrimSpace(sourceRow.cells[headers.reference].text)
		}
		if debitPresent {
			row.Type = "expense"
			row.Amount = debit
		} else {
			row.Type = "income"
			row.Amount = credit
		}
		result = append(result, row)
	}
	return result, nil
}

func parseAmount(cell cellValue) (int64, bool, error) {
	raw := strings.TrimSpace(strings.ReplaceAll(cell.text, "\u00a0", " "))
	if raw == "" {
		return 0, false, nil
	}
	if cell.numeric {
		rational, ok := new(big.Rat).SetString(raw)
		if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
			return 0, true, errors.New("non-integer numeric amount")
		}
		amount := rational.Num().Int64()
		if amount < 0 || amount > maxCanonicalAmount {
			return 0, true, errors.New("amount out of range")
		}
		return amount, true, nil
	}
	raw = strings.ReplaceAll(raw, " ", "")
	if raw == "" {
		return 0, false, nil
	}
	separator := rune(0)
	for _, current := range raw {
		if current == ',' || current == '.' {
			if separator == 0 {
				separator = current
			} else if separator != current {
				return 0, true, errors.New("mixed amount separators")
			}
		} else if current < '0' || current > '9' {
			return 0, true, errors.New("invalid amount character")
		}
	}
	if separator != 0 {
		parts := strings.Split(raw, string(separator))
		if len(parts) < 2 || len(parts[0]) < 1 || len(parts[0]) > 3 {
			return 0, true, errors.New("invalid grouped amount")
		}
		for _, part := range parts[1:] {
			if len(part) != 3 {
				return 0, true, errors.New("fractional amount is not supported")
			}
		}
		raw = strings.Join(parts, "")
	}
	amount, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || amount < 0 || amount > maxCanonicalAmount {
		return 0, true, errors.New("amount out of range")
	}
	return amount, true, nil
}

func parseDate(cell cellValue, location *time.Location, date1904 bool, dateStyles map[int]bool) (time.Time, error) {
	raw := strings.TrimSpace(cell.text)
	if cell.numeric {
		if !dateStyles[cell.style] && cell.style != 0 {
			return time.Time{}, errors.New("numeric date lacks date style")
		}
		serial, err := strconv.ParseFloat(raw, 64)
		if err != nil || math.IsNaN(serial) || math.IsInf(serial, 0) || serial < 1 || serial > 2_958_465 {
			return time.Time{}, errors.New("invalid Excel date")
		}
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, location)
		if date1904 {
			base = time.Date(1904, 1, 1, 0, 0, 0, 0, location)
		}
		whole, fraction := math.Modf(serial)
		return base.AddDate(0, 0, int(whole)).Add(time.Duration(math.Round(fraction * float64(24*time.Hour)))), nil
	}
	for _, layout := range []string{
		"02/01/2006", "02-01-2006", "2006-01-02", "02/01/2006 15:04:05", "02-01-2006 15:04:05", time.RFC3339,
	} {
		parsed, err := time.ParseInLocation(layout, raw, location)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("unsupported date")
}

func sourceID(name, sheetName string, rowNumber int) string {
	digest := sha256.Sum256([]byte(name + "\x00" + sheetName + "\x00" + strconv.Itoa(rowNumber)))
	return "xlsx_" + hex.EncodeToString(digest[:16])
}

func builtInDateFormat(id int) bool {
	return id >= 14 && id <= 22 || id >= 27 && id <= 36 || id >= 45 && id <= 47 || id >= 50 && id <= 58
}

func looksLikeDateFormat(format string) bool {
	format = strings.ToLower(format)
	var cleaned strings.Builder
	inQuote := false
	for _, current := range format {
		switch current {
		case '"':
			inQuote = !inQuote
		case '\\':
			continue
		default:
			if !inQuote {
				cleaned.WriteRune(current)
			}
		}
	}
	value := cleaned.String()
	return strings.ContainsAny(value, "dy") && strings.Contains(value, "m")
}

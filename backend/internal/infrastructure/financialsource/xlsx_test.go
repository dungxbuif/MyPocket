package financialsource

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"strings"
	"testing"
	"time"
)

func TestParseWorkbookTechcombankHeaders(t *testing.T) {
	data := testWorkbook(t, workbookFixture{
		rows: [][]string{
			{"Ngày giao dịch", "Số tham chiếu", "Diễn giải", "Ghi nợ", "Ghi có", "Số dư"},
			{"26/10/2026", "FT2630012345", "THANH TOAN QR", "1.250.000", "", "8.750.000"},
			{"27/10/2026", "FT2630012346", "HOAN TIEN", "", "250,000", "9.000.000"},
		},
	})

	doc, err := ParseWorkbook("techcombank.xlsx", data, "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatalf("ParseWorkbook() error = %v", err)
	}
	if doc.Name != "techcombank.xlsx" {
		t.Fatalf("Document.Name = %q, want techcombank.xlsx", doc.Name)
	}
	if len(doc.Rows) != 2 {
		t.Fatalf("len(Document.Rows) = %d, want 2", len(doc.Rows))
	}

	wantDates := []string{"2026-10-26T00:00:00+07:00", "2026-10-27T00:00:00+07:00"}
	wantTypes := []string{"expense", "income"}
	wantAmounts := []int64{1250000, 250000}
	for i, row := range doc.Rows {
		if got := row.OccurredAt.Format(time.RFC3339); got != wantDates[i] {
			t.Errorf("Rows[%d].OccurredAt = %q, want %q", i, got, wantDates[i])
		}
		if row.Type != wantTypes[i] {
			t.Errorf("Rows[%d].Type = %q, want %q", i, row.Type, wantTypes[i])
		}
		if row.Amount != wantAmounts[i] {
			t.Errorf("Rows[%d].Amount = %d, want %d", i, row.Amount, wantAmounts[i])
		}
		if row.SourceCurrency != "VND" {
			t.Errorf("Rows[%d].SourceCurrency = %q, want VND", i, row.SourceCurrency)
		}
		if row.SourceName != "techcombank.xlsx" || row.SheetName != "Transactions" || row.RowNumber != i+2 {
			t.Errorf("Rows[%d] source = %q/%q/%d, want techcombank.xlsx/Transactions/%d", i, row.SourceName, row.SheetName, row.RowNumber, i+2)
		}
		if row.SourceID == "" {
			t.Errorf("Rows[%d].SourceID is empty", i)
		}
	}
	if doc.Rows[0].Description != "THANH TOAN QR" || doc.Rows[0].Reference != "FT2630012345" {
		t.Fatalf("first row description/reference = %q/%q", doc.Rows[0].Description, doc.Rows[0].Reference)
	}
}

func TestParseWorkbookTechcombankExportHeaders(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: [][]string{
		{"NGAY", "DIEN GIAI", "CHI TIET", "NO", "CO", "SO DU"},
		{"26/10/2026", "THANH TOAN QR", "FT2630012345", "1,250,000", "", "8,750,000"},
	}})
	document, err := ParseWorkbook("techcombank-export.xlsx", data, "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatalf("ParseWorkbook() error = %v", err)
	}
	if len(document.Rows) != 1 || document.Rows[0].Description != "THANH TOAN QR" || document.Rows[0].Reference != "FT2630012345" {
		t.Fatalf("unexpected Techcombank export row: %+v", document.Rows)
	}
}

func TestParseWorkbookReadsSharedStrings(t *testing.T) {
	data := testSharedStringsWorkbook(t)
	doc, err := ParseWorkbook("shared.xlsx", data, "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatalf("ParseWorkbook() error = %v", err)
	}
	if len(doc.Rows) != 1 {
		t.Fatalf("len(Document.Rows) = %d, want 1", len(doc.Rows))
	}
	row := doc.Rows[0]
	if row.Amount != 1250000 || row.Type != "expense" || row.Reference != "FT2630012345" || row.Description != "THANH TOAN QR" {
		t.Fatalf("row = %#v, want shared-string Techcombank expense", row)
	}
}

func TestParseWorkbookPreservesIdenticalRowsWithStableDistinctSourceIDs(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: [][]string{
		{"Ngày giao dịch", "Số tham chiếu", "Nội dung giao dịch", "Số tiền ghi nợ", "Số tiền ghi có"},
		{"26/10/2026", "same", "same", "50,000", ""},
		{"26/10/2026", "same", "same", "50,000", ""},
	}})

	first, err := ParseWorkbook("duplicate.xlsx", data, "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatalf("first ParseWorkbook() error = %v", err)
	}
	second, err := ParseWorkbook("duplicate.xlsx", data, "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatalf("second ParseWorkbook() error = %v", err)
	}
	if len(first.Rows) != 2 {
		t.Fatalf("len(Rows) = %d, want 2", len(first.Rows))
	}
	if first.Rows[0].SourceID == first.Rows[1].SourceID {
		t.Fatalf("identical source rows share SourceID %q", first.Rows[0].SourceID)
	}
	for i := range first.Rows {
		if first.Rows[i].SourceID != second.Rows[i].SourceID {
			t.Errorf("Rows[%d].SourceID changed between parses: %q != %q", i, first.Rows[i].SourceID, second.Rows[i].SourceID)
		}
	}
}

func TestParseWorkbookRejectsAmbiguousAmounts(t *testing.T) {
	tests := []struct {
		name string
		row  []string
	}{
		{name: "debit and credit", row: []string{"26/10/2026", "ref", "ambiguous", "10,000", "20,000", "VND"}},
		{name: "neither debit nor credit", row: []string{"26/10/2026", "ref", "missing", "", "", "VND"}},
		{name: "fractional VND", row: []string{"26/10/2026", "ref", "fractional", "10.5", "", "VND"}},
		{name: "non VND", row: []string{"26/10/2026", "ref", "foreign", "10", "", "USD"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := testWorkbook(t, workbookFixture{rows: [][]string{
				{"Ngày giao dịch", "Số tham chiếu", "Diễn giải", "Ghi nợ", "Ghi có", "Loại tiền"},
				tt.row,
			}})
			if _, err := ParseWorkbook("unsafe.xlsx", data, "Asia/Ho_Chi_Minh"); err == nil {
				t.Fatal("ParseWorkbook() error = nil, want rejection")
			}
		})
	}
}

func TestParseWorkbookRejectsMonetaryRowWithoutDate(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: [][]string{
		{"NGAY", "DIEN GIAI", "CHI TIET", "NO", "CO"},
		{"", "MISSING DATE", "FT2630012345", "10,000", ""},
	}})
	if _, err := ParseWorkbook("missing-date.xlsx", data, "Asia/Ho_Chi_Minh"); err == nil {
		t.Fatal("ParseWorkbook() error = nil, want blank-date monetary row rejection")
	}
}

func TestParseWorkbookSkipsEmptyAndFooterRowsWithoutDate(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: [][]string{
		{"NGAY", "DIEN GIAI", "CHI TIET", "NO", "CO"},
		{"26/10/2026", "THANH TOAN QR", "FT2630012345", "10,000", ""},
		{"", "", "", "", ""},
		{"", "Generated by Techcombank", "", "", ""},
	}})
	document, err := ParseWorkbook("footer.xlsx", data, "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatalf("ParseWorkbook() error = %v", err)
	}
	if len(document.Rows) != 1 {
		t.Fatalf("len(Document.Rows) = %d, want 1", len(document.Rows))
	}
}

func TestParseWorkbookRejectsDuplicateRecognizedHeaders(t *testing.T) {
	tests := []struct {
		name      string
		duplicate string
		value     string
	}{
		{name: "date", duplicate: "NGAY", value: "26/10/2026"},
		{name: "description", duplicate: "DIEN GIAI", value: "THANH TOAN QR"},
		{name: "debit", duplicate: "NO", value: "10,000"},
		{name: "credit", duplicate: "CO", value: ""},
		{name: "reference", duplicate: "CHI TIET", value: "FT2630012346"},
		{name: "currency", duplicate: "LOAI TIEN", value: "VND"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := testWorkbook(t, workbookFixture{rows: [][]string{
				{"NGAY", "DIEN GIAI", "CHI TIET", "NO", "CO", "LOAI TIEN", tt.duplicate},
				{"26/10/2026", "THANH TOAN QR", "FT2630012345", "10,000", "", "VND", tt.value},
			}})
			if _, err := ParseWorkbook("duplicate-header.xlsx", data, "Asia/Ho_Chi_Minh"); err == nil {
				t.Fatal("ParseWorkbook() error = nil, want duplicate recognized header rejection")
			}
		})
	}
}

func TestParseWorkbookRejectsUnsupportedLayoutAndInvalidEnvelope(t *testing.T) {
	t.Run("unsupported headers", func(t *testing.T) {
		data := testWorkbook(t, workbookFixture{rows: [][]string{{"When", "What", "Value"}, {"today", "coffee", "1"}}})
		if _, err := ParseWorkbook("unknown.xlsx", data, "UTC"); err == nil {
			t.Fatal("ParseWorkbook() error = nil, want unsupported layout")
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		data := testWorkbook(t, workbookFixture{rows: validRows()})
		if _, err := ParseWorkbook("statement.xlsx", data, "Not/A_Zone"); err == nil {
			t.Fatal("ParseWorkbook() error = nil, want invalid timezone rejection")
		}
	})

	t.Run("not zip", func(t *testing.T) {
		if _, err := ParseWorkbook("statement.xlsx", []byte("not an xlsx"), "UTC"); err == nil {
			t.Fatal("ParseWorkbook() error = nil, want byte validation rejection")
		}
	})

	t.Run("zip without OOXML content types", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		writeZipFile(t, zw, "payload.txt", "hello")
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := ParseWorkbook("statement.xlsx", buf.Bytes(), "UTC"); err == nil {
			t.Fatal("ParseWorkbook() error = nil, want MIME rejection")
		}
	})
}

func TestParseWorkbookRejectsActiveContent(t *testing.T) {
	tests := []struct {
		name    string
		fixture workbookFixture
	}{
		{name: "formula", fixture: workbookFixture{rows: validRows(), formulaCell: "D2"}},
		{name: "macro", fixture: workbookFixture{rows: validRows(), macro: true}},
		{name: "renamed macro relationship", fixture: workbookFixture{rows: validRows(), renamedMacro: true}},
		{name: "external relationship", fixture: workbookFixture{rows: validRows(), externalRelationship: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := testWorkbook(t, tt.fixture)
			if _, err := ParseWorkbook("active.xlsx", data, "UTC"); err == nil {
				t.Fatal("ParseWorkbook() error = nil, want active content rejection")
			}
		})
	}
}

func TestParseWorkbookRejectsUnsafeOOXMLPartsAndRelationships(t *testing.T) {
	tests := []struct {
		name    string
		fixture workbookFixture
	}{
		{name: "connections", fixture: workbookFixture{rows: validRows(), extraParts: map[string]string{"xl/connections.xml": "<connections/>"}}},
		{name: "ActiveX", fixture: workbookFixture{rows: validRows(), extraParts: map[string]string{"xl/activeX/activeX1.xml": "<activeX/>"}}},
		{name: "OLE embedding", fixture: workbookFixture{rows: validRows(), extraParts: map[string]string{"xl/embeddings/oleObject1.bin": "ole"}}},
		{name: "custom UI", fixture: workbookFixture{rows: validRows(), extraParts: map[string]string{"customUI/customUI.xml": "<customUI/>"}}},
		{name: "macrosheet", fixture: workbookFixture{rows: validRows(), extraParts: map[string]string{"xl/macrosheets/sheet1.xml": "<worksheet/>"}}},
		{name: "unknown relationship", fixture: workbookFixture{rows: validRows(), extraRelationshipType: "http://schemas.microsoft.com/office/2006/relationships/attachedToolbars"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := testWorkbook(t, tt.fixture)
			if _, err := ParseWorkbook("unsafe-parts.xlsx", data, "UTC"); err == nil {
				t.Fatal("ParseWorkbook() error = nil, want unsafe OOXML rejection")
			}
		})
	}
}

func TestParseWorkbookAllowsSafeThemeStylesAndDocumentProperties(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: validRows(), safeMetadata: true})
	if _, err := ParseWorkbook("safe-parts.xlsx", data, "UTC"); err != nil {
		t.Fatalf("ParseWorkbook() rejected safe OOXML parts: %v", err)
	}
}

func TestParseWorkbookRejectsPathTraversal(t *testing.T) {
	tests := []struct {
		name    string
		fixture workbookFixture
	}{
		{name: "ZIP member", fixture: workbookFixture{rows: validRows(), extraParts: map[string]string{"xl/../evil.xml": "<evil/>"}}},
		{name: "root relationship", fixture: workbookFixture{rows: validRows(), rootWorkbookTarget: "xl/../xl/workbook.xml"}},
		{name: "workbook relationship", fixture: workbookFixture{rows: validRows(), worksheetTarget: "worksheets/../worksheets/sheet1.xml"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := testWorkbook(t, tt.fixture)
			if _, err := ParseWorkbook("traversal.xlsx", data, "UTC"); err == nil {
				t.Fatal("ParseWorkbook() error = nil, want path traversal rejection")
			}
		})
	}
}

func TestParseWorkbookEnforcesArchiveMemberBound(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: validRows(), extraMembers: maxArchiveMembers})
	if _, err := ParseWorkbook("oversized.xlsx", data, "UTC"); err == nil {
		t.Fatal("ParseWorkbook() error = nil, want archive member bound rejection")
	}
}

func TestParseWorkbookEnforcesArchiveMemberByteBound(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: validRows(), extraMemberContents: strings.Repeat("x", maxArchiveMemberBytes+1)})
	if _, err := ParseWorkbook("oversized.xlsx", data, "UTC"); err == nil {
		t.Fatal("ParseWorkbook() error = nil, want archive member byte bound rejection")
	}
}

func TestParseWorkbookAllowsSafeZIPDirectoryEntries(t *testing.T) {
	data := testWorkbook(t, workbookFixture{rows: validRows(), directoryEntries: []string{"_rels/", "xl/", "xl/worksheets/"}})
	if _, err := ParseWorkbook("directories.xlsx", data, "UTC"); err != nil {
		t.Fatalf("ParseWorkbook() rejected safe directory entries: %v", err)
	}
}

func validRows() [][]string {
	return [][]string{
		{"Ngày giao dịch", "Số tham chiếu", "Diễn giải", "Ghi nợ", "Ghi có"},
		{"26/10/2026", "ref", "description", "10,000", ""},
	}
}

type workbookFixture struct {
	rows                  [][]string
	formulaCell           string
	macro                 bool
	renamedMacro          bool
	externalRelationship  bool
	extraMembers          int
	extraMemberContents   string
	directoryEntries      []string
	extraParts            map[string]string
	extraRelationshipType string
	safeMetadata          bool
	rootWorkbookTarget    string
	worksheetTarget       string
}

func testWorkbook(t *testing.T, fixture workbookFixture) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range fixture.directoryEntries {
		if _, err := zw.Create(name); err != nil {
			t.Fatalf("create ZIP directory %q: %v", name, err)
		}
	}

	contentTypes := `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="vml" ContentType="application/vnd.openxmlformats-officedocument.vmlDrawing"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`
	if fixture.macro {
		contentTypes = strings.Replace(contentTypes, "</Types>", `<Override PartName="/xl/vbaProject.bin" ContentType="application/vnd.ms-office.vbaProject"/></Types>`, 1)
	}
	if fixture.renamedMacro {
		contentTypes = strings.Replace(contentTypes, "</Types>", `<Default Extension="bin" ContentType="application/vnd.ms-office.vbaProject"/></Types>`, 1)
	}
	if fixture.safeMetadata {
		contentTypes = strings.Replace(contentTypes, "</Types>", `<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/><Override PartName="/xl/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/><Override PartName="/docProps/custom.xml" ContentType="application/vnd.openxmlformats-officedocument.custom-properties+xml"/></Types>`, 1)
	}
	writeZipFile(t, zw, "[Content_Types].xml", contentTypes)
	rootWorkbookTarget := fixture.rootWorkbookTarget
	if rootWorkbookTarget == "" {
		rootWorkbookTarget = "xl/workbook.xml"
	}
	rootRels := `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="` + html.EscapeString(rootWorkbookTarget) + `"/>`
	if fixture.safeMetadata {
		rootRels += `<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/><Relationship Id="rId4" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/custom-properties" Target="docProps/custom.xml"/>`
	}
	rootRels += `</Relationships>`
	writeZipFile(t, zw, "_rels/.rels", rootRels)
	writeZipFile(t, zw, "xl/workbook.xml", `<?xml version="1.0"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Transactions" sheetId="1" r:id="rId1"/></sheets></workbook>`)
	worksheetTarget := fixture.worksheetTarget
	if worksheetTarget == "" {
		worksheetTarget = "worksheets/sheet1.xml"
	}
	rels := `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="` + html.EscapeString(worksheetTarget) + `"/>`
	if fixture.externalRelationship {
		rels += `<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/externalLink" Target="https://example.invalid/book.xlsx" TargetMode="External"/>`
	}
	if fixture.renamedMacro {
		rels += `<Relationship Id="rId3" Type="http://schemas.microsoft.com/office/2006/relationships/vbaProject" Target="active.bin"/>`
	}
	if fixture.safeMetadata {
		rels += `<Relationship Id="rId4" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rId5" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>`
	}
	if fixture.extraRelationshipType != "" {
		rels += `<Relationship Id="rId9" Type="` + html.EscapeString(fixture.extraRelationshipType) + `" Target="worksheets/sheet1.xml"/>`
	}
	rels += `</Relationships>`
	writeZipFile(t, zw, "xl/_rels/workbook.xml.rels", rels)
	writeZipFile(t, zw, "xl/worksheets/sheet1.xml", fixtureWorksheetXML(fixture.rows, fixture.formulaCell))
	if fixture.macro {
		writeZipFile(t, zw, "xl/vbaProject.bin", "macro")
	}
	if fixture.renamedMacro {
		writeZipFile(t, zw, "xl/active.bin", "macro")
	}
	if fixture.safeMetadata {
		writeZipFile(t, zw, "xl/styles.xml", `<?xml version="1.0"?><styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><cellXfs count="1"><xf numFmtId="0"/></cellXfs></styleSheet>`)
		writeZipFile(t, zw, "xl/theme/theme1.xml", `<?xml version="1.0"?><theme xmlns="http://schemas.openxmlformats.org/drawingml/2006/main"/>`)
		writeZipFile(t, zw, "docProps/core.xml", `<?xml version="1.0"?><coreProperties xmlns="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"/>`)
		writeZipFile(t, zw, "docProps/app.xml", `<?xml version="1.0"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"/>`)
		writeZipFile(t, zw, "docProps/custom.xml", `<?xml version="1.0"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/custom-properties"/>`)
	}
	for name, contents := range fixture.extraParts {
		writeZipFile(t, zw, name, contents)
	}
	for i := 0; i < fixture.extraMembers; i++ {
		writeZipFile(t, zw, fmt.Sprintf("xl/extra/%04d.xml", i), "<x/>")
	}
	if fixture.extraMemberContents != "" {
		writeZipFile(t, zw, "xl/extra/large.xml", fixture.extraMemberContents)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close workbook: %v", err)
	}
	return buf.Bytes()
}

func testSharedStringsWorkbook(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeZipFile(t, zw, "[Content_Types].xml", `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/></Types>`)
	writeZipFile(t, zw, "_rels/.rels", `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`)
	writeZipFile(t, zw, "xl/workbook.xml", `<?xml version="1.0"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Transactions" sheetId="1" r:id="rId1"/></sheets></workbook>`)
	writeZipFile(t, zw, "xl/_rels/workbook.xml.rels", `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/></Relationships>`)
	values := []string{"Ngày giao dịch", "Số tham chiếu", "Diễn giải", "Ghi nợ", "Ghi có", "26/10/2026", "FT2630012345", "THANH TOAN QR", "1.250.000"}
	var shared strings.Builder
	shared.WriteString(`<?xml version="1.0"?><sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	for _, value := range values {
		fmt.Fprintf(&shared, `<si><t>%s</t></si>`, html.EscapeString(value))
	}
	shared.WriteString(`</sst>`)
	writeZipFile(t, zw, "xl/sharedStrings.xml", shared.String())
	writeZipFile(t, zw, "xl/worksheets/sheet1.xml", `<?xml version="1.0"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c><c r="C1" t="s"><v>2</v></c><c r="D1" t="s"><v>3</v></c><c r="E1" t="s"><v>4</v></c></row><row r="2"><c r="A2" t="s"><v>5</v></c><c r="B2" t="s"><v>6</v></c><c r="C2" t="s"><v>7</v></c><c r="D2" t="s"><v>8</v></c><c r="E2" t="inlineStr"><is><t></t></is></c></row></sheetData></worksheet>`)
	if err := zw.Close(); err != nil {
		t.Fatalf("close shared-string workbook: %v", err)
	}
	return buf.Bytes()
}

func fixtureWorksheetXML(rows [][]string, formulaCell string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for rowIndex, values := range rows {
		fmt.Fprintf(&b, `<row r="%d">`, rowIndex+1)
		for columnIndex, value := range values {
			ref := fmt.Sprintf("%s%d", excelColumn(columnIndex), rowIndex+1)
			fmt.Fprintf(&b, `<c r="%s" t="inlineStr">`, ref)
			if ref == formulaCell {
				b.WriteString(`<f>1+1</f>`)
			}
			fmt.Fprintf(&b, `<is><t>%s</t></is></c>`, html.EscapeString(value))
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData></worksheet>`)
	return b.String()
}

func excelColumn(index int) string {
	index++
	var out string
	for index > 0 {
		index--
		out = string(rune('A'+index%26)) + out
		index /= 26
	}
	return out
}

func writeZipFile(t *testing.T, zw *zip.Writer, name, contents string) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create ZIP member %q: %v", name, err)
	}
	if _, err := w.Write([]byte(contents)); err != nil {
		t.Fatalf("write ZIP member %q: %v", name, err)
	}
}

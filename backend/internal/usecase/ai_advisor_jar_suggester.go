package usecase

import (
	"fmt"
	"regexp"
	"strings"
)

// JarSuggestExpense is one unassigned ordinary expense (account-local month).
type JarSuggestExpense struct {
	ID    string
	Note  string
	Month string
}

// JarSuggestHistory is an already-assigned ordinary expense of the same owner.
type JarSuggestHistory struct {
	Note  string
	JarID string
}

// JarSuggestJar is a jar with an ACTIVE config in a given month.
type JarSuggestJar struct{ ID, Name string }

// JarSuggestion is one evidence-backed proposal.
type JarSuggestion struct {
	TransactionID string `json:"transaction_id"`
	JarID         string `json:"jar_id"`
	JarName       string `json:"jar_name"`
	Reason        string `json:"reason"`
	Confidence    string `json:"confidence"` // high | medium
}

// JarSuggestResult separates evidence-backed rows from rows that need the user.
type JarSuggestResult struct {
	Suggested          []JarSuggestion `json:"suggested_assignments"`
	NeedsClarification []string        `json:"needs_clarification"`
}

// JarSuggester proposes jars from evidence only (hashtag, repeated history).
// It is pure: callers pass owner-scoped data and no I/O happens here.
type JarSuggester struct {
	// ActiveJars maps an account-local month (YYYY-MM) to the jars with an active config in it.
	ActiveJars map[string][]JarSuggestJar
	History    []JarSuggestHistory
}

var (
	jarHashtagPattern = regexp.MustCompile(`#([\p{L}\p{N}_-]+)`)
	jarDigitsPattern  = regexp.MustCompile(`[0-9]+(?:[.,/:-][0-9]+)*\s*(?:k|tr|d|đ|vnd|vnđ|ng|nghin|nghìn|trieu|triệu)?`)
	jarNonAlnum       = regexp.MustCompile(`[^a-z0-9]+`)
)

const jarAccentFrom = "àáạảãâầấậẩẫăằắặẳẵèéẹẻẽêềếệểễìíịỉĩòóọỏõôồốộổỗơờớợởỡùúụủũưừứựửữỳýỵỷỹđ"
const jarAccentTo = "aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd"

var jarAccentMap = func() map[rune]rune {
	from, to := []rune(jarAccentFrom), []rune(jarAccentTo)
	m := make(map[rune]rune, len(from)*2)
	for i, r := range from {
		m[r] = to[i]
		upper := []rune(strings.ToUpper(string(r)))
		m[upper[0]] = to[i]
	}
	return m
}()

// NormalizeJarText lowercases, strips Vietnamese accents (including đ/Đ) and
// turns every run of non-alphanumerics into a single '-'.
func NormalizeJarText(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		if mapped, ok := jarAccentMap[r]; ok {
			r = mapped
		}
		b.WriteString(strings.ToLower(string(r)))
	}
	return strings.Trim(jarNonAlnum.ReplaceAllString(b.String(), "-"), "-")
}

// normalizeJarNote also drops amounts, dates and other digits and any hashtags.
func normalizeJarNote(note string) string {
	note = jarHashtagPattern.ReplaceAllString(note, " ")
	note = jarDigitsPattern.ReplaceAllString(strings.ToLower(note), " ")
	return NormalizeJarText(note)
}

// Suggest evaluates the rules for at most MaxJarAssignmentRows expenses.
func (s JarSuggester) Suggest(expenses []JarSuggestExpense) JarSuggestResult {
	result := JarSuggestResult{Suggested: []JarSuggestion{}, NeedsClarification: []string{}}
	if len(expenses) > MaxJarAssignmentRows {
		expenses = expenses[:MaxJarAssignmentRows]
	}
	jarName := map[string]string{}
	for _, jars := range s.ActiveJars {
		for _, jar := range jars {
			jarName[jar.ID] = jar.Name
		}
	}
	byNote := map[string]map[string]int{}
	for _, row := range s.History {
		note := normalizeJarNote(row.Note)
		if note == "" || row.JarID == "" {
			continue
		}
		if byNote[note] == nil {
			byNote[note] = map[string]int{}
		}
		byNote[note][row.JarID]++
	}
	for _, expense := range expenses {
		if suggestion, ok := s.suggestOne(expense, byNote, jarName); ok {
			result.Suggested = append(result.Suggested, suggestion)
		} else {
			result.NeedsClarification = append(result.NeedsClarification, expense.ID)
		}
	}
	return result
}

func (s JarSuggester) suggestOne(expense JarSuggestExpense, byNote map[string]map[string]int, jarName map[string]string) (JarSuggestion, bool) {
	active := s.ActiveJars[expense.Month]
	// Rule a: hashtag equal to an active jar name.
	for _, match := range jarHashtagPattern.FindAllStringSubmatch(expense.Note, -1) {
		tag := NormalizeJarText(match[1])
		for _, jar := range active {
			if tag != "" && tag == NormalizeJarText(jar.Name) {
				return JarSuggestion{TransactionID: expense.ID, JarID: jar.ID, JarName: jar.Name, Reason: fmt.Sprintf("Ghi chú có thẻ #%s khớp hũ %s", match[1], jar.Name), Confidence: "high"}, true
			}
		}
	}
	// Rule b: the normalised note repeats in history and always went to one jar.
	note := normalizeJarNote(expense.Note)
	if note == "" {
		return JarSuggestion{}, false
	}
	counts := map[string]int{}
	for historyNote, jars := range byNote {
		if historyNote == note || (len(note) >= 4 && strings.Contains(historyNote, note)) {
			for jarID, n := range jars {
				counts[jarID] += n
			}
		}
	}
	if len(counts) != 1 {
		return JarSuggestion{}, false
	}
	for jarID, n := range counts {
		name, isActive := "", false
		for _, jar := range active {
			if jar.ID == jarID {
				name, isActive = jar.Name, true
			}
		}
		if n < 2 || !isActive {
			return JarSuggestion{}, false
		}
		confidence := "medium"
		if n >= 3 {
			confidence = "high"
		}
		return JarSuggestion{TransactionID: expense.ID, JarID: jarID, JarName: name, Reason: fmt.Sprintf("%d khoản '%s' trước đây gắn hũ %s", n, strings.TrimSpace(expense.Note), name), Confidence: confidence}, true
	}
	return JarSuggestion{}, false
}

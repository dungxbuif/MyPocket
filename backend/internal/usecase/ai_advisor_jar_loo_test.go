package usecase

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestJarSuggesterLeaveOneOut measures the suggester on real already-assigned
// rows. It needs two TSV exports (JAR_LOO_DIR/tx.tsv: id, month, jar_id, note;
// JAR_LOO_DIR/jars.tsv: month, jar_id, name) and is skipped without them.
// For each assigned expense the jar is hidden and the other rows are history.
func TestJarSuggesterLeaveOneOut(t *testing.T) {
	dir := os.Getenv("JAR_LOO_DIR")
	if dir == "" {
		t.Skip("JAR_LOO_DIR not set")
	}
	read := func(name string) [][]string {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		var out [][]string
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			cols := strings.Split(sc.Text(), "\t")
			for len(cols) < 4 {
				cols = append(cols, "")
			}
			out = append(out, cols)
		}
		return out
	}
	active := map[string][]JarSuggestJar{}
	for _, c := range read("jars.tsv") {
		active[c[0]] = append(active[c[0]], JarSuggestJar{ID: c[1], Name: c[2]})
	}
	tag := regexp.MustCompile(`\s*(\[[^\]]*\]|•\s*Mẫu)\s*$`)
	for _, strip := range []bool{false, true} {
		type row struct{ id, month, jar, note string }
		var rows []row
		for _, c := range read("tx.tsv") {
			if c[2] == "" {
				continue
			}
			n := c[3]
			if strip {
				n = tag.ReplaceAllString(n, "")
			}
			rows = append(rows, row{c[0], c[1], c[2], n})
		}
		stats := map[string][2]int{} // confidence -> {proposed, correct}
		reasons := map[string]int{}
		wrong := map[string]int{}
		for i, r := range rows {
			s := JarSuggester{ActiveJars: active}
			for j, h := range rows {
				if j != i {
					s.History = append(s.History, JarSuggestHistory{Note: h.note, JarID: h.jar})
				}
			}
			res := s.Suggest([]JarSuggestExpense{{ID: r.id, Note: r.note, Month: r.month}})
			if len(res.Suggested) == 0 {
				reasons[looReason(s, r.note, r.month)]++
				continue
			}
			g := res.Suggested[0]
			st := stats[g.Confidence]
			st[0]++
			if g.JarID == r.jar {
				st[1]++
			} else {
				wrong[strings.TrimSpace(r.note)]++
			}
			stats[g.Confidence] = st
		}
		t.Logf("strip_tag=%v rows=%d", strip, len(rows))
		prop := stats["high"][0] + stats["medium"][0]
		t.Logf("  coverage=%d/%d (%.1f%%)", prop, len(rows), 100*float64(prop)/float64(len(rows)))
		for _, k := range []string{"high", "medium"} {
			st := stats[k]
			p := 0.0
			if st[0] > 0 {
				p = 100 * float64(st[1]) / float64(st[0])
			}
			t.Logf("  %s: proposed=%d correct=%d precision=%.1f%%", k, st[0], st[1], p)
		}
		t.Logf("  no-proposal reasons: %v", reasons)
		keys := make([]string, 0, len(wrong))
		for k := range wrong {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(a, b int) bool { return wrong[keys[a]] > wrong[keys[b]] })
		for i, k := range keys {
			if i < 5 {
				t.Logf("  wrong-note %q x%d", k, wrong[k])
			}
		}
	}
}

func looReason(s JarSuggester, rawNote, month string) string {
	note := normalizeJarNote(rawNote)
	if note == "" {
		return "empty-after-normalisation"
	}
	counts := map[string]int{}
	for _, h := range s.History {
		hn := normalizeJarNote(h.Note)
		if hn == note || (len(note) >= 4 && strings.Contains(hn, note)) {
			counts[h.JarID]++
		}
	}
	switch {
	case len(counts) == 0:
		return "no-matching-history"
	case len(counts) > 1:
		return "history-conflicts-between-jars"
	}
	for id, n := range counts {
		on := false
		for _, j := range s.ActiveJars[month] {
			on = on || j.ID == id
		}
		if !on {
			return "true-jar-not-active-in-month"
		}
		if n < 2 {
			return "only-1-matching-row"
		}
	}
	return "other"
}

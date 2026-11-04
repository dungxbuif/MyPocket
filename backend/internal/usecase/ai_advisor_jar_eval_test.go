package usecase

import (
	"fmt"
	"testing"
)

// TestJarSuggesterSyntheticBenchmark measures the deterministic jar suggester on a fixed, labelled set of
// realistic Vietnamese notes. The notes are SYNTHETIC (written for this benchmark, not the owner's data), so the
// numbers show the rules behave as designed, not how they perform on the owner's real notes. It checks three
// things: proposals are never wrong, repeated notes are covered, and anything without evidence is left to the user.
func TestJarSuggesterSyntheticBenchmark(t *testing.T) {
	const month = "2026-11"
	jars := []JarSuggestJar{
		{ID: "food", Name: "Ăn uống"}, {ID: "move", Name: "Đi lại"}, {ID: "home", Name: "Nhà cửa"},
		{ID: "fun", Name: "Giải trí"}, {ID: "health", Name: "Sức khỏe"}, {ID: "learn", Name: "Học tập"},
	}
	// "retired" exists in history but has no active config in the viewed month.
	history := []JarSuggestHistory{
		{"Cà phê sáng 35k", "food"}, {"Cà phê sáng 40k", "food"}, {"Cà phê sáng", "food"},
		{"Ăn trưa văn phòng 55k", "food"}, {"Ăn trưa văn phòng 60k", "food"},
		{"Đi chợ cuối tuần 300k", "food"}, {"Đi chợ cuối tuần 250k", "food"},
		{"Đổ xăng 80k", "move"}, {"Đổ xăng 90k", "move"}, {"Đổ xăng 100k", "move"},
		{"Gửi xe 5k", "move"}, {"Gửi xe 5k", "move"},
		{"Grab đi làm 45k", "move"}, {"Grab đi làm 52k", "move"},
		{"Tiền điện tháng 10", "home"}, {"Tiền điện tháng 9", "home"},
		{"Tiền nước 120k", "home"}, {"Tiền nước 110k", "home"},
		{"Phí chung cư", "home"}, {"Phí chung cư", "home"},
		{"Xem phim 120k", "fun"}, {"Xem phim 150k", "fun"},
		{"Netflix 260k", "fun"}, {"Netflix 260k", "fun"},
		{"Khám răng 500k", "health"}, {"Khám răng 600k", "health"},
		{"Thuốc cảm 90k", "health"}, {"Thuốc cảm 70k", "health"},
		{"Khóa tiếng Anh tháng 11", "learn"}, {"Khóa tiếng Anh tháng 10", "learn"},
		{"Mua sách 180k", "learn"}, {"Mua sách 220k", "learn"},
		// Ambiguous on purpose: the same note went to two different jars.
		{"Đi chơi với bạn 200k", "fun"}, {"Đi chơi với bạn 150k", "food"},
		{"Đi chơi với bạn", "fun"}, {"Đi chơi với bạn", "food"},
		// Only one earlier row: not enough evidence.
		{"Sửa xe máy 400k", "move"},
		// Went to a jar that is not active this month.
		{"Quà tặng sinh nhật 300k", "retired"}, {"Quà tặng sinh nhật 500k", "retired"},
	}
	type tc struct {
		class, note, want string // want "" means the suggester must leave it to the user
	}
	cases := []tc{
		// Same note, different amount/date: must be covered and correct.
		{"repeat", "Cà phê sáng 45k", "food"}, {"repeat", "cà phê sáng", "food"},
		{"repeat", "Ăn trưa văn phòng 58.000", "food"}, {"repeat", "Đi chợ cuối tuần 410k", "food"},
		{"repeat", "Đổ xăng 95k", "move"}, {"repeat", "Gửi xe 5.000đ", "move"}, {"repeat", "Grab đi làm 61k", "move"},
		{"repeat", "Tiền điện tháng 11", "home"}, {"repeat", "Tiền nước 135k", "home"}, {"repeat", "Phí chung cư", "home"},
		{"repeat", "Xem phim 135k", "fun"}, {"repeat", "Netflix 260k", "fun"},
		{"repeat", "Khám răng 450k", "health"}, {"repeat", "Thuốc cảm 85k", "health"},
		{"repeat", "Khóa tiếng Anh tháng 12", "learn"}, {"repeat", "Mua sách 199k", "learn"},
		// No accents or extra case: normalisation must still match.
		{"variant", "do xang 100k", "move"}, {"variant", "CA PHE SANG", "food"}, {"variant", "xem phim", "fun"},
		// Explicit hashtag: always wins and works without accents.
		{"hashtag", "Mua giày #hoc-tap", "learn"},
		{"hashtag", "Tiền gì đó #Sức khỏe", ""}, // a tag stops at the space, so "#Sức" matches no jar
		{"hashtag", "Mua vitamin #suc-khoe", "health"}, {"hashtag", "Vé concert #giai-tri", "fun"}, {"hashtag", "Bún bò #an-uong", "food"},
		// Must abstain: new notes, one-off history, ambiguous history, retired jar, empty/number-only notes.
		{"abstain-new", "Mua laptop mới", ""}, {"abstain-new", "Tiền gửi bố mẹ", ""}, {"abstain-new", "Highlands 65k", ""},
		{"abstain-one-off", "Sửa xe máy 700k", ""},
		{"abstain-ambiguous", "Đi chơi với bạn 300k", ""},
		{"abstain-retired", "Quà tặng sinh nhật 400k", ""},
		{"abstain-empty", "", ""}, {"abstain-empty", "150k", ""},
	}
	suggester := JarSuggester{ActiveJars: map[string][]JarSuggestJar{month: jars}, History: history}
	expenses := make([]JarSuggestExpense, len(cases))
	for i, c := range cases {
		expenses[i] = JarSuggestExpense{ID: fmt.Sprintf("tx-%02d", i), Note: c.note, Month: month}
	}
	got := map[string]string{}
	for _, s := range suggester.Suggest(expenses).Suggested {
		got[s.TransactionID] = s.JarID
	}
	byClass := map[string][3]int{} // total, proposed, correct-or-rightly-abstained
	wrong, proposed, answerable, answered := 0, 0, 0, 0
	for i, c := range cases {
		id := fmt.Sprintf("tx-%02d", i)
		jar, did := got[id]
		stat := byClass[c.class]
		stat[0]++
		if did {
			stat[1]++
			proposed++
		}
		switch {
		case c.want == "" && !did:
			stat[2]++
		case c.want != "" && did && jar == c.want:
			stat[2]++
		default:
			wrong++
			t.Errorf("[%s] %q: want %q, got %q (proposed=%v)", c.class, c.note, c.want, jar, did)
		}
		if c.want != "" {
			answerable++
			if did {
				answered++
			}
		}
		byClass[c.class] = stat
	}
	for class, stat := range byClass {
		t.Logf("%-18s total=%2d proposed=%2d right=%2d", class, stat[0], stat[1], stat[2])
	}
	t.Logf("cases=%d proposals=%d wrong=%d  coverage on answerable=%d/%d", len(cases), proposed, wrong, answered, answerable)
	if wrong != 0 {
		t.Fatalf("the suggester must never propose a wrong jar or fill a row without evidence: %d mistakes", wrong)
	}
}

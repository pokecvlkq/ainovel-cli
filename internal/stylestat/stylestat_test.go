package stylestat

import (
	"strings"
	"testing"
)

func chapterWith(body string) string {
	return "# 标题\n" + body
}

func TestComputeBelowMinChapters(t *testing.T) {
	in := Input{Chapters: []string{"a", "b", "c", "d"}}
	if Compute(in) != nil {
		t.Fatal("below minChapters should return nil")
	}
}

func TestComputePatterns(t *testing.T) {
	body := "Hắn không phải do dự, mà là sợ hãi. Trầm mặc vài nhịp thở. Như một đạo ánh sáng.\nChính văn。\n"
	chapters := make([]string, 6)
	for i := range chapters {
		chapters[i] = chapterWith(body)
	}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	want := map[string]int{
		"Câu chỉnh hướng『không phải… mà là…』":                     6,
		"Từ chỉ thời gian nhanh『nhịp thở/khoảnh khắc』":           6,
		"So sánh tu từ『như một/giống như/tựa như』":               6,
		"Tiết tấu im lặng『im lặng/không nói gì/không ngoảnh lại』": 6,
	}
	for _, p := range s.Patterns {
		if w, ok := want[p.Name]; ok && p.Total != w {
			t.Errorf("%s total: got %d want %d", p.Name, p.Total, w)
		}
		if p.PerChapter != 1.0 {
			t.Errorf("%s per_chapter: got %v want 1.0", p.Name, p.PerChapter)
		}
	}
	if len(s.Patterns) != 4 {
		t.Errorf("want 4 pattern classes, got %d: %+v", len(s.Patterns), s.Patterns)
	}
}

func TestComputeTopPhrasesWithStopwords(t *testing.T) {
	// 「青云山巅」高频出现；「陆九渊」是角色名应被过滤
	line := "众人望向青云山巅，陆九渊负手而立。\n"
	chapters := make([]string, 10)
	for i := range chapters {
		chapters[i] = chapterWith(strings.Repeat(line, 3))
	}
	s := Compute(Input{Chapters: chapters, Stopwords: []string{"陆九渊"}})
	if s == nil {
		t.Fatal("expected stats")
	}
	var hasMountain, hasName bool
	for _, p := range s.TopPhrases {
		if strings.Contains(p.Text, "青云山") {
			hasMountain = true
		}
		if strings.Contains(p.Text, "九渊") || strings.Contains(p.Text, "陆九") {
			hasName = true
		}
	}
	if !hasMountain {
		t.Errorf("expected 青云山 phrase mined, got %+v", s.TopPhrases)
	}
	if hasName {
		t.Errorf("character name should be filtered, got %+v", s.TopPhrases)
	}
}

func TestComputeRepeatedSentences(t *testing.T) {
	motto := "Kiếp này chưa thể đi xa, mong ngươi đi xem non sông xa xôi giúp ta."
	chapters := make([]string, 6)
	for i := range chapters {
		body := "平常正文，没有什么重复。\n"
		if i%2 == 0 {
			body += motto + "\n"
		}
		chapters[i] = chapterWith(body)
	}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	if len(s.RepeatedSentences) == 0 {
		t.Fatalf("expected repeated sentence, got none")
	}
	got := s.RepeatedSentences[0]
	if got.Chapters != 3 || got.Count != 3 {
		t.Errorf("repeated sentence: %+v", got)
	}
	if !strings.HasPrefix(got.Text, "Kiếp này chưa thể đi xa") {
		t.Errorf("text: %q", got.Text)
	}
}

func TestComputeEndingAndOpening(t *testing.T) {
	short := chapterWith("Cả một đêm không ngủ。\nChính văn rất dài rất dài rất dài。\nHắn đi rồi。")
	long := chapterWith("Chuyện ban ngày。\nChính văn。\nĐây là một câu kết vô cùng vô cùng vô cùng dài, vượt xa ngưỡng ba mươi ký tự, dùng để test trung vị。")
	chapters := []string{short, short, short, long, long}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	if s.Ending.ShortRatio != 0.6 {
		t.Errorf("short_ratio: got %v want 0.6", s.Ending.ShortRatio)
	}
	if s.OpeningTimeRate != 0.6 {
		t.Errorf("opening_time_rate: got %v want 0.6", s.OpeningTimeRate)
	}
}

func TestComputeTitleFormats(t *testing.T) {
	chapters := make([]string, 5)
	for i := range chapters {
		chapters[i] = chapterWith("正文。")
	}
	// 混用 → 上报
	s := Compute(Input{Chapters: chapters, Titles: []string{"Chương 1 Phong khởi", "Vân dũng", "Chương 3 Lôi động"}})
	if s.TitleFormats == nil || s.TitleFormats.WithPrefix != 2 || s.TitleFormats.WithoutPrefix != 1 {
		t.Errorf("title formats: %+v", s.TitleFormats)
	}
	// 统一 → 不上报
	s = Compute(Input{Chapters: chapters, Titles: []string{"风起", "云涌"}})
	if s.TitleFormats != nil {
		t.Errorf("uniform titles should not report: %+v", s.TitleFormats)
	}
}

package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/han/go-ecommerce/internal/common/enums"
	"github.com/han/go-ecommerce/internal/learning"
	"golang.org/x/net/html"
)

type bank struct {
	Path          string
	License       enums.LicenseType
	Count         int
	CriticalCount int
}

var defaultBanks = []bank{{"/hang/a1", enums.LicenseA1, 250, 20}, {"/hang/b", enums.LicenseB, 600, 60}}
var chapterNames = []string{
	"Quy định chung và quy tắc giao thông đường bộ",
	"Văn hóa giao thông, đạo đức người lái xe, PCCC và cứu hộ, cứu nạn",
	"Kỹ thuật lái xe",
	"Cấu tạo và sửa chữa",
	"Báo hiệu đường bộ",
	"Giải thế sa hình và kỹ năng xử lý tình huống giao thông",
}
var questionArray = regexp.MustCompile(`"questions"\s*:`)
var flightPush = regexp.MustCompile(`self\.__next_f\.push\((\[.*?\])\)\s*;?`)

type remoteQuestion struct {
	ID       int      `json:"id"`
	Chapter  int      `json:"chapter"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answer   *int     `json:"answer"`
	Image    *string  `json:"image"`
	Critical *bool    `json:"critical"`
	Explain  string   `json:"explain"`
}

// Next.js gửi JSON qua nhiều script. Ghép các text chunk trước khi decode để
// không làm mất câu khi một object bị chia giữa hai chunk. Không chạy JavaScript.
func parseBank(body []byte, b bank) ([]learning.CrawledQuestion, error) {
	z := html.NewTokenizer(bytes.NewReader(body))
	var stream strings.Builder
	inScript := false
	for {
		switch z.Next() {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return nil, z.Err()
			}
			return parseStream(stream.String(), b)
		case html.StartTagToken:
			name, _ := z.TagName()
			inScript = string(name) == "script"
		case html.EndTagToken:
			inScript = false
		case html.TextToken:
			if !inScript {
				continue
			}
			for _, m := range flightPush.FindAllSubmatch(z.Text(), -1) {
				var frame []json.RawMessage
				if err := json.Unmarshal(m[1], &frame); err != nil {
					return nil, fmt.Errorf("invalid page data frame: %w", err)
				}
				if len(frame) != 2 {
					continue
				}
				var kind int
				if json.Unmarshal(frame[0], &kind) != nil || kind != 1 {
					continue
				}
				var chunk string
				if err := json.Unmarshal(frame[1], &chunk); err != nil {
					return nil, err
				}
				stream.WriteString(chunk)
			}
		}
	}
}
func parseStream(stream string, b bank) ([]learning.CrawledQuestion, error) {
	locations := questionArray.FindAllStringIndex(stream, -1)
	if len(locations) != 1 {
		return nil, fmt.Errorf("expected one question array, found %d; source layout changed", len(locations))
	}
	var questions []remoteQuestion
	if err := json.NewDecoder(strings.NewReader(stream[locations[0][1]:])).Decode(&questions); err != nil {
		return nil, fmt.Errorf("decode questions: %w", err)
	}
	if len(questions) != b.Count {
		return nil, fmt.Errorf("%s: expected %d questions, got %d", b.License, b.Count, len(questions))
	}
	seen := map[int]bool{}
	critical := 0
	out := make([]learning.CrawledQuestion, 0, len(questions))
	for _, q := range questions {
		if q.ID < 1 || q.ID > 600 || seen[q.ID] {
			return nil, fmt.Errorf("invalid or duplicate source id %d", q.ID)
		}
		seen[q.ID] = true
		if q.Chapter != chapterFor(q.ID) {
			return nil, fmt.Errorf("question %d has unexpected chapter", q.ID)
		}
		if strings.TrimSpace(q.Question) == "" || strings.TrimSpace(q.Explain) == "" || q.Critical == nil || q.Answer == nil || len(q.Options) < 2 || len(q.Options) > 4 || *q.Answer < 1 || *q.Answer > len(q.Options) {
			return nil, fmt.Errorf("question %d has incomplete content or answer key", q.ID)
		}
		if *q.Critical {
			critical++
		}
		image := ""
		if q.Image != nil && *q.Image != "" {
			base, _ := url.Parse(sourceOrigin)
			ref, err := url.Parse(*q.Image)
			if err != nil {
				return nil, err
			}
			resolved := base.ResolveReference(ref)
			if resolved.Scheme != "https" || resolved.Host != base.Host || resolved.User != nil || !strings.HasPrefix(resolved.Path, "/cau-hoi/") {
				return nil, fmt.Errorf("question %d has unsupported image URL", q.ID)
			}
			image = resolved.String()
		}
		in := learning.QuestionInput{Title: fmt.Sprintf("Câu %d", q.ID), Content: strings.TrimSpace(q.Question), Explanation: strings.TrimSpace(q.Explain), Image: image, LicenseType: b.License, IsCritical: *q.Critical, Published: false}
		unique := map[string]bool{}
		for i, text := range q.Options {
			text = strings.TrimSpace(text)
			key := strings.ToLower(text)
			if text == "" || unique[key] {
				return nil, fmt.Errorf("question %d has empty or duplicate options", q.ID)
			}
			unique[key] = true
			in.Answers = append(in.Answers, learning.AnswerInput{Content: text, Label: string(rune('A' + i)), Order: i + 1, IsCorrect: i+1 == *q.Answer})
		}
		out = append(out, learning.CrawledQuestion{Source: sourceKey, RemoteID: q.ID, SourceURL: fmt.Sprintf("%s/cau-hoi/%d", sourceOrigin, q.ID), CategoryName: fmt.Sprintf("CSGT 2025 - Chương %d: %s", q.Chapter, chapterNames[q.Chapter-1]), Question: in})
	}
	if critical != b.CriticalCount {
		return nil, fmt.Errorf("%s: expected %d critical questions, got %d", b.License, b.CriticalCount, critical)
	}
	return out, nil
}
func chapterFor(id int) int {
	switch {
	case id <= 180:
		return 1
	case id <= 205:
		return 2
	case id <= 263:
		return 3
	case id <= 300:
		return 4
	case id <= 485:
		return 5
	default:
		return 6
	}
}

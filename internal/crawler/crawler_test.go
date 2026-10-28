package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/database"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	"gorm.io/gorm"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type memoryImporter struct {
	calls int
	items []learning.CrawledQuestion
}

func (m *memoryImporter) ImportCrawled(ctx context.Context, items []learning.CrawledQuestion) (learning.CrawlSummary, error) {
	m.calls++
	m.items = items
	return learning.CrawlSummary{Received: len(items), Created: len(items)}, nil
}
func fixtureQuestions(n int) []remoteQuestion {
	out := []remoteQuestion{}
	for i := 1; i <= n; i++ {
		a := 2
		critical := i <= 20 || i >= 251 && i <= 290
		image := "/cau-hoi/demo.webp"
		out = append(out, remoteQuestion{ID: i, Chapter: chapterFor(i), Question: fmt.Sprintf("Demo question %d", i), Options: []string{"Choice A", "Choice B"}, Answer: &a, Critical: &critical, Explain: "Synthetic explanation", Image: &image})
	}
	return out
}
func fixturePage(qs []remoteQuestion) string {
	raw, _ := json.Marshal(map[string]interface{}{"questions": qs})
	// Flight record cố ý bị cắt đôi giữa object JSON và chứa dấu ngoặc trong string.
	stream := "0:" + string(raw) + "\n"
	mid := len(stream) / 2
	var out strings.Builder
	out.WriteString("<html><body>")
	for _, chunk := range []string{stream[:mid], stream[mid:]} {
		frame, _ := json.Marshal([]interface{}{1, chunk})
		out.WriteString("<script>self.__next_f.push(" + string(frame) + ")</script>")
	}
	out.WriteString("</body></html>")
	return out.String()
}
func stubClient(pages map[string]string, calls *atomic.Int32) *http.Client {
	return &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.Scheme != "https" || r.URL.Host != "hoclaixeonline.com" {
			return nil, fmt.Errorf("unexpected URL %s", r.URL)
		}
		if r.Header.Get("User-Agent") != userAgent {
			return nil, fmt.Errorf("missing crawler identity")
		}
		body, ok := pages[r.URL.Path]
		status := 200
		if !ok {
			status = 404
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"text/html; charset=utf-8"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}
}
func validPages() map[string]string {
	return map[string]string{"/robots.txt": "User-agent: *\nAllow: /\nDisallow: /api/\nDisallow: /auth/", "/hang/a1": fixturePage(fixtureQuestions(250)), "/hang/b": fixturePage(fixtureQuestions(600))}
}
func TestCrawlDefaultsAndCooldown(t *testing.T) {
	var calls atomic.Int32
	m := &memoryImporter{}
	s := New(m)
	s.client = stubClient(validPages(), &calls)
	s.interval = 0
	out, err := s.Crawl(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.Created != 850 || out.Published || calls.Load() != 3 || m.calls != 1 {
		t.Fatalf("unexpected result: %+v, requests=%d", out, calls.Load())
	}
	if m.items[0].Question.Answers[0].IsCorrect || !m.items[0].Question.Answers[1].IsCorrect {
		t.Fatal("source one-based answer mapped incorrectly")
	}
	if m.items[0].Question.Image != "https://hoclaixeonline.com/cau-hoi/demo.webp" {
		t.Fatal("image URL not resolved")
	}
	if _, err = s.Crawl(context.Background()); !errors.Is(err, ErrCooldown) {
		t.Fatalf("missing cooldown: %v", err)
	}
	if calls.Load() != 3 {
		t.Fatal("cooldown still fetched source")
	}
}
func TestCrawlFailsBeforeImport(t *testing.T) {
	for _, name := range []string{"missing page", "layout change", "incomplete bank", "bad answer", "inconsistent banks", "robots block", "bad image"} {
		t.Run(name, func(t *testing.T) {
			pages := validPages()
			qs := fixtureQuestions(600)
			switch name {
			case "missing page":
				delete(pages, "/hang/b")
			case "layout change":
				pages["/hang/b"] = "<html>Maintenance</html>"
			case "incomplete bank":
				pages["/hang/b"] = fixturePage(qs[:599])
			case "bad answer":
				v := 0
				qs[0].Answer = &v
				pages["/hang/b"] = fixturePage(qs)
			case "inconsistent banks":
				qs[0].Explain = "Changed explanation"
				pages["/hang/b"] = fixturePage(qs)
			case "robots block":
				pages["/robots.txt"] = "User-agent: *\nDisallow: /hang/"
			case "bad image":
				v := "http://127.0.0.1/private"
				qs[0].Image = &v
				pages["/hang/b"] = fixturePage(qs)
			}
			var calls atomic.Int32
			m := &memoryImporter{}
			s := New(m)
			s.client = stubClient(pages, &calls)
			s.interval = 0
			if _, err := s.Crawl(context.Background()); !errors.Is(err, ErrUpstream) {
				t.Fatalf("expected source failure, got %v", err)
			}
			if m.calls != 0 {
				t.Fatal("partial or invalid bank imported")
			}
		})
	}
}
func TestRobotsRules(t *testing.T) {
	for _, tc := range []struct {
		body, path string
		allowed    bool
	}{
		{"User-agent: *\nAllow: /\nDisallow: /api/", "/hang/b", true},
		{"User-agent: *\nDisallow: /", "/hang/a1", false},
		{"User-agent: *\nDisallow: /hang/*\nAllow: /hang/b$", "/hang/b", true},
		{"User-agent: OtherBot\nDisallow: /\nUser-agent: *\nAllow: /", "/hang/b", true},
		{"User-agent: *\nAllow: /\nUser-agent: DrivingQuestionImporter\nDisallow: /", "/hang/b", false},
	} {
		if got := robotsAllowed(tc.body, tc.path); got != tc.allowed {
			t.Fatalf("robots %q on %q: %v", tc.body, tc.path, got)
		}
	}
}
func TestConcurrentCrawlAndCancellation(t *testing.T) {
	var calls atomic.Int32
	s := New(&memoryImporter{})
	s.client = stubClient(validPages(), &calls)
	s.interval = 0
	s.mu.Lock()
	_, err := s.Crawl(context.Background())
	s.mu.Unlock()
	if !errors.Is(err, ErrBusy) {
		t.Fatal("concurrent crawl not rejected")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = s.Crawl(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("context cancellation lost: %v", err)
	}
}
func TestFetchLimitsAndRedirectPolicy(t *testing.T) {
	s := New(&memoryImporter{})
	var calls atomic.Int32
	s.client.Transport = stubClient(map[string]string{"/hang/b": "too large"}, &calls).Transport
	if _, err := s.fetch(context.Background(), "/hang/b", 3); !errors.Is(err, ErrUpstream) {
		t.Fatal("size limit not applied")
	}
	for _, target := range []string{"http://hoclaixeonline.com/hang/b", "https://example.com/", "https://hoclaixeonline.com:8443/"} {
		req, _ := http.NewRequest("GET", target, nil)
		if err := s.client.CheckRedirect(req, nil); err == nil {
			t.Fatalf("unsafe redirect accepted %s", target)
		}
	}
}
func TestEmptyBodyHTTP(t *testing.T) {
	var calls atomic.Int32
	s := New(&memoryImporter{})
	s.client = stubClient(validPages(), &calls)
	s.interval = 0
	r := gin.New()
	pass := func(c *gin.Context) { c.Next() }
	RegisterRoutes(r, s, pass, pass)
	for _, tc := range []struct {
		path, body string
		status     int
	}{{"/api/v1/admin/questions/crawl?url=x", "", 400}, {"/api/v1/admin/questions/crawl", "{}", 400}, {"/api/v1/admin/questions/crawl", "", 200}} {
		req := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	}
}
func tempDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := database.InitDB(filepath.Join(t.TempDir(), "crawl.db"))
	if err != nil {
		t.Fatal(err)
	}
	sql, _ := db.DB()
	t.Cleanup(func() { _ = sql.Close() })
	return db
}
func TestImportDedupPreservesEditsAndDeletion(t *testing.T) {
	db := tempDB(t)
	s := learning.New(db)
	qs, err := parseBank([]byte(fixturePage(fixtureQuestions(250))), defaultBanks[0])
	if err != nil {
		t.Fatal(err)
	}
	items := qs[:2]
	result, err := s.ImportCrawled(context.Background(), items)
	if err != nil {
		t.Fatal(err)
	}
	if result.Created != 2 || result.CategoriesCreated != 1 {
		t.Fatalf("bad import: %+v", result)
	}
	var q question.Question
	if err = db.Where("source_url = ?", items[0].SourceURL).First(&q).Error; err != nil {
		t.Fatal(err)
	}
	if q.Published || q.SourceURL == "" {
		t.Fatal("draft/provenance missing")
	}
	if err = db.Model(&q).Update("content", "Admin edited").Error; err != nil {
		t.Fatal(err)
	}
	changed := append([]learning.CrawledQuestion(nil), items...)
	changed[0].Question.Content = "New remote content"
	result, err = s.ImportCrawled(context.Background(), changed)
	if err != nil {
		t.Fatal(err)
	}
	if result.Created != 0 || result.Skipped != 2 || result.SourceChanged != 1 {
		t.Fatalf("dedup: %+v", result)
	}
	if err = db.First(&q, q.ID).Error; err != nil {
		t.Fatal(err)
	}
	if q.Content != "Admin edited" {
		t.Fatal("admin edits overwritten")
	}
	if err = db.Delete(&q).Error; err != nil {
		t.Fatal(err)
	}
	result, err = s.ImportCrawled(context.Background(), items)
	if err != nil {
		t.Fatal(err)
	}
	if result.Created != 0 || result.Skipped != 2 {
		t.Fatal("deleted question resurrected")
	}
	var count int64
	db.Model(&question.Question{}).Count(&count)
	if count != 1 {
		t.Fatal("unexpected active count")
	}
}
func TestImportRollback(t *testing.T) {
	db := tempDB(t)
	s := learning.New(db)
	items, err := parseBank([]byte(fixturePage(fixtureQuestions(250))), defaultBanks[0])
	if err != nil {
		t.Fatal(err)
	}
	items = items[:2]
	items[1].Question.Answers[0].IsCorrect = true
	if _, err = s.ImportCrawled(context.Background(), items); err == nil {
		t.Fatal("invalid answer accepted")
	}
	var n int64
	db.Model(&question.Question{}).Count(&n)
	if n != 0 {
		t.Fatal("partial questions committed")
	}
	db.Model(&learning.CrawlRecord{}).Count(&n)
	if n != 0 {
		t.Fatal("partial provenance committed")
	}
}

// Opt-in: tải thật hoặc đọc bản HTML vừa tải để kiểm chứng parser với nguồn hiện
// tại. Luôn ghi DB trong t.TempDir, không đụng cmd/data/gorm.db và không chạy ở CI mặc định.
func TestLiveCrawl(t *testing.T) {
	if os.Getenv("CRAWLER_LIVE") != "1" && os.Getenv("CRAWLER_SNAPSHOT_DIR") == "" {
		t.Skip("set CRAWLER_LIVE=1 or CRAWLER_SNAPSHOT_DIR for source smoke test")
	}
	db := tempDB(t)
	s := New(learning.New(db))
	if dir := os.Getenv("CRAWLER_SNAPSHOT_DIR"); dir != "" {
		files := map[string]string{"/robots.txt": "driving-robots.txt", "/hang/a1": "driving-bank-a1.html", "/hang/b": "driving-bank.html"}
		pages := map[string]string{}
		for path, name := range files {
			b, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			pages[path] = string(b)
		}
		var calls atomic.Int32
		s.client = stubClient(pages, &calls)
		s.interval = 0
	}
	result, err := s.Crawl(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Created != 850 || result.CategoriesCreated != 6 || result.Published {
		t.Fatalf("unexpected live import %+v", result)
	}
	// Bỏ cooldown chỉ trong test để xác nhận retry không tạo trùng.
	s.lastSuccess = time.Time{}
	result, err = s.Crawl(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Created != 0 || result.Skipped != 850 || result.SourceChanged != 0 {
		t.Fatalf("duplicate live import %+v", result)
	}
	t.Log("Verified A1=250, B=600, six categories, 850 drafts; repeat creates zero questions")
}

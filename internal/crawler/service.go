// Package crawler đọc các trang HTML công khai đã định trước, không thực thi
// script và không nhận URL từ client. Cấu trúc nguồn đổi thì dừng, không đoán đáp án.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/han/go-ecommerce/internal/learning"
)

const (
	sourceOrigin = "https://hoclaixeonline.com"
	sourceKey    = "hoclaixeonline-csgt-2025"
	userAgent    = "DrivingQuestionImporter/1.0"
	maxPageBytes = 8 << 20
	cooldown     = time.Minute
)

var (
	ErrBusy     = errors.New("a crawl is already running")
	ErrCooldown = errors.New("please wait 60 seconds after a successful crawl before trying again")
	ErrUpstream = errors.New("crawl source unavailable or incompatible")
)

type Importer interface {
	ImportCrawled(context.Context, []learning.CrawledQuestion) (learning.CrawlSummary, error)
}
type Service struct {
	importer    Importer
	client      *http.Client
	mu          sync.Mutex
	lastSuccess time.Time
	interval    time.Duration
}

func New(importer Importer) *Service {
	return &Service{importer: importer, interval: 500 * time.Millisecond, client: &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Không đi theo redirect sang host tùy ý hoặc hạ HTTPS xuống HTTP.
			if len(via) >= 3 || req.URL.Scheme != "https" || req.URL.Host != "hoclaixeonline.com" || req.URL.User != nil {
				return errors.New("unexpected source redirect")
			}
			return nil
		},
	}}
}

type BankResult struct {
	LicenseType string `json:"license_type"`
	Received    int    `json:"received"`
	URL         string `json:"url"`
}
type Result struct {
	learning.CrawlSummary
	Source       string       `json:"source"`
	Banks        []BankResult `json:"banks"`
	ImageStorage string       `json:"image_storage"`
	Message      string       `json:"message"`
}

func (s *Service) fetch(ctx context.Context, path string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceOrigin+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,text/plain;q=0.9")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrUpstream, path, err)
	}
	defer resp.Body.Close()
	// robots.txt 404 nghĩa là nguồn không công bố quy tắc; các lỗi khác không
	// được dùng làm lý do bỏ qua robots hoặc cố vượt CAPTCHA/đăng nhập.
	if path == "/robots.txt" && resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%w: %s returned HTTP %d", ErrUpstream, path, resp.StatusCode)
	}
	if path != "/robots.txt" && !strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		return nil, fmt.Errorf("%w: expected HTML for %s", ErrUpstream, path)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstream, err)
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("%w: response exceeds size limit", ErrUpstream)
	}
	return body, nil
}
func (s *Service) pause(ctx context.Context) error {
	timer := time.NewTimer(s.interval)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (s *Service) Crawl(ctx context.Context) (*Result, error) {
	if !s.mu.TryLock() {
		return nil, ErrBusy
	}
	defer s.mu.Unlock()
	if !s.lastSuccess.IsZero() && time.Since(s.lastSuccess) < cooldown {
		return nil, ErrCooldown
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	robots, err := s.fetch(ctx, "/robots.txt", 256<<10)
	if err != nil {
		return nil, err
	}
	all := []learning.CrawledQuestion{}
	banks := []BankResult{}
	// Mỗi lần chỉ ba request: robots + hai trang tổng hợp. Không cào 850 trang
	// chi tiết và không tải hàng trăm ảnh; image giữ URL tuyệt đối của nguồn.
	for _, b := range defaultBanks {
		if !robotsAllowed(string(robots), b.Path) {
			return nil, fmt.Errorf("%w: robots.txt disallows %s", ErrUpstream, b.Path)
		}
		if err = s.pause(ctx); err != nil {
			return nil, err
		}
		body, e := s.fetch(ctx, b.Path, maxPageBytes)
		if e != nil {
			return nil, e
		}
		items, e := parseBank(body, b)
		if e != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrUpstream, b.Path, e)
		}
		all = append(all, items...)
		banks = append(banks, BankResult{string(b.License), len(items), sourceOrigin + b.Path})
	}
	// Câu A1 phải thuộc bộ 600 và nội dung/đáp án phải khớp giữa hai trang.
	// Không tự suy luận rằng toàn bộ câu ô tô đều áp dụng cho xe máy.
	byID := map[int]learning.CrawledQuestion{}
	for _, item := range all {
		if item.Question.LicenseType == "B" {
			byID[item.RemoteID] = item
		}
	}
	for _, item := range all {
		if item.Question.LicenseType != "A1" {
			continue
		}
		other, ok := byID[item.RemoteID]
		other.Question.LicenseType = item.Question.LicenseType
		// Nguồn đánh dấu điểm liệt theo từng hạng (A1:20, B:60),
		// nên không ép cờ của câu A1 giống cờ trong bộ ô tô.
		other.Question.IsCritical = item.Question.IsCritical
		if !ok || !reflect.DeepEqual(item, other) {
			return nil, fmt.Errorf("%w: inconsistent A1/B question %d", ErrUpstream, item.RemoteID)
		}
	}
	summary, err := s.importer.ImportCrawled(ctx, all)
	if err != nil {
		return nil, err
	}
	s.lastSuccess = time.Now()
	return &Result{CrawlSummary: summary, Source: sourceKey, Banks: banks, ImageStorage: "remote_url", Message: "Imported as drafts. Review questions before publishing. Existing questions are not overwritten."}, nil
}

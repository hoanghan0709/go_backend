package learning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	category "github.com/han/go-ecommerce/internal/category/model"
	"gorm.io/gorm"
)

// CrawlRecord không xóa cùng câu hỏi: gọi lại crawler không được hồi sinh câu mà
// admin đã chủ động xóa, hoặc ghi đè nội dung đã được admin biên tập.
type CrawlRecord struct {
	ID          uint   `gorm:"primaryKey"`
	Source      string `gorm:"not null;uniqueIndex:idx_crawl_source_question"`
	LicenseType string `gorm:"not null;uniqueIndex:idx_crawl_source_question"`
	RemoteID    int    `gorm:"not null;uniqueIndex:idx_crawl_source_question"`
	QuestionID  uint   `gorm:"not null;index"`
	SourceURL   string
	Fingerprint string `gorm:"not null"`
	CreatedAt   time.Time
}

type CrawledQuestion struct {
	Source       string
	RemoteID     int
	SourceURL    string
	CategoryName string
	Question     QuestionInput
}

type CrawlSummary struct {
	Received          int  `json:"received"`
	Created           int  `json:"created"`
	Skipped           int  `json:"skipped"`
	SourceChanged     int  `json:"source_changed"`
	CategoriesCreated int  `json:"categories_created"`
	Published         bool `json:"published"`
}

// ImportCrawled dùng một transaction cho tất cả các hạng. Nếu một nguồn hỏng,
// caller phải dừng trước bước này; nếu lỗi ghi DB, cả lô được rollback.
func (s *Service) ImportCrawled(ctx context.Context, items []CrawledQuestion) (CrawlSummary, error) {
	result := CrawlSummary{Received: len(items)}
	if len(items) == 0 || len(items) > 2000 {
		return CrawlSummary{}, Invalid("invalid crawl batch size")
	}
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		categories := map[string]uint{}
		seen := map[string]bool{}
		for _, item := range items {
			if item.Source == "" || item.RemoteID < 1 || item.SourceURL == "" || item.CategoryName == "" {
				return Invalid("missing crawl provenance")
			}
			key := fmt.Sprintf("%s/%s/%d", item.Source, item.Question.LicenseType, item.RemoteID)
			if seen[key] {
				return Invalid("duplicate source question in crawl batch")
			}
			seen[key] = true
			// CategoryID được cấp tại DB đích, không phải dữ liệu nguồn để tính hash.
			item.Question.CategoryID = 0
			item.Question.Published = false
			raw, err := json.Marshal(item)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(raw)
			fingerprint := hex.EncodeToString(sum[:])
			var record CrawlRecord
			lookup := tx.Where("source = ? AND license_type = ? AND remote_id = ?", item.Source, item.Question.LicenseType, item.RemoteID).Limit(1).Find(&record)
			if lookup.Error != nil {
				return lookup.Error
			}
			if lookup.RowsAffected > 0 {
				result.Skipped++
				if record.Fingerprint != fingerprint {
					result.SourceChanged++
				}
				continue
			}
			cid, ok := categories[item.CategoryName]
			if !ok {
				var c category.Category
				lookup = tx.Where("name = ?", item.CategoryName).Order("id").Limit(1).Find(&c)
				if lookup.Error != nil {
					return lookup.Error
				}
				if lookup.RowsAffected == 0 {
					c = category.Category{Name: item.CategoryName, Description: "Nhập từ " + item.Source + "; cần rà soát trước khi xuất bản."}
					if err = tx.Create(&c).Error; err != nil {
						return err
					}
					result.CategoriesCreated++
				}
				cid = c.ID
				categories[item.CategoryName] = cid
			}
			item.Question.CategoryID = cid
			q, err := saveQuestion(tx, 0, item.Question)
			if err != nil {
				return fmt.Errorf("source question %s: %w", key, err)
			}
			if err = tx.Model(q).Update("source_url", item.SourceURL).Error; err != nil {
				return err
			}
			record = CrawlRecord{Source: item.Source, LicenseType: string(item.Question.LicenseType), RemoteID: item.RemoteID, QuestionID: q.ID, SourceURL: item.SourceURL, Fingerprint: fingerprint}
			if err = tx.Create(&record).Error; err != nil {
				return err
			}
			result.Created++
		}
		return nil
	})
	if err != nil {
		return CrawlSummary{}, err
	}
	return result, nil
}

package dto

import (
	"github.com/han/go-ecommerce/internal/common/enums"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"time"
)

// HistoryItem chủ động chọn trường công khai, không serialize quan hệ User,
// refresh token hoặc snapshot chứa đáp án trong danh sách lịch sử.
type HistoryItem struct {
	ID            uint             `json:"id"`
	TestID        uint             `json:"test_id"`
	Status        enums.ExamStatus `json:"status"`
	StartedAt     time.Time        `json:"started_at"`
	ExpiresAt     time.Time        `json:"expires_at"`
	FinishedAt    *time.Time       `json:"finished_at,omitempty"`
	Score         int              `json:"score"`
	PassingScore  int              `json:"passing_score"`
	IsPassed      bool             `json:"is_passed"`
	FailureReason string           `json:"failure_reason,omitempty"`
}

func ToHistory(sessions []exam.ExamSession) []HistoryItem {
	out := make([]HistoryItem, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, HistoryItem{ID: s.ID, TestID: s.TestID, Status: s.Status, StartedAt: s.StartedAt, ExpiresAt: s.ExpiresAt, FinishedAt: s.FinishedAt, Score: s.Score, PassingScore: s.PassingScore, IsPassed: s.IsPassed, FailureReason: s.FailureReason})
	}
	return out
}

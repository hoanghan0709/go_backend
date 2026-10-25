package dto

import (
	"github.com/han/go-ecommerce/internal/common/enums"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"time"
)

type View struct {
	ID             uint             `json:"id"`
	TestID         uint             `json:"test_id"`
	Status         enums.ExamStatus `json:"status"`
	StartedAt      time.Time        `json:"started_at"`
	ExpiresAt      time.Time        `json:"expires_at"`
	FinishedAt     *time.Time       `json:"finished_at,omitempty"`
	ServerTime     time.Time        `json:"server_time"`
	CorrectAnswer  int              `json:"correct_answers"`
	TotalQuestions int              `json:"total_questions"`
	Score          int              `json:"score"`
	PassingScore   int              `json:"passing_score"`
	IsPassed       bool             `json:"is_passed"`
	FailureReason  string           `json:"failure_reason,omitempty"`
	Exam           exam.Snapshot    `json:"exam"`
}

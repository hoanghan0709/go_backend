package exam

import (
	"time"

	auth "github.com/han/go-ecommerce/internal/auth/model"
	"github.com/han/go-ecommerce/internal/common/enums"
	common "github.com/han/go-ecommerce/internal/common/model"
	test "github.com/han/go-ecommerce/internal/test/model"
)

type ExamSession struct {
	common.BaseModel
	// Người làm bài
	UserID uint `gorm:"index:idx_exam_user_status_expiry"`
	User   auth.User
	// Đề thi
	TestID *uint
	Test   test.Test
	// Random sessions have no shared Test row; the source template ID is metadata.
	TemplateID uint `json:"template_id,omitempty"`
	// Số câu đúng
	Snapshot      string    `gorm:"type:text" json:"-"`
	ExpiresAt     time.Time `gorm:"index:idx_exam_user_status_expiry"`
	PassingScore  int
	CriticalFail  bool
	FailureReason string
	CorrectAnswer int
	Score         int
	IsPassed      bool

	StartedAt  time.Time
	FinishedAt *time.Time

	// Trạng thái
	Status enums.ExamStatus `gorm:"type:varchar(20);default:'DOING';index:idx_exam_user_status_expiry"`

	// Thời gian làm bài (giây)
	Duration    int
	UserAnswers []UserAnswer
}

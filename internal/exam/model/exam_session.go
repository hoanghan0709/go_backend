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
	UserID uint
	User   auth.User
	// Đề thi
	TestID uint
	Test   test.Test
	// Số câu đúng
	CorrectAnswer int
	Score         int
	IsPassed      bool

	StartedAt  time.Time
	FinishedAt *time.Time

	// Trạng thái
	Status enums.ExamStatus `gorm:"type:varchar(20);default:'DOING'"`

	// Thời gian làm bài (giây)
	Duration    int
	UserAnswers []UserAnswer
}

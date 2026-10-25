package usecase

import exam "github.com/han/go-ecommerce/internal/exam/model"

// Delivery phụ thuộc hành vi nghiệp vụ, không cần biết DB hay cách chấm điểm.
type Interface interface {
	Start(user, testID uint) (*View, error)
	Get(user, id uint) (*View, error)
	SaveAnswer(user, id, questionID, answerID uint) (*View, error)
	Submit(user, id uint) (*View, error)
	History(user uint, page, limit int) ([]exam.ExamSession, int64, error)
	ExpireUser(user uint) error
}

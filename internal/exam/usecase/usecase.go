package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/han/go-ecommerce/internal/exam/dto"
	"time"

	"github.com/han/go-ecommerce/internal/common/enums"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"github.com/han/go-ecommerce/internal/exam/repository"
	"github.com/han/go-ecommerce/internal/learning"
	"github.com/han/go-ecommerce/internal/question/validation"
	test "github.com/han/go-ecommerce/internal/test/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrClosed = errors.New("exam is completed or expired")

type Usecase struct {
	repo repository.Interface
	Now  func() time.Time
}

func New(repo repository.Interface) *Usecase { return &Usecase{repo: repo, Now: time.Now} }

// View giữ tương thích tên trả về của service; DTO thực nằm ở package dto.
type View = dto.View

func decode(session *exam.ExamSession) (exam.Snapshot, error) {
	var snap exam.Snapshot
	err := json.Unmarshal([]byte(session.Snapshot), &snap)
	return snap, err
}
func view(tx *gorm.DB, s *exam.ExamSession, now time.Time) (*View, error) {
	snap, err := decode(s)
	if err != nil {
		return nil, err
	}
	var answers []exam.UserAnswer
	if err = tx.Where("exam_session_id = ?", s.ID).Find(&answers).Error; err != nil {
		return nil, err
	}
	selected := map[uint]exam.UserAnswer{}
	for _, a := range answers {
		selected[a.QuestionID] = a
	}
	for i := range snap.Questions {
		q := &snap.Questions[i]
		a, ok := selected[q.ID]
		if ok {
			q.SelectedAnswerID = a.AnswerID
		}
		if s.Status == enums.ExamDoing {
			q.Explanation = ""
			q.IsCorrect = nil
			for j := range q.Answers {
				q.Answers[j].IsCorrect = nil
			}
		} else {
			correct := ok && a.IsCorrect
			q.IsCorrect = &correct
		}
	}
	return &View{ID: s.ID, TestID: s.TestID, Status: s.Status, StartedAt: s.StartedAt, ExpiresAt: s.ExpiresAt, FinishedAt: s.FinishedAt, ServerTime: now, CorrectAnswer: s.CorrectAnswer, TotalQuestions: len(snap.Questions), Score: s.Score, PassingScore: s.PassingScore, IsPassed: s.IsPassed, FailureReason: s.FailureReason, Exam: snap}, nil
}
func (u *Usecase) Start(user, tid uint) (out *View, err error) {
	err = u.repo.Transaction(func(tx *gorm.DB) error {
		var t test.Test
		if e := tx.Preload("Questions", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).Preload("Questions.Question.Answers", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).Where("published = ?", true).First(&t, tid).Error; e != nil {
			return e
		}
		if len(t.Questions) == 0 {
			return learning.Invalid(fmt.Sprintf("test_id %d: test contains no questions", tid))
		}
		if len(t.Questions) != t.TotalQuestion {
			return learning.Invalid(fmt.Sprintf("test_id %d: total_question=%d but test contains %d questions", tid, t.TotalQuestion, len(t.Questions)))
		}
		if t.Duration < 1 || t.Duration > 180 {
			return learning.Invalid(fmt.Sprintf("test_id %d: duration_minutes must be between 1 and 180", tid))
		}
		if t.PassingScore < 1 || t.PassingScore > len(t.Questions) {
			return learning.Invalid(fmt.Sprintf("test_id %d: passing_score must be between 1 and %d", tid, len(t.Questions)))
		}

		snap := exam.Snapshot{Name: t.Name, LicenseType: t.LicenseType, Version: t.Version}
		seen := map[uint]bool{}
		for _, tq := range t.Questions {
			q := tq.Question
			if q.ID == 0 {
				return learning.Invalid(fmt.Sprintf("test_id %d: question %d does not exist or was deleted", tid, tq.QuestionID))
			}
			if seen[q.ID] {
				return learning.Invalid(fmt.Sprintf("test_id %d: question %d is duplicated", tid, q.ID))
			}
			if !q.Published {
				return learning.Invalid(fmt.Sprintf("test_id %d: question %d is unpublished", tid, q.ID))
			}
			if string(q.LicenseType) != t.LicenseType {
				return learning.Invalid(fmt.Sprintf("test_id %d: question %d has license_type=%s; expected %s", tid, q.ID, q.LicenseType, t.LicenseType))
			}

			seen[q.ID] = true
			if e := validation.Answers(q.Answers); e != nil {
				return learning.Invalid(fmt.Sprintf("test_id %d: question %d: %s", tid, q.ID, e))
			}
			sq := exam.SnapshotQuestion{ID: q.ID, Title: q.Title, Content: q.Content, Image: q.Image, CategoryID: q.CategoryID, IsCritical: q.IsCritical, Explanation: q.Explanation}
			for _, a := range q.Answers {
				correct := a.IsCorrect
				sq.Answers = append(sq.Answers, exam.SnapshotAnswer{ID: a.ID, Content: a.Content, Label: a.Label, Order: a.Order, IsCorrect: &correct})
			}
			snap.Questions = append(snap.Questions, sq)
		}
		encoded, e := json.Marshal(snap)
		if e != nil {
			return e
		}
		now := u.Now()
		s := exam.ExamSession{UserID: user, TestID: tid, Snapshot: string(encoded), StartedAt: now, ExpiresAt: now.Add(time.Duration(t.Duration) * time.Minute), PassingScore: t.PassingScore, CriticalFail: t.CriticalFail, Status: enums.ExamDoing}
		if e = tx.Omit("User", "Test", "UserAnswers").Create(&s).Error; e != nil {
			return e
		}
		out, e = view(tx, &s, now)
		return e
	})
	return
}
func load(tx *gorm.DB, user, id uint) (*exam.ExamSession, error) {
	var s exam.ExamSession
	// Scope theo user ngay trong query: người khác không đọc, sửa hoặc nộp hộ bài.
	err := tx.Where("user_id = ?", user).First(&s, id).Error
	return &s, err
}
func finish(tx *gorm.DB, s *exam.ExamSession, now time.Time, timeout bool) error {
	if s.Status != enums.ExamDoing {
		return nil
	}
	snap, err := decode(s)
	if err != nil {
		return err
	}
	var answers []exam.UserAnswer
	if err = tx.Where("exam_session_id = ?", s.ID).Find(&answers).Error; err != nil {
		return err
	}
	selected := map[uint]exam.UserAnswer{}
	for _, a := range answers {
		selected[a.QuestionID] = a
	}
	criticalWrong := false
	score := 0
	for _, q := range snap.Questions {
		a, ok := selected[q.ID]
		correct := false
		for _, option := range q.Answers {
			if ok && option.ID == a.AnswerID && option.IsCorrect != nil && *option.IsCorrect {
				correct = true
			}
		}
		if correct {
			score++
		} else if q.IsCritical {
			criticalWrong = true
		}
		if ok {
			if err = tx.Model(&exam.UserAnswer{}).Where("id = ?", a.ID).Update("is_correct", correct).Error; err != nil {
				return err
			}
		}
	}
	s.CorrectAnswer = score
	s.Score = score
	s.IsPassed = score >= s.PassingScore && !(s.CriticalFail && criticalWrong)
	if !s.IsPassed {
		if s.CriticalFail && criticalWrong {
			s.FailureReason = "critical_question"
		} else {
			s.FailureReason = "insufficient_score"
		}
	}
	s.Status = enums.ExamCompleted
	end := now
	if timeout {
		s.Status = enums.ExamTimeout
		end = s.ExpiresAt
	}
	s.FinishedAt = &end
	s.Duration = int(end.Sub(s.StartedAt).Seconds())
	// Chấm một lần, kết quả và trạng thái commit cùng nhau. Nộp lại chỉ đọc kết quả.
	return tx.Omit("User", "Test", "UserAnswers").Save(s).Error
}
func (u *Usecase) Get(user, id uint) (out *View, err error) {
	err = u.repo.Transaction(func(tx *gorm.DB) error {
		s, e := load(tx, user, id)
		if e != nil {
			return e
		}
		now := u.Now()
		if s.Status == enums.ExamDoing && !now.Before(s.ExpiresAt) {
			if e = finish(tx, s, now, true); e != nil {
				return e
			}
		}
		out, e = view(tx, s, now)
		return e
	})
	return
}
func (u *Usecase) SaveAnswer(user, id, qid, aid uint) (out *View, err error) {
	closed := false
	err = u.repo.Transaction(func(tx *gorm.DB) error {
		s, e := load(tx, user, id)
		if e != nil {
			return e
		}
		now := u.Now()
		if s.Status == enums.ExamDoing && !now.Before(s.ExpiresAt) {
			if e = finish(tx, s, now, true); e != nil {
				return e
			}
		}
		if s.Status != enums.ExamDoing {
			closed = true
			return nil
		}
		snap, e := decode(s)
		if e != nil {
			return e
		}
		foundQuestion, foundAnswer := false, false
		for _, q := range snap.Questions {
			if q.ID == qid {
				foundQuestion = true
				for _, a := range q.Answers {
					if a.ID == aid {
						foundAnswer = true
					}
				}
			}
		}
		if !foundQuestion {
			return learning.Invalid(fmt.Sprintf("question_id %d does not belong to this exam", qid))
		}
		if aid != 0 && !foundAnswer {
			return learning.Invalid(fmt.Sprintf("answer_id %d does not belong to question %d in this exam; use an answer ID from the exam response", aid, qid))
		}
		// answer_id=0 xóa lựa chọn. Upsert giúp retry lưu đáp án không tạo bản ghi trùng.
		if aid == 0 {
			e = tx.Unscoped().Where("exam_session_id = ? AND question_id = ?", id, qid).Delete(&exam.UserAnswer{}).Error
		} else {
			a := exam.UserAnswer{ExamSessionID: id, QuestionID: qid, AnswerID: aid}
			e = tx.Omit("Question", "Answer").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "exam_session_id"}, {Name: "question_id"}}, DoUpdates: clause.Assignments(map[string]interface{}{"answer_id": aid, "is_correct": false, "updated_at": now})}).Create(&a).Error
		}
		if e != nil {
			return e
		}
		out, e = view(tx, s, now)
		return e
	})
	if err == nil && closed {
		err = ErrClosed
	}
	return
}
func (u *Usecase) Submit(user, id uint) (out *View, err error) {
	err = u.repo.Transaction(func(tx *gorm.DB) error {
		s, e := load(tx, user, id)
		if e != nil {
			return e
		}
		now := u.Now()
		if e = finish(tx, s, now, !now.Before(s.ExpiresAt)); e != nil {
			return e
		}
		out, e = view(tx, s, now)
		return e
	})
	return
}

// ExpireUser hoàn tất bài hết giờ khi đọc lịch sử/thống kê. Kết quả không phụ
// thuộc việc app còn mở hay không; chưa cần một background worker cho MVP.
func (u *Usecase) ExpireUser(user uint) error {
	ids, err := u.repo.ExpiredIDs(user, u.Now())
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := u.Get(user, id); err != nil {
			return err
		}
	}
	return nil
}
func (u *Usecase) History(user uint, page, limit int) ([]exam.ExamSession, int64, error) {
	if err := u.ExpireUser(user); err != nil {
		return nil, 0, err
	}
	page, limit = (learning.Filter{Page: page, Limit: limit}).Bounds()
	return u.repo.History(user, page, limit)
}

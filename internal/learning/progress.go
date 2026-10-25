package learning

import (
	"encoding/json"
	"gorm.io/gorm"
	"sort"
	"time"

	"github.com/han/go-ecommerce/internal/common/enums"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	question "github.com/han/go-ecommerce/internal/question/model"
)

type CategoryProgress struct {
	CategoryID uint    `json:"category_id"`
	Attempts   int     `json:"attempts"`
	Correct    int     `json:"correct"`
	Accuracy   float64 `json:"accuracy"`
}
type ExamTrend struct {
	ID         uint       `json:"id"`
	FinishedAt *time.Time `json:"finished_at"`
	Score      int        `json:"score"`
	Total      int        `json:"total"`
	Passed     bool       `json:"passed"`
}
type Progress struct {
	CompletedExams      int                `json:"completed_exams"`
	PassedExams         int                `json:"passed_exams"`
	PassRate            float64            `json:"pass_rate"`
	Categories          []CategoryProgress `json:"categories"`
	RecentExams         []ExamTrend        `json:"recent_exams"`
	SuggestedCategoryID uint               `json:"suggested_category_id,omitempty"`
	WrongQuestionIDs    []uint             `json:"wrong_question_ids"`
}
type event struct {
	qid, category uint
	correct       bool
	at            time.Time
}

func (s *Service) Progress(user uint, license string) (*Progress, error) {
	out := &Progress{Categories: []CategoryProgress{}, RecentExams: []ExamTrend{}, WrongQuestionIDs: []uint{}}
	var practice []PracticeAttempt
	db := s.DB.Where("user_id = ?", user)
	if license != "" {
		db = db.Where("license_type = ?", license)
	}
	if err := db.Find(&practice).Error; err != nil {
		return nil, err
	}
	events := []event{}
	for _, p := range practice {
		events = append(events, event{p.QuestionID, p.CategoryID, p.IsCorrect, p.CreatedAt})
	}
	var sessions []exam.ExamSession
	if err := s.DB.Where("user_id = ? AND status IN ?", user, []enums.ExamStatus{enums.ExamCompleted, enums.ExamTimeout}).Order("finished_at DESC, id DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}
	for _, session := range sessions {
		var snap exam.Snapshot
		if err := json.Unmarshal([]byte(session.Snapshot), &snap); err != nil {
			return nil, err
		}
		if license != "" && snap.LicenseType != license {
			continue
		}
		out.CompletedExams++
		if session.IsPassed {
			out.PassedExams++
		}
		if len(out.RecentExams) < 20 {
			out.RecentExams = append(out.RecentExams, ExamTrend{session.ID, session.FinishedAt, session.Score, len(snap.Questions), session.IsPassed})
		}
		var answers []exam.UserAnswer
		if err := s.DB.Where("exam_session_id = ?", session.ID).Find(&answers).Error; err != nil {
			return nil, err
		}
		correct := map[uint]bool{}
		for _, a := range answers {
			correct[a.QuestionID] = a.IsCorrect
		}
		at := session.StartedAt
		if session.FinishedAt != nil {
			at = *session.FinishedAt
		}
		for _, q := range snap.Questions {
			events = append(events, event{q.ID, q.CategoryID, correct[q.ID], at})
		}
	}
	categories := map[uint]*CategoryProgress{}
	latest := map[uint]event{}
	for _, e := range events {
		c := categories[e.category]
		if c == nil {
			c = &CategoryProgress{CategoryID: e.category}
			categories[e.category] = c
		}
		c.Attempts++
		if e.correct {
			c.Correct++
		}
		previous, ok := latest[e.qid]
		if !ok || e.at.After(previous.at) {
			latest[e.qid] = e
		}
	}
	for _, c := range categories {
		c.Accuracy = 100 * float64(c.Correct) / float64(c.Attempts)
		out.Categories = append(out.Categories, *c)
	}
	sort.Slice(out.Categories, func(i, j int) bool {
		if out.Categories[i].Accuracy == out.Categories[j].Accuracy {
			return out.Categories[i].CategoryID < out.Categories[j].CategoryID
		}
		return out.Categories[i].Accuracy < out.Categories[j].Accuracy
	})
	if len(out.Categories) > 0 && out.Categories[0].Accuracy < 100 {
		out.SuggestedCategoryID = out.Categories[0].CategoryID
	}
	for qid, e := range latest {
		if !e.correct {
			out.WrongQuestionIDs = append(out.WrongQuestionIDs, qid)
		}
	}
	sort.Slice(out.WrongQuestionIDs, func(i, j int) bool { return out.WrongQuestionIDs[i] < out.WrongQuestionIDs[j] })
	if out.CompletedExams > 0 {
		out.PassRate = 100 * float64(out.PassedExams) / float64(out.CompletedExams)
	}
	return out, nil
}

// WrongQuestions lấy câu có lần làm gần nhất bị sai hoặc bỏ trống. Trả lời đúng
// ở lần ôn tiếp theo sẽ tự bỏ câu khỏi danh sách; chỉ lấy nội dung còn xuất bản.
func (s *Service) WrongQuestions(user uint, f Filter) ([]question.Question, int64, error) {
	p, err := s.Progress(user, f.License)
	if err != nil {
		return nil, 0, err
	}
	out := []question.Question{}
	if len(p.WrongQuestionIDs) == 0 {
		return out, 0, nil
	}
	db := filterQuestions(s.DB.Model(&question.Question{}), f).Where("id IN ? AND published = ?", p.WrongQuestionIDs, true)
	var total int64
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, limit := f.Bounds()
	err = db.Preload("Answers", func(tx *gorm.DB) *gorm.DB { return tx.Order("\"order\" ASC") }).Order("id").Offset((page - 1) * limit).Limit(limit).Find(&out).Error
	return out, total, err
}

package learning

import (
	"encoding/json"
	"sort"
	"time"

	category "github.com/han/go-ecommerce/internal/category/model"
	"gorm.io/gorm"

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
	LicenseType          string             `json:"license_type"`
	TotalQuestions       int                `json:"total_questions"`
	AttemptedQuestions   int                `json:"attempted_questions"`
	MasteredQuestions    int                `json:"mastered_questions"`
	WrongQuestions       int                `json:"wrong_questions"`
	UnattemptedQuestions int                `json:"unattempted_questions"`
	CompletionPercentage float64            `json:"completion_percentage"`
	CompletedExams       int                `json:"completed_exams"`
	PassedExams          int                `json:"passed_exams"`
	PassRate             float64            `json:"pass_rate"`
	Categories           []CategoryProgress `json:"categories"`
	RecentExams          []ExamTrend        `json:"recent_exams"`
	SuggestedCategoryID  uint               `json:"suggested_category_id,omitempty"`
	WrongQuestionIDs     []uint             `json:"wrong_question_ids"`
}
type event struct {
	qid, category uint
	correct       bool
	at            time.Time
	source, id    uint
}

// Equal timestamps prefer exam results, then the larger ID within a source.
func (e event) newerThan(other event) bool {
	return e.at.After(other.at) || (e.at.Equal(other.at) && (e.source > other.source || (e.source == other.source && e.id > other.id)))
}

func (s *Service) Progress(user uint, license string) (*Progress, error) {
	return s.ProgressFiltered(user, Filter{License: license})
}
func (s *Service) ProgressFiltered(user uint, f Filter) (*Progress, error) {
	license := f.License
	out := &Progress{LicenseType: license, Categories: []CategoryProgress{}, RecentExams: []ExamTrend{}, WrongQuestionIDs: []uint{}}
	var practice []PracticeAttempt
	db := s.DB.Where("user_id = ?", user)
	if license != "" {
		db = db.Where("license_type = ?", license)
	}
	if err := db.Find(&practice).Error; err != nil {
		return nil, err
	}
	events := []event{}
	allEvents := []event{}
	var categoryIDs []uint
	if f.CategoryType != "" {
		if err := s.DB.Model(&category.Category{}).Where("type = ?", f.CategoryType).Pluck("id", &categoryIDs).Error; err != nil {
			return nil, err
		}
	}
	matchesCategory := func(id uint) bool {
		if f.CategoryType == "" {
			return true
		}
		for _, candidate := range categoryIDs {
			if candidate == id {
				return true
			}
		}
		return false
	}
	for _, p := range practice {
		e := event{p.QuestionID, p.CategoryID, p.IsCorrect, p.CreatedAt, 0, p.ID}
		allEvents = append(allEvents, e)
		if matchesCategory(p.CategoryID) {
			events = append(events, e)
		}
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
		matchesExam := f.CategoryType == "" || snap.CategoryType == f.CategoryType
		if matchesExam {
			out.CompletedExams++
			if session.IsPassed {
				out.PassedExams++
			}
			if len(out.RecentExams) < 20 {
				out.RecentExams = append(out.RecentExams, ExamTrend{session.ID, session.FinishedAt, session.Score, len(snap.Questions), session.IsPassed})
			}
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
			e := event{q.ID, q.CategoryID, correct[q.ID], at, 1, session.ID}
			allEvents = append(allEvents, e)
			if matchesExam {
				events = append(events, e)
			}
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
		if !ok || e.newerThan(previous) {
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
	// Count distinct questions against the current published bank, regardless of
	// the exam mode in which the learner encountered each question.
	var questionIDs []uint
	bankFilter := f
	published := true
	bankFilter.Published = &published
	if err := filterQuestions(s.DB.Model(&question.Question{}), bankFilter).Pluck("id", &questionIDs).Error; err != nil {
		return nil, err
	}
	latestAll := map[uint]event{}
	for _, e := range allEvents {
		previous, ok := latestAll[e.qid]
		if !ok || e.newerThan(previous) {
			latestAll[e.qid] = e
		}
	}
	out.TotalQuestions = len(questionIDs)
	for _, id := range questionIDs {
		if e, ok := latestAll[id]; ok {
			out.AttemptedQuestions++
			if e.correct {
				out.MasteredQuestions++
			} else {
				out.WrongQuestions++
			}
		}
	}
	out.UnattemptedQuestions = out.TotalQuestions - out.AttemptedQuestions
	if out.TotalQuestions > 0 {
		out.CompletionPercentage = 100 * float64(out.MasteredQuestions) / float64(out.TotalQuestions)
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

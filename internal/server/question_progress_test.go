package server_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/han/go-ecommerce/internal/learning"
	"github.com/han/go-ecommerce/internal/server"
)

func assertQuestionProgress(t *testing.T, p *learning.Progress, total, attempted, mastered, wrong int) {
	t.Helper()
	percent := float64(0)
	if total > 0 {
		percent = 100 * float64(mastered) / float64(total)
	}
	if p.TotalQuestions != total || p.AttemptedQuestions != attempted || p.MasteredQuestions != mastered || p.WrongQuestions != wrong || p.UnattemptedQuestions != total-attempted || p.CompletionPercentage != percent {
		t.Fatalf("unexpected question progress: %+v", p)
	}
}

func TestQuestionProgressLatestAttemptAndCurrentBank(t *testing.T) {
	f := setup(t)
	check := func(total, attempted, mastered, wrong int) {
		t.Helper()
		p, err := f.service.Progress(f.user.ID, "A1")
		requireNoError(t, err)
		assertQuestionProgress(t, p, total, attempted, mastered, wrong)
	}
	check(2, 0, 0, 0)
	q := f.questions[0]
	// Same timestamp: the larger practice ID is the newer result.
	at := time.Now().Add(-time.Hour)
	for _, correct := range []bool{false, true, true, false} {
		requireNoError(t, f.db.Create(&learning.PracticeAttempt{UserID: f.user.ID, QuestionID: q.ID, CategoryID: q.CategoryID, LicenseType: "A1", IsCorrect: correct, CreatedAt: at}).Error)
		mastered, wrong := 0, 1
		if correct {
			mastered, wrong = 1, 0
		}
		check(2, 1, mastered, wrong)
	}
	p, err := f.service.Progress(f.other.ID, "A1")
	requireNoError(t, err)
	assertQuestionProgress(t, p, 2, 0, 0, 0)
	requireNoError(t, f.db.Model(&q).Update("published", false).Error)
	check(1, 0, 0, 0)
	requireNoError(t, f.db.Model(&q).Update("published", true).Error)
	check(2, 1, 0, 1)
	requireNoError(t, f.db.Delete(&q).Error)
	check(1, 0, 0, 0)
	requireNoError(t, f.db.Delete(&f.questions[1]).Error)
	check(0, 0, 0, 0)
}

func TestQuestionProgressAcrossPracticeAndExams(t *testing.T) {
	f := setup(t)
	now := time.Now().Add(-time.Hour)
	f.exams.Now = func() time.Time { return now }
	v := start(t, f)
	q := f.questions[0]
	_, err := f.exams.SaveAnswer(f.user.ID, v.ID, q.ID, q.Answers[0].ID)
	requireNoError(t, err)
	p, err := f.service.Progress(f.user.ID, "A1")
	requireNoError(t, err)
	assertQuestionProgress(t, p, 2, 0, 0, 0) // In-progress exams do not contribute.
	_, err = f.exams.Submit(f.user.ID, v.ID)
	requireNoError(t, err)
	p, err = f.service.Progress(f.user.ID, "A1")
	requireNoError(t, err)
	assertQuestionProgress(t, p, 2, 2, 1, 1) // The unanswered question counts as wrong.
	p, err = f.service.ProgressFiltered(f.user.ID, learning.Filter{License: "A1", CategoryType: "UNCATEGORIZED"})
	requireNoError(t, err)
	assertQuestionProgress(t, p, 2, 2, 1, 1) // Mixed exams still contribute to question mastery.
	if p.CompletedExams != 0 {
		t.Fatal("legacy exam mode filter changed")
	}
	_, _, err = f.service.Practice(f.user.ID, q.ID, q.Answers[1].ID)
	requireNoError(t, err)
	p, err = f.service.Progress(f.user.ID, "A1")
	requireNoError(t, err)
	assertQuestionProgress(t, p, 2, 2, 0, 2)
	// A later timed-out exam supersedes practice, including skipped answers.
	now = time.Now().Add(time.Minute)
	v = start(t, f)
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, q.ID, q.Answers[0].ID)
	requireNoError(t, err)
	now = v.ExpiresAt
	_, err = f.exams.Get(f.user.ID, v.ID)
	requireNoError(t, err)
	p, err = f.service.Progress(f.user.ID, "A1")
	requireNoError(t, err)
	assertQuestionProgress(t, p, 2, 2, 1, 1)
}

func TestQuestionProgressHTTPPreferenceAndFilters(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	jwt, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	decodeBody(t, request(t, r, "", "GET", "/api/v1/progress", nil), 401, nil)
	decodeBody(t, request(t, r, jwt, "PUT", "/api/v1/me/license", map[string]string{"license_type": "A1"}), 200, nil)
	for _, tc := range []struct {
		query, license string
		total          int
	}{
		{"", "A1", 2},
		{"?license_type=B", "B", 0},
		{"?category_type=SIGN", "A1", 0},
		{"?is_critical=true", "A1", 1},
		{"?published=false&page=2&limit=1", "A1", 2},
		{fmt.Sprintf("?category_id=%d", f.questions[0].CategoryID), "A1", 2},
	} {
		var p learning.Progress
		decodeBody(t, request(t, r, jwt, "GET", "/api/v1/progress"+tc.query, nil), 200, &p)
		if p.LicenseType != tc.license {
			t.Fatalf("license: %+v", p)
		}
		assertQuestionProgress(t, &p, tc.total, 0, 0, 0)
	}
}

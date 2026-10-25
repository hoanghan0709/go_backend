package server_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	answer "github.com/han/go-ecommerce/internal/answer/model"
	answerRepo "github.com/han/go-ecommerce/internal/answer/repository"
	auth "github.com/han/go-ecommerce/internal/auth/model"
	"github.com/han/go-ecommerce/internal/common/enums"
	"github.com/han/go-ecommerce/internal/database"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	examRepo "github.com/han/go-ecommerce/internal/exam/repository"
	examUsecase "github.com/han/go-ecommerce/internal/exam/usecase"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/server"
	test "github.com/han/go-ecommerce/internal/test/model"
	"github.com/han/go-ecommerce/internal/token"
	"gorm.io/gorm"
)

type fixture struct {
	db                 *gorm.DB
	service            *learning.Service
	exams              *examUsecase.Usecase
	user, other, admin auth.User
	questions          []question.Question
	test               test.Test
	input              learning.QuestionInput
	jwt                *token.Service
}

func setup(t *testing.T) *fixture {
	t.Helper()
	db, err := database.InitDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	sql, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sql.Close() })
	f := &fixture{db: db, service: learning.New(db), exams: examUsecase.New(examRepo.New(db))}
	f.user = auth.User{Name: "Learner", Email: "learner@example.com"}
	f.other = auth.User{Name: "Other", Email: "other@example.com"}
	f.admin = auth.User{Name: "Admin", Email: "admin@example.com", Role: "admin"}
	for _, u := range []*auth.User{&f.user, &f.other, &f.admin} {
		if err = db.Create(u).Error; err != nil {
			t.Fatal(err)
		}
	}
	cat, err := f.service.SaveCategory(0, learning.CategoryInput{Name: "Biển báo"})
	if err != nil {
		t.Fatal(err)
	}
	f.input = learning.QuestionInput{Title: "Câu hỏi mẫu", Content: "Chọn đáp án đúng", Explanation: "Giải thích mẫu", LicenseType: enums.LicenseA1, CategoryID: cat.ID, Published: true, Answers: []learning.AnswerInput{{Content: "Đúng", Label: "A", Order: 1, IsCorrect: true}, {Content: "Sai", Label: "B", Order: 2}}}
	for _, critical := range []bool{true, false} {
		in := f.input
		in.IsCritical = critical
		q, e := f.service.SaveQuestion(0, in)
		if e != nil {
			t.Fatal(e)
		}
		f.questions = append(f.questions, *q)
	}
	examTest, err := f.service.SaveTest(0, learning.TestInput{Name: "Đề mẫu", LicenseType: "A1", Version: "test-only", Duration: 1, PassingScore: 1, CriticalFail: true, Published: true, QuestionIDs: []uint{f.questions[0].ID, f.questions[1].ID}})
	if err != nil {
		t.Fatal(err)
	}
	f.test = *examTest
	f.jwt, err = token.New("test-secret-only", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func start(t *testing.T, f *fixture) *examUsecase.View {
	t.Helper()
	v, err := f.exams.Start(f.user.ID, f.test.ID)
	requireNoError(t, err)
	return v
}
func TestExamOwnershipSnapshotAndIdempotentSubmit(t *testing.T) {
	f := setup(t)
	v := start(t, f)
	body, _ := json.Marshal(v)
	if bytes.Contains(body, []byte("is_correct")) || bytes.Contains(body, []byte("Giải thích mẫu")) {
		t.Fatalf("answer leaked: %s", body)
	}
	_, err := f.exams.Get(f.other.ID, v.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other user read: %v", err)
	}
	_, err = f.exams.SaveAnswer(f.other.ID, v.ID, f.questions[0].ID, f.questions[0].Answers[0].ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other user write: %v", err)
	}
	_, err = f.exams.Submit(f.other.ID, v.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other user submit: %v", err)
	}
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, f.questions[0].ID, f.questions[1].Answers[0].ID)
	if !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("foreign answer accepted: %v", err)
	}
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, 999, 0)
	if !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("foreign question accepted: %v", err)
	}
	for _, q := range f.questions {
		_, err = f.exams.SaveAnswer(f.user.ID, v.ID, q.ID, q.Answers[0].ID)
		requireNoError(t, err)
	}
	// Thay toàn bộ đáp án trong ngân hàng sau khi bài thi đã bắt đầu.
	in := f.input
	in.Content = "Nội dung mới"
	in.Answers = append([]learning.AnswerInput(nil), f.input.Answers...)
	in.Answers[0].IsCorrect = false
	in.Answers[1].IsCorrect = true
	_, err = f.service.SaveQuestion(f.questions[0].ID, in)
	requireNoError(t, err)
	out, err := f.exams.Submit(f.user.ID, v.ID)
	requireNoError(t, err)
	if !out.IsPassed || out.Score != 2 || out.Exam.Questions[0].Content != "Chọn đáp án đúng" {
		t.Fatalf("snapshot changed: %+v", out)
	}
	again, err := f.exams.Submit(f.user.ID, v.ID)
	requireNoError(t, err)
	a, _ := json.Marshal(out.Exam)
	b, _ := json.Marshal(again.Exam)
	if out.Score != again.Score || !out.FinishedAt.Equal(*again.FinishedAt) || !bytes.Equal(a, b) {
		t.Fatal("submit is not idempotent")
	}
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, f.questions[0].ID, 0)
	if !errors.Is(err, examUsecase.ErrClosed) {
		t.Fatalf("closed exam changed: %v", err)
	}
}
func TestCriticalFailureAndTimeout(t *testing.T) {
	f := setup(t)
	for _, skip := range []bool{false, true} {
		v := start(t, f)
		if !skip {
			_, err := f.exams.SaveAnswer(f.user.ID, v.ID, f.questions[0].ID, f.questions[0].Answers[1].ID)
			requireNoError(t, err)
		}
		_, err := f.exams.SaveAnswer(f.user.ID, v.ID, f.questions[1].ID, f.questions[1].Answers[0].ID)
		requireNoError(t, err)
		out, err := f.exams.Submit(f.user.ID, v.ID)
		requireNoError(t, err)
		if out.Score != 1 || out.IsPassed || out.FailureReason != "critical_question" {
			t.Fatalf("critical rule failed: %+v", out)
		}
	}
	now := time.Now()
	f.exams.Now = func() time.Time { return now }
	v := start(t, f)
	_, err := f.exams.SaveAnswer(f.user.ID, v.ID, f.questions[0].ID, f.questions[0].Answers[0].ID)
	requireNoError(t, err)
	now = v.ExpiresAt
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, f.questions[1].ID, f.questions[1].Answers[0].ID)
	if !errors.Is(err, examUsecase.ErrClosed) {
		t.Fatalf("late answer accepted: %v", err)
	}
	out, err := f.exams.Get(f.user.ID, v.ID)
	requireNoError(t, err)
	if out.Status != enums.ExamTimeout || out.Score != 1 || !out.FinishedAt.Equal(v.ExpiresAt) {
		t.Fatalf("timeout did not persist: %+v", out)
	}
}
func TestResumeClearAndConcurrentSubmit(t *testing.T) {
	f := setup(t)
	v := start(t, f)
	q := f.questions[0]
	for i := 0; i < 3; i++ {
		_, err := f.exams.SaveAnswer(f.user.ID, v.ID, q.ID, q.Answers[0].ID)
		requireNoError(t, err)
	}
	var n int64
	requireNoError(t, f.db.Model(&exam.UserAnswer{}).Where("exam_session_id = ?", v.ID).Count(&n).Error)
	if n != 1 {
		t.Fatal("duplicate saved answers")
	}
	resumed, err := f.exams.Get(f.user.ID, v.ID)
	requireNoError(t, err)
	if resumed.Exam.Questions[0].SelectedAnswerID != q.Answers[0].ID {
		t.Fatal("resume lost answer")
	}
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, q.ID, 0)
	requireNoError(t, err)
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, q.ID, q.Answers[0].ID)
	requireNoError(t, err)
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, e := f.exams.Submit(f.user.ID, v.ID)
			if e == nil && out.Score != 1 {
				e = fmt.Errorf("inconsistent score")
			}
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		requireNoError(t, e)
	}
}
func TestImportRollbackAndAnswerInvariant(t *testing.T) {
	f := setup(t)
	bad := f.input
	bad.Answers = []learning.AnswerInput{{Content: "A", Label: "A", Order: 1, IsCorrect: true}, {Content: "B", Label: "B", Order: 2, IsCorrect: true}}
	var before, after int64
	requireNoError(t, f.db.Model(&question.Question{}).Count(&before).Error)
	_, err := f.service.ImportQuestions([]learning.QuestionInput{f.input, bad})
	if !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("bad import accepted: %v", err)
	}
	requireNoError(t, f.db.Model(&question.Question{}).Count(&after).Error)
	if before != after {
		t.Fatal("partial import committed")
	}
	repo := answerRepo.New(f.db)
	q := f.questions[0]
	a := q.Answers[1]
	a.IsCorrect = true
	if err = repo.Update(&a); !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("two correct answers allowed: %v", err)
	}
	if err = repo.Delete(q.ID, q.Answers[0].ID); !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("invalid delete allowed: %v", err)
	}
	extra := answer.Answer{QuestionID: q.ID, Content: "C", Label: "C", Order: 3}
	requireNoError(t, repo.Create(&extra))
	requireNoError(t, repo.Delete(q.ID, extra.ID))
	extra.ID = 0
	requireNoError(t, repo.Create(&extra)) // nhãn/thứ tự của đáp án xóa mềm dùng lại được.
	if err = f.service.DeleteQuestion(q.ID); !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("test question deleted: %v", err)
	}
}
func TestGeneratePracticeAndProgress(t *testing.T) {
	f := setup(t)
	in := learning.GenerateInput{TestInput: learning.TestInput{Name: "Random", LicenseType: "A1", Version: "test-only", Duration: 1, PassingScore: 1, Published: true}, Quotas: []learning.Quota{{CategoryID: f.input.CategoryID, Count: 2, CriticalCount: 1}}}
	generated, err := f.service.GenerateTest(in)
	requireNoError(t, err)
	if len(generated.Questions) != 2 || generated.Questions[0].QuestionID == generated.Questions[1].QuestionID {
		t.Fatal("invalid generated test")
	}
	in.Quotas[0].Count = 3
	_, err = f.service.GenerateTest(in)
	if !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("insufficient bank accepted: %v", err)
	}
	q := f.questions[0]
	_, _, err = f.service.Practice(f.user.ID, q.ID, f.questions[1].Answers[0].ID)
	if !errors.Is(err, learning.ErrInvalid) {
		t.Fatalf("foreign practice answer: %v", err)
	}
	_, _, err = f.service.Practice(f.user.ID, q.ID, q.Answers[1].ID)
	requireNoError(t, err)
	wrong, n, err := f.service.WrongQuestions(f.user.ID, learning.Filter{})
	requireNoError(t, err)
	if n != 1 || wrong[0].ID != q.ID {
		t.Fatal("wrong question missing")
	}
	_, _, err = f.service.Practice(f.user.ID, q.ID, q.Answers[0].ID)
	requireNoError(t, err)
	_, n, err = f.service.WrongQuestions(f.user.ID, learning.Filter{})
	requireNoError(t, err)
	if n != 0 {
		t.Fatal("corrected question still wrong")
	}
	progress, err := f.service.Progress(f.user.ID, "A1")
	requireNoError(t, err)
	if len(progress.Categories) != 1 || progress.Categories[0].Accuracy != 50 {
		t.Fatalf("bad progress: %+v", progress)
	}
	other, err := f.service.Progress(f.other.ID, "")
	requireNoError(t, err)
	if len(other.Categories) != 0 {
		t.Fatal("progress leaked across users")
	}
}
func request(t *testing.T, r http.Handler, jwt string, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		requireNoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}
func TestHTTPPermissionsAndExamFlow(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	learner, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	admin, err := f.jwt.GenerateAccessToken(f.admin.ID)
	requireNoError(t, err)
	cases := []struct {
		token, method, path string
		body                interface{}
		status              int
	}{
		{"", "GET", "/api/v1/tests", nil, 401},
		{learner, "POST", "/api/v1/admin/questions", f.input, 403},
		{learner, "POST", "/question/create", f.input, 403},
		{learner, "GET", "/questions/1/answers", nil, 403},
		{admin, "POST", "/api/v1/admin/questions", f.input, 201},
		{learner, "GET", "/api/v1/practice/questions?limit=1&page=1&license_type=A1", nil, 200},
		{learner, "GET", "/api/v1/practice/questions?page=-1", nil, 400},
		{learner, "GET", "/api/v1/tests?license_type=invalid", nil, 400},
		{learner, "GET", "/api/v1/exams/0", nil, 400},
	}
	for _, tc := range cases {
		rec := request(t, r, tc.token, tc.method, tc.path, tc.body)
		if rec.Code != tc.status {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
	rec := request(t, r, learner, "GET", "/api/v1/practice/questions", nil)
	if strings.Contains(rec.Body.String(), "is_correct") || strings.Contains(rec.Body.String(), "IsCorrect") {
		t.Fatal("practice list leaked answer")
	}
	rec = request(t, r, learner, "POST", "/api/v1/exams", map[string]uint{"test_id": f.test.ID})
	if rec.Code != 201 {
		t.Fatal(rec.Body.String())
	}
	var v examUsecase.View
	requireNoError(t, json.Unmarshal(rec.Body.Bytes(), &v))
	for _, q := range f.questions {
		rec = request(t, r, learner, "PUT", fmt.Sprintf("/api/v1/exams/%d/answers/%d", v.ID, q.ID), map[string]uint{"answer_id": q.Answers[0].ID})
		if rec.Code != 200 {
			t.Fatal(rec.Body.String())
		}
	}
	rec = request(t, r, learner, "POST", fmt.Sprintf("/api/v1/exams/%d/submit", v.ID), nil)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	requireNoError(t, json.Unmarshal(rec.Body.Bytes(), &v))
	if !v.IsPassed || v.Score != 2 {
		t.Fatal("HTTP grading failed")
	}
	// Quyền bị thu hồi có hiệu lực với chính token cũ.
	requireNoError(t, f.db.Model(&f.admin).Update("role", "learner").Error)
	rec = request(t, r, admin, "POST", "/api/v1/admin/questions", f.input)
	if rec.Code != 403 {
		t.Fatal("revoked admin still authorized")
	}
}

func TestHistoryExpiresAndIncludesUnansweredInWrongQuestions(t *testing.T) {
	f := setup(t)
	now := time.Now()
	f.exams.Now = func() time.Time { return now }
	v := start(t, f)
	now = v.ExpiresAt.Add(time.Second)
	history, total, err := f.exams.History(f.user.ID, 1, 20)
	requireNoError(t, err)
	if total != 1 || history[0].Status != enums.ExamTimeout {
		t.Fatal("history did not expire exam")
	}
	progress, err := f.service.Progress(f.user.ID, "")
	requireNoError(t, err)
	if progress.CompletedExams != 1 || len(progress.WrongQuestionIDs) != 2 || len(progress.RecentExams) != 1 {
		t.Fatalf("incomplete progress: %+v", progress)
	}
	_, total, err = f.exams.History(f.other.ID, 1, 20)
	requireNoError(t, err)
	if total != 0 {
		t.Fatal("history leaked")
	}
}
func TestDraftAndInvalidConfiguration(t *testing.T) {
	f := setup(t)
	in := f.input
	in.Explanation = ""
	if _, err := f.service.SaveQuestion(0, in); !errors.Is(err, learning.ErrInvalid) {
		t.Fatal("published question without explanation")
	}
	in.Published = false
	draft, err := f.service.SaveQuestion(0, in)
	requireNoError(t, err)
	_, total, err := f.service.Questions(learning.Filter{}, false)
	requireNoError(t, err)
	if total != 2 {
		t.Fatal("draft visible to learner")
	}
	for _, input := range []learning.TestInput{
		{Name: "bad", LicenseType: "A1", Version: "v1", Duration: 1, PassingScore: 1, QuestionIDs: []uint{draft.ID}},
		{Name: "bad", LicenseType: "B", Version: "v1", Duration: 1, PassingScore: 1, QuestionIDs: []uint{f.questions[0].ID}},
		{Name: "bad", LicenseType: "A1", Version: "v1", Duration: 1, PassingScore: 1, QuestionIDs: []uint{f.questions[0].ID, f.questions[0].ID}},
		{Name: "bad", LicenseType: "A1", Version: "v1", Duration: 1, PassingScore: 3, QuestionIDs: []uint{f.questions[0].ID}},
	} {
		if _, err = f.service.SaveTest(0, input); !errors.Is(err, learning.ErrInvalid) {
			t.Fatalf("bad test accepted: %+v (%v)", input, err)
		}
	}
	// Tắt quy tắc điểm liệt: điểm đạt vẫn được xét theo cấu hình đề.
	requireNoError(t, f.db.Model(&f.test).Update("critical_fail", false).Error)
	v := start(t, f)
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, f.questions[1].ID, f.questions[1].Answers[0].ID)
	requireNoError(t, err)
	out, err := f.exams.Submit(f.user.ID, v.ID)
	requireNoError(t, err)
	if !out.IsPassed {
		t.Fatal("critical_fail=false ignored")
	}
}
func TestAuthRegistrationAndRefresh(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	in := map[string]interface{}{"name": "New learner", "email": "new@example.com", "password": "testing-password", "role": "admin"}
	rec := request(t, r, "", "POST", "/auth/register", in)
	if rec.Code != 201 {
		t.Fatal(rec.Body.String())
	}
	var stored auth.User
	requireNoError(t, f.db.Where("email = ?", "new@example.com").First(&stored).Error)
	if stored.Role != "learner" || stored.Password == "testing-password" {
		t.Fatal("registration privilege or hashing failed")
	}
	rec = request(t, r, "", "POST", "/auth/login", in)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var login struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}
	}
	requireNoError(t, json.Unmarshal(rec.Body.Bytes(), &login))
	rec = request(t, r, login.Data.AccessToken, "GET", "/auth/getProfile", nil)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var profile struct {
		ID   uint   `json:"id"`
		Role string `json:"role"`
	}
	requireNoError(t, json.Unmarshal(rec.Body.Bytes(), &profile))
	if profile.ID != stored.ID || profile.Role != "learner" {
		t.Fatal("profile missing ID or role")
	}
	rec = request(t, r, "", "POST", "/auth/refresh", map[string]string{"refresh_token": login.Data.RefreshToken})
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	rec = request(t, r, "", "POST", "/auth/refresh", map[string]string{"refresh_token": login.Data.RefreshToken})
	if rec.Code != 401 {
		t.Fatal("rotated refresh token reusable")
	}
}

func TestAdminContentHTTP(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	jwt, err := f.jwt.GenerateAccessToken(f.admin.ID)
	requireNoError(t, err)
	rec := request(t, r, jwt, "POST", "/api/v1/admin/categories", learning.CategoryInput{Name: "Chủ đề mới"})
	if rec.Code != 201 {
		t.Fatal(rec.Body.String())
	}
	var category struct {
		ID uint `json:"id"`
	}
	requireNoError(t, json.Unmarshal(rec.Body.Bytes(), &category))
	in := f.input
	in.CategoryID = category.ID
	rec = request(t, r, jwt, "POST", "/api/v1/admin/questions/import", map[string]interface{}{"questions": []learning.QuestionInput{in}})
	if rec.Code != 201 {
		t.Fatal(rec.Body.String())
	}
	var imported struct {
		Data []question.Question `json:"data"`
	}
	requireNoError(t, json.Unmarshal(rec.Body.Bytes(), &imported))
	qid := imported.Data[0].ID
	rec = request(t, r, jwt, "GET", fmt.Sprintf("/api/v1/admin/questions/%d", qid), nil)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	in.Content = "Nội dung đã chỉnh"
	rec = request(t, r, jwt, "PUT", fmt.Sprintf("/api/v1/admin/questions/%d", qid), in)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	testInput := learning.TestInput{Name: "Đề thay đổi", LicenseType: "A1", Version: "v2", Duration: 2, PassingScore: 1, Published: true, QuestionIDs: []uint{qid}}
	rec = request(t, r, jwt, "PUT", fmt.Sprintf("/api/v1/admin/tests/%d", f.test.ID), testInput)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	// Update lần nữa không vướng unique index của liên kết đề cũ.
	rec = request(t, r, jwt, "PUT", fmt.Sprintf("/api/v1/admin/tests/%d", f.test.ID), testInput)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	rec = request(t, r, jwt, "DELETE", fmt.Sprintf("/api/v1/admin/categories/%d", category.ID), nil)
	if rec.Code != 400 {
		t.Fatalf("nonempty category deleted: %s", rec.Body.String())
	}
	v := start(t, f)
	rec = request(t, r, jwt, "DELETE", fmt.Sprintf("/api/v1/admin/tests/%d", f.test.ID), nil)
	if rec.Code != 204 {
		t.Fatal(rec.Body.String())
	}
	rec = request(t, r, jwt, "DELETE", fmt.Sprintf("/api/v1/admin/questions/%d", qid), nil)
	if rec.Code != 204 {
		t.Fatal(rec.Body.String())
	}
	rec = request(t, r, jwt, "DELETE", fmt.Sprintf("/api/v1/admin/categories/%d", category.ID), nil)
	if rec.Code != 204 {
		t.Fatal(rec.Body.String())
	}
	// Bài đã bắt đầu vẫn đọc/làm được sau khi đề, câu hỏi và chủ đề bị xóa mềm.
	out, err := f.exams.Get(f.user.ID, v.ID)
	requireNoError(t, err)
	_, err = f.exams.SaveAnswer(f.user.ID, v.ID, qid, out.Exam.Questions[0].Answers[0].ID)
	requireNoError(t, err)
	out, err = f.exams.Submit(f.user.ID, v.ID)
	requireNoError(t, err)
	if !out.IsPassed {
		t.Fatal("deletion changed existing exam")
	}
}

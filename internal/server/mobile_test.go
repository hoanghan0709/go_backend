package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	auth "github.com/han/go-ecommerce/internal/auth/model"
	"github.com/han/go-ecommerce/internal/common/enums"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	examusecase "github.com/han/go-ecommerce/internal/exam/usecase"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/server"
	testmodel "github.com/han/go-ecommerce/internal/test/model"
)

func decodeBody(t *testing.T, res *httptest.ResponseRecorder, status int, out interface{}) {
	t.Helper()
	if res.Code != status {
		t.Fatalf("expected %d, got %d: %s", status, res.Code, res.Body.String())
	}
	if out != nil {
		requireNoError(t, json.Unmarshal(res.Body.Bytes(), out))
	}
}

func TestMobileLicensePreferenceAndCategoryFilters(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	learner, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	admin, err := f.jwt.GenerateAccessToken(f.admin.ID)
	requireNoError(t, err)
	decodeBody(t, request(t, r, "", "PUT", "/api/v1/me/license", map[string]string{"license_type": "A1"}), 401, nil)
	decodeBody(t, request(t, r, learner, "PUT", "/api/v1/me/license", map[string]string{"license_type": "invalid"}), 400, nil)
	decodeBody(t, request(t, r, learner, "PUT", "/api/v1/me/license", map[string]interface{}{"license_type": "A1", "user_id": f.other.ID, "role": "admin"}), 200, nil)
	var me auth.User
	requireNoError(t, f.db.First(&me, f.user.ID).Error)
	if me.CurrentLicenseType != "A1" || me.Role != "learner" {
		t.Fatal("preference changed role or was not persisted")
	}
	var other auth.User
	requireNoError(t, f.db.First(&other, f.other.ID).Error)
	if other.CurrentLicenseType != "" {
		t.Fatal("another user's license changed")
	}
	var profile struct {
		CurrentLicenseType string `json:"current_license_type"`
	}
	decodeBody(t, request(t, r, learner, "GET", "/auth/getProfile", nil), 200, &profile)
	if profile.CurrentLicenseType != "A1" {
		t.Fatal("legacy profile missing preference")
	}
	c, err := f.service.SaveCategory(0, learning.CategoryInput{Name: "Rule B", Type: "RULE"})
	requireNoError(t, err)
	in := f.input
	in.CategoryID = c.ID
	in.LicenseType = enums.LicenseType("B")
	_, err = f.service.SaveQuestion(0, in)
	requireNoError(t, err)
	var page struct {
		Total int                 `json:"total"`
		Data  []question.Question `json:"data"`
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/practice/questions", nil), 200, &page)
	if page.Total != 2 {
		t.Fatal("saved license not applied")
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/practice/questions?license_type=B&category_type=RULE", nil), 200, &page)
	if page.Total != 1 {
		t.Fatalf("explicit license/type filter: %+v", page)
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/practice/questions?license_type=B&category_type=SIGN", nil), 200, &page)
	if page.Total != 0 {
		t.Fatal("category type filter leaked questions")
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/practice/questions?category_type=BAD", nil), 400, nil)
	decodeBody(t, request(t, r, admin, "GET", "/api/v1/admin/questions?published=false", nil), 200, &page)
	if page.Total != 0 {
		t.Fatal("published filter ignored")
	}
	var cats struct {
		Data []struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/categories?license_type=B&category_type=RULE", nil), 200, &cats)
	if len(cats.Data) != 1 || cats.Data[0].ID != c.ID {
		t.Fatal("category tabs not scoped by license/type")
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/category-types", nil), 200, nil)
}

func TestRoadSignCRUDQuestionsAndRandomExam(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	admin, err := f.jwt.GenerateAccessToken(f.admin.ID)
	requireNoError(t, err)
	learner, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	other, err := f.jwt.GenerateAccessToken(f.other.ID)
	requireNoError(t, err)
	c, err := f.service.SaveCategory(0, learning.CategoryInput{Name: "Signs", Type: "SIGN"})
	requireNoError(t, err)
	signInput := learning.RoadSignInput{Code: "DEMO-01", Name: "Biển minh họa", Meaning: "Ý nghĩa minh họa", Image: "/demo.png", SignType: "WARNING", LicenseType: "A1", CategoryID: c.ID, Published: true}
	decodeBody(t, request(t, r, learner, "POST", "/api/v1/admin/road-signs", signInput), 403, nil)
	var sign struct {
		ID uint `json:"id"`
	}
	decodeBody(t, request(t, r, admin, "POST", "/api/v1/admin/road-signs", signInput), 201, &sign)
	signPath := fmt.Sprintf("/api/v1/admin/road-signs/%d", sign.ID)
	decodeBody(t, request(t, r, admin, "POST", "/api/v1/admin/road-signs", signInput), 400, nil)
	signInput.Name = "Đổi tên"
	decodeBody(t, request(t, r, admin, "PUT", signPath, signInput), 200, nil)
	// An unrelated B sign and an unpublished A1 sign must not leak into A1 lessons.
	bInput := signInput
	bInput.LicenseType = "B"
	bInput.Code = "B-01"
	_, err = f.service.SaveRoadSign(0, bInput)
	requireNoError(t, err)
	draftInput := signInput
	draftInput.Code = "DRAFT"
	draftInput.Published = false
	draft, err := f.service.SaveRoadSign(0, draftInput)
	requireNoError(t, err)
	_, err = f.service.SetLicense(f.user.ID, "A1")
	requireNoError(t, err)
	var list struct {
		Total int `json:"total"`
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/road-signs?sign_type=WARNING", nil), 200, &list)
	if list.Total != 1 {
		t.Fatal("road signs not filtered by preference/publication")
	}
	decodeBody(t, request(t, r, learner, "GET", fmt.Sprintf("/api/v1/road-signs/%d", draft.ID), nil), 404, nil)
	input := f.input
	input.CategoryID = c.ID
	input.RoadSignID = &sign.ID
	input.IsCritical = false
	first, err := f.service.SaveQuestion(0, input)
	requireNoError(t, err)
	input.Title = "Second sign question"
	input.IsCritical = true
	second, err := f.service.SaveQuestion(0, input)
	requireNoError(t, err)
	wrongLicense := input
	wrongLicense.LicenseType = enums.LicenseType("B")
	if _, err = f.service.SaveQuestion(0, wrongLicense); err == nil {
		t.Fatal("cross-license sign link allowed")
	}
	draftLink := input
	draftLink.RoadSignID = &draft.ID
	if _, err = f.service.SaveQuestion(0, draftLink); err == nil {
		t.Fatal("published question linked to draft sign")
	}
	qpath := fmt.Sprintf("/api/v1/road-signs/%d/questions", sign.ID)
	rec := request(t, r, learner, "GET", qpath, nil)
	decodeBody(t, rec, 200, &list)
	if list.Total != 2 || bytes.Contains(rec.Body.Bytes(), []byte("is_correct")) {
		t.Fatal("sign question list invalid/leaks answer key")
	}
	decodeBody(t, request(t, r, learner, "GET", fmt.Sprintf("/api/v1/practice/questions?road_sign_id=%d&category_type=SIGN", sign.ID), nil), 200, &list)
	if list.Total != 2 {
		t.Fatal("sign question filters ignored")
	}
	decodeBody(t, request(t, r, learner, "POST", fmt.Sprintf("/api/v1/practice/questions/%d/answer", first.ID), map[string]uint{"answer_id": first.Answers[0].ID}), 200, nil)
	decodeBody(t, request(t, r, admin, "DELETE", signPath, nil), 400, nil)
	signInput.Published = false
	decodeBody(t, request(t, r, admin, "PUT", signPath, signInput), 400, nil)
	config := learning.GenerateInput{TestInput: learning.TestInput{Name: "Demo sign exam", LicenseType: "A1", CategoryType: "SIGN", Version: "demo-v1", Duration: 10, PassingScore: 2, CriticalFail: true, Published: true}, Quotas: []learning.Quota{{CategoryID: c.ID, Count: 2, CriticalCount: 1}}}
	decodeBody(t, request(t, r, learner, "POST", "/api/v1/admin/exam-templates", config), 403, nil)
	var template learning.ExamTemplate
	decodeBody(t, request(t, r, admin, "POST", "/api/v1/admin/exam-templates", config), 201, &template)
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/exam-templates?category_type=SIGN", nil), 200, &list)
	if list.Total != 1 {
		t.Fatal("template listing")
	}
	var beforeTests, beforeExams int64
	requireNoError(t, f.db.Model(&testmodel.Test{}).Count(&beforeTests).Error)
	requireNoError(t, f.db.Model(&exam.ExamSession{}).Count(&beforeExams).Error)
	var v examusecase.View
	rec = request(t, r, learner, "POST", "/api/v1/exams/random", map[string]interface{}{"category_type": "SIGN", "passing_score": 1, "duration_minutes": 999, "user_id": f.other.ID})
	decodeBody(t, rec, 201, &v)
	if v.TestID != nil || v.TemplateID != template.ID || v.PassingScore != 2 || len(v.Exam.Questions) != 2 || v.Exam.CategoryType != "SIGN" {
		t.Fatalf("random result/config: %+v", v)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("is_correct")) || bytes.Contains(rec.Body.Bytes(), []byte("explanation")) {
		t.Fatal("random exam leaked key")
	}
	var afterTests int64
	requireNoError(t, f.db.Model(&testmodel.Test{}).Count(&afterTests).Error)
	if afterTests != beforeTests {
		t.Fatal("random exam polluted shared tests")
	}
	decodeBody(t, request(t, r, other, "GET", fmt.Sprintf("/api/v1/exams/%d", v.ID), nil), 404, nil)
	decodeBody(t, request(t, r, other, "POST", fmt.Sprintf("/api/v1/exams/%d/submit", v.ID), nil), 404, nil)
	for _, q := range v.Exam.Questions {
		var aid uint
		if q.ID == first.ID {
			aid = first.Answers[0].ID
		} else if q.ID == second.ID {
			aid = second.Answers[0].ID
		} else {
			t.Fatal("unrelated question in sign test")
		}
		decodeBody(t, request(t, r, learner, "PUT", fmt.Sprintf("/api/v1/exams/%d/answers/%d", v.ID, q.ID), map[string]uint{"answer_id": aid}), 200, nil)
	}
	// Snapshot survives editing the template and bank while an exam is in flight.
	input.Explanation = "New explanation"
	input.Answers[0].IsCorrect = false
	input.Answers[1].IsCorrect = true
	_, err = f.service.SaveQuestion(second.ID, input)
	requireNoError(t, err)
	config.PassingScore = 1
	decodeBody(t, request(t, r, admin, "PUT", fmt.Sprintf("/api/v1/admin/exam-templates/%d", template.ID), config), 200, nil)
	decodeBody(t, request(t, r, learner, "POST", fmt.Sprintf("/api/v1/exams/%d/submit", v.ID), nil), 200, &v)
	if !v.IsPassed || v.Score != 2 || v.PassingScore != 2 {
		t.Fatal("snapshot did not preserve grading")
	}
	// Insufficient bank rolls back; no orphaned session is created.
	config.Quotas[0].Count = 3
	config.Quotas[0].CriticalCount = 1
	decodeBody(t, request(t, r, admin, "PUT", fmt.Sprintf("/api/v1/admin/exam-templates/%d", template.ID), config), 200, nil)
	decodeBody(t, request(t, r, learner, "POST", "/api/v1/exams/random", map[string]string{"category_type": "SIGN"}), 400, nil)
	var count int64
	requireNoError(t, f.db.Model(&exam.ExamSession{}).Count(&count).Error)
	if count != beforeExams+1 {
		t.Fatal("failed random generation left a session")
	}
	decodeBody(t, request(t, r, admin, "DELETE", fmt.Sprintf("/api/v1/admin/exam-templates/%d", template.ID), nil), 204, nil)
	decodeBody(t, request(t, r, learner, "GET", fmt.Sprintf("/api/v1/exams/%d", v.ID), nil), 200, nil)
	decodeBody(t, request(t, r, admin, "DELETE", fmt.Sprintf("/api/v1/admin/road-signs/%d", draft.ID), nil), 204, nil)
}

func TestRandomDefaultsRetryAndHistoryFilters(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	learner, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	decodeBody(t, request(t, r, learner, "POST", "/api/v1/exams/random", map[string]interface{}{}), 400, nil)
	_, err = f.service.SetLicense(f.user.ID, "A1")
	requireNoError(t, err)
	config := learning.GenerateInput{TestInput: learning.TestInput{Name: "Random full demo", LicenseType: "A1", Version: "demo", Duration: 1, PassingScore: 1, Published: true}, Quotas: []learning.Quota{{CategoryID: f.input.CategoryID, Count: 2, CriticalCount: 1}}}
	template, err := f.service.SaveTemplate(0, config)
	requireNoError(t, err)
	var first, second examusecase.View
	decodeBody(t, request(t, r, learner, "POST", "/api/v1/exams/random", map[string]interface{}{}), 201, &first)
	decodeBody(t, request(t, r, learner, "POST", "/api/v1/exams/random", map[string]uint{"template_id": template.ID}), 201, &second)
	if first.ID == second.ID {
		t.Fatal("random attempts reused a session")
	}
	ids := map[uint]bool{}
	for _, q := range second.Exam.Questions {
		if ids[q.ID] {
			t.Fatal("duplicate random question")
		}
		ids[q.ID] = true
	}
	_, err = f.service.SetLicense(f.user.ID, "B")
	requireNoError(t, err)
	decodeBody(t, request(t, r, learner, "POST", "/api/v1/exams/random", map[string]uint{"template_id": template.ID}), 400, nil)
	decodeBody(t, request(t, r, learner, "GET", fmt.Sprintf("/api/v1/exams/%d", first.ID), nil), 200, &first)
	if first.Exam.LicenseType != "A1" {
		t.Fatal("preference change altered existing exam")
	}
	var history struct {
		Total int `json:"total"`
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/exams?license_type=A1&status=DOING", nil), 200, &history)
	if history.Total != 2 {
		t.Fatal("history filtering failed")
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/exams?license_type=B", nil), 200, &history)
	if history.Total != 0 {
		t.Fatal("history license filter ignored")
	}
	decodeBody(t, request(t, r, learner, "GET", "/api/v1/exams?status=INVALID", nil), 400, nil)
	// Retrying submit retains one result, including random sessions with NULL test_id.
	decodeBody(t, request(t, r, learner, "POST", fmt.Sprintf("/api/v1/exams/%d/submit", first.ID), nil), 200, nil)
	decodeBody(t, request(t, r, learner, "POST", fmt.Sprintf("/api/v1/exams/%d/submit", first.ID), nil), 200, nil)
}

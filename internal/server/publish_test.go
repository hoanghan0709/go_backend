package server_test

import (
	"encoding/json"
	"reflect"
	"testing"

	answer "github.com/han/go-ecommerce/internal/answer/model"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/server"
)

func TestBulkPublishQuestions(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	admin, err := f.jwt.GenerateAccessToken(f.admin.ID)
	requireNoError(t, err)
	learner, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	input := f.input
	input.Published = false
	good, err := f.service.SaveQuestion(0, input)
	requireNoError(t, err)
	input.Explanation = ""
	missingExplanation, err := f.service.SaveQuestion(0, input)
	requireNoError(t, err)
	input.Explanation = "Lời giải"
	badAnswers, err := f.service.SaveQuestion(0, input)
	requireNoError(t, err)
	requireNoError(t, f.db.Model(&answer.Answer{}).Where("question_id = ?", badAnswers.ID).Update("is_correct", false).Error)
	deleted, err := f.service.SaveQuestion(0, input)
	requireNoError(t, err)
	requireNoError(t, f.service.DeleteQuestion(deleted.ID))
	const path = "/api/v1/admin/questions/publish"
	body := map[string]interface{}{"question_ids": []uint{good.ID, missingExplanation.ID, badAnswers.ID, deleted.ID, 999999}}
	for _, auth := range []struct {
		token string
		code  int
	}{{"", 401}, {learner, 403}} {
		res := request(t, r, auth.token, "POST", path, body)
		if res.Code != auth.code {
			t.Fatalf("permissions: got %d, want %d", res.Code, auth.code)
		}
	}
	for _, ids := range [][]uint{{}, {0}, {good.ID, good.ID}, make([]uint, 501)} {
		res := request(t, r, admin, "POST", path, map[string]interface{}{"question_ids": ids})
		if res.Code != 400 {
			t.Fatalf("invalid IDs accepted: %d", res.Code)
		}
	}
	res := request(t, r, admin, "POST", path, body)
	if res.Code != 200 {
		t.Fatalf("publish: %d %s", res.Code, res.Body.String())
	}
	var result learning.PublishQuestionsResult
	requireNoError(t, json.Unmarshal(res.Body.Bytes(), &result))
	if !reflect.DeepEqual(result.PublishedIDs, []uint{good.ID}) || len(result.Failed) != 4 {
		t.Fatalf("unexpected partial result: %+v", result)
	}
	var stored question.Question
	requireNoError(t, f.db.Preload("Answers").First(&stored, good.ID).Error)
	if !stored.Published || stored.Title != good.Title || stored.Explanation != good.Explanation {
		t.Fatal("publication changed content or failed")
	}
	if len(stored.Answers) != len(good.Answers) {
		t.Fatal("answers changed")
	}
	for i, a := range stored.Answers {
		if a.ID != good.Answers[i].ID || a.Content != good.Answers[i].Content {
			t.Fatal("answer ID/content changed")
		}
	}
	for _, id := range []uint{missingExplanation.ID, badAnswers.ID} {
		var q question.Question
		requireNoError(t, f.db.First(&q, id).Error)
		if q.Published {
			t.Fatal("invalid question was published")
		}
	}
	// Repeating the operation is safe; publication does not recreate answers.
	res = request(t, r, admin, "POST", path, map[string]interface{}{"question_ids": []uint{good.ID}})
	if res.Code != 200 {
		t.Fatal(res.Body.String())
	}
	var count int64
	requireNoError(t, f.db.Unscoped().Model(&answer.Answer{}).Where("question_id = ?", good.ID).Count(&count).Error)
	if count != int64(len(good.Answers)) {
		t.Fatal("publishing recreated answers")
	}
	// The question now appears in learner practice queries.
	qs, _, err := f.service.Questions(learning.Filter{Search: good.Title, Limit: 100}, false)
	requireNoError(t, err)
	found := false
	for _, q := range qs {
		if q.ID == good.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("published question absent from learner practice")
	}
}

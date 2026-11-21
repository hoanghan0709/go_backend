package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/han/go-ecommerce/internal/learning"
	"github.com/han/go-ecommerce/internal/server"
)

func TestPublishedQuestionPickerPagination(t *testing.T) {
	f := setup(t)
	cat, err := f.service.SaveCategory(0, learning.CategoryInput{Name: "Picker pagination"})
	requireNoError(t, err)
	inputs := make([]learning.QuestionInput, 255)
	for i := range inputs {
		inputs[i] = f.input
		inputs[i].CategoryID = cat.ID
		inputs[i].Published = i >= 250
	}
	created, err := f.service.ImportQuestions(inputs)
	requireNoError(t, err)
	r := server.NewRouter(f.db, f.jwt)
	admin, err := f.jwt.GenerateAccessToken(f.admin.ID)
	requireNoError(t, err)
	learner, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	var envelope struct {
		Data []struct {
			ID uint `json:"id"`
		} `json:"data"`
		Total int64 `json:"total"`
	}
	for _, token := range []string{admin, learner} {
		res := request(t, r, token, "GET", fmt.Sprintf("/api/v1/practice/questions?license_type=A1&category_id=%d&page=1&limit=20", cat.ID), nil)
		if res.Code != 200 {
			t.Fatal(res.Body.String())
		}
		requireNoError(t, json.Unmarshal(res.Body.Bytes(), &envelope))
		if envelope.Total != 5 || len(envelope.Data) != 5 || envelope.Data[0].ID != created[250].ID {
			t.Fatalf("drafts distorted pagination: %+v", envelope)
		}
		if bytes.Contains(res.Body.Bytes(), []byte("is_correct")) {
			t.Fatal("picker/learner list leaked answer key")
		}
	}
	// The full admin bank must still include drafts for editing/publication.
	res := request(t, r, admin, "GET", fmt.Sprintf("/api/v1/admin/questions?category_id=%d&limit=20", cat.ID), nil)
	requireNoError(t, json.Unmarshal(res.Body.Bytes(), &envelope))
	if envelope.Total != 255 || len(envelope.Data) != 20 {
		t.Fatal("admin bank no longer includes drafts")
	}
}

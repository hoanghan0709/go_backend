package httpapi

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/learning"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBindSpecificErrors(t *testing.T) {
	for _, tc := range []struct{ name, body, field, rule, fragment string }{
		{"empty", "", "body", "required", "request body is required"},
		{"malformed", `{"title":!}`, "body", "json", "malformed JSON"},
		{"type", `{"category_id":"abc"}`, "category_id", "type", "uint"},
		{"nested", `{"title":"T","content":"C","category_id":1,"license_type":"A1","answers":[{"label":"A","order":1},{"content":"B","label":"B","order":2}]}`, "answers[0].content", "required", "answers[0].content is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			var in learning.QuestionInput
			if Bind(c, &in) {
				t.Fatal("invalid body accepted")
			}
			var result struct {
				Message string
				Errors  []struct{ Field, Rule, Message string }
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if rec.Code != 400 || len(result.Errors) == 0 || result.Errors[0].Field != tc.field || result.Errors[0].Rule != tc.rule || !strings.Contains(result.Message, tc.fragment) {
				t.Fatalf("unexpected response: %s", rec.Body)
			}
		})
	}
}

func TestEmbeddedValidationPath(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"license_type":"A1","version":"v1","duration_minutes":10,"passing_score":1,"quotas":[{"category_id":1,"count":1}]}`))
	var in learning.GenerateInput
	if Bind(c, &in) {
		t.Fatal("missing name accepted")
	}
	if !strings.Contains(rec.Body.String(), `"field":"name"`) || strings.Contains(rec.Body.String(), "TestInput") {
		t.Fatal(rec.Body.String())
	}
}

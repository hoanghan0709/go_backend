package server_test

import (
	"fmt"
	"testing"

	"github.com/han/go-ecommerce/internal/exam/dto"
	"github.com/han/go-ecommerce/internal/server"
)

func TestLicenseCatalogAndExamHistoryAfterSwitch(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	jwt, err := f.jwt.GenerateAccessToken(f.user.ID)
	requireNoError(t, err)
	decodeBody(t, request(t, r, "", "GET", "/api/v1/license-types", nil), 401, nil)
	var catalog struct {
		Data []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, request(t, r, jwt, "GET", "/api/v1/license-types", nil), 200, &catalog)
	expected := []string{"A1", "A", "B1", "B", "C1", "C", "D1", "D2", "D", "BE", "C1E", "CE", "D1E", "D2E", "DE"}
	if len(catalog.Data) != len(expected) {
		t.Fatalf("unexpected catalog: %+v", catalog)
	}
	for i, item := range catalog.Data {
		if item.Code != expected[i] || item.Name == "" {
			t.Fatalf("invalid item: %+v", item)
		}
	}
	decodeBody(t, request(t, r, jwt, "PUT", "/api/v1/me/license", map[string]string{"license_type": "A1"}), 200, nil)
	var exam dto.View
	decodeBody(t, request(t, r, jwt, "POST", "/api/v1/exams", map[string]uint{"test_id": f.test.ID}), 201, &exam)
	// Switching before submission must never relabel the existing A1 exam as B.
	decodeBody(t, request(t, r, jwt, "PUT", "/api/v1/me/license", map[string]string{"license_type": "B"}), 200, nil)
	decodeBody(t, request(t, r, jwt, "POST", fmt.Sprintf("/api/v1/exams/%d/submit", exam.ID), nil), 200, nil)
	for _, tc := range []struct {
		query string
		total int
	}{
		{"?license_type=A1&status=COMPLETED", 1},
		{"?license_type=B", 0},
		{"", 1},
	} {
		var page struct {
			Data  []dto.HistoryItem `json:"data"`
			Total int               `json:"total"`
		}
		decodeBody(t, request(t, r, jwt, "GET", "/api/v1/exams"+tc.query, nil), 200, &page)
		if page.Total != tc.total || len(page.Data) != tc.total {
			t.Fatalf("unexpected history: %+v", page)
		}
		for _, item := range page.Data {
			if item.LicenseType != "A1" || item.ID != exam.ID {
				t.Fatalf("history relabeled: %+v", item)
			}
		}
	}
	other, err := f.jwt.GenerateAccessToken(f.other.ID)
	requireNoError(t, err)
	var page struct {
		Total int `json:"total"`
	}
	decodeBody(t, request(t, r, other, "GET", "/api/v1/exams?license_type=A1", nil), 200, &page)
	if page.Total != 0 {
		t.Fatal("another user's exam leaked")
	}
}

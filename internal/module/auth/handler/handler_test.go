package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	dto "github.com/han/go-ecommerce/internal/module/auth/dto/request"
	"github.com/han/go-ecommerce/internal/module/auth/model"
	"github.com/han/go-ecommerce/internal/module/auth/service"
	"gorm.io/gorm"
)

type serviceStub struct {
	register func(dto.RegisterRequest) (*model.User, error)
}

func (s *serviceStub) Register(req dto.RegisterRequest) (*model.User, error) {
	return s.register(req)
}

func performRegister(t *testing.T, authService AuthService, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", NewHandler(authService).Register)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestRegisterReturnsCreatedUserWithoutPassword(t *testing.T) {
	authService := &serviceStub{register: func(req dto.RegisterRequest) (*model.User, error) {
		return &model.User{Model: gormModel(3), Name: req.Name, Email: req.Email, Password: "secret"}, nil
	}}
	recorder := performRegister(t, authService, `{"name":"Nguyen Van An","email":"an@example.com","password":"password123"}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	data := body["data"].(map[string]interface{})
	if _, exists := data["password"]; exists {
		t.Fatal("response exposes password")
	}
}

func TestRegisterValidatesRequest(t *testing.T) {
	called := false
	authService := &serviceStub{register: func(dto.RegisterRequest) (*model.User, error) {
		called = true
		return nil, nil
	}}
	recorder := performRegister(t, authService, `{"name":"A","email":"invalid","password":"short"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if called {
		t.Fatal("service was called for an invalid request")
	}
}

func TestRegisterMapsExistingEmailToConflict(t *testing.T) {
	authService := &serviceStub{register: func(dto.RegisterRequest) (*model.User, error) {
		return nil, service.ErrEmailAlreadyExists
	}}
	recorder := performRegister(t, authService, `{"name":"Nguyen Van An","email":"an@example.com","password":"password123"}`)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}

func gormModel(id uint) gorm.Model {
	return gorm.Model{ID: id}
}

package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/module/auth/dto/mapper"
	dto "github.com/han/go-ecommerce/internal/module/auth/dto/request"
	"github.com/han/go-ecommerce/internal/module/auth/model"
	sv "github.com/han/go-ecommerce/internal/module/auth/service"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*model.User, error)
	GetUser(id string) (*model.User, error)
}

type Handler struct {
	service AuthService
}

type response struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

func NewHandler(service AuthService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request",
		})
		return
	}

	user, err := h.service.Register(req)
	if errors.Is(err, sv.ErrEmailAlreadyExists) {
		c.JSON(http.StatusConflict, response{
			StatusCode: http.StatusConflict,
			Message:    "Email already exists",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			StatusCode: http.StatusInternalServerError,
			Message:    "Cannot create user",
		})
		return
	}

	c.JSON(http.StatusCreated, response{
		StatusCode: http.StatusCreated,
		Message:    "Register success",
		Data: gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, mapper.ToUserResponse(user))
}

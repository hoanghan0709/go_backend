package delivery

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	commonError "github.com/han/go-ecommerce/internal/auth/common_error"
	"github.com/han/go-ecommerce/internal/auth/dto"
	"github.com/han/go-ecommerce/internal/auth/usecase"
)

type Handler struct {
	service usecase.AuthService
}

type response struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

func New(service usecase.AuthService) *Handler {
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
	if errors.Is(err, commonError.ErrEmailAlreadyExists) {
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
		Data:       dto.ToUserResponse(user),
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
	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

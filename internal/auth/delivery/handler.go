package delivery

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	commonError "github.com/han/go-ecommerce/internal/auth/common_error"
	"github.com/han/go-ecommerce/internal/auth/dto"
	"github.com/han/go-ecommerce/internal/auth/usecase"
	common "github.com/han/go-ecommerce/internal/common/model"
)

type Handler struct {
	service usecase.AuthService
}

func New(service usecase.AuthService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request",
		})
		return
	}

	user, err := h.service.Register(req)
	if errors.Is(err, commonError.ErrEmailAlreadyExists) {
		c.JSON(http.StatusConflict, common.Response{
			StatusCode: http.StatusConflict,
			Message:    "Email already exists",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "Cannot create user",
		})
		return
	}

	c.JSON(http.StatusCreated, common.Response{
		StatusCode: http.StatusCreated,
		Message:    "Register success",
		Data:       dto.ToUserResponse(user),
	})
}
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

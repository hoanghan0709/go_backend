package delivery

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	commonError "github.com/han/go-ecommerce/internal/auth/common_error"
	"github.com/han/go-ecommerce/internal/auth/dto"
	"github.com/han/go-ecommerce/internal/auth/usecase"
	common "github.com/han/go-ecommerce/internal/common/model"
	"github.com/han/go-ecommerce/internal/middleware"
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

	value, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized",
		})
		return
	}

	userID, ok := value.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "Invalid user context",
		})
		return
	}
	user, err := h.service.GetUser(strconv.FormatUint(uint64(userID), 10))

	// id := c.Param("id")
	// user, err := h.service.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

func (h *Handler) Login(c *gin.Context) {
	fmt.Println("🔥 LOGIN HANDLER CALLED")
	var request dto.LoginRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid login request",
		})
		return
	}
	result, errLogin := h.service.Login(request)
	if errors.Is(errLogin, commonError.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, common.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid email or password",
		})
		return
	}
	if errLogin != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    errLogin.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, common.Response{
		StatusCode: http.StatusOK,
		Message:    "Login success",
		Data: dto.ToLoginResponse(
			result.User,
			result.AccessToken,
			result.RefreshToken,
		),
	})
}
func (h *Handler) Refresh(c *gin.Context) {
	var request dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid refresh request",
		})
		return
	}

	result, err :=
		h.service.Refresh(request.RefreshToken)

	if errors.Is(err, usecase.ErrInvalidRefreshToken) ||
		errors.Is(err, usecase.ErrExpiredRefreshToken) {
		c.JSON(http.StatusUnauthorized, common.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid or expired refresh token",
		})
		return
	}

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			common.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			},
		)
		return
	}

	c.JSON(http.StatusOK, common.Response{
		StatusCode: http.StatusOK,
		Message:    "Refresh token success",
		Data: dto.TokenResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresIn:    12 * 60 * 60,
		},
	})
}

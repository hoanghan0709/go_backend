package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/han/go-ecommerce/internal/common/model"
	commonValidator "github.com/han/go-ecommerce/internal/common/validator"
	"github.com/han/go-ecommerce/internal/question/dto"
	"github.com/han/go-ecommerce/internal/question/usecase"
)

type Handler struct {
	usecase usecase.QuestionService
}

func New(usecase usecase.QuestionService) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) GetListQuestion(c *gin.Context) {
	listQuestion, err := h.usecase.GetListQuestion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToListQuestionResponse(listQuestion))
}

func (h *Handler) CreateQuestion(c *gin.Context) {
	var req dto.CreateQuestionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if messages, ok := commonValidator.ValidationMessages(err); ok {
			c.JSON(http.StatusBadRequest, common.Response{
				StatusCode: http.StatusBadRequest,
				Message:    "Validation failed",
				Data:       messages,
			})
			return
		}

		c.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid JSON request",
		})
		return
	}

	if err := h.usecase.CreateQuestion(&req); err != nil {
		c.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, common.Response{
		StatusCode: http.StatusCreated,
		Message:    "Create question success",
	})
}

package delivery

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/han/go-ecommerce/internal/common/model"
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
	fmt.Print("listQuestion", listQuestion)
	c.JSON(http.StatusOK, listQuestion)

}

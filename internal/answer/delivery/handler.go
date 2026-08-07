package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/answer/dto"
	usecaseAnswer "github.com/han/go-ecommerce/internal/answer/usecase"
	common "github.com/han/go-ecommerce/internal/common/model"
)

type Handler struct {
	usecase usecaseAnswer.AnswerI
}

func New(usecase usecaseAnswer.AnswerI) *Handler {
	return &Handler{usecase: usecase}
}
func (h *Handler) GetList(ctx *gin.Context) {

	listAnswer, err := h.usecase.GetList()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, common.Response{
		StatusCode: http.StatusOK,
		Message:    "Success",
		Data:       dto.ToListAnswer(listAnswer),
	})
}

func (h *Handler) Create(ctx *gin.Context) {

	var req dto.AnswerRequest
	errInvalid := ctx.ShouldBindJSON(&req)
	if errInvalid != nil {
		ctx.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid JSON request",
		})
		return
	}
	if err := h.usecase.Create(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, common.Response{
		StatusCode: http.StatusCreated,
		Message:    "Create answer success",
	})
}

package delivery

import (
	"net/http"
	"strconv"

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

// func (h *Handler) GetList(ctx *gin.Context) {

// 	listAnswer, err := h.usecase.GetList()
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, common.Response{
// 			StatusCode: http.StatusInternalServerError,
// 			Message:    err.Error(),
// 		})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, common.Response{
// 		StatusCode: http.StatusOK,
// 		Message:    "Success",
// 		Data:       dto.ToListAnswer(listAnswer),
// 	})
// }

// func (h *Handler) Create(ctx *gin.Context) {

// 	var req dto.AnswerRequest
// 	errInvalid := ctx.ShouldBindJSON(&req)
// 	if errInvalid != nil {
// 		ctx.JSON(http.StatusBadRequest, common.Response{
// 			StatusCode: http.StatusBadRequest,
// 			Message:    "Invalid JSON request",
// 		})
// 		return
// 	}
// 	if err := h.usecase.Create(&req); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, common.Response{
// 			StatusCode: http.StatusInternalServerError,
// 			Message:    err.Error(),
// 		})
// 		return
// 	}

//		ctx.JSON(http.StatusCreated, common.Response{
//			StatusCode: http.StatusCreated,
//			Message:    "Create answer success",
//		})
//	}
func parseUintParam(
	ctx *gin.Context,
	name string,
) (uint, error) {
	value, err := strconv.ParseUint(
		ctx.Param(name),
		10,
		64,
	)
	if err != nil {
		return 0, err
	}

	return uint(value), nil
}

func (h *Handler) Create(ctx *gin.Context) {
	questionID, err := parseUintParam(ctx, "question_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid question ID",
		})
		return
	}

	var req dto.AnswerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid JSON request",
		})
		return
	}

	if err := h.usecase.Create(questionID, &req); err != nil {
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

func (h *Handler) GetListByQuestionID(ctx *gin.Context) {
	questionID, err := parseUintParam(ctx, "question_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid question ID",
		})
		return
	}

	answers, err := h.usecase.GetListByQuestionID(questionID)
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
		Data:       answers,
	})
}

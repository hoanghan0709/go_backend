package delivery

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/common/httpapi"
	"github.com/han/go-ecommerce/internal/exam/dto"
	"github.com/han/go-ecommerce/internal/exam/usecase"
)

func RegisterRoutes(r *gin.Engine, h *Handler, auth gin.HandlerFunc) {
	api := r.Group("/api/v1/exams", auth)
	api.POST("", func(c *gin.Context) {
		var in dto.StartRequest
		if !httpapi.Bind(c, &in) {
			return
		}
		out, err := h.usecase.Start(httpapi.User(c), in.TestID)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(201, out)
	})
	api.GET("", func(c *gin.Context) {
		f, ok := httpapi.Filter(c)
		if !ok {
			return
		}
		out, n, err := h.usecase.History(httpapi.User(c), f.Page, f.Limit)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		httpapi.Page(c, dto.ToHistory(out), n, f)
	})
	api.GET("/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		out, err := h.usecase.Get(httpapi.User(c), id)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
	api.PUT("/:id/answers/:question_id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		qid, ok := httpapi.ID(c, "question_id")
		if !ok {
			return
		}
		var in dto.AnswerRequest
		if !httpapi.Bind(c, &in) {
			return
		}
		out, err := h.usecase.SaveAnswer(httpapi.User(c), id, qid, *in.AnswerID)
		if errors.Is(err, usecase.ErrClosed) {
			c.JSON(409, gin.H{"message": err.Error()})
			return
		}
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
	api.POST("/:id/submit", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		out, err := h.usecase.Submit(httpapi.User(c), id)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
}

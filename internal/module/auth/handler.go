// Handler chỉ làm 4 việc.

// 1. Nhận Request

// 2. Parse JSON

// 3. Gọi Service

// 4. Trả Response
package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// logic
func (h *Handler) Register(c *gin.Context) {

	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest,

			gin.H{

				"message": err.Error(),
			})

		return
	}

	if err := h.service.Register(req); err != nil {

		c.JSON(http.StatusBadRequest,

			gin.H{

				"message": err.Error(),
			})

		return
	}

	c.JSON(

		http.StatusCreated,

		gin.H{

			"message": "register success",
		},
	)
}

// Handler chỉ làm 4 việc.

// 1. Nhận Request

// 2. Parse JSON

// 3. Gọi Service

// 4. Trả Response
package handler

import (
	// "net/http"
	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/module/auth/dto"
	sv "github.com/han/go-ecommerce/internal/module/auth/service"
)

type Handler struct {
	service *sv.Service
}

func NewHandler(service *sv.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// logic
func (h *Handler) Register(c *gin.Context) {

	var req dto.RegisterRequest

	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest,
	// 		gin.H{
	// 			"message": err.Error(),
	// 		})
	// 	return
	// }

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, sv.
			BaseResponse{
			StatusCode: 400,
			Message:    "Invalid request",
		})
		return
	}
	response := h.service.Register(req)
	c.JSON(response.StatusCode, response)
	// if err := h.service.Register(req); err != nil {
	// 	c.JSON(http.StatusBadRequest,
	// 		gin.H{
	// 			"message": err.Error(),
	// 		})
	// 	return
	// }
	// c.JSON(
	// 	http.StatusCreated,
	// 	gin.H{
	// 		"message": "register success",
	// 	},
	// )
}

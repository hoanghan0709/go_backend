package delivery

//
import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	question := r.Group("/question")

	question.GET("/getListQuestion",
		h.GetListQuestion)
}

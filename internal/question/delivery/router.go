package delivery

//
import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	question := r.Group("/question")
	//get List
	question.GET("/getListQuestion",
		h.GetListQuestion)
	//create new Question
	question.POST("/create",
		h.CreateQuestion)
}

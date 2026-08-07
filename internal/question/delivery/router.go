package delivery

//
import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler,
	authMiddleware gin.HandlerFunc) {

	question := r.Group("/question")

	protected := question.Group("")

	protected.Use(authMiddleware)
	{
		//get List
		protected.GET("/getList",
			h.GetListQuestion)
		//create new Question
		protected.POST("/create",
			h.CreateQuestion)

	}
}

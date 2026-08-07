package delivery

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware gin.HandlerFunc) {

	answer := r.Group("/answer")

	protected := answer.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/getList", handler.GetList)
		protected.POST("/create", handler.Create)
		// protected.DELETE("/delete", handler.Delete)
	}
}

//POST /questions/:question_id/answers
// PUT  /questions/:question_id/answers/:answer_id
// DELETE /questions/:question_id/answers/:answer_id

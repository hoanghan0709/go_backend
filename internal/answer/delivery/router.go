package delivery

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	r *gin.Engine,
	handler *Handler,
	authMiddleware gin.HandlerFunc,
) {
	questions := r.Group("/questions")
	questions.Use(authMiddleware)

	{
		//CREATE answer by question
		questions.POST(
			"/:question_id/answers",
			handler.Create,
		)
		//GET answer by question
		questions.GET(
			"/:question_id/answers",
			handler.GetListByQuestionID,
		)

		//PUT answer by question
		questions.PUT(
			"/:question_id/answers/:answer_id",
			handler.Update,
		)
		//delete answer by question
		questions.DELETE(
			"/:question_id/answers/:answer_id",
			handler.Delete,
		)
	}
}

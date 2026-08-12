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
		questions.POST(
			"/:question_id/answers",
			handler.Create,
		)

		questions.GET(
			"/:question_id/answers",
			handler.GetListByQuestionID,
		)

		questions.PUT(
			"/:question_id/answers/:answer_id",
			handler.Update,
		)

		// questions.DELETE(
		//     "/:question_id/answers/:answer_id",
		//     handler.Delete,
		// )
	}
}

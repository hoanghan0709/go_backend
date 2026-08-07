package dto

type AnswerRequest struct {
	Content    string `json:"content" binding:"required"`
	Label      string `json:"label" binding:"required,max=2"`
	Order      int    `json:"order" binding:"required,min=1"`
	IsCorrect  bool   `json:"is_correct"`
	QuestionID uint   `json:"question_id" binding:"required"`
}

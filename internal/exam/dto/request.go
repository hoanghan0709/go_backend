package dto

type StartRequest struct {
	TestID uint `json:"test_id" binding:"required"`
}

// Pointer phân biệt thiếu answer_id với answer_id=0 (bỏ chọn).
type AnswerRequest struct {
	AnswerID *uint `json:"answer_id" binding:"required"`
}

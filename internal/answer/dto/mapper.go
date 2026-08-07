package dto

import "github.com/han/go-ecommerce/internal/answer/model"

func ToListAnswer(answer []model.Answer) []AnswerResponse {
	res := make([]AnswerResponse, 0, len(answer))
	for i := range answer {
		res = append(res, toAnswerResponse(&answer[i]))
	}
	return res
}

func toAnswerResponse(t *model.Answer) AnswerResponse {
	return AnswerResponse{
		Content:    t.Content,
		Label:      t.Label,
		Order:      t.Order,
		IsCorrect:  t.IsCorrect,
		QuestionID: t.QuestionID,
	}
}

package dto

import "github.com/han/go-ecommerce/internal/question/model"

func ToListQuestionResponse(questions []model.Question) []ListQuestionResponse {
	res := make([]ListQuestionResponse, 0, len(questions))
	for i := range questions {
		res = append(res, toQuestionResponse(&questions[i]))
	}
	return res
}

func toQuestionResponse(t *model.Question) ListQuestionResponse {
	return ListQuestionResponse{
		Title:       t.Title,
		Content:     t.Content,
		Image:       t.Image,
		IsCritical:  t.IsCritical,
		Position:    t.Position,
		LicenseType: t.LicenseType,
		CategoryID:  t.CategoryID,
		Category:    t.Category,
		Answers:     t.Answers,
	}
}

package delivery

import "github.com/han/go-ecommerce/internal/exam/usecase"

type Handler struct{ usecase usecase.Interface }

func New(u usecase.Interface) *Handler { return &Handler{usecase: u} }

.PHONY: new

new:
	@mkdir -p internal/$(name)/{delivery,usecase,repository,model,dto,errors}

	@printf 'package delivery\n\nimport "github.com/gin-gonic/gin"\n\n\
type Usecase interface{}\n\n\
type Handler struct {\n\tusecase Usecase\n}\n\n\
func New(usecase Usecase) *Handler {\n\treturn &Handler{usecase: usecase}\n}\n' \
> internal/$(name)/delivery/handler.go

	@printf 'package delivery\n\nimport "github.com/gin-gonic/gin"\n\n\
func RegisterRoutes(r *gin.Engine, h *Handler) {\n\n}\n' \
> internal/$(name)/delivery/router.go

	@printf 'package usecase\n\n\
type Repository interface{}\n\n\
type Usecase struct {\n\trepo Repository\n}\n\n\
func New(repo Repository) *Usecase {\n\treturn &Usecase{repo: repo}\n}\n' \
> internal/$(name)/usecase/usecase.go

	@printf 'package usecase\n\n\
type Interface interface{}\n' \
> internal/$(name)/usecase/interface.go

	@printf 'package repository\n\n\
type Repository struct {}\n\n\
func New() *Repository {\n\treturn &Repository{}\n}\n' \
> internal/$(name)/repository/repository.go

	@printf 'package repository\n\n\
type Interface interface{}\n' \
> internal/$(name)/repository/interface.go

	@printf 'package model\n\n\
type Model struct {\n\tID uint `gorm:"primaryKey"`\n}\n' \
> internal/$(name)/model/model.go

	@printf 'package dto\n\n\
type Request struct{}\n' \
> internal/$(name)/dto/request.go

	@printf 'package dto\n\n\
type Response struct{}\n' \
> internal/$(name)/dto/response.go

	@printf 'package dto\n\n' \
> internal/$(name)/dto/mapper.go

	@printf 'package errors\n\nimport "errors"\n\n\
var ErrNotFound = errors.New("not found")\n' \
> internal/$(name)/errors/errors.go

	@echo "✅ Feature $(name) created."
	.PHONY: common

common:
	@mkdir -p internal/common/{response,validator,utils,constants,logger,pagination}

	@printf 'package response\n\n\
type Response struct {\n\
\tStatusCode int    `json:"statusCode"`\n\
\tMessage    string `json:"message"`\n\
\tData       any    `json:"data,omitempty"`\n\
}\n' \
> internal/common/response/response.go

	@printf 'package response\n' \
> internal/common/response/success.go

	@printf 'package validator\n' \
> internal/common/validator/validator.go

	@printf 'package validator\n' \
> internal/common/validator/custom.go

	@printf 'package utils\n' \
> internal/common/utils/password.go

	@printf 'package utils\n' \
> internal/common/utils/random.go

	@printf 'package utils\n' \
> internal/common/utils/string.go

	@printf 'package utils\n' \
> internal/common/utils/time.go

	@printf 'package utils\n' \
> internal/common/utils/uuid.go

	@printf 'package constants\n' \
> internal/common/constants/role.go

	@printf 'package constants\n' \
> internal/common/constants/pagination.go

	@printf 'package constants\n' \
> internal/common/constants/status.go

	@printf 'package logger\n' \
> internal/common/logger/logger.go

	@printf 'package pagination\n' \
> internal/common/pagination/pagination.go

	@echo "✅ Common package created."
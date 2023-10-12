package validate

import (
	"event-validation/internal/repo/git"

	"github.com/labstack/echo"
)

type Service interface {
	Pull(c echo.Context) error
	Validate(c echo.Context) error
}

type service struct {
	gitRepo git.GitRepository
}

func NewService(gitRepo git.GitRepository) Service {
	return &service{gitRepo: gitRepo}
}

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResp struct {
	Error *Resp `json:"error"`
}

func buildErr(code int, msg string) *ErrorResp {
	resp := &Resp{
		Code: code, Message: msg,
	}
	return &ErrorResp{resp}
}

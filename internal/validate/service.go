package validate

import (
	"event-validation/internal/config"
	"event-validation/internal/repo/git"

	"github.com/labstack/echo"
)

type Service interface {
	CloneSpec(c echo.Context) error
	Validate(c echo.Context) error
}

type service struct {
	gitRepo          git.Git
	eventsFileConfig *config.EventsFile
	eventData        *eventYamlModel
}

func NewService(git git.Git, evf *config.EventsFile) Service {
	return &service{
		gitRepo:          git,
		eventsFileConfig: evf,
	}
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

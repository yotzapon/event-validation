package validate

import (
	"github.com/labstack/echo"
	"log"
)

func (s *service) CloneSpec(c echo.Context) error {
	err := s.gitRepo.Clone()
	if err != nil {
		log.Printf("git clone failed: %v", err.Error())
		return c.JSON(400, buildErr(400, err.Error()))
	}

	return c.JSON(200, Resp{200, "Ok"})
}

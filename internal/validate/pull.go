package validate

import (
	"fmt"
	"log"

	"github.com/labstack/echo"
)

func (s *service) Pull(c echo.Context) error {
	fmt.Print("Pullll")
	err := s.gitRepo.Pull()
	if err != nil {
		log.Printf("Pull failed: %v", err.Error())
		return c.JSON(400, buildErr(400, err.Error()))
	}

	log.Print("Pull Success")
	return c.JSON(200, Resp{200, "Ok"})
}

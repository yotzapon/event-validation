package main

import (
	"context"
	"event-validation/internal/repo/git"
	"event-validation/internal/validate"
	"fmt"
	"os"
	"os/signal"

	"time"

	cfg "event-validation/internal/config"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	// Echo instance
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Logger.SetHeader(`${time_rfc3339_nano} | ${level} | ${long_file}:${line} |`)

	conf := new(cfg.Config)
	if err := cfg.ReadConfigFile(conf, "ENV"); err != nil {
		e.Logger.Fatal("Unexpected error to read configuration: %v.", err)
	}

	gitRepo := git.NewGit(&conf.Repo.Git)

	v := validate.NewService(gitRepo, &conf.Repo.EventsFile)

	v1 := e.Group("/v1")
	v1.GET("/api-spec", v.Download).Name = "clone"
	v1.POST("/api-spec/event", v.Validate).Name = "validate"

	// Start server
	go func() {
		port := fmt.Sprintf(":%v", conf.Server.Port)
		if err := e.Start(port); err != nil {
			e.Logger.Infof("shutting down the server (%v)", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	e.Logger.Info("receive interrupt signal")
	ctx, cancel := context.WithTimeout(context.Background(), conf.Server.ShutdownTimeout*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	e.Logger.Info("server exited properly")

	return nil
}

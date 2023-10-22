package git

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"os"
)

type Git interface {
	Clone() error
}

type gitRepository struct {
	config *Config
}

func NewGit(conf *Config) Git {
	return &gitRepository{
		config: conf,
	}
}

const gitToken = "GIT_TOKEN"

func (g *gitRepository) Clone() error {

	path := g.config.Destination

	// remove the existing repository
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		_ = os.RemoveAll(path)
	}

	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL: g.config.Url,
		Auth: &http.BasicAuth{
			Username: g.config.Auth.Username,
			Password: os.Getenv(gitToken),
		},
	})

	if err != nil {
		return err
	}

	return nil
}

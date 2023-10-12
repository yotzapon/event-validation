package git

import (
	"github.com/labstack/gommon/log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type GitRepository interface {
	Pull() error
}

type gitRepository struct {
	config  *Config
	gitRepo git.Repository
}

func NewGit(conf *Config) GitRepository {
	return &gitRepository{
		config: conf,
	}
}

func (g *gitRepository) clone() error {

	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: g.config.Url,
		Auth: &http.BasicAuth{
			Username: g.config.Auth.Username,
			Password: g.config.Auth.Password,
		},
	})

	if err != nil {
		return err
	}

	g.gitRepo = *repo

	return nil
}

func (g *gitRepository) Pull() error {

	err := g.clone()
	if err != nil {
		return err
	}

	// Get the working directory for the repository
	w, err := g.gitRepo.Worktree()
	if err != nil {
		return err
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.Pull(&git.PullOptions{RemoteName: g.config.RemoteName})
	if err != nil {
		return err
	}

	// Print the latest commit that was just pulled
	ref, err := g.gitRepo.Head()
	if err != nil {
		return err
	}
	commit, err := g.gitRepo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	log.Infof("----last commit: %v ----", commit)

	return nil
}

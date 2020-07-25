package controller

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
)

func RepoUpdate(dir, url string) (*git.Repository, error) {
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})
	if err == nil {
		log.Printf("found new repo: %q\n", url)
		return repo, nil
	}
	if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return nil, fmt.Errorf("plain clone: %w", err)
	}
	repo, err = git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("plain open: %w", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("open worktree: %w", err)
	}
	err = worktree.Pull(&git.PullOptions{Force: true})
	if err == nil {
		log.Printf("found new update: %q\n", url)
		return repo, nil
	}
	if !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil, fmt.Errorf("pull worktree: %w", err)
	}
	return repo, nil
}

func RepoGetHEAD(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("get head: %w", err)
	}
	return head.Hash().String()[:8], nil
}

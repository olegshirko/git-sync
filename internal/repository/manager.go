package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// Manager управляет операциями с Git-репозиториями
type Manager struct {
	tempDir string
}

// NewManager создает новый экземпляр Manager
func NewManager(tempDir string) *Manager {
	return &Manager{
		tempDir: tempDir,
	}
}

// Clone клонирует репозиторий по URL в указанную директорию
func (m *Manager) Clone(repoURL, path, token, sshKeyPath string) (*git.Repository, error) {
	cloneOptions := &git.CloneOptions{
		URL: repoURL,
	}

	if token != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: "oauth2", // Для GitLab Personal Access Token
			Password: token,
		}
	} else if sshKeyPath != "" {
		sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
		if err != nil {
			return nil, fmt.Errorf("не удалось создать SSH-аутентификацию: %w", err)
		}
		cloneOptions.Auth = sshAuth
	}

	repo, err := git.PlainClone(path, false, cloneOptions)
	if err != nil {
		return nil, fmt.Errorf("не удалось клонировать репозиторий %s: %w", repoURL, err)
	}
	return repo, nil
}

// Pull обновляет репозиторий
func (m *Manager) Pull(repo *git.Repository, token, sshKeyPath string) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("не удалось получить Worktree: %w", err)
	}

	pullOptions := &git.PullOptions{}
	if token != "" {
		pullOptions.Auth = &http.BasicAuth{
			Username: "oauth2",
			Password: token,
		}
	} else if sshKeyPath != "" {
		sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
		if err != nil {
			return fmt.Errorf("не удалось создать SSH-аутентификацию: %w", err)
		}
		pullOptions.Auth = sshAuth
	}

	err = w.Pull(pullOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("не удалось выполнить pull: %w", err)
	}
	return nil
}

// Push отправляет изменения в удаленный репозиторий
func (m *Manager) Push(repo *git.Repository, token, sshKeyPath string) error {
	pushOptions := &git.PushOptions{}
	if token != "" {
		pushOptions.Auth = &http.BasicAuth{
			Username: "oauth2",
			Password: token,
		}
	} else if sshKeyPath != "" {
		sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
		if err != nil {
			return fmt.Errorf("не удалось создать SSH-аутентификацию: %w", err)
		}
		pushOptions.Auth = sshAuth
	}

	err := repo.Push(pushOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("не удалось выполнить push: %w", err)
	}
	return nil
}

// CleanTempDir очищает временную директорию
func (m *Manager) CleanTempDir() error {
	return os.RemoveAll(m.tempDir)
}

// CreateTempRepoPath создает путь для временного репозитория
func (m *Manager) CreateTempRepoPath(repoName string) string {
	return filepath.Join(m.tempDir, repoName)
}

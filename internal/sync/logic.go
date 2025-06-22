package sync

import (
	"fmt"
	"log"
	"strings"

	"git-sync/internal/repository"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// Logic содержит логику синхронизации репозиториев
type Logic struct {
	repoManager *repository.Manager
}

// NewLogic создает новый экземпляр Logic
func NewLogic(repoManager *repository.Manager) *Logic {
	return &Logic{
		repoManager: repoManager,
	}
}

// Synchronize выполняет двустороннюю синхронизацию между двумя репозиториями
func (l *Logic) Synchronize(gitlabURL, privateRepoURL, gitlabToken, sshKeyPath string) error {
	gitlabRepoName := getRepoNameFromURL(gitlabURL)
	privateRepoName := getRepoNameFromURL(privateRepoURL)

	gitlabLocalPath := l.repoManager.CreateTempRepoPath(gitlabRepoName)
	privateLocalPath := l.repoManager.CreateTempRepoPath(privateRepoName)

	defer func() {
		log.Printf("Очистка временных директорий: %s, %s", gitlabLocalPath, privateLocalPath)
		if err := l.repoManager.CleanTempDir(); err != nil {
			log.Printf("Ошибка очистки временной директории: %v", err)
		}
	}()

	// Клонирование/обновление GitLab репозитория
	log.Printf("Клонирование/обновление GitLab репозитория: %s в %s", gitlabURL, gitlabLocalPath)
	gitlabRepo, err := l.repoManager.Clone(gitlabURL, gitlabLocalPath, gitlabToken, "")
	if err != nil {
		return fmt.Errorf("не удалось клонировать/обновить GitLab репозиторий: %w", err)
	}
	if err := l.repoManager.Pull(gitlabRepo, gitlabToken, ""); err != nil {
		log.Printf("Предупреждение: не удалось выполнить pull для GitLab репозитория: %v", err)
	}

	// Клонирование/обновление приватного репозитория
	log.Printf("Клонирование/обновление приватного репозитория: %s в %s", privateRepoURL, privateLocalPath)
	privateRepo, err := l.repoManager.Clone(privateRepoURL, privateLocalPath, "", sshKeyPath)
	if err != nil {
		return fmt.Errorf("не удалось клонировать/обновить приватный репозиторий: %w", err)
	}
	if err := l.repoManager.Pull(privateRepo, "", sshKeyPath); err != nil {
		log.Printf("Предупреждение: не удалось выполнить pull для приватного репозитория: %v", err)
	}

	// Синхронизация GitLab -> Private
	log.Printf("Синхронизация GitLab -> Private для %s", gitlabURL)
	if err := l.syncBranches(gitlabRepo, privateRepo, gitlabToken, sshKeyPath); err != nil {
		return fmt.Errorf("ошибка синхронизации GitLab -> Private: %w", err)
	}

	// Синхронизация Private -> GitLab
	log.Printf("Синхронизация Private -> GitLab для %s", privateRepoURL)
	if err := l.syncBranches(privateRepo, gitlabRepo, "", gitlabToken); err != nil {
		return fmt.Errorf("ошибка синхронизации Private -> GitLab: %w", err)
	}

	return nil
}

// syncBranches синхронизирует ветки между source и destination репозиториями
func (l *Logic) syncBranches(sourceRepo, destinationRepo *git.Repository, sourceToken, destToken string) error {
	// Получаем remote для source репозитория
	sourceRemote, err := sourceRepo.Remote("origin")
	if err != nil {
		return fmt.Errorf("не удалось получить remote 'origin' для source репозитория: %w", err)
	}

	// Получаем URL source репозитория
	sourceURL := sourceRemote.Config().URLs[0]
	log.Printf("Source URL: %s", sourceURL)

	// Получаем метод аутентификации для source репозитория
	sourceAuth, err := l.getAuthMethod(sourceToken, "")
	if err != nil {
		return fmt.Errorf("не удалось получить метод аутентификации для source: %w", err)
	}

	// Получаем метод аутентификации для destination репозитория
	destAuth, err := l.getAuthMethod(destToken, "")
	if err != nil {
		return fmt.Errorf("не удалось получить метод аутентификации для destination: %w", err)
	}

	// Добавляем source репозиторий как remote в destination репозиторий
	remoteName := "sync-source"
	_, err = destinationRepo.CreateRemote(&gitconfig.RemoteConfig{
		Name: remoteName,
		URLs: []string{sourceURL},
	})
	if err != nil && err != git.ErrRemoteExists {
		return fmt.Errorf("не удалось создать remote %s: %w", remoteName, err)
	}

	// Получаем remote для fetch
	syncRemote, err := destinationRepo.Remote(remoteName)
	if err != nil {
		return fmt.Errorf("не удалось получить remote %s: %w", remoteName, err)
	}

	// Выполняем fetch из source репозитория
	log.Printf("Выполняем fetch из source репозитория")
	err = syncRemote.Fetch(&git.FetchOptions{
		Auth: sourceAuth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("не удалось выполнить fetch из source репозитория: %w", err)
	}

	// Получаем ветки из source репозитория
	sourceBranches, err := sourceRepo.Branches()
	if err != nil {
		return fmt.Errorf("не удалось получить ветки из source репозитория: %w", err)
	}

	// Получаем worktree для destination репозитория
	destWorktree, err := destinationRepo.Worktree()
	if err != nil {
		return fmt.Errorf("не удалось получить worktree для destination репозитория: %w", err)
	}

	err = sourceBranches.ForEach(func(branchRef *plumbing.Reference) error {
		branchName := branchRef.Name().Short()
		if strings.HasPrefix(branchName, "HEAD") {
			return nil // Пропускаем HEAD
		}

		log.Printf("Синхронизация ветки: %s", branchName)

		// Получаем reference из fetched данных
		fetchedRef := plumbing.NewRemoteReferenceName(remoteName, branchName)
		fetchedBranchRef, err := destinationRepo.Reference(fetchedRef, true)
		if err != nil {
			log.Printf("Предупреждение: не удалось найти fetched ветку %s: %v", branchName, err)
			return nil // Пропускаем эту ветку
		}

		// Проверяем, существует ли ветка в destination репозитории
		destBranchRef, err := destinationRepo.Reference(branchRef.Name(), true)
		needsUpdate := false

		if err == plumbing.ErrReferenceNotFound {
			// Ветка не существует в destination репозитории
			log.Printf("Ветка %s не существует в destination репозитории, создаем новую", branchName)
			needsUpdate = true
		} else if err != nil {
			return fmt.Errorf("ошибка при поиске ветки %s в destination репозитории: %w", branchName, err)
		} else {
			// Ветка существует, проверяем, нужно ли обновление
			if destBranchRef.Hash() != fetchedBranchRef.Hash() {
				log.Printf("Ветка %s требует обновления (dest: %s, source: %s)", branchName, destBranchRef.Hash().String(), fetchedBranchRef.Hash().String())
				needsUpdate = true
			} else {
				log.Printf("Ветка %s уже синхронизирована", branchName)
			}
		}

		if needsUpdate {
			// Переключаемся на ветку или создаем её
			checkoutOptions := &git.CheckoutOptions{
				Branch: branchRef.Name(),
				Create: err == plumbing.ErrReferenceNotFound,
				Force:  true,
			}

			if checkoutOptions.Create {
				// Для новой ветки устанавливаем hash из fetched данных
				checkoutOptions.Hash = fetchedBranchRef.Hash()
			}

			err = destWorktree.Checkout(checkoutOptions)
			if err != nil {
				return fmt.Errorf("не удалось переключиться на ветку %s в destination репозитории: %w", branchName, err)
			}

			// Если ветка уже существует, обновляем её до нужного коммита
			if !checkoutOptions.Create {
				err = destWorktree.Reset(&git.ResetOptions{
					Commit: fetchedBranchRef.Hash(),
					Mode:   git.HardReset,
				})
				if err != nil {
					return fmt.Errorf("не удалось обновить ветку %s до коммита %s: %w", branchName, fetchedBranchRef.Hash().String(), err)
				}
			}

			// Выполняем push в remote репозиторий
			pushOptions := &git.PushOptions{
				Auth: destAuth,
			}

			err = destinationRepo.Push(pushOptions)
			if err != nil {
				if err == git.ErrNonFastForwardUpdate {
					log.Printf("Предупреждение: не удалось выполнить fast-forward push для ветки %s. Пропускаем синхронизацию этой ветки, чтобы избежать принудительной перезаписи.", branchName)
					return nil // Пропускаем эту ветку
				}
				if err != git.NoErrAlreadyUpToDate {
					return fmt.Errorf("не удалось выполнить push ветки %s в destination репозиторий: %w", branchName, err)
				}
			}

			log.Printf("Ветка %s успешно синхронизирована", branchName)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("ошибка при итерации по веткам source репозитория: %w", err)
	}

	return nil
}

// getAuthMethod возвращает метод аутентификации на основе токена или SSH-ключа
func (l *Logic) getAuthMethod(token, sshKeyPath string) (transport.AuthMethod, error) {
	if token != "" {
		return &http.BasicAuth{
			Username: "oauth2", // Для GitLab Personal Access Token
			Password: token,
		}, nil
	} else if sshKeyPath != "" {
		sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
		if err != nil {
			return nil, fmt.Errorf("не удалось создать SSH-аутентификацию: %w", err)
		}
		return sshAuth, nil
	}
	return nil, nil // Нет аутентификации
}

// getRepoNameFromURL извлекает имя репозитория из URL
func getRepoNameFromURL(repoURL string) string {
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		nameWithExt := parts[len(parts)-1]
		return strings.TrimSuffix(nameWithExt, ".git")
	}
	return repoURL
}

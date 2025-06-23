package gitlab

import (
	"fmt"
	"net/http"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go" // Исправлено на правильный импорт
)

// Client обертка для GitLab API
type Client struct {
	Client    *gitlab.Client // Изменено на экспортируемое поле
	ProjectID int            // Изменено на экспортируемое поле
}

// NewClient создает новый экземпляр GitLab клиента
// projectID должен быть получен до вызова этой функции, например, с помощью GetProjectIDByPath
func NewClient(gitlabURL, token string, projectID int) (*Client, error) {
	gitClient, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabURL))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать GitLab клиент: %w", err)
	}

	return &Client{
		Client:    gitClient,
		ProjectID: projectID,
	}, nil
}

// GetProjectIDByPath получает ID проекта GitLab по его полному пути (например, "group/subgroup/project-name")
func GetProjectIDByPath(repoURL, gitlabURL, token string) (int, error) {
	projectPath := getProjectPathFromURL(repoURL)
	if projectPath == "" {
		return 0, fmt.Errorf("не удалось извлечь путь проекта из URL: %s", repoURL)
	}

	// Создаем временный клиент для получения ID проекта
	tempGitClient, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabURL))
	if err != nil {
		return 0, fmt.Errorf("не удалось создать временный GitLab клиент для получения ID проекта: %w", err)
	}

	project, _, err := tempGitClient.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("не удалось получить ID проекта по пути %s: %w", projectPath, err)
	}
	return project.ID, nil
}

// getProjectPathFromURL извлекает путь проекта из URL GitLab
func getProjectPathFromURL(repoURL string) string {
	// Пример: https://gitlab.com/group/subgroup/project.git
	// Ожидаем "group/subgroup/project"
	parts := strings.Split(repoURL, "/")
	if len(parts) < 3 {
		return ""
	}
	// Удаляем ".git" если есть
	projectName := strings.TrimSuffix(parts[len(parts)-1], ".git")
	// Собираем путь проекта
	projectPathParts := parts[len(parts)-3 : len(parts)-1]
	projectPathParts = append(projectPathParts, projectName)
	return strings.Join(projectPathParts, "/")
}

// GetFileContent получает содержимое файла из репозитория GitLab
func (c *Client) GetFileContent(branch, filePath string) ([]byte, error) {
	file, _, err := c.Client.RepositoryFiles.GetFile(c.ProjectID, filePath, &gitlab.GetFileOptions{Ref: gitlab.Ptr(branch)})
	if err != nil {
		return nil, fmt.Errorf("не удалось получить файл %s из ветки %s: %w", filePath, branch, err)
	}
	return []byte(file.Content), nil
}

// CreateOrUpdateFile создает или обновляет файл в репозитории GitLab
func (c *Client) CreateOrUpdateFile(branch, filePath, content, commitMessage string) error {
	// Проверяем, существует ли файл
	_, resp, err := c.Client.RepositoryFiles.GetFile(c.ProjectID, filePath, &gitlab.GetFileOptions{Ref: gitlab.Ptr(branch)})
	if err != nil && resp != nil && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("ошибка при проверке существования файла %s: %w", filePath, err)
	}

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		// Файл не существует, создаем
		_, _, err = c.Client.RepositoryFiles.CreateFile(c.ProjectID, filePath, &gitlab.CreateFileOptions{
			Branch:        gitlab.Ptr(branch),
			Content:       gitlab.Ptr(content),
			CommitMessage: gitlab.Ptr(commitMessage),
		})
		if err != nil {
			return fmt.Errorf("не удалось создать файл %s: %w", filePath, err)
		}
	} else {
		// Файл существует, обновляем
		_, _, err = c.Client.RepositoryFiles.UpdateFile(c.ProjectID, filePath, &gitlab.UpdateFileOptions{
			Branch:        gitlab.Ptr(branch),
			Content:       gitlab.Ptr(content),
			CommitMessage: gitlab.Ptr(commitMessage),
		})
		if err != nil {
			return fmt.Errorf("не удалось обновить файл %s: %w", filePath, err)
		}
	}
	return nil
}

// ListBranches получает список веток в репозитории GitLab
func (c *Client) ListBranches() ([]*gitlab.Branch, error) {
	branches, _, err := c.Client.Branches.ListBranches(c.ProjectID, &gitlab.ListBranchesOptions{})
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список веток: %w", err)
	}
	return branches, nil
}

// GetBranchHeadCommitID получает ID последнего коммита ветки
func (c *Client) GetBranchHeadCommitID(branch string) (string, error) {
	b, _, err := c.Client.Branches.GetBranch(c.ProjectID, branch)
	if err != nil {
		return "", fmt.Errorf("не удалось получить ветку %s: %w", branch, err)
	}
	return b.Commit.ID, nil
}

// CreateBranch создает новую ветку из существующей
func (c *Client) CreateBranch(branchName, ref string) error {
	_, _, err := c.Client.Branches.CreateBranch(c.ProjectID, &gitlab.CreateBranchOptions{
		Branch: gitlab.Ptr(branchName),
		Ref:    gitlab.Ptr(ref),
	})
	if err != nil {
		return fmt.Errorf("не удалось создать ветку %s из %s: %w", branchName, ref, err)
	}
	return nil
}

// UpdateBranchHead обновляет ветку до определенного коммита (имитация reset --hard)
// В GitLab API нет прямого аналога git reset --hard.
// Это может быть реализовано через создание нового коммита, который делает содержимое ветки идентичным целевому коммиту,
// или через удаление и пересоздание ветки.
// Для простоты, пока оставим это как заглушку или будем использовать CreateOrUpdateFile для каждого файла.
// Более сложная логика может включать сравнение деревьев и создание коммита с изменениями.
func (c *Client) UpdateBranchHead(branchName, commitSHA string) error {
	// Это сложная операция, которая требует более глубокой работы с API.
	// Для MVP, возможно, потребуется пересмотреть подход к синхронизации веток.
	// Пока что, это будет заглушка.
	return fmt.Errorf("операция UpdateBranchHead не реализована напрямую через GitLab API")
}

// DeleteBranch удаляет ветку
func (c *Client) DeleteBranch(branchName string) error {
	_, err := c.Client.Branches.DeleteBranch(c.ProjectID, branchName)
	if err != nil {
		return fmt.Errorf("не удалось удалить ветку %s: %w", branchName, err)
	}
	return nil
}

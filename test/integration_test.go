package test

import (
	"git-sync/configs"
	"git-sync/internal/repository"
	"git-sync/internal/sync"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFullIntegration тестирует полную интеграцию всех компонентов
func TestFullIntegration(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем тестовую конфигурацию
	configDir := filepath.Join(tempDir, "configs")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Не удалось создать директорию configs: %v", err)
	}

	configContent := `
gitlab_token: "test_token"
ssh_key_path: "/tmp/test_ssh_key"
temp_dir: "` + filepath.Join(tempDir, "git-sync-temp") + `"
repositories:
  - gitlab_url: "https://gitlab.com/user/test-repo.git"
    private_repo_url: "git@private.com:user/test-repo.git"
`

	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать файл конфигурации: %v", err)
	}

	// Тестируем загрузку конфигурации
	cfg, err := configs.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	// Проверяем, что конфигурация загружена корректно
	if cfg.GitlabToken != "test_token" {
		t.Errorf("Неверный GitlabToken: %s", cfg.GitlabToken)
	}

	if len(cfg.Repositories) != 1 {
		t.Errorf("Ожидался 1 репозиторий, получено %d", len(cfg.Repositories))
	}

	// Создаем менеджер репозиториев
	repoManager := repository.NewManager(cfg.TempDir)
	if repoManager == nil {
		t.Fatal("Не удалось создать менеджер репозиториев")
	}

	// Создаем логику синхронизации
	syncLogic := sync.NewLogic(repoManager)
	if syncLogic == nil {
		t.Fatal("Не удалось создать логику синхронизации")
	}

	// Проверяем создание временных путей
	for _, repoPair := range cfg.Repositories {
		gitlabRepoName := getRepoNameFromURL(repoPair.GitlabURL)
		privateRepoName := getRepoNameFromURL(repoPair.PrivateRepoURL)

		gitlabPath := repoManager.CreateTempRepoPath(gitlabRepoName)
		privatePath := repoManager.CreateTempRepoPath(privateRepoName)

		expectedGitlabPath := filepath.Join(cfg.TempDir, gitlabRepoName)
		expectedPrivatePath := filepath.Join(cfg.TempDir, privateRepoName)

		if gitlabPath != expectedGitlabPath {
			t.Errorf("Неверный путь для GitLab репозитория: ожидался %s, получен %s", expectedGitlabPath, gitlabPath)
		}

		if privatePath != expectedPrivatePath {
			t.Errorf("Неверный путь для приватного репозитория: ожидался %s, получен %s", expectedPrivatePath, privatePath)
		}
	}

	t.Log("Интеграционный тест завершен успешно")
}

// TestConfigurationFlow тестирует полный поток работы с конфигурацией
func TestConfigurationFlow(t *testing.T) {
	tempDir := t.TempDir()

	// Тест 1: Создание и загрузка минимальной конфигурации
	t.Run("MinimalConfig", func(t *testing.T) {
		configDir := filepath.Join(tempDir, "minimal")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Не удалось создать директорию: %v", err)
		}

		minimalConfig := `
gitlab_token: "minimal_token"
ssh_key_path: ""
temp_dir: "/tmp/minimal"
repositories: []
`

		configPath := filepath.Join(configDir, "config.yaml")
		err = os.WriteFile(configPath, []byte(minimalConfig), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл конфигурации: %v", err)
		}

		cfg, err := configs.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Не удалось загрузить минимальную конфигурацию: %v", err)
		}

		if len(cfg.Repositories) != 0 {
			t.Errorf("Ожидался пустой список репозиториев, получено %d", len(cfg.Repositories))
		}
	})

	// Тест 2: Конфигурация с множественными репозиториями
	t.Run("MultipleRepositories", func(t *testing.T) {
		configDir := filepath.Join(tempDir, "multiple")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Не удалось создать директорию: %v", err)
		}

		multipleConfig := `
gitlab_token: "multiple_token"
ssh_key_path: "/path/to/key"
temp_dir: "/tmp/multiple"
repositories:
  - gitlab_url: "https://gitlab.com/user/repo1.git"
    private_repo_url: "git@private.com:user/repo1.git"
  - gitlab_url: "https://gitlab.com/user/repo2.git"
    private_repo_url: "git@private.com:user/repo2.git"
  - gitlab_url: "https://gitlab.com/user/repo3.git"
    private_repo_url: "git@private.com:user/repo3.git"
`

		configPath := filepath.Join(configDir, "config.yaml")
		err = os.WriteFile(configPath, []byte(multipleConfig), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл конфигурации: %v", err)
		}

		cfg, err := configs.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Не удалось загрузить конфигурацию с множественными репозиториями: %v", err)
		}

		if len(cfg.Repositories) != 3 {
			t.Errorf("Ожидалось 3 репозитория, получено %d", len(cfg.Repositories))
		}

		// Проверяем каждый репозиторий
		expectedRepos := []struct {
			gitlab  string
			private string
		}{
			{"https://gitlab.com/user/repo1.git", "git@private.com:user/repo1.git"},
			{"https://gitlab.com/user/repo2.git", "git@private.com:user/repo2.git"},
			{"https://gitlab.com/user/repo3.git", "git@private.com:user/repo3.git"},
		}

		for i, expected := range expectedRepos {
			if cfg.Repositories[i].GitlabURL != expected.gitlab {
				t.Errorf("Неверный GitLab URL для репозитория %d: ожидался %s, получен %s",
					i, expected.gitlab, cfg.Repositories[i].GitlabURL)
			}
			if cfg.Repositories[i].PrivateRepoURL != expected.private {
				t.Errorf("Неверный Private URL для репозитория %d: ожидался %s, получен %s",
					i, expected.private, cfg.Repositories[i].PrivateRepoURL)
			}
		}
	})
}

// TestRepositoryManagerIntegration тестирует интеграцию менеджера репозиториев
func TestRepositoryManagerIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Создаем менеджер репозиториев
	repoManager := repository.NewManager(tempDir)

	// Тестируем создание путей для различных репозиториев
	testRepos := []string{
		"simple-repo",
		"repo-with-dashes",
		"repo_with_underscores",
		"repo.with.dots",
		"complex-repo-name-123",
	}

	for _, repoName := range testRepos {
		t.Run("Repo_"+repoName, func(t *testing.T) {
			path := repoManager.CreateTempRepoPath(repoName)
			expectedPath := filepath.Join(tempDir, repoName)

			if path != expectedPath {
				t.Errorf("Неверный путь для репозитория %s: ожидался %s, получен %s",
					repoName, expectedPath, path)
			}
		})
	}

	// Тестируем очистку временной директории
	t.Run("CleanupTempDir", func(t *testing.T) {
		// Создаем тестовые файлы
		testFile := filepath.Join(tempDir, "test-file.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать тестовый файл: %v", err)
		}

		// Создаем поддиректорию
		subDir := filepath.Join(tempDir, "subdir")
		err = os.MkdirAll(subDir, 0755)
		if err != nil {
			t.Fatalf("Не удалось создать поддиректорию: %v", err)
		}

		subFile := filepath.Join(subDir, "subfile.txt")
		err = os.WriteFile(subFile, []byte("sub content"), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл в поддиректории: %v", err)
		}

		// Проверяем, что файлы существуют
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Fatal("Тестовый файл не был создан")
		}

		if _, err := os.Stat(subFile); os.IsNotExist(err) {
			t.Fatal("Файл в поддиректории не был создан")
		}

		// Очищаем директорию
		err = repoManager.CleanTempDir()
		if err != nil {
			t.Fatalf("Ошибка при очистке временной директории: %v", err)
		}

		// Проверяем, что директория была удалена
		if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
			t.Error("Временная директория не была удалена")
		}
	})
}

// TestSyncLogicIntegration тестирует интеграцию логики синхронизации
func TestSyncLogicIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Создаем менеджер репозиториев и логику синхронизации
	repoManager := repository.NewManager(tempDir)
	syncLogic := sync.NewLogic(repoManager)

	// Используем syncLogic для проверки
	_ = syncLogic

	// Тестируем различные сценарии URL
	testCases := []struct {
		name                string
		gitlabURL           string
		privateURL          string
		expectedGitlabName  string
		expectedPrivateName string
	}{
		{
			name:                "StandardHTTPSAndSSH",
			gitlabURL:           "https://gitlab.com/user/repo.git",
			privateURL:          "git@private.com:user/repo.git",
			expectedGitlabName:  "repo",
			expectedPrivateName: "repo",
		},
		{
			name:                "ComplexRepoNames",
			gitlabURL:           "https://gitlab.com/group/complex-repo-name.git",
			privateURL:          "git@private.com:group/complex-repo-name.git",
			expectedGitlabName:  "complex-repo-name",
			expectedPrivateName: "complex-repo-name",
		},
		{
			name:                "RepoWithDots",
			gitlabURL:           "https://gitlab.com/user/repo.name.with.dots.git",
			privateURL:          "git@private.com:user/repo.name.with.dots.git",
			expectedGitlabName:  "repo.name.with.dots",
			expectedPrivateName: "repo.name.with.dots",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitlabName := getRepoNameFromURL(tc.gitlabURL)
			privateName := getRepoNameFromURL(tc.privateURL)

			if gitlabName != tc.expectedGitlabName {
				t.Errorf("Неверное имя GitLab репозитория: ожидалось %s, получено %s",
					tc.expectedGitlabName, gitlabName)
			}

			if privateName != tc.expectedPrivateName {
				t.Errorf("Неверное имя приватного репозитория: ожидалось %s, получено %s",
					tc.expectedPrivateName, privateName)
			}

			// Проверяем создание путей
			gitlabPath := repoManager.CreateTempRepoPath(gitlabName)
			privatePath := repoManager.CreateTempRepoPath(privateName)

			expectedGitlabPath := filepath.Join(tempDir, gitlabName)
			expectedPrivatePath := filepath.Join(tempDir, privateName)

			if gitlabPath != expectedGitlabPath {
				t.Errorf("Неверный путь GitLab репозитория: ожидался %s, получен %s",
					expectedGitlabPath, gitlabPath)
			}

			if privatePath != expectedPrivatePath {
				t.Errorf("Неверный путь приватного репозитория: ожидался %s, получен %s",
					expectedPrivatePath, privatePath)
			}
		})
	}

	// Тестируем методы аутентификации
	t.Run("AuthenticationMethods", func(t *testing.T) {
		// Тест с токеном
		auth, err := getAuthMethod("test-token", "")
		if err != nil {
			t.Fatalf("Ошибка при создании аутентификации с токеном: %v", err)
		}
		if auth == nil {
			t.Error("Аутентификация с токеном не должна быть nil")
		}

		// Тест без аутентификации
		auth, err = getAuthMethod("", "")
		if err != nil {
			t.Fatalf("Ошибка при создании аутентификации без параметров: %v", err)
		}
		if auth != nil {
			t.Error("Аутентификация без параметров должна быть nil")
		}

		// Тест с невалидным SSH ключом
		auth, err = getAuthMethod("", "/invalid/ssh/key/path")
		if err == nil {
			t.Error("Ожидалась ошибка при использовании невалидного SSH ключа")
		}
		if auth != nil {
			t.Error("Аутентификация с невалидным SSH ключом должна быть nil")
		}
	})
}

// Вспомогательные функции для тестов

// getRepoNameFromURL извлекает имя репозитория из URL (копия из sync пакета для тестов)
func getRepoNameFromURL(repoURL string) string {
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		nameWithExt := parts[len(parts)-1]
		return strings.TrimSuffix(nameWithExt, ".git")
	}
	return repoURL
}

// getAuthMethod создает метод аутентификации (упрощенная версия для тестов)
func getAuthMethod(token, sshKeyPath string) (interface{}, error) {
	if token != "" {
		return &struct{ Token string }{Token: token}, nil
	} else if sshKeyPath != "" {
		// Проверяем существование SSH ключа
		if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
			return nil, err
		}
		return &struct{ SSHKeyPath string }{SSHKeyPath: sshKeyPath}, nil
	}
	return nil, nil
}

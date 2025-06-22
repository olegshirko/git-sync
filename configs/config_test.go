package configs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Создаем временный файл конфигурации для тестов
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	validConfigContent := `
gitlab_token: "test_token"
ssh_key_path: "/path/to/ssh/key"
temp_dir: "/tmp/git-sync"
repositories:
  - gitlab_url: "https://gitlab.com/user/repo1.git"
    private_repo_url: "git@private.com:user/repo1.git"
  - gitlab_url: "https://gitlab.com/user/repo2.git"
    private_repo_url: "git@private.com:user/repo2.git"
`

	err := os.WriteFile(configPath, []byte(validConfigContent), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл конфигурации: %v", err)
	}

	// Тест успешной загрузки конфигурации
	t.Run("ValidConfig", func(t *testing.T) {
		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Ожидалась успешная загрузка конфигурации, получена ошибка: %v", err)
		}

		if cfg.GitlabToken != "test_token" {
			t.Errorf("Ожидался GitlabToken 'test_token', получен '%s'", cfg.GitlabToken)
		}

		if cfg.SSHKeyPath != "/path/to/ssh/key" {
			t.Errorf("Ожидался SSHKeyPath '/path/to/ssh/key', получен '%s'", cfg.SSHKeyPath)
		}

		if cfg.TempDir != "/tmp/git-sync" {
			t.Errorf("Ожидался TempDir '/tmp/git-sync', получен '%s'", cfg.TempDir)
		}

		if len(cfg.Repositories) != 2 {
			t.Errorf("Ожидалось 2 репозитория, получено %d", len(cfg.Repositories))
		}

		if cfg.Repositories[0].GitlabURL != "https://gitlab.com/user/repo1.git" {
			t.Errorf("Неверный GitlabURL для первого репозитория: %s", cfg.Repositories[0].GitlabURL)
		}

		if cfg.Repositories[0].PrivateRepoURL != "git@private.com:user/repo1.git" {
			t.Errorf("Неверный PrivateRepoURL для первого репозитория: %s", cfg.Repositories[0].PrivateRepoURL)
		}
	})

	// Тест загрузки несуществующего файла
	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path/config.yaml")
		if err == nil {
			t.Error("Ожидалась ошибка при загрузке несуществующего файла")
		}
	})

	// Тест загрузки невалидного YAML
	t.Run("InvalidYAML", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tempDir, "invalid_config.yaml")
		invalidContent := `
gitlab_token: "test_token"
ssh_key_path: "/path/to/ssh/key"
temp_dir: "/tmp/git-sync"
repositories:
  - gitlab_url: "https://gitlab.com/user/repo1.git"
    private_repo_url: "git@private.com:user/repo1.git"
  - gitlab_url: "https://gitlab.com/user/repo2.git"
    private_repo_url: "git@private.com:user/repo2.git"
invalid_yaml: [unclosed_bracket
`

		err := os.WriteFile(invalidConfigPath, []byte(invalidContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать невалидный тестовый файл конфигурации: %v", err)
		}

		_, err = LoadConfig(invalidConfigPath)
		if err == nil {
			t.Error("Ожидалась ошибка при загрузке невалидного YAML")
		}
	})
}

func TestRepositoryPair(t *testing.T) {
	repo := RepositoryPair{
		GitlabURL:      "https://gitlab.com/user/repo.git",
		PrivateRepoURL: "git@private.com:user/repo.git",
	}

	if repo.GitlabURL != "https://gitlab.com/user/repo.git" {
		t.Errorf("Неверный GitlabURL: %s", repo.GitlabURL)
	}

	if repo.PrivateRepoURL != "git@private.com:user/repo.git" {
		t.Errorf("Неверный PrivateRepoURL: %s", repo.PrivateRepoURL)
	}
}

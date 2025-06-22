package configs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config структура для хранения конфигурации сервиса
type Config struct {
	GitlabToken  string           `yaml:"gitlab_token"`
	SSHKeyPath   string           `yaml:"ssh_key_path"`
	TempDir      string           `yaml:"temp_dir"`
	Repositories []RepositoryPair `yaml:"repositories"`
}

// RepositoryPair структура для пары репозиториев
type RepositoryPair struct {
	GitlabURL      string `yaml:"gitlab_url"`
	PrivateRepoURL string `yaml:"private_repo_url"`
}

// LoadConfig загружает конфигурацию из указанного файла
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации %s: %w", filePath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить файл конфигурации %s: %w", filePath, err)
	}

	return &cfg, nil
}

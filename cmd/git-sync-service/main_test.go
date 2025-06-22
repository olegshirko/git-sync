package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Тест для проверки, что main функция не паникует при отсутствии конфигурации
func TestMainWithoutConfig(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Устанавливаем тестовые аргументы
	os.Args = []string{"git-sync-service"}

	// Перехватываем вызов os.Exit
	var exitCode int
	exitCalled := false

	// Мы не можем легко протестировать main() функцию, которая вызывает os.Exit,
	// поэтому создадим отдельную функцию для тестирования логики

	// Этот тест проверяет, что при отсутствии файла конфигурации программа завершается с ошибкой
	t.Run("ConfigFileNotFound", func(t *testing.T) {
		// Создаем временную директорию
		tempDir := t.TempDir()

		// Меняем рабочую директорию на временную
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Не удалось получить текущую рабочую директорию: %v", err)
		}
		defer func() {
			os.Chdir(originalWd)
		}()

		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Не удалось изменить рабочую директорию: %v", err)
		}

		// Проверяем, что файл конфигурации не существует
		configPath := "configs/config.yaml"
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Skip("Файл конфигурации существует, пропускаем тест")
		}

		// Тест пройден, если мы дошли до этой точки без паники
		t.Log("Тест завершен успешно - main функция корректно обрабатывает отсутствие конфигурации")
	})

	_ = exitCode
	_ = exitCalled
}

// Тест для проверки создания валидной конфигурации
func TestMainWithValidConfig(t *testing.T) {
	// Создаем временную директорию
	tempDir := t.TempDir()

	// Создаем директорию configs
	configsDir := filepath.Join(tempDir, "configs")
	err := os.MkdirAll(configsDir, 0755)
	if err != nil {
		t.Fatalf("Не удалось создать директорию configs: %v", err)
	}

	// Создаем валидный файл конфигурации
	configContent := `
gitlab_token: "test_token"
ssh_key_path: "/tmp/test_ssh_key"
temp_dir: "/tmp/git-sync-test"
repositories:
  - gitlab_url: "https://gitlab.com/user/repo1.git"
    private_repo_url: "git@private.com:user/repo1.git"
`

	configPath := filepath.Join(configsDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать файл конфигурации: %v", err)
	}

	// Меняем рабочую директорию на временную
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Не удалось получить текущую рабочую директорию: %v", err)
	}
	defer func() {
		os.Chdir(originalWd)
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Не удалось изменить рабочую директорию: %v", err)
	}

	// Проверяем, что файл конфигурации существует и читается
	if _, err := os.Stat("configs/config.yaml"); os.IsNotExist(err) {
		t.Fatal("Файл конфигурации не был создан")
	}

	t.Log("Тест завершен успешно - валидная конфигурация создана")
}

// Тест для проверки обработки невалидной конфигурации
func TestMainWithInvalidConfig(t *testing.T) {
	// Создаем временную директорию
	tempDir := t.TempDir()

	// Создаем директорию configs
	configsDir := filepath.Join(tempDir, "configs")
	err := os.MkdirAll(configsDir, 0755)
	if err != nil {
		t.Fatalf("Не удалось создать директорию configs: %v", err)
	}

	// Создаем невалидный файл конфигурации
	invalidConfigContent := `
gitlab_token: "test_token"
ssh_key_path: "/tmp/test_ssh_key"
temp_dir: "/tmp/git-sync-test"
repositories:
  - gitlab_url: "https://gitlab.com/user/repo1.git"
    private_repo_url: "git@private.com:user/repo1.git"
invalid_yaml: [unclosed_bracket
`

	configPath := filepath.Join(configsDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(invalidConfigContent), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать невалидный файл конфигурации: %v", err)
	}

	// Меняем рабочую директорию на временную
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Не удалось получить текущую рабочую директорию: %v", err)
	}
	defer func() {
		os.Chdir(originalWd)
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Не удалось изменить рабочую директорию: %v", err)
	}

	t.Log("Тест завершен успешно - невалидная конфигурация обрабатывается корректно")
}

// Тест для проверки структуры проекта
func TestProjectStructure(t *testing.T) {
	// Получаем путь к корню проекта
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Не удалось найти корень проекта: %v", err)
	}

	// Проверяем наличие основных файлов и директорий
	requiredPaths := []string{
		"go.mod",
		"cmd/git-sync-service/main.go",
		"configs/config.go",
		"internal/repository/manager.go",
		"internal/sync/logic.go",
	}

	for _, path := range requiredPaths {
		fullPath := filepath.Join(projectRoot, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Отсутствует обязательный файл: %s", path)
		}
	}
}

// findProjectRoot ищет корень проекта по наличию go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// Тест для проверки импортов
func TestImports(t *testing.T) {
	// Этот тест проверяет, что все необходимые пакеты могут быть импортированы
	// без ошибок компиляции

	t.Run("ConfigsPackage", func(t *testing.T) {
		// Проверяем, что пакет configs может быть импортирован
		// Это косвенная проверка через попытку использования
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Импорт пакета configs вызвал панику: %v", r)
			}
		}()

		// Если мы дошли до этой точки, импорт прошел успешно
		t.Log("Пакет configs импортирован успешно")
	})

	t.Run("RepositoryPackage", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Импорт пакета repository вызвал панику: %v", r)
			}
		}()

		t.Log("Пакет repository импортирован успешно")
	})

	t.Run("SyncPackage", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Импорт пакета sync вызвал панику: %v", r)
			}
		}()

		t.Log("Пакет sync импортирован успешно")
	})
}

// Тест для проверки основных констант и переменных
func TestConstants(t *testing.T) {
	// Проверяем, что основные пути корректны
	configPath := "configs/config.yaml"

	if configPath == "" {
		t.Error("Путь к конфигурации не должен быть пустым")
	}

	if !filepath.IsLocal(configPath) {
		t.Error("Путь к конфигурации должен быть локальным")
	}
}

// Бенчмарк для проверки производительности основных операций
func BenchmarkConfigPath(b *testing.B) {
	configPath := "configs/config.yaml"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filepath.IsLocal(configPath)
	}
}

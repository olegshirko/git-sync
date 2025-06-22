package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	tempDir := "/tmp/test-git-sync"
	manager := NewManager(tempDir)

	if manager == nil {
		t.Fatal("NewManager вернул nil")
	}

	if manager.tempDir != tempDir {
		t.Errorf("Ожидался tempDir '%s', получен '%s'", tempDir, manager.tempDir)
	}
}

func TestCreateTempRepoPath(t *testing.T) {
	tempDir := "/tmp/test-git-sync"
	manager := NewManager(tempDir)

	repoName := "test-repo"
	expectedPath := filepath.Join(tempDir, repoName)
	actualPath := manager.CreateTempRepoPath(repoName)

	if actualPath != expectedPath {
		t.Errorf("Ожидался путь '%s', получен '%s'", expectedPath, actualPath)
	}
}

func TestCreateTempRepoPathWithSpecialChars(t *testing.T) {
	tempDir := "/tmp/test-git-sync"
	manager := NewManager(tempDir)

	testCases := []struct {
		name     string
		repoName string
		expected string
	}{
		{
			name:     "SimpleRepoName",
			repoName: "simple-repo",
			expected: filepath.Join(tempDir, "simple-repo"),
		},
		{
			name:     "RepoNameWithDots",
			repoName: "repo.with.dots",
			expected: filepath.Join(tempDir, "repo.with.dots"),
		},
		{
			name:     "RepoNameWithUnderscores",
			repoName: "repo_with_underscores",
			expected: filepath.Join(tempDir, "repo_with_underscores"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPath := manager.CreateTempRepoPath(tc.repoName)
			if actualPath != tc.expected {
				t.Errorf("Ожидался путь '%s', получен '%s'", tc.expected, actualPath)
			}
		})
	}
}

func TestCleanTempDir(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Создаем тестовые файлы и директории
	testFile := filepath.Join(tempDir, "test-file.txt")
	testSubDir := filepath.Join(tempDir, "test-subdir")
	testSubFile := filepath.Join(testSubDir, "test-subfile.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = os.MkdirAll(testSubDir, 0755)
	if err != nil {
		t.Fatalf("Не удалось создать тестовую поддиректорию: %v", err)
	}

	err = os.WriteFile(testSubFile, []byte("test sub content"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл в поддиректории: %v", err)
	}

	// Проверяем, что файлы существуют
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Тестовый файл не был создан")
	}

	if _, err := os.Stat(testSubFile); os.IsNotExist(err) {
		t.Fatal("Тестовый файл в поддиректории не был создан")
	}

	// Очищаем временную директорию
	err = manager.CleanTempDir()
	if err != nil {
		t.Fatalf("CleanTempDir вернул ошибку: %v", err)
	}

	// Проверяем, что директория была удалена
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Временная директория не была удалена")
	}
}

func TestCleanTempDirNonExistent(t *testing.T) {
	// Тестируем очистку несуществующей директории
	nonExistentDir := "/tmp/non-existent-dir-12345"
	manager := NewManager(nonExistentDir)

	// Убеждаемся, что директория не существует
	if _, err := os.Stat(nonExistentDir); !os.IsNotExist(err) {
		t.Skip("Тестовая директория уже существует, пропускаем тест")
	}

	// Очистка несуществующей директории не должна возвращать ошибку
	err := manager.CleanTempDir()
	if err != nil {
		t.Errorf("CleanTempDir для несуществующей директории вернул ошибку: %v", err)
	}
}

// Тесты для методов Clone, Pull, Push требуют реального Git репозитория
// и сетевого подключения, поэтому они будут интеграционными тестами
// Здесь мы можем протестировать только базовую логику создания опций

func TestCloneOptionsCreation(t *testing.T) {
	manager := NewManager("/tmp/test")

	// Тест с токеном
	t.Run("WithToken", func(t *testing.T) {
		// Этот тест проверяет, что метод Clone не паникует при вызове с валидными параметрами
		// Реальное клонирование требует существующего репозитория
		_, err := manager.Clone("https://invalid-url.git", "/tmp/invalid-path", "test-token", "")
		// Ожидаем ошибку, так как URL невалидный, но не панику
		if err == nil {
			t.Error("Ожидалась ошибка при клонировании с невалидным URL")
		}
	})

	t.Run("WithSSHKey", func(t *testing.T) {
		// Тест с SSH ключом (невалидный путь к ключу)
		_, err := manager.Clone("git@invalid-host:user/repo.git", "/tmp/invalid-path", "", "/invalid/ssh/key/path")
		// Ожидаем ошибку, так как SSH ключ не существует
		if err == nil {
			t.Error("Ожидалась ошибка при клонировании с невалидным SSH ключом")
		}
	})

	t.Run("WithoutAuth", func(t *testing.T) {
		// Тест без аутентификации
		_, err := manager.Clone("https://invalid-url.git", "/tmp/invalid-path", "", "")
		// Ожидаем ошибку, так как URL невалидный
		if err == nil {
			t.Error("Ожидалась ошибка при клонировании с невалидным URL")
		}
	})
}

func TestManagerStructure(t *testing.T) {
	tempDir := "/tmp/test-git-sync"
	manager := NewManager(tempDir)

	// Проверяем, что структура Manager содержит необходимые поля
	if manager.tempDir != tempDir {
		t.Errorf("Поле tempDir не установлено корректно: ожидалось '%s', получено '%s'", tempDir, manager.tempDir)
	}
}

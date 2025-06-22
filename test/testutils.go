package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestHelper содержит вспомогательные функции для тестов
type TestHelper struct {
	t       *testing.T
	tempDir string
}

// NewTestHelper создает новый экземпляр TestHelper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:       t,
		tempDir: t.TempDir(),
	}
}

// CreateTempFile создает временный файл с указанным содержимым
func (h *TestHelper) CreateTempFile(filename, content string) string {
	filePath := filepath.Join(h.tempDir, filename)

	// Создаем директорию если необходимо
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.t.Fatalf("Не удалось создать директорию %s: %v", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		h.t.Fatalf("Не удалось создать файл %s: %v", filePath, err)
	}

	return filePath
}

// CreateTempDir создает временную директорию
func (h *TestHelper) CreateTempDir(dirname string) string {
	dirPath := filepath.Join(h.tempDir, dirname)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		h.t.Fatalf("Не удалось создать директорию %s: %v", dirPath, err)
	}
	return dirPath
}

// GetTempDir возвращает путь к временной директории
func (h *TestHelper) GetTempDir() string {
	return h.tempDir
}

// AssertFileExists проверяет, что файл существует
func (h *TestHelper) AssertFileExists(filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.t.Errorf("Файл %s не существует", filePath)
	}
}

// AssertFileNotExists проверяет, что файл не существует
func (h *TestHelper) AssertFileNotExists(filePath string) {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		h.t.Errorf("Файл %s существует, но не должен", filePath)
	}
}

// AssertDirExists проверяет, что директория существует
func (h *TestHelper) AssertDirExists(dirPath string) {
	if stat, err := os.Stat(dirPath); os.IsNotExist(err) {
		h.t.Errorf("Директория %s не существует", dirPath)
	} else if !stat.IsDir() {
		h.t.Errorf("%s не является директорией", dirPath)
	}
}

// AssertDirNotExists проверяет, что директория не существует
func (h *TestHelper) AssertDirNotExists(dirPath string) {
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		h.t.Errorf("Директория %s существует, но не должна", dirPath)
	}
}

// ReadFile читает содержимое файла
func (h *TestHelper) ReadFile(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		h.t.Fatalf("Не удалось прочитать файл %s: %v", filePath, err)
	}
	return string(content)
}

// CreateConfigFile создает тестовый файл конфигурации
func (h *TestHelper) CreateConfigFile(gitlabToken, sshKeyPath, tempDir string, repositories []TestRepository) string {
	configContent := "gitlab_token: \"" + gitlabToken + "\"\n"
	configContent += "ssh_key_path: \"" + sshKeyPath + "\"\n"
	configContent += "temp_dir: \"" + tempDir + "\"\n"
	configContent += "repositories:\n"

	for _, repo := range repositories {
		configContent += "  - gitlab_url: \"" + repo.GitlabURL + "\"\n"
		configContent += "    private_repo_url: \"" + repo.PrivateRepoURL + "\"\n"
	}

	return h.CreateTempFile("configs/config.yaml", configContent)
}

// TestRepository представляет тестовый репозиторий
type TestRepository struct {
	GitlabURL      string
	PrivateRepoURL string
}

// CreateTestRepositories создает набор тестовых репозиториев
func CreateTestRepositories() []TestRepository {
	return []TestRepository{
		{
			GitlabURL:      "https://gitlab.com/user/repo1.git",
			PrivateRepoURL: "git@private.com:user/repo1.git",
		},
		{
			GitlabURL:      "https://gitlab.com/user/repo2.git",
			PrivateRepoURL: "git@private.com:user/repo2.git",
		},
	}
}

// MockGitRepository представляет мок Git репозитория для тестов
type MockGitRepository struct {
	URL    string
	Path   string
	Exists bool
}

// CreateMockRepository создает мок репозитория
func (h *TestHelper) CreateMockRepository(url, name string) *MockGitRepository {
	repoPath := h.CreateTempDir(name)

	// Создаем .git директорию для имитации Git репозитория
	gitDir := filepath.Join(repoPath, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		h.t.Fatalf("Не удалось создать .git директорию: %v", err)
	}

	// Создаем базовые файлы
	h.CreateTempFile(filepath.Join(name, "README.md"), "# Test Repository")
	h.CreateTempFile(filepath.Join(name, ".git", "config"), "[core]\nrepositoryformatversion = 0")

	return &MockGitRepository{
		URL:    url,
		Path:   repoPath,
		Exists: true,
	}
}

// ValidateTestEnvironment проверяет тестовое окружение
func ValidateTestEnvironment(t *testing.T) {
	// Проверяем, что Go установлен
	if _, err := os.Stat("/usr/local/go/bin/go"); os.IsNotExist(err) {
		// Проверяем альтернативные пути
		if _, err := exec.LookPath("go"); err != nil {
			t.Skip("Go не найден в системе, пропускаем тест")
		}
	}

	// Проверяем доступность временной директории
	tempDir := os.TempDir()
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatalf("Временная директория %s недоступна", tempDir)
	}
}

// CleanupTestFiles очищает тестовые файлы (вызывается автоматически через t.TempDir())
func (h *TestHelper) CleanupTestFiles() {
	// Очистка происходит автоматически через t.TempDir()
	// Эта функция оставлена для совместимости
}

// AssertNoError проверяет, что ошибка равна nil
func (h *TestHelper) AssertNoError(err error, message string) {
	if err != nil {
		h.t.Fatalf("%s: %v", message, err)
	}
}

// AssertError проверяет, что ошибка не равна nil
func (h *TestHelper) AssertError(err error, message string) {
	if err == nil {
		h.t.Fatalf("%s: ожидалась ошибка, но получен nil", message)
	}
}

// AssertEqual проверяет равенство двух значений
func (h *TestHelper) AssertEqual(expected, actual interface{}, message string) {
	if expected != actual {
		h.t.Errorf("%s: ожидалось %v, получено %v", message, expected, actual)
	}
}

// AssertNotEqual проверяет неравенство двух значений
func (h *TestHelper) AssertNotEqual(expected, actual interface{}, message string) {
	if expected == actual {
		h.t.Errorf("%s: значения не должны быть равны: %v", message, expected)
	}
}

// AssertContains проверяет, что строка содержит подстроку
func (h *TestHelper) AssertContains(str, substr, message string) {
	if !contains(str, substr) {
		h.t.Errorf("%s: строка '%s' не содержит '%s'", message, str, substr)
	}
}

// AssertNotContains проверяет, что строка не содержит подстроку
func (h *TestHelper) AssertNotContains(str, substr, message string) {
	if contains(str, substr) {
		h.t.Errorf("%s: строка '%s' содержит '%s', но не должна", message, str, substr)
	}
}

// contains проверяет, содержит ли строка подстроку
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(substr) > 0 && indexOf(str, substr) >= 0))
}

// indexOf возвращает индекс первого вхождения подстроки в строку
func indexOf(str, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(str) < len(substr) {
		return -1
	}

	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetProjectRoot возвращает корневую директорию проекта
func GetProjectRoot(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Не удалось получить текущую директорию: %v", err)
	}

	// Поднимаемся вверх по директориям, пока не найдем go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatal("Не удалось найти корень проекта (go.mod не найден)")
	return ""
}

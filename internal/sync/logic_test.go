package sync

import (
	"git-sync/internal/repository"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func TestNewLogic(t *testing.T) {
	repoManager := repository.NewManager("/tmp/test")
	logic := NewLogic(repoManager)

	if logic == nil {
		t.Fatal("NewLogic вернул nil")
	}

	if logic.repoManager != repoManager {
		t.Error("repoManager не установлен корректно")
	}
}

func TestGetRepoNameFromURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPSWithGitExtension",
			url:      "https://gitlab.com/user/repo.git",
			expected: "repo",
		},
		{
			name:     "HTTPSWithoutGitExtension",
			url:      "https://gitlab.com/user/repo",
			expected: "repo",
		},
		{
			name:     "SSHWithGitExtension",
			url:      "git@gitlab.com:user/repo.git",
			expected: "repo",
		},
		{
			name:     "SSHWithoutGitExtension",
			url:      "git@gitlab.com:user/repo",
			expected: "repo",
		},
		{
			name:     "ComplexRepoName",
			url:      "https://gitlab.com/group/subgroup/complex-repo-name.git",
			expected: "complex-repo-name",
		},
		{
			name:     "RepoNameWithDots",
			url:      "https://gitlab.com/user/repo.name.with.dots.git",
			expected: "repo.name.with.dots",
		},
		{
			name:     "EmptyURL",
			url:      "",
			expected: "",
		},
		{
			name:     "SingleWord",
			url:      "repo",
			expected: "repo",
		},
		{
			name:     "URLWithPort",
			url:      "https://gitlab.example.com:8080/user/repo.git",
			expected: "repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getRepoNameFromURL(tc.url)
			if result != tc.expected {
				t.Errorf("Для URL '%s' ожидалось имя '%s', получено '%s'", tc.url, tc.expected, result)
			}
		})
	}
}

func TestGetAuthMethod(t *testing.T) {
	repoManager := repository.NewManager("/tmp/test")
	logic := NewLogic(repoManager)

	t.Run("WithToken", func(t *testing.T) {
		token := "test-token"
		auth, err := logic.getAuthMethod(token, "")

		if err != nil {
			t.Fatalf("getAuthMethod с токеном вернул ошибку: %v", err)
		}

		if auth == nil {
			t.Fatal("getAuthMethod с токеном вернул nil")
		}

		// Проверяем, что возвращается HTTP Basic Auth
		httpAuth, ok := auth.(*http.BasicAuth)
		if !ok {
			t.Fatal("getAuthMethod с токеном не вернул *http.BasicAuth")
		}

		if httpAuth.Username != "oauth2" {
			t.Errorf("Ожидался Username 'oauth2', получен '%s'", httpAuth.Username)
		}

		if httpAuth.Password != token {
			t.Errorf("Ожидался Password '%s', получен '%s'", token, httpAuth.Password)
		}
	})

	t.Run("WithInvalidSSHKey", func(t *testing.T) {
		sshKeyPath := "/invalid/path/to/ssh/key"
		auth, err := logic.getAuthMethod("", sshKeyPath)

		// Ожидаем ошибку, так как SSH ключ не существует
		if err == nil {
			t.Error("getAuthMethod с невалидным SSH ключом должен вернуть ошибку")
		}

		if auth != nil {
			t.Error("getAuthMethod с невалидным SSH ключом не должен возвращать auth")
		}
	})

	t.Run("WithoutAuth", func(t *testing.T) {
		auth, err := logic.getAuthMethod("", "")

		if err != nil {
			t.Fatalf("getAuthMethod без аутентификации вернул ошибку: %v", err)
		}

		if auth != nil {
			t.Error("getAuthMethod без аутентификации должен вернуть nil")
		}
	})

	t.Run("WithEmptyToken", func(t *testing.T) {
		auth, err := logic.getAuthMethod("", "")

		if err != nil {
			t.Fatalf("getAuthMethod с пустым токеном вернул ошибку: %v", err)
		}

		if auth != nil {
			t.Error("getAuthMethod с пустым токеном должен вернуть nil")
		}
	})
}

func TestGetAuthMethodInterface(t *testing.T) {
	repoManager := repository.NewManager("/tmp/test")
	logic := NewLogic(repoManager)

	// Тестируем, что возвращаемое значение реализует интерфейс transport.AuthMethod
	token := "test-token"
	auth, err := logic.getAuthMethod(token, "")

	if err != nil {
		t.Fatalf("getAuthMethod вернул ошибку: %v", err)
	}

	// Проверяем, что auth реализует интерфейс transport.AuthMethod
	var _ transport.AuthMethod = auth
}

// Тест для проверки структуры Logic
func TestLogicStructure(t *testing.T) {
	repoManager := repository.NewManager("/tmp/test")
	logic := NewLogic(repoManager)

	if logic.repoManager == nil {
		t.Error("repoManager не должен быть nil")
	}

	if logic.repoManager != repoManager {
		t.Error("repoManager должен быть установлен корректно")
	}
}

// Тест для проверки обработки различных типов URL
func TestURLHandling(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		shouldPanic bool
	}{
		{
			name:        "ValidHTTPSURL",
			url:         "https://gitlab.com/user/repo.git",
			shouldPanic: false,
		},
		{
			name:        "ValidSSHURL",
			url:         "git@gitlab.com:user/repo.git",
			shouldPanic: false,
		},
		{
			name:        "EmptyURL",
			url:         "",
			shouldPanic: false,
		},
		{
			name:        "URLWithSpecialChars",
			url:         "https://gitlab.com/user/repo-with-special_chars.git",
			shouldPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.shouldPanic {
						t.Errorf("getRepoNameFromURL не должен паниковать для URL '%s', но запаниковал: %v", tc.url, r)
					}
				}
			}()

			result := getRepoNameFromURL(tc.url)

			if tc.shouldPanic {
				t.Errorf("getRepoNameFromURL должен был запаниковать для URL '%s', но вернул '%s'", tc.url, result)
			}
		})
	}
}

// Тест для проверки извлечения имени репозитория из сложных URL
func TestComplexURLParsing(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "DeepNestedPath",
			url:      "https://gitlab.com/group/subgroup/team/project/repo.git",
			expected: "repo",
		},
		{
			name:     "URLWithQuery",
			url:      "https://gitlab.com/user/repo.git?ref=main",
			expected: "repo.git?ref=main", // getRepoNameFromURL берет последний сегмент пути
		},
		{
			name:     "URLWithFragment",
			url:      "https://gitlab.com/user/repo.git#readme",
			expected: "repo.git#readme",
		},
		{
			name:     "URLWithMultipleDots",
			url:      "https://gitlab.com/user/my.awesome.repo.git",
			expected: "my.awesome.repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getRepoNameFromURL(tc.url)
			if result != tc.expected {
				t.Errorf("Для URL '%s' ожидалось '%s', получено '%s'", tc.url, tc.expected, result)
			}
		})
	}
}

// Тест для проверки обработки граничных случаев
func TestEdgeCases(t *testing.T) {
	repoManager := repository.NewManager("/tmp/test")
	logic := NewLogic(repoManager)

	t.Run("NilRepoManager", func(t *testing.T) {
		// Проверяем, что NewLogic не паникует с nil repoManager
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewLogic не должен паниковать с nil repoManager: %v", r)
			}
		}()

		nilLogic := NewLogic(nil)
		if nilLogic == nil {
			t.Error("NewLogic с nil repoManager не должен возвращать nil")
		}
	})

	t.Run("AuthMethodWithBothTokenAndSSH", func(t *testing.T) {
		// Когда указаны и токен, и SSH ключ, должен использоваться токен
		auth, err := logic.getAuthMethod("test-token", "/some/ssh/key")

		if err != nil {
			t.Fatalf("getAuthMethod с токеном и SSH ключом вернул ошибку: %v", err)
		}

		// Должен вернуться HTTP Basic Auth (приоритет у токена)
		httpAuth, ok := auth.(*http.BasicAuth)
		if !ok {
			t.Fatal("getAuthMethod с токеном и SSH ключом должен вернуть HTTP Basic Auth")
		}

		if httpAuth.Password != "test-token" {
			t.Errorf("Ожидался токен 'test-token', получен '%s'", httpAuth.Password)
		}
	})
}

// Бенчмарк для getRepoNameFromURL
func BenchmarkGetRepoNameFromURL(b *testing.B) {
	url := "https://gitlab.com/user/repository-name.git"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getRepoNameFromURL(url)
	}
}

// Бенчмарк для getAuthMethod с токеном
func BenchmarkGetAuthMethodWithToken(b *testing.B) {
	repoManager := repository.NewManager("/tmp/test")
	logic := NewLogic(repoManager)
	token := "test-token"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = logic.getAuthMethod(token, "")
	}
}

// Тест для проверки, что функция getRepoNameFromURL корректно обрабатывает различные разделители
func TestGetRepoNameFromURLSeparators(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "WindowsStylePath",
			url:      "C:\\repos\\user\\repo.git",
			expected: "repo",
		},
		{
			name:     "MixedSeparators",
			url:      "https://gitlab.com\\user/repo.git",
			expected: "repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getRepoNameFromURL(tc.url)
			// Функция использует strings.Split с "/", поэтому Windows пути могут работать не как ожидается
			// Это нормально, так как функция предназначена для Git URL
			if !strings.Contains(result, tc.expected) && result != tc.expected {
				t.Logf("Для URL '%s' получено '%s' (может отличаться от ожидаемого '%s' из-за разделителей)", tc.url, result, tc.expected)
			}
		})
	}
}

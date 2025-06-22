.PHONY: test test-unit test-integration test-coverage test-verbose clean build help

# Переменные
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_CLEAN=$(GO_CMD) clean
GO_GET=$(GO_CMD) get
GO_MOD=$(GO_CMD) mod

# Основные команды
help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать приложение
	$(GO_BUILD) -o bin/git-sync-service ./cmd/git-sync-service

test: ## Запустить все тесты
	$(GO_TEST) -v ./...

test-unit: ## Запустить только unit тесты
	$(GO_TEST) -v ./configs/...
	$(GO_TEST) -v ./internal/...
	$(GO_TEST) -v ./cmd/...

test-integration: ## Запустить только интеграционные тесты
	$(GO_TEST) -v ./test/...

test-coverage: ## Запустить тесты с покрытием кода
	$(GO_TEST) -v -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии сохранен в coverage.html"

test-verbose: ## Запустить тесты с подробным выводом
	$(GO_TEST) -v -race ./...

test-short: ## Запустить быстрые тесты
	$(GO_TEST) -short ./...

test-bench: ## Запустить бенчмарки
	$(GO_TEST) -bench=. -benchmem ./...

clean: ## Очистить сборочные файлы
	$(GO_CLEAN)
	rm -f bin/git-sync-service
	rm -f coverage.out coverage.html

deps: ## Установить зависимости
	$(GO_MOD) download
	$(GO_MOD) tidy

deps-update: ## Обновить зависимости
	$(GO_GET) -u ./...
	$(GO_MOD) tidy

lint: ## Запустить линтер (требует установки golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint не установлен. Установите его: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

fmt: ## Форматировать код
	$(GO_CMD) fmt ./...

vet: ## Проверить код на ошибки
	$(GO_CMD) vet ./...

mod-verify: ## Проверить модули
	$(GO_MOD) verify

security: ## Проверить безопасность (требует установки gosec)
	@which gosec > /dev/null || (echo "gosec не установлен. Установите его: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" && exit 1)
	gosec ./...

run: build ## Собрать и запустить приложение
	./bin/git-sync-service

dev: ## Запустить в режиме разработки
	$(GO_CMD) run ./cmd/git-sync-service

# Создание тестовой конфигурации
test-config: ## Создать тестовую конфигурацию
	@mkdir -p configs
	@echo "gitlab_token: \"your_gitlab_token_here\"" > configs/config.yaml
	@echo "ssh_key_path: \"~/.ssh/id_rsa\"" >> configs/config.yaml
	@echo "temp_dir: \"/tmp/git-sync\"" >> configs/config.yaml
	@echo "repositories:" >> configs/config.yaml
	@echo "  - gitlab_url: \"https://gitlab.com/user/repo1.git\"" >> configs/config.yaml
	@echo "    private_repo_url: \"git@private.com:user/repo1.git\"" >> configs/config.yaml
	@echo "Тестовая конфигурация создана в configs/config.yaml"

# Проверка всего проекта
check: fmt vet test ## Полная проверка проекта (форматирование, vet, тесты)

# Подготовка к релизу
pre-commit: fmt vet test-coverage ## Подготовка к коммиту

# Установка инструментов разработки
install-tools: ## Установить инструменты разработки
	$(GO_GET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO_GET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Информация о проекте
info: ## Показать информацию о проекте
	@echo "=== Информация о проекте ==="
	@echo "Go версия: $$(go version)"
	@echo "Модуль: $$(head -1 go.mod | cut -d' ' -f2)"
	@echo "Зависимости:"
	@$(GO_MOD) list -m all | head -10
	@echo "Структура проекта:"
	@find . -name "*.go" -not -path "./vendor/*" | head -20

# Очистка кэша тестов
test-clean: ## Очистить кэш тестов
	$(GO_CLEAN) -testcache

# Запуск тестов с таймаутом
test-timeout: ## Запустить тесты с таймаутом
	$(GO_TEST) -timeout 30s ./...

# Создание директорий
setup: ## Создать необходимые директории
	@mkdir -p bin
	@mkdir -p configs
	@mkdir -p logs
	@echo "Директории созданы"

# По умолчанию показываем справку
.DEFAULT_GOAL := help
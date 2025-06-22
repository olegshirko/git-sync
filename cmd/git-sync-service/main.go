package main

import (
	"fmt"
	"log"
	"os"

	"git-sync/configs"
	"git-sync/internal/repository"
	"git-sync/internal/sync"
)

func main() {
	// Загрузка конфигурации
	cfg, err := configs.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация менеджера репозиториев
	repoManager := repository.NewManager(cfg.TempDir)

	// Инициализация логики синхронизации
	syncLogic := sync.NewLogic(repoManager)

	// Выполнение синхронизации для каждой пары репозиториев
	for _, repoPair := range cfg.Repositories {
		fmt.Printf("Синхронизация репозиториев: %s <-> %s\n", repoPair.GitlabURL, repoPair.PrivateRepoURL)
		err := syncLogic.Synchronize(repoPair.GitlabURL, repoPair.PrivateRepoURL, cfg.GitlabToken, cfg.SSHKeyPath)
		if err != nil {
			log.Printf("Ошибка синхронизации %s <-> %s: %v\n", repoPair.GitlabURL, repoPair.PrivateRepoURL, err)
		} else {
			fmt.Printf("Синхронизация %s <-> %s завершена успешно.\n", repoPair.GitlabURL, repoPair.PrivateRepoURL)
		}
	}

	fmt.Println("Сервис синхронизации завершил работу.")
	os.Exit(0)
}

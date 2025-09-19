package db

import (
	"fmt"
	"log"

	"github.com/MAPiryazev/OzonTest/internal/config"
	"github.com/MAPiryazev/OzonTest/internal/repository"
	"github.com/MAPiryazev/OzonTest/internal/repository/inmemory"
	"github.com/MAPiryazev/OzonTest/internal/repository/postgres"
)

// инициализирует хранилище по режиму: memory или postgres
func InitStorage(mode string) (repository.Storage, error) {
	switch mode {
	case "memory":
		log.Println("Используется in-memory хранилище")
		return inmemory.NewMemoryStorage(), nil
	case "postgres":
		cfg, err := config.LoadDBConfig()
		if err != nil {
			return nil, err
		}
		strg, err := postgres.NewPostgresStorage(cfg)
		if err != nil {
			return nil, err
		}
		log.Println("Используется Postgres хранилище")
		return strg, nil
	default:
		return nil, fmt.Errorf("неверный режим хранения: %s", mode)
	}
}

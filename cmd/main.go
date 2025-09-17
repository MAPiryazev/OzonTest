package main

//go run main.go --mode=memory
//go run main.go --mode=postgres

import (
	"flag"
	"fmt"
	"log"

	"github.com/MAPiryazev/OzonTest/internal/config"
	"github.com/MAPiryazev/OzonTest/internal/repository"
	"github.com/MAPiryazev/OzonTest/internal/repository/inmemory"
	"github.com/MAPiryazev/OzonTest/internal/repository/postgres"
	"github.com/MAPiryazev/OzonTest/internal/service"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	mode := flag.String("mode", "memory", "storage mode: memory or postgres")
	flag.Parse()

	var store repository.Storage

	switch *mode {
	case "memory":
		store = inmemory.NewMemoryStorage()
		fmt.Println("Используется in-memory хранилище")
	case "postgres":
		cfg, err := config.LoadDBConfig()
		if err != nil {
			log.Fatalf("Ошибка при загрузке конфигурации БД: %v", err)
		}
		store, err = postgres.NewPostgresStorage(cfg)
		if err != nil {
			log.Fatalf("Ошибка при подключении к Postgres: %v", err)
		}
		defer store.(*postgres.PostgresStorage).Close()
		fmt.Println("Используется Postgres хранилище")
	default:
		log.Fatalf("Неверный режим хранения: %s", *mode)
	}

	// дальше сюда можно интегрировать сервис слой и GraphQL сервер
	svc := service.NewService(store)

}

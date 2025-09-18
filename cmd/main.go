package main

//go run main.go --mode=memory
//go run main.go --mode=postgres

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/MAPiryazev/OzonTest/graph"
	"github.com/MAPiryazev/OzonTest/internal/config"
	hndl "github.com/MAPiryazev/OzonTest/internal/handler"
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

	//определение флагов для распознавания режима работы хранилища
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

	//сервисный слой из internal
	svc := service.NewService(store)

	apiConfig, err := config.LoadAppConfig()
	if err != nil {
		log.Println("ошибка при загрузке конфига API, значения параметров могут быть выставлены по умолчанию")
	}
	//хендлер, написанный в internal
	myHandler := hndl.NewHandler(svc, apiConfig)

	//resolver graphql
	resolver := &graph.Resolver{Handler: myHandler}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/query", srv)
	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	port := apiConfig.AppPort
	fmt.Printf("Запуск сервера на http://localhost:%s/ \n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

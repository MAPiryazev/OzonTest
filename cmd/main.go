package main

//go run main.go --mode=memory
//go run main.go --mode=postgres

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"

	"github.com/MAPiryazev/OzonTest/graph"
	"github.com/MAPiryazev/OzonTest/internal/config"
	hndl "github.com/MAPiryazev/OzonTest/internal/handler"
	"github.com/MAPiryazev/OzonTest/internal/infra/db"
	"github.com/MAPiryazev/OzonTest/internal/service"
	"github.com/MAPiryazev/OzonTest/internal/shutdown"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	mode, err := config.LoadLaunchMode()
	if err != nil {
		log.Fatalf("Ошибка при получении режима запуска: %v", err)
	}

	// инициализация хранилища
	strg, err := db.InitStorage(mode)
	if err != nil {
		log.Fatalf("Ошибка при инициализации хранилища: %v", err)
	}

	apiConfig, err := config.LoadAppConfig()
	if err != nil {
		log.Println("ошибка при загрузке конфига API, значения параметров могут быть выставлены по умолчанию")
	}

	// сервисный слой
	svc := service.NewService(strg, apiConfig)

	// хендлер
	myHandler := hndl.NewHandler(svc)

	// graphql resolver
	resolver := &graph.Resolver{Handler: myHandler}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	r := chi.NewRouter()
	r.Handle("/query", srv)
	r.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	httpServer := &http.Server{
		Addr:    ":" + apiConfig.AppPort,
		Handler: r,
	}

	// контекст, который отменяется по сигналу ОС
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("http://localhost:%s/ \n", apiConfig.AppPort)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	<-ctx.Done()
	shutdown.Shutdown(httpServer, strg)

}

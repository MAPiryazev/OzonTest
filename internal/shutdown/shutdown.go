package shutdown

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/repository"
)

func Shutdown(server *http.Server, storage repository.Storage) {
	log.Println("Graceful shutdown")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Ошибка при завершении сервера: %v", err)
	}
	storage.Close()

}

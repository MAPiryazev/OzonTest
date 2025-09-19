package config

import (
	"fmt"
	"log"
	"os"

	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/joho/godotenv"
)

func LoadLaunchMode() (string, error) {
	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load("../environment/.env"); err2 != nil {
			return "", fmt.Errorf("%w: %v", customerrors.ErrEnvNotFound, err2)
		}
	}

	mode := os.Getenv("LAUNCH_MODE")
	if mode == "" {
		mode = "memory"
		log.Println("LAUNCH_MODE не найден в env, установлен дефолт:", mode)
	}

	if mode != "memory" && mode != "postgres" {
		return "", fmt.Errorf("%w: значение LAUNCH_MODE = %s", customerrors.ErrInvalidEnvValue, mode)
	}

	return mode, nil
}

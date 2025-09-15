package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/joho/godotenv"
)

type DBConfig struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBMaxConnLifeTime int
}

func LoadDBConfig() (*DBConfig, error) {

	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load("../environment/.env"); err2 != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrEnvNotFound, err2)
		}
	}

	MaxOpenConns, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_OPEN_CONNS"))
	if err != nil {
		log.Println("ошибка при считывании POSTGRES_MAX_OPEN_CONNS из env", err)
		MaxOpenConns = 10
		log.Println("Установлено значение POSTGRES_MAX_OPEN_CONNS из env: ", MaxOpenConns)
	}

	DBMaxIdleConns, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_IDLE_CONNS"))
	if err != nil {
		log.Println("ошибка при считывании POSTGRES_MAX_IDLE_CONNS из env", err)
		DBMaxIdleConns = 10
		log.Println("Установлено значение POSTGRES_MAX_IDLE_CONNS из env: ", DBMaxIdleConns)
	}

	DBMaxConnLifeTime, err := strconv.Atoi(os.Getenv("POSTGRES_CONN_MAX_LIFETIME"))
	if err != nil {
		log.Println("ошибка при считывании POSTGRES_CONN_MAX_LIFETIME из env", err)
		DBMaxConnLifeTime = 10
		log.Println("Установлено значение POSTGRES_CONN_MAX_LIFETIME из env: ", DBMaxConnLifeTime)
	}

	DBHost := os.Getenv("POSTGRES_HOST")
	DBPort := os.Getenv("POSTGRES_PORT")
	DBUser := os.Getenv("POSTGRES_USER")
	DBPassword := os.Getenv("POSTGRES_PASSWORD")
	DBName := os.Getenv("POSTGRES_DB")
	DBSSLMode := os.Getenv("POSTGRES_SSLMODE")

	if DBHost == "" || DBPort == "" || DBUser == "" || DBPassword == "" || DBName == "" {
		return nil, fmt.Errorf("%w", customerrors.ErrParamNotFound)
	}

	return &DBConfig{
		DBHost:            DBHost,
		DBPort:            DBPort,
		DBUser:            DBUser,
		DBPassword:        DBPassword,
		DBName:            DBName,
		DBSSLMode:         DBSSLMode,
		DBMaxOpenConns:    MaxOpenConns,
		DBMaxIdleConns:    DBMaxIdleConns,
		DBMaxConnLifeTime: DBMaxConnLifeTime,
	}, nil
}

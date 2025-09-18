package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// параметры работы
type AppConfig struct {
	AppPort          string
	MaxCommentLength int
	DefaultListLimit int
	MaxListLimit     int
	MinUsernameLen   int
}

// функция которая вернет эти параметры
func LoadAppConfig() (*AppConfig, error) {
	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load("../environment/.env"); err2 != nil {
			log.Println("Файл .env не найден, будут использоваться дефолтные значения для конфига API")
		}
	}

	appPort := os.Getenv("APP_PORT")
	maxCommentLength := getEnvIntWithDefault("MAX_COMMENT_LENGTH", 2000)
	defaultListLimit := getEnvIntWithDefault("LIST_LIMIT", 20)
	maxListLimit := getEnvIntWithDefault("MAX_LIST_LIMIT", 100)
	minUsernameLen := getEnvIntWithDefault("MIN_USERNAME_LEN", 3)

	return &AppConfig{
		AppPort:          appPort,
		MaxCommentLength: maxCommentLength,
		DefaultListLimit: defaultListLimit,
		MaxListLimit:     maxListLimit,
		MinUsernameLen:   minUsernameLen,
	}, nil
}

func getEnvIntWithDefault(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("Переменная %s не найдена, используется значение по умолчанию: %d", key, defaultVal)
		return defaultVal
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Переменная %s имеет некорректное значение (%s), используется значение по умолчанию: %d", key, val, defaultVal)
		return defaultVal
	}
	return num
}

package customerrors

import "errors"

var (
	ErrEnvNotFound     = errors.New("файл .env не найден")
	ErrParamNotFound   = errors.New("один или несколько критически важных параметров не были найдены в env, проверьте: POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB")
	ErrInvalidEnvValue = errors.New("некорректное значение переменной окружения")
)

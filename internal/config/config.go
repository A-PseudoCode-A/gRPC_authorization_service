package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env" envDefault:"local"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-default:"1h"`
	GRPC        GRPCConfig    `yaml:"grpc"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

	// Проверка на наличие пути
	if path == "" {
		panic("config path is empty")
	}

	// Проверка на наличие файла
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file is not exists: " + path)
	}

	// Переменная, в которую сохранится конфиг
	var cfg Config

	// Проверка на считывание конфига
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var result string

	// Создание флага config, а также парсинг этого флага
	flag.StringVar(&result, "config", "", "path to config file")
	flag.Parse()

	// Если флаг не найден - то используется переменная окружения
	if result == "" {
		result = os.Getenv("CONFIG_PATH")
	}

	return result
}

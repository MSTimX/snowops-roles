package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/MSTimX/Snowops-roles/internal/models"
)

// DB хранит глобальное подключение к базе данных.
var DB *gorm.DB

// Init загружает конфигурацию и инициализирует подключение к PostgreSQL.
func Init() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("не удалось загрузить .env файл: %v", err)
	}

	required := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
	}

	config := make(map[string]string, len(required))
	for _, key := range required {
		value := os.Getenv(key)
		if value == "" {
			log.Fatalf("переменная окружения %s не установлена", key)
		}
		config[key] = value
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Almaty",
		config["DB_HOST"],
		config["DB_USER"],
		config["DB_PASSWORD"],
		config["DB_NAME"],
		config["DB_PORT"],
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
	}

	DB = db
}

// Migrate выполняет авто-миграции для всех моделей.
func Migrate() {
	if DB == nil {
		log.Fatalf("подключение к базе данных не инициализировано")
	}

	if err := DB.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Driver{},
		&models.Vehicle{},
	); err != nil {
		log.Fatalf("ошибка авто-миграции: %v", err)
	}
}

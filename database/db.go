package database

import (
	"fmt"
	"lea/entities"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	err := godotenv.Load()

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		host, user, password, dbname, port,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v\n", err)
	}
	fmt.Println("Banco conectado")

	err = DB.AutoMigrate(
		&entities.User{},
		&entities.Post{},
		&entities.Follower{},
		&entities.Role{},
		&entities.Category{},
		&entities.Message{},
		&entities.Group{},
		&entities.GroupMember{},
		&entities.Comment{},
	)

	if err != nil {
		log.Fatalf("Erro migracao: %v", err)
	}
	fmt.Println("Migracoes executadas")
}

package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dbUser := os.Getenv("DB_USER")

	dbHost := os.Getenv("DB_HOST")
	
	dbPort := os.Getenv("DB_PORT")
	
	dbName := os.Getenv("DB_NAME")
	

	dsn := fmt.Sprintf("%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbHost, dbPort, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("cant connect to the database: %v", err)
	}

	DB = db
	log.Print("Connected to the database successfully")
}

func NewDB() *gorm.DB {
	if DB == nil {
		Connect()
	}
	return DB
}

package main

import (
	"backend_institutions/internal/database"
	"backend_institutions/internal/grpc"
	"backend_institutions/internal/model"
	"backend_institutions/internal/seeds"
	"backend_institutions/internal/wire"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		err = godotenv.Load("../.env")
		if err != nil {
			log.Println("Warning: Error loading .env file from current or parent directory, relying on environment variables or defaults")
		}
	}

	database.Connect()

	

	err = database.DB.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.User{},
		&model.Institutions{},
		&model.Department{},
		&model.Faculty{},
		&model.Student{},
		&model.DepartmentPayment{},
		&model.StudentPayment{},
		&model.Session{},
		&model.Institution_Admins{},
	)
	if err != nil {
		log.Fatal(err)
	}
	

	seeds.RunSeeders()

	err = grpc.ConnectService()
	if err != nil {
		log.Println("Failed to connect to notification gRPC service:", err)
	}

	app, err := wire.InitializeApp()
	if err != nil {
		log.Fatal("Failed to initialize application: ", err)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8090"
	}
	log.Printf("Server starting on :%s", port)

	log.Fatal(app.Listen(":" + port))
	
}

	

	

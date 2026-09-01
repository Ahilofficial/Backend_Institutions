package main

import (
	"backend_institutions/internal/database"
	"backend_institutions/internal/grpc"
	"backend_institutions/internal/model"
	"backend_institutions/internal/seeds"
	"backend_institutions/internal/wire"
	"github.com/joho/godotenv"
	"log"
	"os"
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
		&model.Institutions{},
		&model.Department{},
		&model.Faculty{},
		&model.Student{},
		&model.Fees{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.Session{},
		&model.Menu{},
		&model.RoleMenu{},
		&model.Payment{},
		&model.StudentPayment{},
		&model.Institution_Admins{},
		&model.StudentVerificationAccess{},
	)
	if err != nil {
		log.Fatal(err)
	}

	seeds.RunSeeders()

	app, err := wire.InitializeApp()
	if err != nil {
		log.Fatal("Failed to initialize application: ", err)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8090"
	}
	log.Printf("Server starting on :%s", port)
	err = grpc.ConnectLogger()
	if err != nil {
		log.Fatal("Failed to connect to Logger Service:", err)
	}

	err = grpc.ConnectService()
	if err != nil {
		log.Fatal("Failed to connect to Notification Service:", err)
	}

	log.Fatal(app.Listen(":" + port))
}

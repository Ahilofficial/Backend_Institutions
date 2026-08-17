package seeds

import (
	"backend_institutions/internal/constants"
	"backend_institutions/internal/database"
	// "backend_institutions/internal/model"
	"log"

	// "golang.org/x/crypto/bcrypt"
)

func RunSeeders() {
	log.Println("Running database seeders...")

	seedPermissions()
	
	
	log.Println("Database seeding completed.")
}



func seedPermissions() {
	db, err := database.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return
	}

	for _, pName := range constants.AllPermissions {
		var count int

		err := db.QueryRow(
			"SELECT COUNT(*) FROM permissions WHERE name = ?",
			pName,
		).Scan(&count)

		if err != nil {
			log.Printf("Failed to check permission %s: %v", pName, err)
			continue
		}

		if count == 0 {
			_, err = db.Exec(
				"INSERT INTO permissions (name) VALUES (?)",
				pName,
			)

			if err != nil {
				log.Printf("Failed to insert permission %s: %v", pName, err)
			}
		}
	}
}




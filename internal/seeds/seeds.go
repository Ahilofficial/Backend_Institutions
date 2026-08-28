package seeds

import (
	"backend_institutions/internal/constants"
	"backend_institutions/internal/database"
	"log"
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
			"SELECT COUNT(*) FROM permissions WHERE LOWER(TRIM(name)) = LOWER(TRIM(?))",
			pName,
		).Scan(&count)

		if err != nil {
			log.Printf("Failed to check permission %s: %v", pName, err)
			continue
		}

		if count == 0 {
			_, err = db.Exec(
				"INSERT INTO permissions (name, created_at, updated_at) VALUES (?, NOW(), NOW())",
				pName,
			)
			if err != nil {
				log.Printf("Failed to insert permission %s: %v", pName, err)
			}
		}
	}
}

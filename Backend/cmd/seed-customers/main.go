package main

import (
	"fmt"
	"log"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
)

type customerSeed struct {
	name    string
	email   string
	contact string
}

var customerSeeds = []customerSeed{
	{name: "Budi Santoso", email: "budi.santoso@example.com", contact: "081234567890"},
	{name: "Siti Rahmawati", email: "siti.rahmawati@example.com", contact: "081298765432"},
	{name: "Andi Wijaya", email: "andi.wijaya@example.com", contact: "081356789012"},
	{name: "Nadia Putri", email: "nadia.putri@example.com", contact: "082112345678"},
	{name: "Rizky Pratama", email: "rizky.pratama@example.com", contact: "081987654321"},
	{name: "Dewi Lestari", email: "dewi.lestari@example.com", contact: "085712345678"},
	{name: "Fajar Hidayat", email: "fajar.hidayat@example.com", contact: "082234567890"},
	{name: "Maya Permata", email: "maya.permata@example.com", contact: "081245678901"},
	{name: "Yoga Saputra", email: "yoga.saputra@example.com", contact: "089512345678"},
	{name: "Intan Maharani", email: "intan.maharani@example.com", contact: "083812345678"},
}

func main() {
	cfg := config.Load()
	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	var businessID string
	err = db.Get(&businessID, `
		SELECT id FROM businesses
		WHERE LOWER(name) = LOWER($1) AND deleted = FALSE
		ORDER BY created_at LIMIT 1`, "Arya Corp")
	if err != nil {
		log.Fatalf("find Arya Corp: %v", err)
	}

	inserted := 0
	skipped := 0
	for _, seed := range customerSeeds {
		var exists bool
		err = db.Get(&exists, `
			SELECT EXISTS(
				SELECT 1 FROM customers
				WHERE business_id = $1 AND LOWER(email) = LOWER($2)
			)`, businessID, seed.email)
		if err != nil {
			log.Fatalf("check customer %s: %v", seed.email, err)
		}
		if exists {
			skipped++
			continue
		}

		_, err = db.Exec(`
			INSERT INTO customers (business_id, name, email, contact, meta, created_at)
			VALUES ($1, $2, $3, $4, '{}'::jsonb, NOW())`,
			businessID, seed.name, seed.email, seed.contact)
		if err != nil {
			log.Fatalf("insert customer %s: %v", seed.email, err)
		}
		inserted++
	}

	fmt.Printf("Arya Corp (%s): %d pelanggan ditambahkan, %d dilewati karena sudah ada\n", businessID, inserted, skipped)
}

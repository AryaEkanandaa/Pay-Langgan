package main

import (
	"flag"
	"log"
	"strings"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	identityrepo "pay-langgan/internal/repositories/identity"
	"pay-langgan/internal/utils"
)

func main() {
	name := flag.String("name", "", "nama Super Admin")
	email := flag.String("email", "", "email Super Admin")
	password := flag.String("password", "", "password Super Admin")
	flag.Parse()

	if strings.TrimSpace(*name) == "" || strings.TrimSpace(*email) == "" || len(*password) < 6 {
		log.Fatal("name and email are required; password must be at least 6 characters")
	}

	cfg := config.Load()
	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := identityrepo.NewUserRepository(db)
	normalizedEmail := strings.ToLower(strings.TrimSpace(*email))
	existing, err := repo.FindByEmail(normalizedEmail)
	if err != nil {
		log.Fatal(err)
	}
	if existing != nil {
		log.Fatalf("email already registered: %s", normalizedEmail)
	}

	hashedPassword, err := utils.HashPassword(*password)
	if err != nil {
		log.Fatal(err)
	}

	user := &models.User{
		Name:     strings.TrimSpace(*name),
		Email:    normalizedEmail,
		Password: hashedPassword,
		Role:     string(models.RoleSuperAdmin),
	}
	if err := repo.Create(user); err != nil {
		log.Fatal(err)
	}

	log.Printf("super admin created: %s", user.Email)
}

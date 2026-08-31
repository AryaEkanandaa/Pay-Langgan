package main

import (
	"fmt"
	"log"
	"time"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
)

type customerSeed struct {
	name string
}

type planSeed struct {
	ID           int     `db:"id"`
	Name         string  `db:"name"`
	Price        float64 `db:"price"`
	BillingCycle string  `db:"billing_cycle"`
	TrialDays    int     `db:"trial_days"`
}

var customerSeeds = []customerSeed{
	{name: "Budi Santoso"},
	{name: "Siti Rahmawati"},
	{name: "Andi Wijaya"},
	{name: "Nadia Putri"},
	{name: "Rizky Pratama"},
	{name: "Dewi Lestari"},
	{name: "Fajar Hidayat"},
	{name: "Maya Permata"},
	{name: "Yoga Saputra"},
	{name: "Intan Maharani"},
}

var statuses = []string{
	"active", "active", "active", "trial", "active",
	"paused", "active", "trial", "active", "cancelled",
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

	var plans []planSeed
	err = db.Select(&plans, `
		SELECT pl.id, pl.name, pl.price, pl.billing_cycle, pl.trial_days
		FROM plans pl
		JOIN products p ON p.id = pl.product_id
		JOIN services s ON s.id = p.service_id
		WHERE s.business_id = $1
		ORDER BY pl.id`, businessID)
	if err != nil {
		log.Fatalf("find Arya Corp plans: %v", err)
	}
	if len(plans) == 0 {
		log.Fatal("Arya Corp belum memiliki plan. Buat katalog plan terlebih dahulu.")
	}

	inserted := 0
	skipped := 0
	now := time.Now()
	for index, customer := range customerSeeds {
		var customerID int
		err = db.Get(&customerID, `
			SELECT id FROM customers
			WHERE business_id = $1 AND LOWER(name) = LOWER($2)
			LIMIT 1`, businessID, customer.name)
		if err != nil {
			log.Fatalf("find customer %s: %v", customer.name, err)
		}

		var exists bool
		err = db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE customer_id = $1)`, customerID)
		if err != nil {
			log.Fatalf("check subscription for %s: %v", customer.name, err)
		}
		if exists {
			skipped++
			continue
		}

		plan := plans[index%len(plans)]
		status := statuses[index]
		startDate := now.AddDate(0, 0, -(index+1)*4)
		var nextBillingDate *time.Time
		var endDate *time.Time
		var trialEndsAt *time.Time

		switch status {
		case "trial":
			trialDays := plan.TrialDays
			if trialDays < 1 {
				trialDays = 7
			}
			startDate = now.AddDate(0, 0, -2)
			trialEnd := startDate.AddDate(0, 0, trialDays)
			trialEndsAt = &trialEnd
			nextBillingDate = &trialEnd
		case "active":
			next := now.AddDate(0, 0, 15+index)
			nextBillingDate = &next
		case "paused":
			next := now.AddDate(0, 0, 30)
			nextBillingDate = &next
		case "cancelled":
			end := startDate.AddDate(0, 1, 0)
			endDate = &end
		}

		tx, err := db.Beginx()
		if err != nil {
			log.Fatalf("begin transaction for %s: %v", customer.name, err)
		}

		var subscriptionID int
		err = tx.QueryRow(`
			INSERT INTO subscriptions
				(customer_id, plan_id, status, start_date, next_billing_date, end_date, trial_ends_at, meta, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, '{"source":"seed","purpose":"demo TA"}'::jsonb, NOW())
			RETURNING id`, customerID, plan.ID, status, startDate, nextBillingDate, endDate, trialEndsAt).Scan(&subscriptionID)
		if err != nil {
			tx.Rollback()
			log.Fatalf("create subscription for %s: %v", customer.name, err)
		}

		invoiceStatus := "pending"
		var paidAt *time.Time
		if status == "cancelled" {
			invoiceStatus = "cancelled"
		} else if index == 0 || index == 4 || index == 8 {
			invoiceStatus = "paid"
			paid := now.AddDate(0, 0, -index)
			paidAt = &paid
		}

		invoiceNumber := fmt.Sprintf("INV-ARYA-SEED-%s-%02d", now.Format("20060102"), index+1)
		var invoiceID int
		err = tx.QueryRow(`
			INSERT INTO invoices (subscription_id, invoice_number, amount, status, due_date, paid_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			RETURNING id`, subscriptionID, invoiceNumber, plan.Price, invoiceStatus, nextBillingDate, paidAt).Scan(&invoiceID)
		if err != nil {
			tx.Rollback()
			log.Fatalf("create invoice for %s: %v", customer.name, err)
		}

		_, err = tx.Exec(`
			INSERT INTO invoice_items
				(invoice_id, item_type, description, quantity, unit_price, subtotal, meta, created_at)
			VALUES ($1, 'plan', $2, 1, $3, $3, '{}'::jsonb, NOW())`, invoiceID, plan.Name, plan.Price)
		if err != nil {
			tx.Rollback()
			log.Fatalf("create invoice item for %s: %v", customer.name, err)
		}

		if invoiceStatus == "paid" {
			_, err = tx.Exec(`
				INSERT INTO payments (invoice_id, provider, provider_transaction_id, amount, status, paid_at, created_at)
				VALUES ($1, 'manual', $2, $3, 'success', $4, NOW())`,
				invoiceID, fmt.Sprintf("SEED-%s", invoiceNumber), plan.Price, paidAt)
			if err != nil {
				tx.Rollback()
				log.Fatalf("create payment for %s: %v", customer.name, err)
			}
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("commit subscription for %s: %v", customer.name, err)
		}
		inserted++
		fmt.Printf("%s -> %s (%.2f, %s)\n", customer.name, plan.Name, plan.Price, status)
	}

	fmt.Printf("Arya Corp (%s): %d langganan ditambahkan, %d dilewati karena sudah ada\n", businessID, inserted, skipped)
}

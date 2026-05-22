package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var merchants = []string{
	"Tokopedia", "Shopee", "Bukalapak", "Lazada", "Blibli",
	"JD.id", "Zalora", "Sociolla", "Alfamart", "Indomaret",
	"Gojek", "Grab", "OVO Store", "Dana Merchant", "LinkAja Shop",
}

var statuses = []string{
	"completed", "completed", "completed", "completed", "completed", "completed",
	"processing", "processing", "processing",
	"failed", "failed",
}

func main() {
	_ = godotenv.Load()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dashboard.db"
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := seedUsers(db); err != nil {
		log.Fatalf("seed users: %v", err)
	}
	if err := seedPayments(db); err != nil {
		log.Fatalf("seed payments: %v", err)
	}
	log.Println("Seed complete.")
}

func seedUsers(db *sql.DB) error {
	var cnt int
	if err := db.QueryRow("SELECT COUNT(1) FROM users").Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		log.Println("users already seeded, skipping")
		return nil
	}
	users := []struct{ email, password, role string }{
		{"cs@test.com", "password", "cs"},
		{"operation@test.com", "password", "operation"},
	}
	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := db.Exec(
			"INSERT INTO users(email, password_hash, role) VALUES (?, ?, ?)",
			u.email, string(hash), u.role,
		); err != nil {
			return err
		}
	}
	log.Printf("seeded %d users", len(users))
	return nil
}

func seedPayments(db *sql.DB) error {
	var cnt int
	if err := db.QueryRow("SELECT COUNT(1) FROM payments").Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		log.Println("payments already seeded, skipping")
		return nil
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	for i := 1; i <= 50; i++ {
		id := fmt.Sprintf("PAY-%03d", i)
		merchant := merchants[rng.Intn(len(merchants))]
		amount := int64((rng.Intn(490)+10)*1000)
		status := statuses[rng.Intn(len(statuses))]
		createdAt := now.Add(-time.Duration(rng.Intn(90*24)) * time.Hour)

		if _, err := db.Exec(
			"INSERT INTO payments(id, merchant, amount, status, created_at) VALUES (?, ?, ?, ?, ?)",
			id, merchant, amount, status, createdAt.Format("2006-01-02 15:04:05"),
		); err != nil {
			return err
		}
	}
	log.Println("seeded 50 payments")
	return nil
}

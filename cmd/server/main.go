package main

import (
	"log"
	"net/http"

	"gopay-recon/internal/api"
	"gopay-recon/internal/db"
)

func main() {
	dbURL := "postgres://postgres:nineleaps@localhost:5432/postgres?sslmode=disable"
	database, err := db.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	app := &api.App{DB: database}

	http.HandleFunc("/", app.HomeHandler)
	http.HandleFunc("/reconciliation/run", app.RunReconciliationHandler)

	// Register the new Day 5 routes
	http.HandleFunc("/dashboard", app.DashboardHandler)
	http.HandleFunc("/export", app.ExportHandler)

	log.Println("GoPay Recon server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

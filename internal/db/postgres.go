package db

import (
	"database/sql"
	"log"

	"gopay-recon/internal/reconciler"

	_ "github.com/lib/pq"
)

// --- DAY 3: SETUP & INSERT ---

func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	createTables(db)
	return db, nil
}

func createTables(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS reconciliation_results (
		id SERIAL PRIMARY KEY,
		transaction_id VARCHAR(100),
		internal_amount NUMERIC(10,2),
		processor_amount NUMERIC(10,2),
		currency VARCHAR(10),
		status VARCHAR(50),
		exception_reason TEXT,
		difference NUMERIC(10,2),
		reconciled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}

func InsertResult(db *sql.DB, r reconciler.ReconciliationResult) error {
	query := `INSERT INTO reconciliation_results 
		(transaction_id, internal_amount, processor_amount, currency, status, exception_reason, difference) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.Exec(query, r.TransactionID, r.InternalAmount, r.ProcessorAmount, r.Currency, r.ReconciliationStatus, r.ExceptionReason, r.Difference)
	return err
}

// --- DAY 5: FETCH FOR DASHBOARD & EXPORT ---

type SummaryStats struct {
	Total            int
	Matched          int
	AmountMismatch   int
	StatusMismatch   int
	MissingProcessor int
	MissingInternal  int
}

func GetSummary(db *sql.DB) (SummaryStats, error) {
	var stats SummaryStats

	query := `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'MATCHED'),
			COUNT(*) FILTER (WHERE status = 'AMOUNT_MISMATCH'),
			COUNT(*) FILTER (WHERE status = 'STATUS_MISMATCH'),
			COUNT(*) FILTER (WHERE status = 'MISSING_PROCESSOR'),
			COUNT(*) FILTER (WHERE status = 'MISSING_INTERNAL')
		FROM reconciliation_results`

	err := db.QueryRow(query).Scan(
		&stats.Total, &stats.Matched, &stats.AmountMismatch,
		&stats.StatusMismatch, &stats.MissingProcessor, &stats.MissingInternal,
	)
	return stats, err
}

func GetAllResults(db *sql.DB) ([]reconciler.ReconciliationResult, error) {
	query := `SELECT transaction_id, internal_amount, processor_amount, currency, status, exception_reason, COALESCE(difference, 0) FROM reconciliation_results ORDER BY status DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []reconciler.ReconciliationResult
	for rows.Next() {
		var r reconciler.ReconciliationResult
		err := rows.Scan(&r.TransactionID, &r.InternalAmount, &r.ProcessorAmount, &r.Currency, &r.ReconciliationStatus, &r.ExceptionReason, &r.Difference)
		if err == nil {
			results = append(results, r)
		}
	}
	return results, nil
}

func ClearResults(db *sql.DB) {
	db.Exec("TRUNCATE TABLE reconciliation_results")
}

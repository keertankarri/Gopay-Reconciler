package api

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"

	"gopay-recon/internal/db"
	"gopay-recon/internal/models"
	"gopay-recon/internal/parser"
	"gopay-recon/internal/reconciler"
	"gopay-recon/internal/worker"
)

type App struct {
	DB *sql.DB
}

func (a *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("ui/templates/index.html"))
	tmpl.Execute(w, nil)
}

func (a *App) RunReconciliationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseMultipartForm(10 << 20)
	workers, _ := strconv.Atoi(r.FormValue("workers"))
	if workers < 1 {
		workers = 4
	}

	internalPath := saveTempFile(r, "internal")
	processorPath := saveTempFile(r, "processor")
	defer os.Remove(internalPath)
	defer os.Remove(processorPath)

	internalRecords, _, _ := parser.ParseInternal(internalPath)
	processorRecords, _, _ := parser.ParseProcessor(processorPath)

	procMap := make(map[string]models.ProcessorPayment)
	for _, p := range processorRecords {
		procMap[p.TransactionID] = p
	}

	results := worker.RunPool(workers, internalRecords, procMap)

	matchedIDs := make(map[string]bool)
	for _, res := range results {
		matchedIDs[res.TransactionID] = true
	}
	for txnID, procPay := range procMap {
		if !matchedIDs[txnID] {
			results = append(results, reconciler.ReconciliationResult{
				TransactionID: txnID, ProcessorAmount: procPay.Amount, ReconciliationStatus: "MISSING_INTERNAL",
			})
		}
	}

	// Clear old run data and insert new
	db.ClearResults(a.DB)
	for _, res := range results {
		db.InsertResult(a.DB, res)
	}

	// Redirect to Dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// NEW: Dashboard Handler
func (a *App) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	stats, _ := db.GetSummary(a.DB)
	allResults, _ := db.GetAllResults(a.DB)

	var exceptions []reconciler.ReconciliationResult
	for _, res := range allResults {
		if res.ReconciliationStatus != "MATCHED" {
			exceptions = append(exceptions, res)
		}
	}

	data := struct {
		Stats      db.SummaryStats
		Exceptions []reconciler.ReconciliationResult
	}{Stats: stats, Exceptions: exceptions}

	tmpl := template.Must(template.ParseFiles("ui/templates/dashboard.html"))
	tmpl.Execute(w, data)
}

// NEW: CSV Export Handler
func (a *App) ExportHandler(w http.ResponseWriter, r *http.Request) {
	results, _ := db.GetAllResults(a.DB)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=reconciliation_report.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"TransactionID", "InternalAmount", "ProcessorAmount", "Currency", "Status", "ExceptionReason", "Difference"})
	for _, res := range results {
		writer.Write([]string{
			res.TransactionID,
			fmt.Sprintf("%.2f", res.InternalAmount),
			fmt.Sprintf("%.2f", res.ProcessorAmount),
			res.Currency,
			res.ReconciliationStatus,
			res.ExceptionReason,
			fmt.Sprintf("%.2f", res.Difference),
		})
	}
}

func saveTempFile(r *http.Request, formKey string) string {
	file, _, err := r.FormFile(formKey)
	if err != nil {
		return ""
	}
	defer file.Close()
	tempFile, _ := os.CreateTemp("", formKey+"-*.csv")
	defer tempFile.Close()
	io.Copy(tempFile, file)
	return tempFile.Name()
}

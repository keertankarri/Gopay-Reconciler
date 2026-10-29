package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	intFile, _ := os.Create("internal_payments.csv")
	procFile, _ := os.Create("processor_payments.csv")
	defer intFile.Close()
	defer procFile.Close()

	intWriter := csv.NewWriter(intFile)
	procWriter := csv.NewWriter(procFile)
	defer intWriter.Flush()
	defer procWriter.Flush()

	// Write Headers
	intWriter.Write([]string{"PaymentID", "TransactionID", "CustomerID", "PaymentTS", "Amount", "Currency", "PaymentMethod", "Status"})
	procWriter.Write([]string{"ProcessorTxnID", "TransactionID", "ProcessorTS", "Amount", "Currency", "Status", "Processor"})

	rand.Seed(time.Now().UnixNano())

	for i := 1; i <= 10000; i++ {
		txnID := fmt.Sprintf("TXN%05d", i)
		amount := float64(rand.Intn(1000) + 100)
		status := "SUCCESS"

		scenario := rand.Intn(100) // Generate a number between 0 and 99

		if scenario < 85 {
			// 85% MATCHED
			writeInternal(intWriter, i, txnID, amount, status)
			writeProcessor(procWriter, i, txnID, amount, status)
		} else if scenario < 88 {
			// 3% AMOUNT_MISMATCH (Processor has 50 less)
			writeInternal(intWriter, i, txnID, amount, status)
			writeProcessor(procWriter, i, txnID, amount-50, status)
		} else if scenario < 91 {
			// 3% STATUS_MISMATCH (Processor says FAILED)
			writeInternal(intWriter, i, txnID, amount, status)
			writeProcessor(procWriter, i, txnID, amount, "FAILED")
		} else if scenario < 95 {
			// 4% MISSING_PROCESSOR (Only in Internal)
			writeInternal(intWriter, i, txnID, amount, status)
		} else {
			// 5% MISSING_INTERNAL (Only in Processor)
			writeProcessor(procWriter, i, txnID, amount, status)
		}
	}
	fmt.Println("Successfully generated 10,000 records of mock data!")
}

func writeInternal(w *csv.Writer, i int, txnID string, amount float64, status string) {
	w.Write([]string{
		fmt.Sprintf("P%d", i), txnID, fmt.Sprintf("C%d", i), time.Now().Format(time.RFC3339),
		fmt.Sprintf("%.2f", amount), "INR", "UPI", status,
	})
}

func writeProcessor(w *csv.Writer, i int, txnID string, amount float64, status string) {
	w.Write([]string{
		fmt.Sprintf("PR%d", i), txnID, time.Now().Format(time.RFC3339),
		fmt.Sprintf("%.2f", amount), "INR", status, "STRIPE",
	})
}

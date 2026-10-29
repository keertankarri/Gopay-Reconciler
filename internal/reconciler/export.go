package reconciler

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ExportToCSV(results []ReconciliationResult, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write Header
	writer.Write([]string{
		"TransactionID", "InternalAmount", "ProcessorAmount", "Currency", "ReconciliationStatus", "ExceptionReason", "Difference",
	})

	// Write Rows
	for _, r := range results {
		writer.Write([]string{
			r.TransactionID,
			fmt.Sprintf("%.2f", r.InternalAmount),
			fmt.Sprintf("%.2f", r.ProcessorAmount),
			r.Currency,
			r.ReconciliationStatus,
			r.ExceptionReason,
			fmt.Sprintf("%.2f", r.Difference),
		})
	}
	return nil
}

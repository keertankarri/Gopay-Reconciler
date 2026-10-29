package parser

import (
	"encoding/csv"
	"os"
	"strconv"
	"gopay-recon/internal/models"
)

// ParseInternal reads the internal CSV and returns valid records + invalid count
func ParseInternal(filePath string) ([]models.InternalPayment, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, err
	}

	var payments []models.InternalPayment
	invalidCount := 0

	for i, row := range records {
		if i == 0 { continue } // Skip header row

		// Basic validation: Check if required fields exist
		if len(row) < 8 || row[1] == "" || row[5] == "" {
			invalidCount++
			continue
		}

		amt, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			invalidCount++
			continue // Skip rows with invalid amounts
		}

		payments = append(payments, models.InternalPayment{
			PaymentID:     row[0],
			TransactionID: row[1],
			CustomerID:    row[2],
			PaymentTS:     row[3],
			Amount:        amt,
			Currency:      row[5],
			PaymentMethod: row[6],
			Status:        row[7],
		})
	}
	return payments, invalidCount, nil
}

// ParseProcessor reads the processor CSV and returns valid records + invalid count
func ParseProcessor(filePath string) ([]models.ProcessorPayment, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, err
	}

	var payments []models.ProcessorPayment
	invalidCount := 0

	for i, row := range records {
		if i == 0 { continue } // Skip header

		if len(row) < 7 || row[1] == "" || row[4] == "" {
			invalidCount++
			continue
		}

		amt, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			invalidCount++
			continue
		}

		payments = append(payments, models.ProcessorPayment{
			ProcessorTxnID: row[0],
			TransactionID:  row[1],
			ProcessorTS:    row[2],
			Amount:         amt,
			Currency:       row[4],
			Status:         row[5],
			Processor:      row[6],
		})
	}
	return payments, invalidCount, nil
}
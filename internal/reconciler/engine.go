package reconciler

import (
	"gopay-recon/internal/models"
)

type ReconciliationResult struct {
	TransactionID        string
	InternalAmount       float64
	ProcessorAmount      float64
	Currency             string
	ReconciliationStatus string
	ExceptionReason      string
	Difference           float64
}

type Summary struct {
	Total            int
	Matched          int
	AmountMismatch   int
	StatusMismatch   int
	MissingProcessor int
	MissingInternal  int
	Duplicates       int
	Invalid          int
}

func Reconcile(internal []models.InternalPayment, processor []models.ProcessorPayment, invalidCount int) ([]ReconciliationResult, Summary) {
	var results []ReconciliationResult
	summary := Summary{Invalid: invalidCount}

	// 1. Build Internal map and detect duplicates
	internalMap := make(map[string]models.InternalPayment)
	seenInternal := make(map[string]bool)

	for _, p := range internal {
		if seenInternal[p.TransactionID] {
			summary.Duplicates++
			results = append(results, ReconciliationResult{
				TransactionID:        p.TransactionID,
				InternalAmount:       p.Amount,
				Currency:             p.Currency,
				ReconciliationStatus: "DUPLICATE",
				ExceptionReason:      "Transaction ID appears multiple times in internal file",
			})
			continue
		}
		seenInternal[p.TransactionID] = true
		internalMap[p.TransactionID] = p
	}

	// 2. Build Processor map
	procMap := make(map[string]models.ProcessorPayment)
	for _, p := range processor {
		procMap[p.TransactionID] = p
	}

	// 3. Compare Internal against Processor
	for txnID, intPay := range internalMap {
		procPay, exists := procMap[txnID]

		if !exists {
			summary.MissingProcessor++
			results = append(results, ReconciliationResult{
				TransactionID:        txnID,
				InternalAmount:       intPay.Amount,
				Currency:             intPay.Currency,
				ReconciliationStatus: "MISSING_PROCESSOR",
				ExceptionReason:      "Not found in processor file",
			})
			continue
		}

		// Check matching rules
		diff := intPay.Amount - procPay.Amount
		if intPay.Amount != procPay.Amount {
			summary.AmountMismatch++
			results = append(results, ReconciliationResult{
				TransactionID:        txnID,
				InternalAmount:       intPay.Amount,
				ProcessorAmount:      procPay.Amount,
				Currency:             intPay.Currency,
				ReconciliationStatus: "AMOUNT_MISMATCH",
				ExceptionReason:      "Amounts do not match",
				Difference:           diff,
			})
		} else if intPay.Status != procPay.Status {
			summary.StatusMismatch++
			results = append(results, ReconciliationResult{
				TransactionID:        txnID,
				InternalAmount:       intPay.Amount,
				ProcessorAmount:      procPay.Amount,
				Currency:             intPay.Currency,
				ReconciliationStatus: "STATUS_MISMATCH",
				ExceptionReason:      "Statuses do not match",
			})
		} else {
			summary.Matched++
			results = append(results, ReconciliationResult{
				TransactionID:        txnID,
				InternalAmount:       intPay.Amount,
				ProcessorAmount:      procPay.Amount,
				Currency:             intPay.Currency,
				ReconciliationStatus: "MATCHED",
			})
		}

		// Remove from processor map so we can find the "Missing Internal" records later
		delete(procMap, txnID)
	}

	// 4. Sweep remaining processor records (Missing Internal)
	for txnID, procPay := range procMap {
		summary.MissingInternal++
		results = append(results, ReconciliationResult{
			TransactionID:        txnID,
			ProcessorAmount:      procPay.Amount,
			Currency:             procPay.Currency,
			ReconciliationStatus: "MISSING_INTERNAL",
			ExceptionReason:      "Not found in internal file",
		})
	}

	summary.Total = summary.Matched + summary.AmountMismatch + summary.StatusMismatch + summary.MissingProcessor + summary.MissingInternal + summary.Duplicates + summary.Invalid
	return results, summary
}

package worker

import (
	"gopay-recon/internal/models"
	"gopay-recon/internal/reconciler"
	"sync"
)

// RunPool distributes reconciliation jobs across multiple goroutines
func RunPool(workers int, internal []models.InternalPayment, procMap map[string]models.ProcessorPayment) []reconciler.ReconciliationResult {
	jobs := make(chan models.InternalPayment, len(internal))
	results := make(chan reconciler.ReconciliationResult, len(internal))
	var wg sync.WaitGroup

	// 1. Start Workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for intPay := range jobs {
				procPay, exists := procMap[intPay.TransactionID]

				var res reconciler.ReconciliationResult
				if !exists {
					res = reconciler.ReconciliationResult{
						TransactionID:        intPay.TransactionID,
						InternalAmount:       intPay.Amount,
						Currency:             intPay.Currency,
						ReconciliationStatus: "MISSING_PROCESSOR",
					}
				} else {
					status := "MATCHED"
					diff := intPay.Amount - procPay.Amount
					if intPay.Amount != procPay.Amount {
						status = "AMOUNT_MISMATCH"
					} else if intPay.Status != procPay.Status {
						status = "STATUS_MISMATCH"
					}

					res = reconciler.ReconciliationResult{
						TransactionID:        intPay.TransactionID,
						InternalAmount:       intPay.Amount,
						ProcessorAmount:      procPay.Amount,
						Currency:             intPay.Currency,
						ReconciliationStatus: status,
						Difference:           diff,
					}
				}
				results <- res
			}
		}()
	}

	// 2. Send Jobs
	for _, p := range internal {
		jobs <- p
	}
	close(jobs)

	// 3. Wait and Close Results
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. Collect Results
	var finalResults []reconciler.ReconciliationResult
	for res := range results {
		finalResults = append(finalResults, res)
	}

	return finalResults
}

package models

type InternalPayment struct {
	PaymentID     string
	TransactionID string
	CustomerID    string
	PaymentTS     string
	Amount        float64
	Currency      string
	PaymentMethod string
	Status        string
}

type ProcessorPayment struct {
	ProcessorTxnID string
	TransactionID  string
	ProcessorTS    string
	Amount         float64
	Currency       string
	Status         string
	Processor      string
}
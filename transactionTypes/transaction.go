// Package transactionTypes provides the "Transaction" and "APIResponse" type structs.
package transactionTypes

import "time"

// Transaction provides the structure of the transaction table in the bitcoin_wallet.db
type Transaction struct {
	TransactionID string
	Amount        float64
	Spent         bool
	CreatedAt     time.Time
}

// APIResponse provides the structure of all of the API endpoint responses in this project.
type APIResponse struct {
	Data   interface{} `json:"data"`
	Errors []string    `json:"errors,omitempty"`
}

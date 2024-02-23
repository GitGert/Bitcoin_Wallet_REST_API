// Package transaction_types provides the "Transaction" and "APIResponse" type structs.
package transaction_types

import "time"

// Transaction provides the structure of the transaction table in the bitcoin_wallet.db
type Transaction struct {
	Transaction_ID string
	Amount         float64
	Spent          bool
	Created_at     time.Time
}

// APIResponse provides the structure of all of the API endpoint responses in this project.
type APIResponse struct {
	Data   interface{} `json:"data"`
	Errors []string    `json:"errors,omitempty"`
}

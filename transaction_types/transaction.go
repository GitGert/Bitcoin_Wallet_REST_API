package transaction_types

import "time"

type Transaction struct {
	Transaction_ID string
	Amount         float64
	Spent          bool
	Created_at     time.Time
}

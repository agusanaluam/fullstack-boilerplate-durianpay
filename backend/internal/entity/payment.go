package entity

import "time"

type Payment struct {
	ID        string    `json:"id"`
	Merchant  string    `json:"merchant"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

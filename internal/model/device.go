package model

type Device struct {
	ID          string  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Description string  `db:"description" json:"description"`
	Category    string  `db:"category" json:"category"`
	PricePerDay float64 `db:"price_per_day" json:"price_per_day"`
	Available   bool    `db:"available" json:"available"`
	ImageURL    string  `db:"image_url" json:"image_url"`
	OwnerID     string  `db:"owner_id" json:"owner_id"`
	City        string  `db:"city" json:"city"`
	Region      string  `db:"region" json:"region"`
	CreatedAt   *string `db:"created_at" json:"created_at"`
	UpdatedAt   *string `db:"updated_at" json:"updated_at"`
}

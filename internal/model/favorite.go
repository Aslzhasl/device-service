package model

type Favorite struct {
	UserID    string `db:"user_id" json:"user_id"`
	DeviceID  string `db:"device_id" json:"device_id"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

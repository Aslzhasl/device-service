package repository

import (
	"context"
	"device-service/internal/model"
	"errors"
	"github.com/jmoiron/sqlx"
)

type FavoriteRepository struct {
	DB *sqlx.DB
}

func NewFavoriteRepository(db *sqlx.DB) *FavoriteRepository {
	return &FavoriteRepository{DB: db}
}

// Add a device to the user's favorites
func (r *FavoriteRepository) AddFavorite(ctx context.Context, userID, deviceID string) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO favorites (user_id, device_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, deviceID,
	)
	return err
}

// Remove a device from the user's favorites
func (r *FavoriteRepository) RemoveFavorite(ctx context.Context, userID, deviceID string) error {
	result, err := r.DB.ExecContext(ctx,
		`DELETE FROM favorites WHERE user_id = $1 AND device_id = $2`,
		userID, deviceID,
	)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("not found")
	}
	return nil
}

// List a user's favorites (returns full device info)
func (r *FavoriteRepository) GetFavorites(ctx context.Context, userID string) ([]model.Device, error) {
	var devices []model.Device
	query := `
      SELECT d.*
      FROM devices d
      JOIN favorites f ON f.device_id = d.id
      WHERE f.user_id = $1
    `
	err := r.DB.SelectContext(ctx, &devices, query, userID)
	return devices, err
}

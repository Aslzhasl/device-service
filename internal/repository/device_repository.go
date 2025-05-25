package repository

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"device-service/config"
	"device-service/internal/model"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/jmoiron/sqlx"
	"time"
)

type DeviceRepository struct {
	DB *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) *DeviceRepository {
	return &DeviceRepository{DB: db}
}

func (r *DeviceRepository) CreateDevice(ctx context.Context, d *model.Device) error {
	query := `
    INSERT INTO devices (name, description, category, price_per_day, available, image_url, owner_id, city, region)
    VALUES (:name, :description, :category, :price_per_day, :available, :image_url, :owner_id, :city, :region)
    RETURNING id, created_at, updated_at
    `
	stmt, err := r.DB.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	return stmt.GetContext(ctx, d, d)
}

func (r *DeviceRepository) GetAllDevices(ctx context.Context, f model.DeviceFilter) ([]model.Device, error) {
	// 1. Build a stable Redis key from the filter
	filterJSON, _ := json.Marshal(f)
	cacheKey := fmt.Sprintf("devices:%x", sha1.Sum(filterJSON))

	// 2. Try cache
	if cached, err := config.RedisClient.Get(ctx, cacheKey).Result(); err == nil {
		var devices []model.Device
		if err := json.Unmarshal([]byte(cached), &devices); err == nil {
			return devices, nil
		}
	}

	// 3. Build dynamic SQL
	baseQuery := `SELECT * FROM devices WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if f.Category != "" {
		baseQuery += fmt.Sprintf(" AND category = $%d", idx)
		args = append(args, f.Category)
		idx++
	}
	if f.Available != nil {
		baseQuery += fmt.Sprintf(" AND available = $%d", idx)
		args = append(args, *f.Available)
		idx++
	}
	if f.MinPrice != nil {
		baseQuery += fmt.Sprintf(" AND price_per_day >= $%d", idx)
		args = append(args, *f.MinPrice)
		idx++
	}
	if f.MaxPrice != nil {
		baseQuery += fmt.Sprintf(" AND price_per_day <= $%d", idx)
		args = append(args, *f.MaxPrice)
		idx++
	}
	if f.City != "" {
		baseQuery += fmt.Sprintf(" AND city ILIKE $%d", idx)
		args = append(args, f.City)
		idx++
	}
	if f.Region != "" {
		baseQuery += fmt.Sprintf(" AND region ILIKE $%d", idx)
		args = append(args, f.Region)
		idx++
	}

	switch f.Sort {
	case "price_asc":
		baseQuery += " ORDER BY price_per_day ASC"
	case "price_desc":
		baseQuery += " ORDER BY price_per_day DESC"
	default:
		baseQuery += " ORDER BY created_at DESC"
	}

	offset := (f.Page - 1) * f.Limit
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, f.Limit, offset)

	// 4. Execute the query
	var devices []model.Device
	if err := r.DB.SelectContext(ctx, &devices, baseQuery, args...); err != nil {
		return nil, err
	}

	// 5. Cache the result for 60s
	if payload, err := json.Marshal(devices); err == nil {
		_ = config.RedisClient.Set(ctx, cacheKey, payload, 60*time.Second).Err()
	}

	return devices, nil
}
func (r *DeviceRepository) GetDeviceByID(ctx context.Context, id string) (*model.Device, error) {
	var device model.Device
	err := r.DB.GetContext(ctx, &device, "SELECT * FROM devices WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &device, nil
}
func (r *DeviceRepository) UpdateDevice(ctx context.Context, device *model.Device) error {
	query := `
        UPDATE devices 
        SET name = :name, description = :description, category = :category,
            price_per_day = :price_per_day, available = :available, image_url = :image_url,
            updated_at = NOW()
        WHERE id = :id AND owner_id = :owner_id
    `
	result, err := r.DB.NamedExecContext(ctx, query, device)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // not found or not owner
	}

	return nil
}
func (r *DeviceRepository) DeleteDevice(ctx context.Context, deviceID string, ownerID string) error {
	query := `DELETE FROM devices WHERE id = $1 AND owner_id = $2`
	result, err := r.DB.ExecContext(ctx, query, deviceID, ownerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (r *DeviceRepository) UpdateAvailability(ctx context.Context, deviceID, ownerID string, available bool) error {
	query := `UPDATE devices SET available = $1, updated_at = NOW() WHERE id = $2 AND owner_id = $3`
	result, err := r.DB.ExecContext(ctx, query, available, deviceID, ownerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetCategories returns all distinct categories
func (r *DeviceRepository) GetCategories(ctx context.Context) ([]string, error) {
	var cats []string
	err := r.DB.SelectContext(ctx, &cats,
		`SELECT DISTINCT category FROM devices WHERE category <> '' ORDER BY category`)
	return cats, err
}

// GetCities returns all distinct non-null cities
func (r *DeviceRepository) GetCities(ctx context.Context) ([]string, error) {
	var cities []string
	err := r.DB.SelectContext(ctx, &cities,
		`SELECT DISTINCT city FROM devices WHERE city IS NOT NULL AND city <> '' ORDER BY city`)
	return cities, err
}

// GetRegions returns all distinct non-null regions
func (r *DeviceRepository) GetRegions(ctx context.Context) ([]string, error) {
	var regions []string
	err := r.DB.SelectContext(ctx, &regions,
		`SELECT DISTINCT region FROM devices WHERE region IS NOT NULL AND region <> '' ORDER BY region`)
	return regions, err
}

// GetTrendingDevices returns the top N devices by # of favorites
func (r *DeviceRepository) GetTrendingDevices(ctx context.Context, limit int) ([]model.Device, error) {
	var devices []model.Device
	query := `
      SELECT d.*
      FROM devices d
      LEFT JOIN favorites f ON f.device_id = d.id
      GROUP BY d.id
      ORDER BY COUNT(f.device_id) DESC
      LIMIT $1
    `
	err := r.DB.SelectContext(ctx, &devices, query, limit)
	return devices, err
}

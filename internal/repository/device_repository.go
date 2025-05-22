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

func (r *DeviceRepository) GetAllDevices(ctx context.Context, filter map[string]interface{}) ([]model.Device, error) {
	// 1. Redis Cache Key Generation
	filterJSON, _ := json.Marshal(filter)
	cacheKey := fmt.Sprintf("devices:%x", sha1.Sum(filterJSON))

	// 2. Check Redis Cache
	cached, err := config.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedDevices []model.Device
		if err := json.Unmarshal([]byte(cached), &cachedDevices); err == nil {
			return cachedDevices, nil
		}
	}

	baseQuery := `SELECT * FROM devices WHERE 1=1`
	args := []interface{}{}
	idx := 1

	// флтр
	if category, ok := filter["category"].(string); ok && category != "" {
		baseQuery += fmt.Sprintf(" AND category = $%d", idx)
		args = append(args, category)
		idx++
	}
	if available, ok := filter["available"].(*bool); ok && available != nil {
		baseQuery += fmt.Sprintf(" AND available = $%d", idx)
		args = append(args, *available)
		idx++
	}
	if minPrice, ok := filter["min_price"].(float64); ok {
		baseQuery += fmt.Sprintf(" AND price_per_day >= $%d", idx)
		args = append(args, minPrice)
		idx++
	}
	if maxPrice, ok := filter["max_price"].(float64); ok {
		baseQuery += fmt.Sprintf(" AND price_per_day <= $%d", idx)
		args = append(args, maxPrice)
		idx++
	}
	if city, ok := filter["city"].(string); ok && city != "" {
		baseQuery += fmt.Sprintf(" AND city ILIKE $%d", idx)
		args = append(args, city)
		idx++
	}
	if region, ok := filter["region"].(string); ok && region != "" {
		baseQuery += fmt.Sprintf(" AND region ILIKE $%d", idx)
		args = append(args, region)
		idx++
	}

	//сорт
	sort := filter["sort"].(string)
	switch sort {
	case "price_asc":
		baseQuery += " ORDER BY price_per_day ASC"
	case "price_desc":
		baseQuery += " ORDER BY price_per_day DESC"
	default:
		baseQuery += " ORDER BY created_at DESC"
	}

	// пагинация
	limit := filter["limit"].(int)
	page := filter["page"].(int)
	offset := (page - 1) * limit
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)
	fmt.Println("🔍 SQL:", baseQuery)
	fmt.Println("📦 Args:", args)
	var devices []model.Device
	err = r.DB.SelectContext(ctx, &devices, baseQuery, args...)
	if err != nil {
		return nil, err
	}

	// сохранение в редис
	payload, _ := json.Marshal(devices)
	_ = config.RedisClient.Set(ctx, cacheKey, payload, 60*time.Second).Err()

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

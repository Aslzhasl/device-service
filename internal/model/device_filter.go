package model

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

type DeviceFilter struct {
	Category  string
	Available *bool
	MinPrice  *float64
	MaxPrice  *float64
	City      string
	Region    string
	Sort      string
	Page      int
	Limit     int
}

func ParseDeviceFilter(c *gin.Context) DeviceFilter {
	f := DeviceFilter{
		Category: c.Query("category"),
		City:     c.Query("city"),
		Region:   c.Query("region"),
		Sort:     c.DefaultQuery("sort", "recent"),
		Page:     func() int { v, _ := strconv.Atoi(c.DefaultQuery("page", "1")); return v }(),
		Limit:    func() int { v, _ := strconv.Atoi(c.DefaultQuery("limit", "10")); return v }(),
	}

	if s := c.Query("available"); s != "" {
		v := s == "true"
		f.Available = &v
	}
	if s := c.Query("min_price"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			f.MinPrice = &v
		}
	}
	if s := c.Query("max_price"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			f.MaxPrice = &v
		}
	}
	return f
}

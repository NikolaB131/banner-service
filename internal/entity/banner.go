package entity

import "time"

type Banner struct {
	ID        int            `db:"id" json:"banner_id"`
	TagIDs    []int          `db:"tag_ids" json:"tag_ids"`
	FeatureID int            `db:"feature_id" json:"feature_id"`
	Content   map[string]any `db:"content" json:"content"`
	IsActive  bool           `db:"is_active" json:"is_active"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}

package sticker

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrInvalidInput = errors.New("invalid input")

// Pack is a catalog of related stickers. ImageUUID references an object in chickchirick-files.
type Pack struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Slug        string    `json:"slug" gorm:"size:80;not null;uniqueIndex"`
	Title       string    `json:"title" gorm:"size:120;not null"`
	Description string    `json:"description" gorm:"size:500;not null;default:''"`
	ImageUUID   string    `json:"image_uuid,omitempty" gorm:"size:36"`
	IsPublished bool      `json:"is_published" gorm:"not null;default:false;index"`
	SortOrder   int       `json:"sort_order" gorm:"not null;default:0;index"`
	Stickers    []Sticker `json:"stickers,omitempty" gorm:"foreignKey:PackID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Pack) TableName() string { return "sticker_packs" }

// ListPublished returns all published packs with stickers ordered by sort_order.
func ListPublishedPacks(ctx context.Context, db *gorm.DB) ([]Pack, error) {
	var packs []Pack
	err := db.WithContext(ctx).
		Model(&Pack{}).
		Where("is_published = ?", true).
		Order("sort_order ASC, id ASC").
		Preload("Stickers", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Find(&packs).Error
	return packs, err
}

// GetPublishedPack returns a single published pack by ID with stickers.
func GetPublishedPack(ctx context.Context, db *gorm.DB, id string) (Pack, error) {
	var pack Pack
	err := db.WithContext(ctx).
		Where("is_published = ?", true).
		Preload("Stickers", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		First(&pack, "id = ?", id).Error
	return pack, err
}

// GetPackByID returns a pack by ID regardless of publish status (for admin).
func GetPackByID(ctx context.Context, db *gorm.DB, id string) (Pack, error) {
	var pack Pack
	err := db.WithContext(ctx).First(&pack, "id = ?", id).Error
	return pack, err
}

// CreatePackInput holds data for creating a pack.
type CreatePackInput struct {
	Slug        string
	Title       string
	Description string
	ImageUUID   string
	IsPublished bool
	SortOrder   int
}

// Create creates a new pack.
func CreatePack(ctx context.Context, db *gorm.DB, in CreatePackInput) (Pack, error) {
	pack := Pack{
		Slug:        strings.TrimSpace(in.Slug),
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		ImageUUID:   in.ImageUUID,
		IsPublished: in.IsPublished,
		SortOrder:   in.SortOrder,
	}
	if pack.Slug == "" || pack.Title == "" {
		return Pack{}, ErrInvalidInput
	}
	err := db.WithContext(ctx).Create(&pack).Error
	return pack, err
}

// UpdatePackInput holds optional fields for updating a pack.
type UpdatePackInput struct {
	Title       *string
	Description *string
	ImageUUID   *string
	IsPublished *bool
	SortOrder   *int
}

// Update applies partial updates to a pack and returns the refreshed entity with stickers.
func (p *Pack) Update(ctx context.Context, db *gorm.DB, in UpdatePackInput) error {
	updates := map[string]interface{}{}
	if in.Title != nil {
		updates["title"] = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.ImageUUID != nil {
		updates["image_uuid"] = *in.ImageUUID
	}
	if in.IsPublished != nil {
		updates["is_published"] = *in.IsPublished
	}
	if in.SortOrder != nil {
		updates["sort_order"] = *in.SortOrder
	}
	if len(updates) == 0 {
		return nil
	}
	if err := db.WithContext(ctx).Model(p).Updates(updates).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Preload("Stickers").First(p, p.ID).Error
}

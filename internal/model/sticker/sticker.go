package sticker

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Sticker struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PackID    uint      `json:"pack_id" gorm:"not null;index;uniqueIndex:idx_pack_sticker_slug,priority:1;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Slug      string    `json:"slug" gorm:"size:80;not null;uniqueIndex:idx_pack_sticker_slug,priority:2"`
	Title     string    `json:"title" gorm:"size:120;not null;default:''"`
	ImageUUID string    `json:"image_uuid" gorm:"size:36;not null"`
	Emoji     string    `json:"emoji,omitempty" gorm:"size:16"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Sticker) TableName() string { return "stickers" }

// ListStickersByPack returns stickers of a published pack.
func ListStickersByPack(ctx context.Context, db *gorm.DB, packID string) ([]Sticker, error) {
	var pack Pack
	if err := db.WithContext(ctx).Select("id").Where("is_published = ?", true).First(&pack, "id = ?", packID).Error; err != nil {
		return nil, err
	}
	var stickers []Sticker
	err := db.WithContext(ctx).
		Where("pack_id = ?", pack.ID).
		Order("sort_order ASC, id ASC").
		Find(&stickers).Error
	return stickers, err
}

// GetPublishedSticker returns a sticker that belongs to a published pack.
func GetPublishedSticker(ctx context.Context, db *gorm.DB, id string) (Sticker, error) {
	var item Sticker
	err := db.WithContext(ctx).
		Joins("JOIN sticker_packs ON sticker_packs.id = stickers.pack_id").
		Where("stickers.id = ? AND sticker_packs.is_published = ?", id, true).
		First(&item).Error
	return item, err
}

// GetStickerByID returns a sticker by ID (admin).
func GetStickerByID(ctx context.Context, db *gorm.DB, id string) (Sticker, error) {
	var item Sticker
	err := db.WithContext(ctx).First(&item, "id = ?", id).Error
	return item, err
}

// CreateStickerInput holds data for creating a sticker.
type CreateStickerInput struct {
	PackID    uint
	Slug      string
	Title     string
	ImageUUID string
	Emoji     string
	SortOrder int
}

// CreateSticker creates a new sticker in the given pack.
func CreateSticker(ctx context.Context, db *gorm.DB, in CreateStickerInput) (Sticker, error) {
	item := Sticker{
		PackID:    in.PackID,
		Slug:      strings.TrimSpace(in.Slug),
		Title:     in.Title,
		ImageUUID: strings.TrimSpace(in.ImageUUID),
		Emoji:     in.Emoji,
		SortOrder: in.SortOrder,
	}
	if item.Slug == "" || item.ImageUUID == "" {
		return Sticker{}, ErrInvalidInput
	}
	err := db.WithContext(ctx).Create(&item).Error
	return item, err
}

// UpdateStickerInput holds optional fields for updating a sticker.
type UpdateStickerInput struct {
	Title     *string
	ImageUUID *string
	Emoji     *string
	SortOrder *int
}

// Update applies partial updates to a sticker and returns the refreshed entity.
func (s *Sticker) Update(ctx context.Context, db *gorm.DB, in UpdateStickerInput) error {
	updates := map[string]interface{}{}
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.ImageUUID != nil {
		updates["image_uuid"] = strings.TrimSpace(*in.ImageUUID)
	}
	if in.Emoji != nil {
		updates["emoji"] = *in.Emoji
	}
	if in.SortOrder != nil {
		updates["sort_order"] = *in.SortOrder
	}
	if len(updates) == 0 {
		return nil
	}
	if err := db.WithContext(ctx).Model(s).Updates(updates).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).First(s, s.ID).Error
}

// DeleteSticker deletes a sticker by ID. Returns rows affected.
func DeleteSticker(ctx context.Context, db *gorm.DB, id string) (int64, error) {
	result := db.WithContext(ctx).Delete(&Sticker{}, "id = ?", id)
	return result.RowsAffected, result.Error
}

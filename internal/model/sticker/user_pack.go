package sticker

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// UserPack stores a user's installed packs. UserUUID is the stable ID returned by auth service.
type UserPack struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserUUID  string    `json:"user_uuid" gorm:"size:64;not null;uniqueIndex:idx_user_pack"`
	PackID    uint      `json:"pack_id" gorm:"not null;uniqueIndex:idx_user_pack;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at"`
	Pack      Pack      `json:"pack,omitempty" gorm:"foreignKey:PackID"`
}

func (UserPack) TableName() string { return "user_sticker_packs" }

// ListInstalledPacks returns published packs installed by the user.
func ListInstalledPacks(ctx context.Context, db *gorm.DB, userUUID string) ([]Pack, error) {
	var rows []UserPack
	err := db.WithContext(ctx).
		Preload("Pack", "is_published = ?", true).
		Where("user_uuid = ?", userUUID).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	packs := make([]Pack, 0, len(rows))
	for _, row := range rows {
		if row.Pack.ID != 0 {
			packs = append(packs, row.Pack)
		}
	}
	return packs, nil
}

// InstallPack installs a published pack for the user (idempotent).
func InstallPack(ctx context.Context, db *gorm.DB, userUUID, packID string) (uint, error) {
	var pack Pack
	if err := db.WithContext(ctx).Where("is_published = ?", true).First(&pack, "id = ?", packID).Error; err != nil {
		return 0, err
	}
	row := UserPack{UserUUID: userUUID, PackID: pack.ID}
	if err := db.WithContext(ctx).Where("user_uuid = ? AND pack_id = ?", row.UserUUID, row.PackID).FirstOrCreate(&row).Error; err != nil {
		return 0, err
	}
	return pack.ID, nil
}

// UninstallPack removes a pack from the user's installed list.
func UninstallPack(ctx context.Context, db *gorm.DB, userUUID, packID string) error {
	return db.WithContext(ctx).
		Where("user_uuid = ? AND pack_id = ?", userUUID, packID).
		Delete(&UserPack{}).Error
}

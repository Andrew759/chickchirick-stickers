package sticker

import "time"

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

// UserPack stores a user's installed packs. UserUUID is the stable ID returned by auth service.
type UserPack struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserUUID  string    `json:"user_uuid" gorm:"size:64;not null;uniqueIndex:idx_user_pack"`
	PackID    uint      `json:"pack_id" gorm:"not null;uniqueIndex:idx_user_pack;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at"`
	Pack      Pack      `json:"pack,omitempty" gorm:"foreignKey:PackID"`
}

func (Pack) TableName() string     { return "sticker_packs" }
func (Sticker) TableName() string  { return "stickers" }
func (UserPack) TableName() string { return "user_sticker_packs" }

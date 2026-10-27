package sticker

import (
	"errors"
	"net/http"
	"strings"

	"chickchirick-stickers/internal/controller/c_controller"
	"chickchirick-stickers/internal/middleware/auth"
	model "chickchirick-stickers/internal/model/sticker"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct{ c_controller.Controller }

func (sc *Controller) RegisterRoutes() {
	api := sc.E.Group("/")
	api.GET("health", sc.Health)
	api.GET("sticker-packs", sc.ListPacks)
	api.GET("sticker-packs/:id", sc.GetPack)
	api.GET("sticker-packs/:id/stickers", sc.ListPackStickers)
	api.GET("stickers/:id", sc.GetSticker)

	user := sc.E.Group("/")
	user.Use(auth.RequireUser(sc.DI.AuthURL, sc.DI.HTTPClient))
	user.GET("me/sticker-packs", sc.ListInstalledPacks)
	user.POST("sticker-packs/:id/install", sc.InstallPack)
	user.DELETE("sticker-packs/:id/install", sc.UninstallPack)

	admin := sc.E.Group("/admin")
	admin.Use(sc.RequireAdmin())
	admin.POST("/sticker-packs", sc.CreatePack)
	admin.PATCH("/sticker-packs/:id", sc.UpdatePack)
	admin.POST("/sticker-packs/:id/stickers", sc.CreateSticker)
	admin.PATCH("/stickers/:id", sc.UpdateSticker)
	admin.DELETE("/stickers/:id", sc.DeleteSticker)
}

func (sc *Controller) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "stickers"})
}

func (sc *Controller) ListPacks(c *gin.Context) {
	var packs []model.Pack
	q := sc.DI.DB.WithContext(c.Request.Context()).Model(&model.Pack{}).Where("is_published = ?", true).Order("sort_order ASC, id ASC")
	if err := q.Preload("Stickers", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).Find(&packs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sticker packs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": packs})
}

func (sc *Controller) GetPack(c *gin.Context) {
	var pack model.Pack
	err := sc.DI.DB.WithContext(c.Request.Context()).Where("is_published = ?", true).Preload("Stickers", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).First(&pack, "id = ?", c.Param("id")).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sticker pack"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": pack})
}

func (sc *Controller) ListPackStickers(c *gin.Context) {
	var pack model.Pack
	err := sc.DI.DB.WithContext(c.Request.Context()).Select("id").Where("is_published = ?", true).First(&pack, "id = ?", c.Param("id")).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sticker pack"})
		return
	}
	var stickers []model.Sticker
	if err := sc.DI.DB.WithContext(c.Request.Context()).Where("pack_id = ?", pack.ID).Order("sort_order ASC, id ASC").Find(&stickers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list stickers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": stickers})
}

func (sc *Controller) GetSticker(c *gin.Context) {
	var item model.Sticker
	err := sc.DI.DB.WithContext(c.Request.Context()).Joins("JOIN sticker_packs ON sticker_packs.id = stickers.pack_id").Where("stickers.id = ? AND sticker_packs.is_published = ?", c.Param("id"), true).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sticker"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": item})
}

func (sc *Controller) ListInstalledPacks(c *gin.Context) {
	userUUID, _ := c.Get(auth.UserUUIDKey)
	var rows []model.UserPack
	err := sc.DI.DB.WithContext(c.Request.Context()).Preload("Pack", "is_published = ?", true).Where("user_uuid = ?", userUUID).Order("created_at DESC").Find(&rows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list installed packs"})
		return
	}
	packs := make([]model.Pack, 0, len(rows))
	for _, row := range rows {
		if row.Pack.ID != 0 {
			packs = append(packs, row.Pack)
		}
	}
	c.JSON(http.StatusOK, gin.H{"payload": packs})
}

func (sc *Controller) InstallPack(c *gin.Context) {
	userUUID, _ := c.Get(auth.UserUUIDKey)
	var pack model.Pack
	err := sc.DI.DB.WithContext(c.Request.Context()).Where("is_published = ?", true).First(&pack, "id = ?", c.Param("id")).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find sticker pack"})
		return
	}
	row := model.UserPack{UserUUID: userUUID.(string), PackID: pack.ID}
	if err := sc.DI.DB.WithContext(c.Request.Context()).Where("user_uuid = ? AND pack_id = ?", row.UserUUID, row.PackID).FirstOrCreate(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to install sticker pack"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": gin.H{"installed": true, "pack_id": pack.ID}})
}

func (sc *Controller) UninstallPack(c *gin.Context) {
	userUUID, _ := c.Get(auth.UserUUIDKey)
	result := sc.DI.DB.WithContext(c.Request.Context()).Where("user_uuid = ? AND pack_id = ?", userUUID, c.Param("id")).Delete(&model.UserPack{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to uninstall sticker pack"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": gin.H{"installed": false}})
}

func (sc *Controller) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		configured := strings.TrimSpace(sc.DI.AdminToken)
		provided := c.GetHeader("X-Admin-Token")
		if configured == "" || provided == "" || provided != configured {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access denied"})
			return
		}
		c.Next()
	}
}

type createPackRequest struct {
	Slug        string `json:"slug" binding:"required,max=80"`
	Title       string `json:"title" binding:"required,max=120"`
	Description string `json:"description"`
	ImageUUID   string `json:"image_uuid"`
	IsPublished bool   `json:"is_published"`
	SortOrder   int    `json:"sort_order"`
}
type updatePackRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ImageUUID   *string `json:"image_uuid"`
	IsPublished *bool   `json:"is_published"`
	SortOrder   *int    `json:"sort_order"`
}
type createStickerRequest struct {
	Slug      string `json:"slug" binding:"required,max=80"`
	Title     string `json:"title"`
	ImageUUID string `json:"image_uuid" binding:"required"`
	Emoji     string `json:"emoji"`
	SortOrder int    `json:"sort_order"`
}
type updateStickerRequest struct {
	Title     *string `json:"title"`
	ImageUUID *string `json:"image_uuid"`
	Emoji     *string `json:"emoji"`
	SortOrder *int    `json:"sort_order"`
}

func (sc *Controller) CreatePack(c *gin.Context) {
	var req createPackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	pack := model.Pack{Slug: strings.TrimSpace(req.Slug), Title: strings.TrimSpace(req.Title), Description: req.Description, ImageUUID: req.ImageUUID, IsPublished: req.IsPublished, SortOrder: req.SortOrder}
	if pack.Slug == "" || pack.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err := sc.DI.DB.WithContext(c.Request.Context()).Create(&pack).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "failed to create sticker pack; slug may already exist"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"payload": pack})
}

func (sc *Controller) UpdatePack(c *gin.Context) {
	var req updatePackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	var pack model.Pack
	if err := sc.DI.DB.WithContext(c.Request.Context()).First(&pack, "id = ?", c.Param("id")).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sticker pack"})
		return
	}
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ImageUUID != nil {
		updates["image_uuid"] = *req.ImageUUID
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if err := sc.DI.DB.WithContext(c.Request.Context()).Model(&pack).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sticker pack"})
		return
	}
	sc.DI.DB.WithContext(c.Request.Context()).Preload("Stickers").First(&pack, pack.ID)
	c.JSON(http.StatusOK, gin.H{"payload": pack})
}

func (sc *Controller) CreateSticker(c *gin.Context) {
	var req createStickerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	var pack model.Pack
	if err := sc.DI.DB.WithContext(c.Request.Context()).First(&pack, "id = ?", c.Param("id")).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sticker pack"})
		return
	}
	item := model.Sticker{PackID: pack.ID, Slug: strings.TrimSpace(req.Slug), Title: req.Title, ImageUUID: strings.TrimSpace(req.ImageUUID), Emoji: req.Emoji, SortOrder: req.SortOrder}
	if item.Slug == "" || item.ImageUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and image_uuid are required"})
		return
	}
	if err := sc.DI.DB.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "failed to create sticker; slug may already exist in this pack"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"payload": item})
}

func (sc *Controller) UpdateSticker(c *gin.Context) {
	var req updateStickerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	var item model.Sticker
	if err := sc.DI.DB.WithContext(c.Request.Context()).First(&item, "id = ?", c.Param("id")).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sticker"})
		return
	}
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.ImageUUID != nil {
		updates["image_uuid"] = strings.TrimSpace(*req.ImageUUID)
	}
	if req.Emoji != nil {
		updates["emoji"] = *req.Emoji
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if err := sc.DI.DB.WithContext(c.Request.Context()).Model(&item).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sticker"})
		return
	}
	sc.DI.DB.WithContext(c.Request.Context()).First(&item, item.ID)
	c.JSON(http.StatusOK, gin.H{"payload": item})
}

func (sc *Controller) DeleteSticker(c *gin.Context) {
	result := sc.DI.DB.WithContext(c.Request.Context()).Delete(&model.Sticker{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete sticker"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

package sticker

import (
	"errors"
	"net/http"
	"strings"

	"chickchirick-stickers/internal/controller/c_controller"
	"chickchirick-stickers/internal/middleware/auth"
	"chickchirick-stickers/internal/model/sticker"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct{ c_controller.Controller }

func (sc *Controller) RegisterRoutes() {
	api := sc.E.Group("/")
	api.GET("sticker-packs", sc.ListPacks)
	api.GET("sticker-packs/:id", sc.GetPack)
	api.GET("sticker-packs/:id/stickers", sc.ListPackStickers)
	api.GET("stickers/:id", sc.GetSticker)

	user := sc.E.Group("/")
	user.Use(auth.RequireUser(sc.DI.AuthURL, sc.DI.HTTPClient))
	user.GET("me/sticker-packs", sc.ListInstalledPacks)
	user.POST("sticker-packs/:id/install", sc.InstallPack)
	user.DELETE("sticker-packs/:id/install", sc.UninstallPack)
	// Создание наборов и стикеров — любой авторизованный пользователь
	user.POST("sticker-packs", sc.CreatePack)
	user.POST("sticker-packs/:id/stickers", sc.CreateSticker)

	admin := sc.E.Group("/admin")
	admin.Use(sc.RequireAdmin())
	admin.PATCH("/sticker-packs/:id", sc.UpdatePack)
	admin.PATCH("/stickers/:id", sc.UpdateSticker)
	admin.DELETE("/stickers/:id", sc.DeleteSticker)
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

func (sc *Controller) ListPacks(c *gin.Context) {
	packs, err := sticker.ListPublishedPacks(c.Request.Context(), sc.DI.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sticker packs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": packs})
}

func (sc *Controller) GetPack(c *gin.Context) {
	pack, err := sticker.GetPublishedPack(c.Request.Context(), sc.DI.DB, c.Param("id"))
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
	stickers, err := sticker.ListStickersByPack(c.Request.Context(), sc.DI.DB, c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list stickers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": stickers})
}

func (sc *Controller) GetSticker(c *gin.Context) {
	item, err := sticker.GetPublishedSticker(c.Request.Context(), sc.DI.DB, c.Param("id"))
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

// --- User ---

func (sc *Controller) ListInstalledPacks(c *gin.Context) {
	userUUID, _ := c.Get(auth.UserUUIDKey)
	packs, err := sticker.ListInstalledPacks(c.Request.Context(), sc.DI.DB, userUUID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list installed packs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": packs})
}

func (sc *Controller) InstallPack(c *gin.Context) {
	userUUID, _ := c.Get(auth.UserUUIDKey)
	packID, err := sticker.InstallPack(c.Request.Context(), sc.DI.DB, userUUID.(string), c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to install sticker pack"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": gin.H{"installed": true, "pack_id": packID}})
}

func (sc *Controller) UninstallPack(c *gin.Context) {
	userUUID, _ := c.Get(auth.UserUUIDKey)
	if err := sticker.UninstallPack(c.Request.Context(), sc.DI.DB, userUUID.(string), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to uninstall sticker pack"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": gin.H{"installed": false}})
}

// --- Admin request DTOs ---

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

// --- Admin handlers ---

func (sc *Controller) CreatePack(c *gin.Context) {
	var req createPackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	pack, err := sticker.CreatePack(c.Request.Context(), sc.DI.DB, sticker.CreatePackInput{
		Slug:        req.Slug,
		Title:       req.Title,
		Description: req.Description,
		ImageUUID:   req.ImageUUID,
		IsPublished: req.IsPublished,
		SortOrder:   req.SortOrder,
	})
	if errors.Is(err, sticker.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err != nil {
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
	pack, err := sticker.GetPackByID(c.Request.Context(), sc.DI.DB, c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sticker pack"})
		return
	}
	if err := pack.Update(c.Request.Context(), sc.DI.DB, sticker.UpdatePackInput{
		Title:       req.Title,
		Description: req.Description,
		ImageUUID:   req.ImageUUID,
		IsPublished: req.IsPublished,
		SortOrder:   req.SortOrder,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sticker pack"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": pack})
}

func (sc *Controller) CreateSticker(c *gin.Context) {
	var req createStickerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	pack, err := sticker.GetPackByID(c.Request.Context(), sc.DI.DB, c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sticker pack"})
		return
	}
	item, err := sticker.CreateSticker(c.Request.Context(), sc.DI.DB, sticker.CreateStickerInput{
		PackID:    pack.ID,
		Slug:      req.Slug,
		Title:     req.Title,
		ImageUUID: req.ImageUUID,
		Emoji:     req.Emoji,
		SortOrder: req.SortOrder,
	})
	if errors.Is(err, sticker.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and image_uuid are required"})
		return
	}
	if err != nil {
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
	item, err := sticker.GetStickerByID(c.Request.Context(), sc.DI.DB, c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sticker"})
		return
	}
	if err := item.Update(c.Request.Context(), sc.DI.DB, sticker.UpdateStickerInput{
		Title:     req.Title,
		ImageUUID: req.ImageUUID,
		Emoji:     req.Emoji,
		SortOrder: req.SortOrder,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sticker"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payload": item})
}

func (sc *Controller) DeleteSticker(c *gin.Context) {
	rows, err := sticker.DeleteSticker(c.Request.Context(), sc.DI.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete sticker"})
		return
	}
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sticker not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

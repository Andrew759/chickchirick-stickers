package factory

import (
	"chickchirick-stickers/internal/controller/c_controller"
	"chickchirick-stickers/internal/controller/service/sticker"

	"github.com/gin-gonic/gin"
)

type StickerServer struct{}

func InitStickerServer(e *gin.Engine, mDIC *c_controller.DIContainer) {
	stickerServer := &StickerServer{}

	stickerServer.initStickerService(e, mDIC)
}

func (ss StickerServer) initStickerService(e *gin.Engine, aDIC *c_controller.DIContainer) {
	stickerService := sticker.Controller{
		E:  e,
		DI: aDIC,
	}

	stickerService.RegisterRoutes()
}

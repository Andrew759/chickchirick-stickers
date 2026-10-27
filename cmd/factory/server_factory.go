package factory

import (
	"chickchirick-stickers/cmd/service"
	"chickchirick-stickers/internal/controller/c_controller"
	"chickchirick-stickers/internal/model/sticker"
	"chickchirick-stickers/pkg/chirick_config"
	"net/http"

	_ "net/http/pprof"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func BuildAndServe(
	dbDecorator *service.DBDecorator,
	httpClient *http.Client,
) {
	err := BuildServer(dbDecorator, httpClient)
	if err != nil {
		//TODO: добавить логирование
		panic(err)
	}
}

func BuildServer(dbDecorator *service.DBDecorator, httpClient *http.Client) error {
	e := gin.Default()

	config := cors.Config{
		AllowCredentials: true,

		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},

		AllowOrigins: []string{
			viper.GetString(chirick_config.UserApp),
			viper.GetString(chirick_config.AuthApp),
			viper.GetString(chirick_config.FrontendApp),
			viper.GetString(chirick_config.MessagesApp),
		},
	}

	InitStickerServer(e, &c_controller.DIContainer{
		DB:         dbDecorator.Gorm,
		HTTPClient: httpClient,
		AuthURL:    viper.GetString(chirick_config.AuthApp),
		AdminToken: viper.GetString(chirick_config.AdminToken),
	})
	e.Use(cors.New(config))

	err := e.Run()
	if err != nil {
		return err
	}

	return nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&sticker.Pack{}, &sticker.Sticker{}, &sticker.UserPack{})
}

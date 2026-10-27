package main

import (
	"chickchirick-stickers/cmd/config"
	"log"

	"chickchirick-stickers/cmd/factory"
	"chickchirick-stickers/cmd/service"
)

func main() {
	factory.InitViper()

	appConfig := config.AppConfiguration{}.NewAppConfiguration()
	dbDecorator := service.InitORM(&appConfig.DatabaseConfig)
	defer dbDecorator.CloseDB()

	//TODO: удалить
	if err := factory.AutoMigrate(dbDecorator.GDB()); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	httpClient := factory.InitHttpClient()

	factory.BuildAndServe(dbDecorator, httpClient)
}

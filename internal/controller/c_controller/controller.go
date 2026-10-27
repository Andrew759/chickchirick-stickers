package c_controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DIContainer struct {
	DB         *gorm.DB
	HTTPClient *http.Client
	AuthURL    string
	AdminToken string
}

type Controller struct {
	E  *gin.Engine
	DI *DIContainer
}

type RequestHandler interface{ RegisterRoutes() }

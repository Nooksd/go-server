package main

import (
	"os"

	routes "github.com/Nooksd/go-server/src/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.AvatarRoutes(router)
	routes.PostRoutes(router)
	routes.UserRoutes(router)

	router.Run(":" + port)
}

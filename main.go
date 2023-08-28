package main

import (
	controllers "backdev_test_task/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.GET("/create-token/:user_id", controllers.CreateToken())
	router.GET("/refresh/:refresh_token", controllers.RefreshToken())
	router.Run(":8080")
}

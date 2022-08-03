package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	api := router.Group("/v1")
	{
		api.POST("/guestbook/:website", POST_Guestbook)
		api.GET("/guestbook/:website", GET_Guestbook)
		api.DELETE("/guestbook/:website/:id", DELETE_Guestbook)
	}

	router.Run("0.0.0.0:8181")
}

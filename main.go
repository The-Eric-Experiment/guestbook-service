package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	api := router.Group("/v1")
	{
		api.POST("/sites", POST_Site)
		api.GET("/sites", GET_Sites)
		api.POST("/guestbook/:website", POST_Guestbook)
		api.POST("/fillup/:website", POST_Fillup)
		api.GET("/guestbook/:website", GET_Guestbook)
		api.DELETE("/guestbook/:website/:id", DELETE_Guestbook)
	}

	router.Run("0.0.0.0:8181")
}

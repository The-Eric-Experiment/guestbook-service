package main

import (
	"math"
	"regexp"
	"strconv"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type GuestbookResponse struct {
	Pages         int64        `json:"pages"`
	Page          int          `json:"page"`
	TotalMessages int64        `json:"total"`
	Messages      []*Guestbook `json:"messages"`
}

func POST_Guestbook(c *gin.Context) {
	var requestBody GuestbookRequest
	website := c.Param("website")

	if err := c.BindJSON(&requestBody); err != nil {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "There was an error with the request.",
		})
		return
	}

	db := GetDB(website)
	defer db.Close()

	lastMessage := db.GetLastMessage(requestBody.UserAgent, requestBody.IP)

	if lastMessage != nil && time.Since(lastMessage.Created).Minutes() <= MESSAGE_INTERVAL_MIN {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "In order to prevent abuse, we're only allowing one message per hour.",
		})
		return
	}

	if requestBody.Message == "" {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "Please write a message before posting.",
		})
		return
	}

	if len(requestBody.Message) >= MAXIMUM_CHARACTERS {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "Please remain within " + strconv.Itoa(MAXIMUM_CHARACTERS) + " characters in your message.",
		})
		return
	}

	if requestBody.Nickname == "" {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "Please insert a nickname before posting.",
		})
		return
	}

	if goaway.IsProfane(requestBody.Message) {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "Please be nice, let's keep the guestbook friendly and clean.",
		})
		return
	}

	if match, _ := regexp.MatchString(`https?:\/\/([^\.]+\.)?[^\.]+\.[^\/\.]{2,4}[\/]?`, requestBody.Message); match == true {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "URLs are not allowed in the Guestbook.",
		})
		return
	}

	db.AddToGuestbook(requestBody)

	c.Status(200)
}

func GET_Guestbook(c *gin.Context) {
	website := c.Param("website")
	qp, ok := c.GetQuery("page")

	if !ok {
		c.Status(400)
		return
	}

	page, err := strconv.Atoi(qp)

	if err != nil {
		c.Status(400)
		return
	}

	db := GetDB(website)
	defer db.Close()

	count := db.GetGuestbookCount()

	totalPages := int64(math.Ceil(float64(count) / float64(PAGE_SIZE)))
	offset := (page - 1) * PAGE_SIZE

	messages := db.GetGuestbook(offset, PAGE_SIZE)

	c.IndentedJSON(200, &GuestbookResponse{
		Pages:         totalPages,
		Page:          page,
		TotalMessages: count,
		Messages:      messages,
	})
}

func DELETE_Guestbook(c *gin.Context) {
	website := c.Param("website")
	id, ok := c.GetQuery("id")

	if !ok {
		c.Status(400)
		return
	}

	db := GetDB(website)
	defer db.Close()

	db.DeleteGuestbook(id)

	c.Status(200)
}

package main

import (
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
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

	if err := checkSiteExists(website); err != nil {
		c.IndentedJSON(400, &ErrorResponse{
			Message: err.Error(),
		})
		return
	}

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

func POST_Fillup(c *gin.Context) {
	var requestBody Guestbook
	website := c.Param("website")

	if err := checkSiteExists(website); err != nil {
		c.IndentedJSON(400, &ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.IndentedJSON(400, &ErrorResponse{
			Message: "There was an error with the request.",
		})
		return
	}

	db := GetDB(website)
	defer db.Close()

	db.AddRawToGuestbook(requestBody)

	c.Status(200)
}

func GET_Guestbook(c *gin.Context) {
	website := c.Param("website")

	if err := checkSiteExists(website); err != nil {
		c.IndentedJSON(400, &ErrorResponse{
			Message: err.Error(),
		})
		return
	}

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

	if err := checkSiteExists(website); err != nil {
		c.IndentedJSON(400, &ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	id := c.Param("id")

	if id == "" {
		c.Status(400)
		return
	}

	db := GetDB(website)
	defer db.Close()

	db.DeleteGuestbook(id)

	c.Status(200)
}

func POST_Site(c *gin.Context) {
	var newSite Site
	if err := c.ShouldBindJSON(&newSite); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid JSON provided",
		})
		return
	}

	var sites []Site
	_, err := os.Stat("data/sites.yaml")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Cannot stat file: %v ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Cannot stat file",
			})
			return
		}
		// If the file does not exist, it's OK - we'll create it later
	} else {
		// If the file exists, read the existing sites
		file, err := ioutil.ReadFile("data/sites.yaml")
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to read file",
			})
			return
		}

		err = yaml.Unmarshal(file, &sites)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to parse file",
			})
			return
		}
	}

	// Append the new site to the list and write back to the file
	sites = append(sites, newSite)
	file, err := yaml.Marshal(&sites)
	if err != nil {
		log.Fatalf("Marshal: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to marshal file",
		})
		return
	}
	err = ioutil.WriteFile("data/sites.yaml", file, 0644)
	if err != nil {
		log.Fatalf("WriteFile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to write to file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Site added successfully",
		"data":    newSite,
	})
}

func GET_Sites(c *gin.Context) {
	file, err := ioutil.ReadFile("data/sites.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to read file",
		})
		return
	}

	var sites []Site
	err = yaml.Unmarshal(file, &sites)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to parse file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   sites,
	})
}

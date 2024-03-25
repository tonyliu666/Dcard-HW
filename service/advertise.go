package service

import (
	"dcardapp/middleware"
	"dcardapp/model"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ADRequest struct {
	Title      string `json:"title"`
	StartAt    string `json:"startAt"`
	EndAt      string `json:"endAt"`
	Conditions struct {
		AgeStart int      `json:"ageStart"`
		AgeEnd   int      `json:"ageEnd"`
		Country  []string `json:"country"`
		Platform []string `json:"platform"`
	} `json:"conditions"`
}

func CreateADs(c *gin.Context) {
	// get the post request parameters
	// create a new ad in the database
	/*
		curl -X POST -H "Content-Type: application/json" \
		"http://<host>/api/v1/ad" \ --data '{
		"title" "AD 55",
		"startAt" "2023-12-10T03:00:00.000Z", "endAt" "2023-12-31T16:00:00.000Z", "conditions": {
		{
		"ageStart": 20,
		"ageEnd": 30,
		"country: ["TW", "JP"], "platform": ["android", "ios"]
		} }
		}
	*/

	var adRequest ADRequest
	if err := c.ShouldBindJSON(&adRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Info(adRequest)

	title := adRequest.Title
	startAtStr := adRequest.StartAt
	endAtStr := adRequest.EndAt
	conditions := adRequest.Conditions

	// convert startAt string to time.Time
	log.Info(startAtStr, endAtStr, endAtStr, conditions)
	// the time format is ISO 8601 format
	startAt, err := time.Parse(time.RFC3339, startAtStr)
	if err != nil {
		log.Error(err)
	}
	// convert endAt string to time.Time
	endAt, err := time.Parse(time.RFC3339, endAtStr)
	if err != nil {
		log.Error(err)
	}

	// create a new ad in the database
	ad := model.User{
		Title:   title,
		StartAt: startAt,
		EndAt:   endAt,
		//convert the conditions to a string
		Conditions: fmt.Sprintf("%v", conditions),
	}
	// get the db variable from the middleware
	db := middleware.GetDB()
	_, err = db.Exec("INSERT INTO advertisement (title, start_at, end_at, conditions) VALUES ($1, $2, $3, $4)", ad.Title, ad.StartAt, ad.EndAt, ad.Conditions)

	if err != nil {
		log.Error(err)
	}

	// send the data to the client
	c.JSON(200, gin.H{
		"message": "Ad created successfully",
		"ad":      ad,
	})
}

func GetADs(c *gin.Context) {
	// get all the ads from the database
	// return the data to the client
	// get the db variable from the middleware
	db := middleware.GetDB()

	// test database connectivity
	err := db.Ping()
	if err != nil {
		log.Error("Error: Could not establish a connection with the database")
		return
	}

	log.Info("Connected to the database")
	rows, err := db.Query("SELECT title, start_at, end_at, conditions FROM advertisement")
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()

	ads := []model.User{}
	for rows.Next() {
		ad := model.User{}
		err := rows.Scan(&ad.Title, &ad.StartAt, &ad.EndAt, &ad.Conditions)
		if err != nil {
			log.Error(err)
		}
		ads = append(ads, ad)
	}

	// send the data to the client
	c.JSON(200, gin.H{
		"ads": ads,
	})
}

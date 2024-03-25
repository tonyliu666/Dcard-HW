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
	var adRequest ADRequest
	if err := c.ShouldBindJSON(&adRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	// // convert endAt string to time.Time
	endAt, err := time.Parse(time.RFC3339, endAtStr)
	if err != nil {
		log.Error(err)
	}
	
	country := "["
	for i, v := range conditions.Country {
		if i == len(conditions.Country)-1 {
			country += fmt.Sprintf(`"%s"`, v)
		} else {
			country += fmt.Sprintf(`"%s",`, v)
		}
	}
	country += "]"

	platform := "["
	for i, v := range conditions.Platform {
		if i == len(conditions.Platform)-1 {
			platform += fmt.Sprintf(`"%s"`, v)
		} else {
			platform += fmt.Sprintf(`"%s",`, v)
		}
	}
	platform += "]"
	
	log.Info(country, platform)
	
	conditionsStr := fmt.Sprintf(`{"ageStart": %d, "ageEnd": %d, "country": %s, "platform": %s}`, conditions.AgeStart, conditions.AgeEnd, country, platform)

	log.Info(conditionsStr)
	//conditionsStr := fmt.Sprintf(`{"ageStart": %d, "ageEnd": %d, "country": %s, "platform": %s}`, conditions.AgeStart, conditions.AgeEnd, conditions.Country, conditions.Platform)
	ad := model.User{
		Title:      title,
		StartAt:    startAt,
		EndAt:      endAt,
		Conditions: conditionsStr,
	}

	log.Info(ad.Title, ad.StartAt, ad.EndAt, ad.Conditions)

	err = CreateDbField(&ad)

	if err != nil {
		log.Error("insert failed: ", err)
	}

	// send the data to the client
	c.JSON(200, gin.H{
		"message": "Ad created successfully",
		"ad":      ad,
	})
}

func CreateDbField(ad *model.User) error {
	// get the db variable from the middleware
	db := middleware.GetDB()
	// create the fields in the database

	insertStmt := `INSERT INTO advertisement (title, start_at, end_at, conditions) VALUES ($1, $2, $3, $4)`
	// change the format of fields country, platform in ad.Conditions to "country": ["TW", "JP","US"], "platform": ["android", "ios"]

	_, err := db.Exec(insertStmt, ad.Title, ad.StartAt, ad.EndAt, ad.Conditions)
	if err != nil {
		return err
	}
	return nil
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

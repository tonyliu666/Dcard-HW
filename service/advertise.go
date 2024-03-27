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
		Gender   string   `json:"gender"`
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
	
	gender := fmt.Sprintf(`"%s"`, conditions.Gender)
	conditionsStr := fmt.Sprintf(`{"ageStart": %d, "ageEnd": %d,"gender": %s, "country": %s, "platform": %s}`, conditions.AgeStart, conditions.AgeEnd, gender, country, platform)

	log.Info(conditionsStr)
	
	ad := model.User{
		Title:      title,
		StartAt:    startAt,
		EndAt:      endAt,
		Conditions: conditionsStr,
	}

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
	
	_, err := db.Exec(insertStmt, ad.Title, ad.StartAt, ad.EndAt, ad.Conditions)
	if err != nil {
		return err
	}
	return nil
}

// get all the advertisements
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

// get the advertisemnet with some conditions
/*
eg: curl -X GET -H "Content-Type: application/json" \
Android iOS，
"http://<hos t>/api/v1/ad?offset =10&limit=3&age=24&gender=F&country=TW&platform=ios"
*/


func GetADsWithConditions(c *gin.Context) {
	// get the ads with some conditions
	// get the db variable from the middleware
	db := middleware.GetDB()

	// get the all the parameters from the client
	// params := c.Params
	// age := params.Get("age")
	// country := params.Get("country")
	// gender := params.Get("gender")
	// platform := params.Get("platform")
	// get the ads with the conditions
	// send the data to the client

	err := db.Ping()
	if err != nil {
		log.Error("Error: Could not establish a connection with the database")
		return
	}


	
}
	

	
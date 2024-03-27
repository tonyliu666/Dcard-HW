package service

import (
	"dcardapp/middleware"
	"dcardapp/model"
	"fmt"
	"net/http"
	"strconv"
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

// get the advertisemnet with some conditions
/*
eg: curl -X GET -H "Content-Type: application/json" \
Android iOS，
"http://<hos t>/api/v1/ad?offset =10&limit=3&age=24&gender=F&country=TW&platform=ios"
*/

func GetADsWithConditions(c *gin.Context) {
	// get the ads with some conditions
	// get the db variable from the middleware
	params := c.Request.URL.Query()

	db := middleware.GetDB()
	// get the all the parameters from the client
	// wrap the parameters in the query

	age := params.Get("age")
	country := params.Get("country")
	gender := params.Get("gender")
	platform := params.Get("platform")
	offset, _ := strconv.Atoi(params.Get("offset"))
	limit, _ := strconv.Atoi(params.Get("limit"))
	

	// check whether country,platform and gender params are in each row of the database
	// if the country and platform are in the conditions of the row, then the row is selected
	// As same as above statement, the age should be between the ageStart and ageEnd
	// country,platform,gender,and age are the variables of the conditions,so pass them to the query

	query := `SELECT title, end_at, conditions FROM advertisement WHERE conditions @> '{"country": ["` + country + `"], "platform": ["` + platform + `"], "gender": "` + gender + `"}'
	AND $1::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int`

	log.Info(query)

	rows, err := db.Query(query, age)
	
	if err != nil {
		log.Error("don't find the suitable advertise for you: ", err)
	}
	defer rows.Close()

	// create a slice to store the satisfy ads
	satisfyADs := []model.User{}
	index := 0
	// select only limit number of rows, the number is equal to limit and the ads start from off
	for rows.Next() {
		if index >= offset {
			ad := model.User{}
			err := rows.Scan(&ad.Title,&ad.EndAt, &ad.Conditions)
			if err != nil {
				log.Error(err)
			}
			satisfyADs = append(satisfyADs, ad)
			// if the length of the satisfyADs is equal to the limit, break the loop
			if len(satisfyADs) == limit {
				break
			}
		}
		index++
	}

	// the result would be like this:
	/*
		{"title": "AD 31",
		"endAt" "2023-12-30T12:00:00.000Z"}
		{"title": "AD 10",
		"endAt" "2023-12-31T16:00:00.000Z"}
	*/

	c.JSON(200, gin.H{
		"outputs": satisfyADs,
	})
}

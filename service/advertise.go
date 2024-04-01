package service

import (
	"database/sql"
	"dcardapp/middleware"
	"dcardapp/model"
	"dcardapp/param"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	CreationCounter int
	counterMutex    sync.Mutex
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


func init() {
	// Start the background goroutine to reset the counter
	go ResetCreationCounter()
}

// restrict this counter value not over 3000
func ResetCreationCounter() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C
		// Reset the counter to 0 at the beginning of each day
		counterMutex.Lock()
		CreationCounter = 0
		counterMutex.Unlock()
	}
}


func CheckADExist(title string) bool {
	db := middleware.GetDB()
	// check the advertisement is created before
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM advertisement WHERE title = $1", title).Scan(&count)
	if err != nil {
		log.Error(err)
	}
	if count > 0 {
		return true
	}
	return false
}

func RequestTransformToUser(adRequest ADRequest) (model.User, error) {
	// process the request

	// the time format is ISO 8601 format
	startAt, err := time.Parse(time.RFC3339, adRequest.StartAt)
	if err != nil {
		return model.User{}, err
	}
	// // convert endAt string to time.Time
	endAt, err := time.Parse(time.RFC3339, adRequest.EndAt)
	if err != nil {
		log.Error(err)
	}

	country := "["
	for i, v := range adRequest.Conditions.Country {
		if i == len(adRequest.Conditions.Country)-1 {
			country += fmt.Sprintf(`"%s"`, v)
		} else {
			country += fmt.Sprintf(`"%s",`, v)
		}
	}
	country += "]"

	platform := "["
	for i, v := range adRequest.Conditions.Platform {
		if i == len(adRequest.Conditions.Platform)-1 {
			platform += fmt.Sprintf(`"%s"`, v)
		} else {
			platform += fmt.Sprintf(`"%s",`, v)
		}
	}

	platform += "]"

	gender := fmt.Sprintf(`"%s"`, adRequest.Conditions.Gender)
	conditionsStr := fmt.Sprintf(`{"ageStart": %d, "ageEnd": %d,"gender": %s, "country": %s, "platform": %s}`, adRequest.Conditions.AgeStart, adRequest.Conditions.AgeEnd, gender, country, platform)

	return model.User{
		Title:      adRequest.Title,
		StartAt:    startAt,
		EndAt:      endAt,
		Conditions: conditionsStr,
	}, nil
}

func CreateADs(c *gin.Context) {
	counterMutex.Lock()
    defer counterMutex.Unlock()

	// counter + 1
	if CreationCounter >= 3000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The advertisement is created over the quota today, please try tomorrow"})
		return
	}
	CreationCounter++


	var adRequest ADRequest

	if err := c.ShouldBindJSON(&adRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := adRequest.Title

	// can't asynchronusly do this function, should check the ad exist first
	if CheckADExist(title) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The advertisement is created before"})
		return
	}

	ad, err := RequestTransformToUser(adRequest)
	if err != nil {
		log.Error("request transform failed: ", err)
	}

	go CreateDbField(&ad)

	// send the data to the client
	c.JSON(200, gin.H{
		"message": "Ad created successfully",
	})
}

func CreateDbField(ad *model.User) error {
	// get the db variable from the middleware
	db := middleware.GetDB()
	// create the fields in the database

	insertStmt := `INSERT INTO advertisement (title, start_at, end_at, conditions) VALUES ($1, $2, $3, $4)`

	// check the advertisement is created before
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

func SearchForYourAds(dbQuery string, query param.Query, db *sql.DB, c *gin.Context) {
	rows, err := db.Query(dbQuery, query.Age, query.Limit, query.Offset)

	if err != nil {
		log.Error("don't find the suitable advertise for you: ", err)
	}
	defer rows.Close()

	// create a slice to store the satisfy ads with the query.Response type

	satisfyADs := []param.Response{}

	index := 1
	// select only limit number of rows, the number is equal to limit and the ads start from off
	// according to how many selected rows, create how many go routines to process the data
	for rows.Next() {
		if index >= query.Offset {
			ad := param.Response{}
			err := rows.Scan(&ad.Title, &ad.EndAt)
			if err != nil {
				log.Error(err)
			}
			satisfyADs = append(satisfyADs, ad)
			// if the length of the satisfyADs is equal to the limit, break the loop
			if len(satisfyADs) == query.Limit {
				break
			}
		}
		index++
	}

	// only return title and endAt to the client
	// send the data to the client
	c.JSON(200, gin.H{
		"items": satisfyADs,
	})
}

func GetADsWithConditions(c *gin.Context) {
	// get the ads with some conditions
	// get the db variable from the middleware
	params := c.Request.URL.Query()
	db := middleware.GetDB()

	// get the all the parameters from the client
	// wrap the parameters in the query
	offset, _ := strconv.Atoi(params.Get("offset"))
	limit := 5 // default limit is 5
	if params.Get("limit") != "" {
		limit, _ = strconv.Atoi(params.Get("limit"))
	}

	query := param.Query{
		Offset:   offset,
		Limit:    limit,
		Age:      params.Get("age"),
		Country:  params.Get("country"),
		Platform: params.Get("platform"),
		Gender:   params.Get("gender"),
	}

	// check whether country,platform and gender params are in each row of the database
	// if the country and platform are in the conditions of the row, then the row is selected
	// As same as above statement, the age should be between the ageStart and ageEnd
	// country,platform,gender,and age are the variables of the conditions,so pass them to the query
	// and sort the result by the endAt

	dbQuery := `SELECT title, end_at FROM advertisement WHERE conditions @> '{"country": ["` + query.Country + `"], "platform": ["` + query.Platform + `"], "gender": "` + query.Gender + `"}'
	AND $1::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int ORDER BY end_at ASC LIMIT $2 OFFSET $3`

	SearchForYourAds(dbQuery, query, db, c)
}

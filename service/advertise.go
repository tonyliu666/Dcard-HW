package service

import (
	"dcardapp/buffer"
	"dcardapp/middleware"
	"dcardapp/model"
	"dcardapp/param"
	"dcardapp/util"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	
	//"dcardapp/buffer"
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

var Enqueuer *work.Enqueuer
var CacheConnection redis.Conn

func init() {
	// Start the background goroutine to reset the counter
	go ResetCreationCounter()
	Enqueuer = buffer.SetupEnqueuer()
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
	db, err := middleware.GetDB()
	if err != nil {
		log.Error("get the database failed: ", err)
	}
	// check the advertisement is created before
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM advertisement WHERE title = $1", title).Scan(&count)
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

	log.Info(adRequest.Title, startAt, endAt, conditionsStr)

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

	//counter + 1
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
	db, err := middleware.GetDB()
	if err != nil {
		log.Error("get the database failed: ", err)
	}
	// create the fields in the database

	insertStmt := `INSERT INTO advertisement (title, start_at, end_at, conditions) VALUES ($1, $2, $3, $4)`

	// check the advertisement is created before
	_, err = db.Exec(insertStmt, ad.Title, ad.StartAt, ad.EndAt, ad.Conditions)
	if err != nil {
		log.Error("create the advertisement failed: ", err)
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
	// db, err := middleware.GetDB()
	// if err != nil {
	// 	log.Error("get the database failed: ", err)
	// }

	// get the all the parameters from the client
	// wrap the parameters in the query
	offset, _ := strconv.Atoi(params.Get("offset"))
	limit := 5 // default limit is 5
	if params.Get("limit") != "" {
		limit, _ = strconv.Atoi(params.Get("limit"))
	}

	// add asynchroneous way, kafka server send the query to the client and let client process them
	query := param.Query{
		Offset:   offset,
		Limit:    limit,
		Age:      params.Get("age"),
		Country:  params.Get("country"),
		Platform: params.Get("platform"),
		Gender:   params.Get("gender"),
	}

	Enqueuer.EnqueueUnique("searchForYourAds", work.Q{"query": query})

	// check whether country,platform and gender params are in each row of the database
	// if the country and platform are in the conditions of the row, then the row is selected
	// As same as above statement, the age should be between the ageStart and ageEnd
	// country,platform,gender,and age are the variables of the conditions,so pass them to the query
	// and sort the result by the endAt

	// check whether the request is in the redis cache or not
	// if the request is in the cache, then return the result from the cache

	CacheConnection = buffer.SetupCacheConnection()
	defer CacheConnection.Close()

	rsp, err := CacheConnection.Do("GET",util.GenerateHash(query))

	
	if err != nil {
		log.Error("get the value from the redis failed: ", err)
	}

	if rsp != nil {
		c.JSON(200, rsp)
		return
	}

	c.JSON(202, gin.H{
		"message": "The query is still processing",
	})

}

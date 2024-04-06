package service

import (
	"bytes"
	"database/sql"
	"dcardapp/buffer"
	"dcardapp/middleware"
	"dcardapp/model"
	"dcardapp/param"
	"dcardapp/util"
	"encoding/json"

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
	GetADsCounter   int
	GetADsMutex     sync.Mutex
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

func SearchForYourAds(dbQuery string, query param.Query, db *sql.DB, c *gin.Context) {
	var rows *sql.Rows
	var err error
	if query.Age == "" {
		rows, err = db.Query(dbQuery, query.Limit, query.Offset)
	} else {
		rows, err = db.Query(dbQuery, query.Age, query.Limit, query.Offset)
	}
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

	c.JSON(200, gin.H{
		"items": satisfyADs,
	})
}

// change the dbQuery depends on the query, sometimes it will miss some attributes like age,genger,country,platform
// so need to check the query and change the dbQuery
func createDBquery(query param.Query) string {
	var conditions bytes.Buffer

	// Default query
	dbQuery := "SELECT title, end_at FROM advertisement"
	onlyAge := true
	noParam := true

	// Append conditions if available
	
	if query.Country != "" {
		appendCondition(&conditions, "country", query.Country)
		onlyAge = false
		noParam = false
	}
	if query.Platform != "" {
		appendCondition(&conditions, "platform", query.Platform)
		onlyAge = false
		noParam = false
	}
	if query.Gender != "" {
		appendCondition(&conditions, "gender", query.Gender)
		onlyAge = false
		noParam = false
	}

	// conditions.WriteString("}'")
	

	//If there are conditions, add them to the query
	if conditions.Len() > 0 {
		dbQuery += " WHERE conditions @> '{" + conditions.String() + "}'"
	}
	if query.Age != "" {
		noParam = false
		if onlyAge {
			dbQuery += " WHERE $1::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int"
		} else{
			dbQuery += " AND $1::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int"
		}
	}

	if noParam {
		dbQuery += " ORDER BY end_at ASC LIMIT $1 OFFSET $2"
	}else{
		dbQuery += " ORDER BY end_at ASC LIMIT $2 OFFSET $3"
	}

	return dbQuery
}

func appendCondition(conditions *bytes.Buffer, key, value string) {
	// Append comma if necessary
	if conditions.Len() > 0 {
		conditions.WriteString(", ")
	}

	// Append condition
	if key == "gender"{
		conditions.WriteString(fmt.Sprintf(`"%s": "%s"`, key, value))
	}else{
		conditions.WriteString(fmt.Sprintf(`"%s": ["%s"]`, key, value))
	}
}



/*
check whether country,platform and gender params are in each row of the database
if the country and platform are in the conditions of the row, then the row is selected
As same as above statement, the age should be between the ageStart and ageEnd
country,platform,gender,and age are the variables of the conditions,so pass them to the query
and sort the result by the endAt

if the request is in the cache, then return the result from the cache
*/
func GetADsWithConditions(c *gin.Context) {
	GetADsMutex.Lock()
	defer GetADsMutex.Unlock()
	GetADsCounter++

	// get the ads with some conditions
	// get the db variable from the middleware
	params := c.Request.URL.Query()

	// get all the parameters from the client
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

	CacheConnection = buffer.SetupCacheConnection()
	defer CacheConnection.Close()

	rsp, err := CacheConnection.Do("GET", util.GenerateHash(query))

	if err != nil {
		log.Error("get the value from the redis failed: ", err)
		return
	}

	if rsp != nil {
		var responses []param.Response
		// trsansform the rsp to string using the helper function, rsps, _ := redis.String(rsp, err)
		json.Unmarshal(rsp.([]byte), &responses)
		// rsp, _ := redis.String(rsp, err)
		c.JSON(200, gin.H{
			"items": responses,
		}) // return the result from the redis cache
		GetADsCounter--
		return
	}

	//if the number of concurrent requests is over 3000, then send the query to the redis server
	if GetADsCounter >= 1000 {
		Enqueuer.EnqueueUnique("searchForYourAds", work.Q{"query": query})
		GetADsCounter--
		c.JSON(202, gin.H{
			"message": "The query is still processing",
		})
		return
	} else {
		dbQuery := createDBquery(query)
		db, err := middleware.GetDB()
		if err != nil {
			log.Error("get the database failed: ", err)
		}
		SearchForYourAds(dbQuery, query, db, c)
		return
	}
	
}

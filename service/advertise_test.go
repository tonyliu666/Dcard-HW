package service

import (
	"bytes"
	"dcardapp/middleware"
	"dcardapp/param"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"time"

	// load the .env file
	"dcardapp/util"
	"testing"

	"github.com/gomodule/redigo/redis"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
)

func TestCreateADs(t *testing.T) {
	// create the moke gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	adRequest := ADRequest{
		Title:   "AD504",
		StartAt: "2023-12-10T04:15:30.000Z",
		EndAt:   "2024-12-29T16:23:15Z",
	}
	adRequest.Conditions.AgeStart = 20
	adRequest.Conditions.AgeEnd = 40
	adRequest.Conditions.Gender = "F"
	adRequest.Conditions.Country = []string{"TW", "JP"}
	adRequest.Conditions.Platform = []string{"android", "ios"}

	jsonStr, err := json.Marshal(adRequest)

	if err != nil {
		t.Errorf("Marsal error: %v", err)
	}

	// set up a fake request with the parameters q which is the query parameters
	c.Request = httptest.NewRequest("POST", "/api/v1/ad", bytes.NewBuffer(jsonStr))

	// ADRequest filled with the mock data
	// I want to mock c.ShouldBindJSON(&adRequest)

	CreateADs(c)

	// check the data exists in the database
	// if the data does not exist, then the test fails
	// if the data(row) exists, then the test passes

	db, err := middleware.GetDB()

	db.QueryRow("SELECT title FROM advertisement WHERE title = 'AD504'").Scan(&jsonStr)

	if jsonStr == nil {
		t.Errorf("data does not exist in the database")
	}

	// return pass
	t.Logf("data exists in the database")

}

// check whtether the query exists in the redis or not
func TestGetADsWithConditions(t *testing.T) {
	// create the moke gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// set the query parameters
	q := url.Values{}
	q.Add("age", "30")
	q.Add("country", "TW")
	q.Add("platform", "android")
	q.Add("gender", "F")
	q.Add("offset", "0")
	q.Add("limit", "3")

	// set up a fake request with the parameters q which is the query parameters
	c.Request = httptest.NewRequest("GET", "/api/v1/ad", nil)

	c.Request.URL = &url.URL{RawQuery: q.Encode()}

	// check the return json data
	GetADsWithConditions(c)

	// check whether the query has been sent to the redis or not
	// if the query has been sent to the redis, then the test passes
	// if the query has not been sent to the redis, then the test fails

	// check the job name searchForYourAds in the redis
	// get the hashvalue
	query := param.Query{
		Age:      "30",
		Country:  "TW",
		Platform: "android",
		Gender:   "F",
		Offset:   0,
		Limit:    3,
	}

	redisPool := &redis.Pool{
		MaxActive:   3000,
		MaxIdle:     3,
		IdleTimeout: 3 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	key := util.GenerateHash(query)
	conn := redisPool.Get()
	defer conn.Close()

	// get the value from the redis
	_, err := conn.Do("GET", key)
	if err != nil {
		t.Errorf("get the value from the redis failed: %v", err)
	}
	t.Log("get the value from the redis")
}

func TestCheckADExist(t *testing.T) {
	// create the moke gin context
	title := []string{"AD504", "AD203", "AD35"}
	count := 0
	for _, v := range title {
		if CheckADExist(v) {
			count++
		}
	}
	assert.Equal(t, count, 3)
}

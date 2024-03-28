package service

import (
	"bytes"
	"dcardapp/middleware"
	"dcardapp/param"
	"encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"time"

	// load the .env file
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

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

	db := middleware.GetDB()

	db.QueryRow("SELECT title FROM advertisement WHERE title = 'AD504'").Scan(&jsonStr)

	if jsonStr == nil {
		t.Errorf("data does not exist in the database")
	}

	// return pass
	t.Logf("data exists in the database")

}
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

	// check the contents of c.JSON
	if c.Writer.Status() != 200 {
		t.Errorf("status code is not 200")
	}

	// how to check the contents of c.JSON

	log.Info("get ad with conditions")

	// check the contents of JSON is equal to the expected value
	// if the contents are not equal, then print the error message
	// and the expected value

	var response struct {
		Items []param.Response `json:"items"`
	}

	b, err := io.ReadAll(w.Body)

	if err != nil {
		t.Errorf("error: %v", err)
	}

	err = json.Unmarshal(b, &response)

	if err != nil {
		t.Errorf("Unmarshal data error: %v", err)
	}

	// response.EndAt needs to be in time.Time format, like 2024-12-27T16:23:15Z
	endAt, _ := time.Parse(time.RFC3339, "2024-12-27T16:23:15Z")

	// check the response
	assert.Equal(t, response.Items[0].Title, "AD403")
	assert.Equal(t, response.Items[0].EndAt, endAt)

	// return pass
	t.Logf("response: %v", response)
}

func TestCheckADExist(t *testing.T) {
	// create the moke gin context
	title := []string{"AD504", "AD203", "AD35"}
	count := 0
	for _, v := range title {
		if CheckADExist(v){
			count++
		}
	}
	assert.Equal(t, count, 3)
}

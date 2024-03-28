package service

import (
	"database/sql"
	"dcardapp/param"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	// load the .env file
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var (
	host = os.Getenv("DB_HOST")
	port = os.Getenv("DB_PORT")
	user = os.Getenv("DB_USERNAME")

	dbName   = os.Getenv("DB_NAME")
	password = os.Getenv("DB_PASSWORD")
	sslmode  = os.Getenv("DB_SSLMODE")
)

func TestDBconnect(t *testing.T) {
	// connect to the database
	// read the connection parameters

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)
	// open a database connection
	var err error
	db, err := sql.Open("postgres", psqlInfo)


	
	if err != nil {
		t.Errorf("Error: Could not establish a connection with the database")
	}

	insertStmt := `INSERT INTO advertisement (title, start_at, end_at, conditions) VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(insertStmt, "Ad1", "2021-01-01", "2021-12-31", `{"ageStart": 25, "ageEnd": 35, "country": ["TW", "JP","US"], "platform": ["android", "ios"]}`)
	if err != nil {
		t.Errorf("insert failed: %v", err)
	}

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
	if response.Items[0].Title != "AD403" || response.Items[0].EndAt != endAt {
		t.Errorf("response: %v", response)
	}

	// return pass
	t.Logf("response: %v", response)
}

package service

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	// load the .env file
	"testing"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
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

	defer db.Close()
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
	query := "http://localhost:8080/api/v1/ad?offset=10&limit=3&age=24&gender=F&country=TW&platform=ios"
	// get the parameters from the query
	myUrl, _ := url.Parse(query)
	params, _ := url.ParseQuery(myUrl.RawQuery)
	fmt.Println(params)

	age := params.Get("age")
	country := params.Get("country")
	//gender := params.Get("gender")
	platform := params.Get("platform")
	// get the db variable from the middleware
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)
	// open a database connection
	var err error
	db, err := sql.Open("postgres", psqlInfo)

	defer db.Close()

	if err != nil {
		t.Errorf("Error: Could not establish a connection with the database")
	}

	// get the data from the database depending on the above params variables

	// find the ad with the conditions that age is between start_at and end_at and country is equal to country variable and platform is equal to platform variable
	queryStats := `SELECT title, start_at, end_at, conditions FROM advertisement WHERE conditions @> $1::jsonb AND $2::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int`
	rows, err := db.Query(queryStats, `{"country": ["`+country+`"], "platform": ["`+platform+`"]}`, age)
	
	if err != nil {
		t.Errorf("don't find the suitable advertise for you: %v", err)
	}
	defer rows.Close()

	// print the rows data
	for rows.Next() {
		var title, start_at, end_at, conditions string
		err := rows.Scan(&title, &start_at, &end_at, &conditions)
		if err != nil {
			t.Errorf("scan failed: %v", err)
		}
		fmt.Println(title, start_at, end_at, conditions)
	}
}

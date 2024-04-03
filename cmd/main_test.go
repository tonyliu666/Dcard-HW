package main

import (
	"dcrad-background/middleware"
	"testing"

	"github.com/gocraft/work"
	_ "github.com/lib/pq"
)

func TestSearchForYourAds(t *testing.T) {
	query := Query{Age: "24", Country: "TW", Platform: "ios",
		Gender: "F", Offset: 0, Limit: 3}
	 
	dbQuery := `SELECT title, end_at FROM advertisement WHERE conditions @> '{"country": ["` + query.Country + `"], "platform": ["` + query.Platform + `"], "gender": "` + query.Gender + `"}'
	AND $1::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int ORDER BY end_at ASC LIMIT $2 OFFSET $3`

	db, err := middleware.GetDB()
	if err != nil {
		t.Error("get the database failed: ", err)
	}

	Ads := query.SearchForYourAds(dbQuery, db)
	if len(Ads) == 0 {
		t.Error("no ads found")
	}
}

func TestCheckTheAdsWithQuery(t *testing.T) {
	// fake the context
	query := Query{Age: "35", Country: "TW", Platform: "ios",
		Gender: "F", Offset: 2, Limit: 3}
	
	job := &work.Job{}
	err := query.CheckTheAdsWithQuery(job)
	if err != nil {
		t.Error("check the ads with query failed: ", err)
	}
	// check the keys in redis, and the keys are hashed by the query

	keys := query.GenerateHash()
	conn := redisPool.Get()
	defer conn.Close()
	_, err = conn.Do("GET", keys)
	if err != nil {
		t.Error("get the value from redis failed: ", err)
	}

}

package main

import (
	"database/sql"
	"dcrad-background/middleware"
	"dcrad-background/param"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

// Make a redis pool
var redisPool = &redis.Pool{
	MaxActive: 7000,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", "redis:6379")
	},
}

// query should be like this:

type Query struct {
	// contains filtered or unexported fields
	Age      string
	Country  string
	Platform string
	Gender   string
	Offset   int
	Limit    int
}

func (c *Query) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Println("Starting job: ", job.Name)
	return next()
}

// func (c *Query) assignQuery(query map[string]interface{}) {

// }

func (c *Query) FindQuery(job *work.Job, next work.NextMiddlewareFunc) error {
	// If there's a user_id param, set it in the context for future middleware and handlers to use.
	if query, ok := job.Args["query"]; ok {
		// fmt.Println("find the query: ", query)
		// fmt.Println("get the type of query: ", fmt.Sprintf("%T", query))
		// assign all the atributes in query to the struct

		for key, value := range query.(map[string]interface{}) {
			switch key {
			case "Age":
				c.Age = value.(string)
			case "Country":
				c.Country = value.(string)
			case "Gender":
				c.Gender = value.(string)
			case "Limit":
				c.Limit = int(value.(float64))
			case "Offset":
				c.Offset = int(value.(float64))
			case "Platform":
				c.Platform = value.(string)
			}
		}

		if err := job.ArgError(); err != nil {
			fmt.Println("arg error: ", err)
			return err
		}
	}
	return next()
}

func (c *Query) SearchForYourAds(dbQuery string, db *sql.DB) []param.Response {
	rows, err := db.Query(dbQuery, c.Age, c.Limit, c.Offset)

	if err != nil {
		log.Error("don't find the suitable advertise for you: ", err)
	}

	//log.Info("search for your ads")
	defer rows.Close()

	// create a slice to store the satisfy ads with the query.Response type

	satisfyADs := []param.Response{}

	index := 1
	// select only limit number of rows, the number is equal to limit and the ads start from off
	// according to how many selected rows, create how many go routines to process the data
	for rows.Next() {
		if index >= c.Offset {
			ad := param.Response{}
			err := rows.Scan(&ad.Title, &ad.EndAt)
			if err != nil {
				log.Error("database scan error: ", err)
			}
			satisfyADs = append(satisfyADs, ad)
			// if the length of the satisfyADs is equal to the limit, break the loop
			if len(satisfyADs) == c.Limit {
				break
			}
		}
		index++
	}

	//rows.Close()

	// only return title and endAt to the client
	// send the data to the client
	return satisfyADs
}

func (c *Query) CheckTheAdsWithQuery(job *work.Job) error {
	// Extract arguments:

	// Extract the query from the job
	db, err := middleware.GetDB()
	if err != nil {
		fmt.Println("get the database failed: ", err)
	}

	dbQuery := `SELECT title, end_at FROM advertisement WHERE conditions @> '{"country": ["` + c.Country + `"], "platform": ["` + c.Platform + `"], "gender": "` + c.Gender + `"}'
	AND $1::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int ORDER BY end_at ASC LIMIT $2 OFFSET $3`

	Ads := c.SearchForYourAds(dbQuery, db)
	// set this result as a value in the redis

	// set these Ads in the redis
	conn := redisPool.Get()
	defer conn.Close()

	key := c.GenerateHash()

	// convert Ads to json string
	AdsJson, err := json.Marshal(Ads)
	if err != nil {
		return err	
	}
	// set the ttl for the key to 30mins if timeout, the key will be deleted
	_, err = conn.Do("SET", key, AdsJson, "EX", 30*60)
	
	if err != nil {
		return err
	}
	return nil
}

func (query Query) GenerateHash() string {
	// Concatenate struct fields into a string
	hash := fmt.Sprintf("%s:%s:%s:%s:%d:%d",
		query.Age, query.Country, query.Platform, query.Gender, query.Offset, query.Limit)
	return hash
}

func main() {
	pool := work.NewWorkerPool(Query{}, 300, "query_namespace", redisPool)

	// Add middleware that will be executed for each job
	pool.Middleware((*Query).Log)
	pool.Middleware((*Query).FindQuery)

	// Map the name of jobs to handler functions
	pool.Job("searchForYourAds", (*Query).CheckTheAdsWithQuery)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	// Stop the pool
	pool.Stop()
}

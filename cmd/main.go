package main

import (
	"database/sql"
	"dcrad-background/middleware"
	"dcrad-background/param"
	"dcrad-background/util"
	"fmt"
	"os"
	"os/signal"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

// Make a redis pool
var redisPool = &redis.Pool{
	MaxActive: 10,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", ":6379")
	},
}

// query should be like this:
/*
type Query struct {
// contains filtered or unexported fields
	Age      string
	Country  string
	Platform string
	Gender   string
	Offset   int
	Limit    int
}
*/

type Context struct {
	Age      string
	Country  string
	Platform string
	Gender   string
	Offset   int
	Limit    int
}

func (c *Context) assignValue(query interface{}) {
	c.Age = query.(param.Query).Age
	c.Country = query.(param.Query).Country
	c.Platform = query.(param.Query).Platform
	c.Gender = query.(param.Query).Gender 
	c.Offset = query.(param.Query).Offset
	c.Limit = query.(param.Query).Limit
}

func (c *Context) transformToQuery() param.Query {
	return param.Query{Age: c.Age, Country: c.Country, Platform: c.Platform, 
		Gender: c.Gender, Offset: c.Offset, Limit: c.Limit}
}



func (c *Context) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Println("Starting job: ", job.Name)
	return next()
}

func (c *Context) FindQuery(job *work.Job, next work.NextMiddlewareFunc) error {
	// If there's a user_id param, set it in the context for future middleware and handlers to use.
	if query, ok := job.Args["query"]; ok {
		c.assignValue(query)
		if err := job.ArgError(); err != nil {
			return err
		}
	}
	return next()
}

func SearchForYourAds(dbQuery string, query param.Query, db *sql.DB) []param.Response {
	rows, err := db.Query(dbQuery, query.Age, query.Limit, query.Offset)

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
		if index >= query.Offset {
			ad := param.Response{}
			err := rows.Scan(&ad.Title, &ad.EndAt)
			if err != nil {
				log.Error("database scan error: ", err)
			}
			satisfyADs = append(satisfyADs, ad)
			// if the length of the satisfyADs is equal to the limit, break the loop
			if len(satisfyADs) == query.Limit {
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

func (c *Context) CheckTheAdsWithQuery(job *work.Job) error {
	// Extract arguments:

	newquery := c.transformToQuery()
	// Extract the query from the job
	db, err := middleware.GetDB()
	if err != nil {
		fmt.Println("get the database failed: ", err)
	}

	dbQuery := `SELECT title, end_at FROM advertisement WHERE conditions @> '{"country": ["` + newquery.Country + `"], "platform": ["` + newquery.Platform + `"], "gender": "` + newquery.Gender + `"}'
	AND $1::int BETWEEN (conditions->>'ageStart')::int AND (conditions->>'ageEnd')::int ORDER BY end_at ASC LIMIT $2 OFFSET $3`

	Ads := SearchForYourAds(dbQuery, newquery, db)
	// set this result as a value in the redis

	// set these Ads in the redis
	conn := redisPool.Get()
	defer conn.Close()

	key := util.GenerateHash(newquery)
	_, err = conn.Do("SET", key, Ads)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// Make a new pool. Arguments:
	// Context{} is a struct that will be the context for the request.
	// 10 is the max concurrency
	// "application_namespace" is the Redis namespace
	// redisPool is a Redis pool
	pool := work.NewWorkerPool(Context{}, 10, "query_namespace", redisPool)

	// Add middleware that will be executed for each job
	pool.Middleware((*Context).Log)
	pool.Middleware((*Context).FindQuery)

	// Map the name of jobs to handler functions
	pool.Job("searchForYourAds", (*Context).CheckTheAdsWithQuery)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
}

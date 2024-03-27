package param

import "time"

// Path: query/param.go

type Query struct {
	// contains filtered or unexported fields
	Age      string
	Country  string
	Platform string
	Gender   string
	Offset   int
	Limit    int
}
type Response struct {
	Title string    `xorm:"pk" json:"title" update:"fixed"`
	EndAt time.Time `json:"endAt" xorm:"updated utc"`
}

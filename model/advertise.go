package model

import "time"

type User struct {
	Title             string    `xorm:"pk" json:"title" update:"fixed"`
	Conditions           string    `json:"condition" binding:"required"`
	StartAt     time.Time `json:"startAt" xorm:"created utc"`
	EndAt     time.Time `json:"endAt" xorm:"updated utc"`
}

func (u *User) TableName() string {
	return "users"
}

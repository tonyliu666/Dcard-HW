package model

import "time"

type User struct {
	Title             string    `xorm:"pk" json:"id" update:"fixed"`
	Conditions           string    `json:"name" binding:"required"`
	StartAt     time.Time `json:"create_time" xorm:"created utc"`
	EndAt     time.Time `json:"update_time" xorm:"updated utc"`
}

func (u *User) TableName() string {
	return "users"
}

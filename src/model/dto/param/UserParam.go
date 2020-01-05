package param

import "time"

type UserParam struct {
	Username    string    `form:"username"     json:"username"     binding:"required,gte=5,lte=30,name"`
	Profile     string    `form:"profile"      json:"profile"      binding:"required,gte=0,lte=255"`
	Sex         string    `form:"sex"          json:"sex"          binding:"required"`
	BirthTime   time.Time `form:"birth_time"   json:"birth_time"   binding:"required" time_format:"2006-01-02"`
	PhoneNumber string    `form:"phone_number" json:"phone_number" binding:"required,len=11,phone"`
}
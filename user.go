package models

type User struct {
	Name string
	Phn_num string
	Start_at string
}

type TicketsShow struct {
	Id string
	Name string
	PhnNumber string
	Number string
	StartAt string
	EndAt string
	Expire string
}

type Timing struct {
	Id int
	Count int
	Start string
	End string
}



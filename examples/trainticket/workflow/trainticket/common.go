package trainticket

import "time"

const CALENDAR_LAYOUT = "2006-01-02 15:04:05"

type OrderTicketsInfo struct {
	AccountID       string
	ContactsID      string
	TripID          string
	SeatType        int
	LoginToken      string
	Date            string
	From            string
	To              string
	Assurance       int
	FoodType        int
	StationName     string
	StoreName       string
	FoodName        string
	FoodPrice       float64
	HandleDate      string
	ConsigneeName   string
	ConsigneePhone  string
	ConsigneeWeight float64
	IsWithin        bool
}

type NotifyInfo struct {
	Email       string
	OrderNumber string
	Username    string
	StartPlace  string
	EndPlace    string
	StartTime   string
	Date        string
	SeatClass   string
	SeatNumber  string
	Price       string
}

func indexOf(list []string, target string) int {
	for i, v := range list {
		if v == target {
			return i
		}
	}
	return -1
}

func isContained(soldTickets []Ticket, seat int) bool {
	for _, t := range soldTickets {
		if t.SeatNo == seat {
			return true
		}
	}
	return false
}

func dateToString() string {
	now := time.Now()
	return now.Format(CALENDAR_LAYOUT)
}

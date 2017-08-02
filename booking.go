package carservice

import "time"

// Booking is the domain model for handling bookings
type Booking struct {
	ID                    int64     `json:"id"`
	CustomerName          string    `json:"customerName"`
	ContactNumber         string    `json:"customerNumber"`
	CarRegistrationNumber string    `json:"carRegistrationNumber"`
	CarMake               string    `json:"carMake"`
	CarModel              string    `json:"carModel"`
	Odometer              int       `json:"odoMeter"`
	ServiceDate           time.Time `json:"serviceDate"`
	BookingDate           time.Time
}

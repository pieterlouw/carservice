package datastore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/pieterlouw/carservice"
)

var (
	// ErrBookingNotFound is used when a specific booking request is not found
	ErrBookingNotFound = errors.New("booking not found")
)

// BookingStore is a database wrapper
type BookingStore struct {
	db *sql.DB
}

// NewBookingStore create a new datastore with underlying sql.DB
func NewBookingStore(db *sql.DB) *BookingStore {
	return &BookingStore{
		db: db,
	}
}

// Add inserts a new booking
func (b BookingStore) Add(ctx context.Context, booking carservice.Booking) (int64, error) {
	result, err := b.db.ExecContext(ctx, `INSERT INTO Bookings(CustomerName,ContactNumber,CarRegistrationNumber,CarMake,CarModel,Odometer,ServiceDate) 
											VALUES (?,?,?,?,?,?,?)`,
		booking.CustomerName, booking.ContactNumber, booking.CarRegistrationNumber, booking.CarMake, booking.CarModel, booking.Odometer, booking.ServiceDate)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Get return a specific booking
func (b BookingStore) Get(ctx context.Context, id int64) (carservice.Booking, error) {
	var booking carservice.Booking

	err := b.db.QueryRowContext(ctx, `SELECT ID,CustomerName,ContactNumber,CarRegistrationNumber,CarMake,
								CarModel,Odometer,ServiceDate
							FROM Bookings WHERE ID = ? `, id).Scan(
		&booking.ID, &booking.CustomerName, &booking.ContactNumber, &booking.CarRegistrationNumber, &booking.CarMake,
		&booking.CarModel, &booking.Odometer, &booking.ServiceDate)

	if err != nil {
		//return zero-valued Booking on error
		if err == sql.ErrNoRows { // own defined error if not found
			return carservice.Booking{}, ErrBookingNotFound
		}
		return carservice.Booking{}, err
	}

	return booking, nil
}

// GetAll return all bookings
func (b BookingStore) GetAll(ctx context.Context) ([]carservice.Booking, error) {
	var bookings []carservice.Booking

	rows, err := b.db.QueryContext(ctx, `SELECT ID,CustomerName,ContactNumber,CarRegistrationNumber,CarMake,
									CarModel,Odometer,ServiceDate
									FROM Bookings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var booking carservice.Booking

		err := rows.Scan(&booking.ID, &booking.CustomerName, &booking.ContactNumber, &booking.CarRegistrationNumber,
			&booking.CarMake, &booking.CarModel, &booking.Odometer, &booking.ServiceDate)

		if err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

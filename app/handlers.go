package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pieterlouw/carservice"
	"github.com/pieterlouw/carservice/datastore"
)

// BookingServer is a web server for car service bookings
type BookingServer struct {
	BookingRepository    *datastore.BookingStore
	NewBookingDispatcher chan carservice.Booking
	context.Context      //embedded
}

func (b BookingServer) showDashBoard(w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, r, "dashboard.html", http.StatusOK, nil)
}

func (b BookingServer) listBookings(w http.ResponseWriter, r *http.Request) error {
	bookings, err := b.BookingRepository.GetAll(b.Context)
	if err != nil {
		return err
	}

	return renderTemplate(w, r, "bookings/list.html", http.StatusOK, struct {
		Bookings []carservice.Booking
	}{
		Bookings: bookings,
	})

}

func (b BookingServer) addBookingForm(w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, r, "bookings/add_form.html", http.StatusOK, struct{}{})
}

func (b BookingServer) handleBooking(w http.ResponseWriter, r *http.Request) error {
	var booking carservice.Booking
	var message string

	// view booking
	if r.Method == http.MethodGet {
		id, err := strconv.ParseInt(r.URL.Path[len("/booking/"):], 10, 64) //retrieve booking id from URL
		if err != nil {
			return err
		}

		booking, err = b.BookingRepository.Get(b.Context, id) // lookup booking from db
		if err != nil {
			return err
		}

	} else if r.Method == http.MethodPost { // add booking via HTTP form POST
		//parse and read form values from request
		err := r.ParseForm()
		if err != nil {
			return nil
		}

		serviceDate, err := time.Parse("2006-01-02", r.Form.Get("ServiceDate"))
		if err != nil {
			return err
		}

		odometer, err := strconv.ParseInt(r.Form.Get("Odometer"), 10, 32)
		if err != nil {
			return err
		}

		booking = carservice.Booking{
			CustomerName:          r.Form.Get("CustomerName"),
			ContactNumber:         r.Form.Get("ContactNumber"),
			CarRegistrationNumber: r.Form.Get("CarRegistrationNumber"),
			CarMake:               r.Form.Get("CarMake"),
			CarModel:              r.Form.Get("CarModel"),
			ServiceDate:           serviceDate,
			Odometer:              int(odometer),
		}

		// call service to add deduction rule
		id, err := b.BookingRepository.Add(b.Context, booking) // add booking to db
		if err != nil {
			return err
		}

		booking.ID = id

		// dispatch the new booking to the dispatcher
		b.NewBookingDispatcher <- booking

		message = fmt.Sprintf("Booking # %d (for %s) added successfully", booking.ID, booking.CarRegistrationNumber)
	} else {
		return errors.New("Invalid Method")
	}

	return renderTemplate(w, r, "bookings/view.html", http.StatusOK, struct {
		carservice.Booking
		Message string
	}{
		Message: message,
		Booking: booking,
	})
}

func (b BookingServer) listBookingsAPI(w http.ResponseWriter, r *http.Request) error {
	bookings, err := b.BookingRepository.GetAll(b.Context)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(bookings)
	if err != nil {
		return err
	}

	return nil
}

func (b BookingServer) handleBookingAPI(w http.ResponseWriter, r *http.Request) error {
	var booking carservice.Booking

	if r.Method == http.MethodGet {
		id, err := strconv.ParseInt(r.URL.Path[len("/api/booking/"):], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid Request: %v", r.URL.Path[len("/api/booking/"):])
			return nil
		}

		booking, err = b.BookingRepository.Get(b.Context, id)
		if err != nil {
			if err == datastore.ErrBookingNotFound {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "Booking requested not found: %v", id)
				return nil
			}

			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Error: %v", err)
			return nil
		}

		w.WriteHeader(http.StatusOK)

	} else if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576)) //read request body to get JSON payload
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Error: %v", err)
			return nil
		}
		defer r.Body.Close()

		err = json.Unmarshal(body, &booking) // unmarshal JSON payload
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "Invalid Booking payload: %v", body)
			return nil
		}

		// add Booking to db
		id, err := b.BookingRepository.Add(b.Context, booking)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Error: %v", err)
			return nil
		}
		booking.ID = id

		// dispatch the new booking to the dispatcher
		b.NewBookingDispatcher <- booking

		w.WriteHeader(http.StatusCreated)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method Not Allowed")
		return nil
	}

	// return booking as JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8") //set HTTP Response Content-Type Header
	err := json.NewEncoder(w).Encode(booking)
	if err != nil {
		return err
	}
	return nil
}

package app

import (
	"bytes"
	"fmt"
	"go/build"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

var (
	// StaticDir is the directory containing static assets.
	StaticDir = filepath.Join(defaultBase("github.com/pieterlouw/carservice/app"), "static")
)

// Handler returns a ServeMux with all handlers related to the webapp
func Handler(b BookingServer) *http.ServeMux {
	m := http.NewServeMux() // create a new HTTP multiplexer

	//handlers that return html
	m.HandleFunc("/", makeHandler(b.showDashBoard)) // root shows a dashboard

	m.HandleFunc("/bookings/", makeHandler(b.listBookings))          // list all bookings
	m.HandleFunc("/booking/add_form", makeHandler(b.addBookingForm)) // show form to fill in booking details
	m.HandleFunc("/booking/", makeHandler(b.handleBooking))          // GET (show) or POST (add) a booking

	//handlers that return json
	m.HandleFunc("/api/bookings/", makeHandler(b.listBookingsAPI)) // list all bookings as a JSON array
	m.HandleFunc("/api/booking/", makeHandler(b.handleBookingAPI)) // GET (show) or POST (add) a booking using JSON objects

	//handler to serve static assests (css/images/javascript/fonts etc.)
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(StaticDir))))

	return m
}

func makeHandler(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loadTemplates()

		start := time.Now()

		err := fn(w, r)
		if err != nil {
			logError(r, err, nil)
			handleError(w, r, http.StatusInternalServerError, err)
		}

		// log middleware for each request
		log.Printf(
			"%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("cache-control", "no-cache")
	err2 := renderTemplate(w, r, "error.html", status, &struct {
		StatusCode int
		Status     string
		Err        error
	}{
		StatusCode: status,
		Status:     http.StatusText(status),
		Err:        err,
	})
	if err2 != nil {
		logError(r, fmt.Errorf("during execution of error template: %s", err2), nil)
	}
}

func logError(req *http.Request, err error, rv interface{}) {
	if err != nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Error serving %s : %s\n", req.URL, err)
		if rv != nil {
			fmt.Fprintln(&buf, rv)
			buf.Write(debug.Stack())
		}
		log.Print(buf.String())
	}
}

func defaultBase(path string) string {
	p, err := build.Default.Import(path, "", build.FindOnly)
	if err != nil {
		log.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	p.Dir, err = filepath.Rel(cwd, p.Dir)
	if err != nil {
		log.Fatal(err)
	}

	return p.Dir
}

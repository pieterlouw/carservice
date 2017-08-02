package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pieterlouw/carservice"
	"github.com/pieterlouw/carservice/app"
	"github.com/pieterlouw/carservice/datastore"
)

const appVersion = "0.1.0"

var (
	webEndPoint = flag.String("w", ":8080", "Endpoint address for web application to listen on")
	dbPath      = flag.String("d", "carservice.db", "Path to sqlite database")
	queueSize   = flag.Int("q", 5, "Size of the queue that holds new bookings")
)

func main() {

	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// setup database connection
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	newBookingChannel := make(chan carservice.Booking, *queueSize)

	// inject database, booking channel and context for our web application
	webapp := app.BookingServer{
		BookingRepository:    datastore.NewBookingStore(db),
		NewBookingDispatcher: newBookingChannel,
		Context:              ctx,
	}

	// Create a new server with handlers and set timeout values.
	webAppServer := http.Server{
		Addr:         *webEndPoint,
		Handler:      app.Handler(webapp),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("startup : Application setup complete.Version=%s\n", appVersion)

	// We want to report when each goroutine is closed.
	var wg sync.WaitGroup

	//start new bookings worker
	go newBookingWorker(ctx, newBookingChannel, &wg)

	//start web server
	go func() {
		wg.Add(1)
		defer wg.Done()

		log.Printf("startup : Web server listening on %s", *webEndPoint)
		log.Printf("shutdown : Web server closed : %v", webAppServer.ListenAndServe())
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// start go routine to wait for any of the signals to be signalled
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done // block until a close signal received

	//graceful shutdown and cleanup code
	close(done)

	// Create a context to attempt a graceful 10 second shutdown.
	const timeout = 10 * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	// Attempt the graceful shutdown by closing the listener and
	// completing all inflight requests.
	if err := webAppServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown : Graceful shutdown did not complete for webapp in %v : %v", timeout, err)

		// Looks like we timedout on the graceful shutdown. Kill it hard.
		if err := webAppServer.Close(); err != nil {
			log.Printf("shutdown : Error killing webapp server : %v", err)
		}
	}

	cancel() // cancel all contexts

	wg.Wait() // Wait for all goroutines to report it is closed.
	log.Printf("shutdown : Application gracefully shutdown.Version=%s\n", appVersion)
}

// blocking worker function to process incoming bookings
func newBookingWorker(ctx context.Context, newBookingChannel chan carservice.Booking, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	log.Println("startup : Worker started")

	for {
		select {
		case booking := <-newBookingChannel:
			log.Println("********************  Booking created!  ********************")
			log.Println("	ID:", booking.ID)
			log.Println("	CustomerName:", booking.CustomerName)
			log.Println("	ContactNumber:", booking.ContactNumber)
			log.Println("	CarRegistrationNumber:", booking.CarRegistrationNumber)
			log.Println("	CarMake:", booking.CarMake)
			log.Println("	CarModel:", booking.CarModel)
			log.Println("	Odometer:", booking.Odometer)
			log.Println("	ServiceDate:", booking.ServiceDate)
			log.Println("************************************************************")

			// the  rest of the code will send the booking to the sales department
			// with a RPC call like JSON RPC/HTTP or gRPC client

		// this is to avoid leaking of this goroutine when ctx is done.
		case <-ctx.Done():
			log.Printf("shutdown : New Bookings Worker")
			return
		}
	}
}

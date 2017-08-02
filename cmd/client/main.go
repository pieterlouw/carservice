package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pieterlouw/carservice"
)

func main() {
	serviceDate, err := time.Parse("2006-01-02", "2017-09-01")
	if err != nil {
		log.Fatalf("Invalid Service Date specified: %v", err)
	}

	booking := carservice.Booking{
		CustomerName:          "Pieter",
		ContactNumber:         "082 123 45678",
		CarRegistrationNumber: "CLI 000 GP",
		CarMake:               "Toyota",
		CarModel:              "Corolla 2004",
		Odometer:              250000,
		ServiceDate:           serviceDate,
	}

	payload, err := json.Marshal(booking)
	//	err = json.NewEncoder(buffer).Encode(booking)
	if err != nil {
		log.Fatalf("Cannot Marshal JSON: %v", err)
	}

	fmt.Printf("Payload: %s\n", payload)

	url := "http://localhost:8080/api/booking/"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))

	req.Header.Add("content-type", "application/json")

	client := &http.Client{}
	res, _ := client.Do(req)

	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
		fmt.Println("Booking created")

		fmt.Println(string(body))
	} else {
		fmt.Printf("Booking failed - Code %d\n", res.StatusCode)

		fmt.Println(res)
	}

}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Receipt ...
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

// Item ...
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// ProcessResponse ...
type ProcessResponse struct {
	ID string `json:"id"`
}

// PointsResponse ...
type PointsResponse struct {
	Points int `json:"points"`
}

// receipts ...
var receipts = make(map[string]int)

// main function to stare
func main() {
	log.Println("Server Started ::")

	// Created Router for routing requests
	router := mux.NewRouter()
	// Request Processor Handler
	router.HandleFunc("/receipts/process", processReceipt).Methods("POST")

	// Request Point Handler
	router.HandleFunc("/receipts/{id}/points", getPoints).Methods("GET")

	// Server runs on :8080 port
	http.ListenAndServe(":8080", router)
}

// processReceipt ...
func processReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	log.Println("Request loggeed ::", r)

	// Decode reqest from byte to json struct
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {

		// if there is invalid request then we have to return bad request 400 status code
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// proceess valid request for generating points
	points := calculatePoints(receipt)

	// generate new request id for every request
	id := uuid.New().String()

	// store requests in cache memory
	receipts[id] = points

	// return with json respoce with process Id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProcessResponse{ID: id})
}

// getPoints ...
func getPoints(w http.ResponseWriter, r *http.Request) {
	log.Println("Request loggeed For Request ::", r)

	fmt.Println("receipts ::", receipts)

	// to read the values from request URL
	vars := mux.Vars(r)

	// get the ID value
	id := vars["id"]

	// check the value is present in cache memory
	points, found := receipts[id]
	if !found {
		// check if the request Id is not present then we have to return the status not found 404 Status code
		http.Error(w, "Receipt ID not found", http.StatusNotFound)
		return
	}

	// return with json respoce with process points
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PointsResponse{Points: points})
}

// calculatePoints ...
func calculatePoints(receipt Receipt) int {
	points := 0

	// Rule 1: One point for every alphanumeric character in the retailer name
	points += countAlphanumeric(receipt.Retailer)

	// Rule 2: 50 points if the total is a round dollar amount with no cents
	if isRoundDollarAmount(receipt.Total) {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25
	if isMultipleOfQuarter(receipt.Total) {
		points += 25
	}

	// Rule 4: 5 points for every two items on the receipt
	points += (len(receipt.Items) / 2) * 5

	// Rule 5: Points based on item description length
	for _, item := range receipt.Items {
		points += calculateItemDescriptionPoints(item)
	}

	// Rule 6: 6 points if the day in the purchase date is odd
	if isPurchaseDateOdd(receipt.PurchaseDate) {
		points += 6
	}

	// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm
	if isWithinTimeRange(receipt.PurchaseTime, "14:00", "16:00") {
		points += 10
	}

	return points
}

// countAlphanumeric ...
func countAlphanumeric(s string) int {
	count := 0
	for _, ch := range s {
		if unicode.IsLetter(ch) || unicode.IsNumber(ch) {
			count++
		}
	}
	return count
}

// isRoundDollarAmount ...
func isRoundDollarAmount(total string) bool {
	// spits the string accoring with separation factor as .
	parts := strings.Split(total, ".")
	return len(parts) == 2 && parts[1] == "00"
}

func isMultipleOfQuarter(total string) bool {
	var amount float64
	// take the value and convert it to float64
	fmt.Sscanf(total, "%f", &amount)
	return math.Mod(amount, 0.25) == 0
}

// isMultipleOfQuarter ...
func calculateItemDescriptionPoints(item Item) int {
	// trim the spaces present string
	description := strings.TrimSpace(item.ShortDescription)
	if len(description)%3 == 0 {
		var price float64
		fmt.Sscanf(item.Price, "%f", &price)
		return int(math.Ceil(price * 0.2))
	}
	return 0
}

// isPurchaseDateOdd ...
func isPurchaseDateOdd(dateStr string) bool {

	// parse date in format of 2006-01-02
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// return error if there given time is not abl to parse
		return false
	}
	return date.Day()%2 != 0
}

// isWithinTimeRange ...
func isWithinTimeRange(timeStr, startStr, endStr string) bool {
	// parse the time
	purchaseTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		// if thre is an error in parsing the time return false
		return false
	}
	startTime, _ := time.Parse("15:04", startStr)
	endTime, _ := time.Parse("15:04", endStr)

	// if time parse correctly, calculate the after and before time.
	return purchaseTime.After(startTime) && purchaseTime.Before(endTime)
}

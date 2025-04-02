package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

func randomString(n int) string {
	if n <= 0 {
		panic("randomString: length must be greater than 0")
	}

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func getCoordinates(address string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := fmt.Sprintf("?q=%s&format=json&limit=1", url.QueryEscape(address))

	// Build the full URL
	fullURL := baseURL + params

	// Create an HTTP GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a non-200 status code
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Parse the JSON response
	var data []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// If no data is found, return nil coordinates
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("no results found for the given address")
	}

	// Convert latitude and longitude strings to float64
	lat, err := strconv.ParseFloat(data[0].Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latitude: %w", err)
	}

	lon, err := strconv.ParseFloat(data[0].Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse longitude: %w", err)
	}

	fmt.Println(lat, lon)

	return lat, lon, nil
}

func isValidEmail(email string) bool {

	validEmailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Validiert eine E-Mail-Adresse anhand von Regex
	return validEmailRegex.MatchString(email)
}

package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Service struct {
	ID          int    `json:"id"`
	AppCode     string `json:"app_code"`
	AppName     string `json:"app_name"`
	Env         string `json:"env"`
	Cloud       string `json:"cloud"`
	Region      string `json:"region"`
	TeamName    string `json:"team_name"`
	PMContact   string `json:"pm_contact"`
	TeamContact string `json:"team_contact"`
}

var (
	services    []Service
	lastService int
	mu          sync.Mutex
	logger      = log.New(os.Stdout, "", log.LstdFlags)
)

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/add", addService)
	http.HandleFunc("/view", viewServices)
	http.HandleFunc("/search", searchServices)
	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func addService(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	service := Service{
		ID:          generateID(),
		AppCode:     r.FormValue("app_code"),
		AppName:     r.FormValue("app_name"),
		Env:         r.FormValue("env"),
		Cloud:       r.FormValue("cloud"),
		Region:      r.FormValue("region"),
		TeamName:    r.FormValue("team_name"),
		PMContact:   r.FormValue("pm_contact"),
		TeamContact: r.FormValue("team_contact"),
	}

	// Validate the fields
	if !validateService(service) {
		http.Error(w, "All fields are required and AppCode must be unique", http.StatusBadRequest)
		return
	}

	// Store services in memory
	mu.Lock()
	services = append(services, service)
	mu.Unlock()

	// Store services in JSON file
	storeServices()

	// Log the added service with timestamp
	logger.Printf("Service added: %+v\n", service)

	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func viewServices(w http.ResponseWriter, r *http.Request) {
	// Load existing services data from file (if available)
	loadServicesFromFile()

	// Render the view template with the services data
	renderServices(w, services)
}

func loadServicesFromFile() {
	// Read the existing services data from file
	data, err := ioutil.ReadFile("services.json")
	if err != nil {
		logger.Println("Error reading services from file:", err)
		return
	}

	// Unmarshal the JSON data into the services slice
	err = json.Unmarshal(data, &services)
	if err != nil {
		logger.Println("Error unmarshalling services:", err)
		return
	}
}

func searchServices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	filteredServices := filterServices(query)

	// Return filtered services as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredServices)
}

func filterServices(query string) []Service {
	var filtered []Service
	for _, service := range services {
		if strings.Contains(strings.ToLower(service.AppCode), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(service.AppName), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(service.Env), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(service.Cloud), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(service.Region), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(service.TeamName), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(service.PMContact), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(service.TeamContact), strings.ToLower(query)) {
			filtered = append(filtered, service)
		}
	}
	return filtered
}

func renderServices(w http.ResponseWriter, svc []Service) {
	tmpl := template.Must(template.ParseFiles("view.html"))
	tmpl.Execute(w, svc)
}

func storeServices() {
	var data []byte
	var err error

	// Check if the file exists
	if _, err := os.Stat("services.json"); err == nil {
		// File exists, read its content
		data, err = ioutil.ReadFile("services.json")
		if err != nil {
			logger.Println("Error reading services.json:", err)
			return
		}
	}

	// Marshal new service data
	newData, err := json.Marshal(services)
	if err != nil {
		logger.Println("Error marshalling services:", err)
		return
	}

	// Append new service data to existing or empty content
	if len(data) > 0 {
		data = append(data[:len(data)-1], ',') // Remove the last ']' and add a comma for JSON array
		data = append(data, newData[1:]...)    // Append the new service data (skipping the initial '[')
	} else {
		data = newData // No existing content, use the new service data directly
	}

	// Write the content to the file
	err = ioutil.WriteFile("services.json", data, 0644)
	if err != nil {
		logger.Println("Error writing services to file:", err)
	}
}

func generateID() int {
	mu.Lock()
	defer mu.Unlock()
	lastService++
	return lastService
}

func validateService(service Service) bool {
	if service.AppCode == "" ||
		service.AppName == "" ||
		service.Env == "" ||
		service.Cloud == "" ||
		service.Region == "" ||
		service.TeamName == "" ||
		service.PMContact == "" ||
		service.TeamContact == "" {
		return false
	}

	for _, s := range services {
		if s.AppCode == service.AppCode {
			return false // Duplicate AppCode found
		}
	}

	return true
}

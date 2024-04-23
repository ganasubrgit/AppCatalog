package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Service struct {
	AppCode     string `json:"app_code"`
	AppName     string `json:"app_name"`
	Env         string `json:"env"`
	Cloud       string `json:"cloud"`
	Region      string `json:"region"`
	TeamName    string `json:"team_name"`
	PMContact   string `json:"pm_contact"`
	TeamContact string `json:"team_contact"`
}

var services []Service

// Logger with timestamp
var logger = log.New(os.Stdout, "", log.LstdFlags)

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/add", addService)
	http.HandleFunc("/view", viewServices)
	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func addService(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	service := Service{
		AppCode:     r.FormValue("app_code"),
		AppName:     r.FormValue("app_name"),
		Env:         r.FormValue("env"),
		Cloud:       r.FormValue("cloud"),
		Region:      r.FormValue("region"),
		TeamName:    r.FormValue("team_name"),
		PMContact:   r.FormValue("pm_contact"),
		TeamContact: r.FormValue("team_contact"),
	}
	services = append(services, service)

	// Store services in JSON file
	storeServices()

	// Log the added service with timestamp
	logger.Printf("Service added: %+v\n", service)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func viewServices(w http.ResponseWriter, r *http.Request) {
	// Display all services or filtered services based on search query
	if r.Method == "POST" {
		r.ParseForm()
		query := r.FormValue("query")
		filteredServices := filterServices(query)
		renderServices(w, filteredServices)
		return
	}

	// If no search query, render all services
	renderServices(w, services)
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
	data, err := json.Marshal(services)
	if err != nil {
		logger.Println("Error marshalling services:", err)
		return
	}

	err = writeFile("services.json", data)
	if err != nil {
		logger.Println("Error writing services to file:", err)
	}
}

func writeFile(filename string, data []byte) error {
	return ioutil.WriteFile(filename, data, 0644)
}

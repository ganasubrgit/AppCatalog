package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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
	jsonFile    = "services.json"
)

func main() {
	loadServicesFromFile()

	http.HandleFunc("/", home)
	http.HandleFunc("/add", addService)
	http.HandleFunc("/view", viewServices)
	http.HandleFunc("/search", searchServices)
	http.HandleFunc("/edit", editService)
	http.HandleFunc("/update", updateService)
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

	if !validateService(service) {
		http.Error(w, "All fields are required and AppCode must be unique", http.StatusBadRequest)
		return
	}

	mu.Lock()
	services = append(services, service)
	mu.Unlock()

	storeServices()

	logger.Printf("Service added: %+v\n", service)

	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func viewServices(w http.ResponseWriter, r *http.Request) {
	loadServicesFromFile()
	renderServices(w, services)
}

func searchServices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	filteredServices := filterServices(query)

	// Check if the request has a query parameter
	if query != "" {
		// Render search results as a table in HTML
		renderServices(w, filteredServices)
		return
	}

	// Check if the request is an HTML form submission
	if r.Header.Get("Accept") == "text/html" {
		// Render search results as a table in HTML
		renderServices(w, filteredServices)
	} else {
		// Reply with JSON data for API request
		jsonData, err := json.Marshal(filteredServices)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}

func editService(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Service ID is required", http.StatusBadRequest)
		return
	}

	serviceID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	var serviceToEdit *Service
	for i, service := range services {
		if service.ID == serviceID {
			serviceToEdit = &services[i]
			break
		}
	}

	if serviceToEdit == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("edit.html"))
	tmpl.Execute(w, serviceToEdit)
}

func updateService(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "Service ID is required", http.StatusBadRequest)
		return
	}

	serviceID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	var serviceToUpdate *Service
	for i, service := range services {
		if service.ID == serviceID {
			serviceToUpdate = &services[i]
			break
		}
	}

	if serviceToUpdate == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	serviceToUpdate.AppCode = r.FormValue("app_code")
	serviceToUpdate.AppName = r.FormValue("app_name")
	serviceToUpdate.Env = r.FormValue("env")
	serviceToUpdate.Cloud = r.FormValue("cloud")
	serviceToUpdate.Region = r.FormValue("region")
	serviceToUpdate.TeamName = r.FormValue("team_name")
	serviceToUpdate.PMContact = r.FormValue("pm_contact")
	serviceToUpdate.TeamContact = r.FormValue("team_contact")

	storeServices()

	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func renderServices(w http.ResponseWriter, svc []Service) {
	tmpl := template.Must(template.ParseFiles("view.html"))
	tmpl.Execute(w, svc)
}

func loadServicesFromFile() {
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Println("Services file does not exist. Starting with a new file.")
			return
		}
		logger.Println("Error reading services from file:", err)
		return
	}

	err = json.Unmarshal(data, &services)
	if err != nil {
		logger.Println("Error unmarshalling services:", err)
		return
	}

	// Find the highest ID from the loaded services
	for _, service := range services {
		if service.ID > lastService {
			lastService = service.ID
		}
	}
}

func storeServices() {
	data, err := json.Marshal(services)
	if err != nil {
		logger.Println("Error marshalling services:", err)
		return
	}

	err = ioutil.WriteFile(jsonFile, data, 0644)
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
		if s.AppCode == service.AppCode && s.ID != service.ID {
			return false // Duplicate AppCode found
		}
	}

	return true
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

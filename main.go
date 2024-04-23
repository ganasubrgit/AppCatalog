package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func viewServices(w http.ResponseWriter, r *http.Request) {
	// Display all services in a table format
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>All Services</h1>")
	fmt.Fprintf(w, "<table border='1'><tr><th>AppCode</th><th>AppName</th><th>Env</th><th>Cloud</th><th>Region</th><th>TeamName</th><th>PMContact</th><th>TeamContact</th></tr>")
	for _, service := range services {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			service.AppCode, service.AppName, service.Env, service.Cloud, service.Region, service.TeamName, service.PMContact, service.TeamContact)
	}
	fmt.Fprintf(w, "</table>")
}

func storeServices() {
	data, err := json.Marshal(services)
	if err != nil {
		fmt.Println("Error marshalling services:", err)
		return
	}

	err = writeFile("services.json", data)
	if err != nil {
		fmt.Println("Error writing services to file:", err)
	}
}

func writeFile(filename string, data []byte) error {
	return ioutil.WriteFile(filename, data, 0644)
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/consul/api"
)

const (
	registrationEndpoint = "/v1/registration/"
)

type (
	// RegistrationResponse is the response from a registration call
	RegistrationResponse struct {
		Hosts []RegistrationHost `json:"hosts"`
	}
	// RegistrationHost is a host in the response
	RegistrationHost struct {
		IPAddress string               `json:"ip_address"`
		Port      int                  `json:"port"`
		Tags      RegistrationHostTags `json:"tags"`
	}
	// RegistrationHostTags are the tages for a registration
	RegistrationHostTags struct {
		AZ                 string `json:"az,omitempty"`
		Canary             *bool  `json:"canary,omitempty"`
		LoadBalacingWeight *int   `json:"load_balancing_weight,omitempty"`
	}
)

func main() {
	http.HandleFunc("/v1/registration/", handleRegistration)

	addr := os.Getenv("ADDRESS")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	log.Println("starting server on ", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("failed to start listener: %v\n", err)
	}
}

func handleRegistration(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Path[len(registrationEndpoint):]

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	services, _, err := client.Catalog().Service(serviceName, "", nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	response := RegistrationResponse{
		Hosts: make([]RegistrationHost, 0, len(services)),
	}
	for _, service := range services {
		response.Hosts = append(response.Hosts, RegistrationHost{
			IPAddress: service.Address,
			Port:      service.ServicePort,
			Tags: RegistrationHostTags{
				AZ: service.Datacenter,
			},
		})
	}

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(response)
}

package main

import (
	"bytes"

	"fmt"
	"log"
	"math"
	"net/url"
	"os"
    "net/http"
	"time"

	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

type DomainStats struct {
	Success int
	Total   int
}

var stats = make(map[string]*DomainStats)

func checkHealth(endpoint Endpoint) {
	var client = &http.Client{}

	bodyBytes := []byte(endpoint.Body)


	reqBody := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest(endpoint.Method, endpoint.URL, reqBody)
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}

	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}
    start := time.Now()
	resp, err := client.Do(req)
	domain := extractDomain(endpoint.URL)

	
    duration := time.Since(start)
	stats[domain].Total++
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && duration <= 500*time.Millisecond {
		defer resp.Body.Close()
		stats[domain].Success++
	}
}

func extractDomain(urls string) string {

	parsedurl, err := url.Parse(urls)
	if err!=nil{
	 fmt.Printf("Error parsing the url:%s\n",urls)
	 return urls
	}
	domain:=parsedurl.Hostname()
	
	return domain
}

func monitorEndpoints(endpoints []Endpoint) {
	for _, endpoint := range endpoints {
		domain := extractDomain(endpoint.URL)

		if stats[domain] == nil {
			stats[domain] = &DomainStats{}
		}
	}

	for {
		for _, endpoint := range endpoints {
			checkHealth(endpoint)
		}
		logResults()
		time.Sleep(15 * time.Second)
	}
}

func logResults() {
	for domain, stat := range stats {
		if stat.Total >0{
			percentage := math.Round(10000 * float64(stat.Success) / float64(stat.Total))/100
			fmt.Printf("%s has %.2f%% availability\n", domain, percentage)
		}else{
			fmt.Printf("No data is available%s\n",domain)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <config_file>")
	}

	filePath := os.Args[1]
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var endpoints []Endpoint
	if err := yaml.Unmarshal(data, &endpoints); err != nil {
		log.Fatal("Error parsing YAML:", err)
	}

	monitorEndpoints(endpoints)
}

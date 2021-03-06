package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var DomainsCORS = "https://redirectinfo.com"

type Redirect struct {
	Url  string `json:"url"`
	Code int    `json:"code"`
}

type Response struct {
	Status   bool       `json:"status"`
	Response []Redirect `json:"response"`
}

func (response *Response) AddRedirect(item Redirect) []Redirect {
	response.Response = append(response.Response, item)
	return response.Response
}

func getDomainFromURI(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal(err)
	}
	parts := strings.Split(u.Hostname(), ".")
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
	return domain
}

func getRedirects(response *Response, site string) {

	nextURL := site
	var i int

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	for i < 10 {

		resp, err := client.Get(nextURL)
		if err != nil {
			fmt.Println(err)
		}

		redirect := Redirect{resp.Request.URL.String(), resp.StatusCode}
		response.AddRedirect(redirect)

		if resp.StatusCode == 200 {
			fmt.Println(resp)
			fmt.Println("Done!")
			break
		} else if resp.StatusCode == 404 {
			fmt.Println("404")
			break
		} else {
			nextURL = resp.Header.Get("Location")

			domain := getDomainFromURI(site)
			if !strings.Contains(nextURL, domain) {
				nextURL = "http://" + domain + nextURL
			}
		}
		i += 1
	}
}

func responseWriter(w http.ResponseWriter, data string) {
	resp := make(map[string]string)
	resp["message"] = data
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Header().Set("Access-Control-Allow-Origin", DomainsCORS)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
	return
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	site := req.URL.Query().Get("site")

	if len(site) < 3 {
		responseWriter(w, "GET parameter `site` is required")
		return
	}

	if !strings.Contains(site, "http://") && !strings.Contains(site, "https://") {
		site = "http://" + site
	}

	_, err := url.ParseRequestURI(site)
	if err != nil {
		responseWriter(w, err.Error())
		return
	}

	_, err = http.Get(site)
	if err != nil {
		responseWriter(w, err.Error())
		return
	}

	response := Response{
		Status: true,
	}

	getRedirects(&response, site)

	redirectsJson, _ := json.Marshal(response)

	w.Header().Set("Access-Control-Allow-Origin", DomainsCORS)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Write(redirectsJson)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":9091", nil)
}

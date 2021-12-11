package main

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
)

type Redirect struct {
	Url		string `json:"url"`
	Code	int `json:"code"`
}

type Response struct {
	Status		bool `json:"status"`
	Response 	[]Redirect `json:"response"`
}

func (response *Response) AddRedirect(item Redirect) []Redirect {
    response.Response = append(response.Response, item)
    return response.Response
}

func getRedirects(response *Response, site string) {

	nextURL := site
	var i int

	client := &http.Client {
	  CheckRedirect: func(req *http.Request, via []*http.Request) error {
	    return http.ErrUseLastResponse
	} }

	for i < 100 {

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
		  i += 1
		}
	}
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	site := req.URL.Query().Get("site")
	if len(site) < 3 {
		resp := make(map[string]string)
		resp["message"] = "GET parameter `site` is required"

		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
		return
	}

	response := Response{
		Status: true,
	}

	getRedirects(&response, site)

	redirectsJson, _ := json.Marshal(response)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Write(redirectsJson)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":9091", nil)
}

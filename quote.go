package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// Quote is data type
type Quote struct {
	Quote    string `json:"quote"`
	Author   string `json:"author"`
	Category string `json:"cat"`
}

var apiURL = "https://talaikis.com/api/quotes/random/"

// GetQuote gets a random quote from Random quotes API, https://talaikis.com/random_quotes_api/
func GetQuote() (quote string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return "四面楚歌"
	}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return "Nothing to say."
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("[WARN] %v", res)
		return "Discretion is the better part of valor."
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("[WARN] %v", err)
		return "No more."
	}

	q := Quote{}
	if err = json.Unmarshal(body, &q); err != nil {
		log.Printf("[WARN] %v", err)
		return "耳泳暴洋 瞳座星原"
	}

	return q.Quote
}

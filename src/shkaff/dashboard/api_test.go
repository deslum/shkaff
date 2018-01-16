package dashboard

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
)

const (
	HOST = "0.0.0.0"
	PORT = 5500
)

func TestGetRequests(t *testing.T) {
	reqString := []string{"GetUser", "GetDatabase", "GetTask"}
	for _, req := range reqString {
		url := fmt.Sprintf("http://%s:%d/api/v1/%s/1?token=12345", HOST, PORT, req)
		resp, err := http.Get(url)
		if err != nil {
			log.Panicln(err)
		}
		if resp.StatusCode != 200 {
			log.Fatal(resp.StatusCode)
		}
	}
}

func TestUpdateUser(t *testing.T) {
	form := url.Values{
		"login":      {"forTest"},
		"api_token":  {"54321"},
		"first_name": {"Mike"},
		"last_name":  {"Polonsky"},
		"is_active":  {"true"},
		"is_admin":   {"false"},
	}
	url := fmt.Sprintf("http://%s:%d/api/v1/UpdateUser/1?token=12345", HOST, PORT)
	body := bytes.NewBufferString(form.Encode())
	resp, err := http.Post(url, "application/x-www-form-urlencoded", body)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalln(resp.StatusCode)
	}
}

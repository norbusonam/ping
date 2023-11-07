package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

type PingBody struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

func main() {
	// load env vars
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// figure out what port to listen on
	port := ":8080"
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	}

	// setup email env
	fromEmail, fromEmailExists := os.LookupEnv("FROM_EMAIL")
	fromPassword, fromPasswordExists := os.LookupEnv("FROM_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	if !fromEmailExists || !fromPasswordExists {
		log.Fatal("Missing email env vars")
	}
	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	// handle send email request
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		// validate request method
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// check api key
		if r.Header.Get("x-api-key") != os.Getenv("API_KEY") {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// check request body is not empty
		if r.Body == nil {
			http.Error(w, "Request body is empty", http.StatusBadRequest)
			return
		}

		// decode request body
		var reqBody PingBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// validate request body
		if len(reqBody.To) == 0 || reqBody.Subject == "" || reqBody.Body == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// send email
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, reqBody.To, []byte("Subject: "+reqBody.Subject+"\r\n\r\n"+reqBody.Body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write([]byte("Message sent"))
		}
	})

	http.ListenAndServe(port, nil)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"utilserver/profile/clients"
	"utilserver/profile/domain"
	"utilserver/profile/storage"
	"utilserver/profile/transport"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		log.Fatal("Error loading .env file")
	}
}

func main() {
	repository, err := storage.New()
	if err != nil {
		log.Fatal(err)
	}
	httpClient := clients.New(5)
	spotifyAuthService := domain.New(repository, httpClient)

	router := transport.Handler(spotifyAuthService)

	fmt.Printf("Starting server at port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", router))
}

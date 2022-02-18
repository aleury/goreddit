package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aleury/goreddit/postgres"
	"github.com/aleury/goreddit/web"
)

func main() {
	dsn := os.Getenv("DATA_SOURCE_NAME")

	store, err := postgres.NewStore(dsn)
	if err != nil {
		log.Fatal(err)
	}

	h := web.NewHandler(store)
	http.ListenAndServe(":3000", h)
}

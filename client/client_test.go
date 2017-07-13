package client

import (
	"fmt"
	"log"
	"os"
)

func ExampleNew() {
	token := os.Getenv("REPLICATED_API_TOKEN")
	appSlug := os.Getenv("REPLICATED_APP_SLUG")

	api := New(token)

	app, err := api.GetAppBySlug(appSlug)
	if err != nil {
		log.Fatal(err)
	}

	channels, err := api.ListChannels(app.Id)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range channels {
		if c.Name == "Stable" {
			fmt.Println("We have a Stable channel")
		}
	}
	// Output: We have a Stable channel
}

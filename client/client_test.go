package client

import (
	"fmt"
	"log"
	"os"
)

func ExampleNew() {
	token := os.Getenv("REPLICATED_API_TOKEN")
	appSlugOrID := os.Getenv("REPLICATED_APP")

	api := New(token)

	app, err := api.GetApp(appSlug)
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

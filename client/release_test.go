package client

import (
	"testing"
)

func TestListReleases(t *testing.T) {
	client := New(apiOrigin, apiKey)
	_, err := client.ListReleases(appID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateRelease(t *testing.T) {
	client := New(apiOrigin, apiKey)
	_, err := client.CreateRelease(appID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPromoteRelease(t *testing.T) {
	client := New(apiOrigin, apiKey)
	release, err := client.CreateRelease(appID)
	if err != nil {
		t.Fatal(err)
	}
	appChannels, err := client.CreateChannel(appID, "name", "Description")
	err = client.PromoteRelease(appID, release.Sequence, "v1-labelx", "bug fixx", false, appChannels[0].Id)
	if err != nil {
		t.Fatal(err)
	}
}

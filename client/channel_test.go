package client

import (
	"testing"
)

func TestListChannels(t *testing.T) {
	client := New(apiOrigin, apiKey)
	appChannels, err := client.ListChannels(appID)
	if err != nil {
		t.Fatal(err)
	}
	if len(appChannels) == 0 {
		t.Error("No channels returned from ListChannels")
	}
}

func TestCreateChannel(t *testing.T) {
	client := New(apiOrigin, apiKey)
	name := "New Channel"
	description := "TestCreateChanel"
	appChannels, err := client.CreateChannel(appID, name, description)
	if err != nil {
		t.Fatal(err)
	}
	if len(appChannels) == 0 {
		t.Error("No channels returned from CreateChannel")
	}
}

func TestArchiveChannel(t *testing.T) {
	client := New(apiOrigin, apiKey)
	// ensure channel exists to delete
	name := "Delete me"
	description := "TestDeleteChannel"
	appChannels, err := client.CreateChannel(appID, name, description)
	if err != nil {
		t.Fatal(err)
	}
	var channelID string
	for _, appChannel := range appChannels {
		if appChannel.Name == name {
			channelID = appChannel.Id
			break
		}
	}
	err = client.ArchiveChannel(appID, channelID)
	if err != nil {
		t.Fatal(err)
	}
	appChannels, err = client.ListChannels(appID)
	if err != nil {
		t.Fatal(err)
	}
	for _, appChannel := range appChannels {
		if appChannel.Id == channelID {
			t.Errorf("Channel %s not successfully archived", channelID)
		}
	}
}

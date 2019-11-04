package platformclient

// this functions as a "does vendor-api accept empty strings" test
// func Test_DeleteApp(t *testing.T) {
// 	var test = func() (err error) {
// 		appId := "cli-delete-app-id"
// 		token := "cli-delete-app-auth"
//
// 		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
// 		client := HTTPClient{
// 			apiKey:    token,
// 			apiOrigin: u,
// 		}
//
// 		err = client.DeleteApp(appId)
// 		assert.Nil(t, err)
// 		return nil
// 	}
//
// 	pact.AddInteraction().
// 		Given("Delete an app cli-delete-app-id").
// 		UponReceiving("A request to delete an app for cli-delete-app-id").
// 		WithRequest(dsl.Request{
// 			Method: "DELETE",
// 			Path:   dsl.String("/v1/app/cli-delete-app-id"),
// 			Headers: dsl.MapMatcher{
// 				"Authorization": dsl.String("cli-delete-app-auth"),
// 			},
// 			Body: nil,
// 		}).
// 		WillRespondWith(dsl.Response{
// 			Status: 204,
// 			Body:   "",
// 		})
//
// 	if err := pact.Verify(test); err != nil {
// 		t.Fatalf("Error on Verify: %v", err)
// 	}
// }

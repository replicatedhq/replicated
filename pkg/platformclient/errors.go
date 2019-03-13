package platformclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ErrNotFound represnets a 404 response from the API.
var ErrNotFound = errors.New("Not found")

// BadRequest represents a 400 response from the API.
type BadRequest struct {
	MessageCode string `json:"messageCode"`
	Message     string `json:"message"`
}

// Error prints the MessageCode and Message returned from the API.
func (br *BadRequest) Error() string {
	return fmt.Sprintf("%s: %s", br.MessageCode, br.Message)
}

type badRequestBody struct {
	Error *BadRequest
}

func unmarshalBadRequest(r io.Reader) (*BadRequest, error) {
	brb := &badRequestBody{}
	if err := json.NewDecoder(r).Decode(brb); err != nil {
		return nil, err
	}
	if brb.Error == nil {
		return nil, errors.New("No error in body")
	}
	return brb.Error, nil
}

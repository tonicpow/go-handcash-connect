package handcash

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mrz1836/go-sanitize"
)

// GetProfile will get the profile for the associated auth token
//
// Specs: https://github.com/HandCash/handcash-connect-sdk-js/blob/master/src/profile/index.js
func (c *Client) GetProfile(ctx context.Context, authToken string) (user *User, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	var signed *signedRequest
	if signed, err = c.getSignedRequest(
		http.MethodGet,
		endpointProfileCurrent,
		authToken,
		&requestBody{authToken: sanitize.AlphaNumeric(authToken, false)},
		currentISOTimestamp(),
	); err != nil {
		return nil, fmt.Errorf("error creating signed request: %w", err)
	}

	log.Println(currentISOTimestamp())
	log.Println(endpointProfileCurrent)
	log.Println(signed.Method)
	log.Println(signed.URI)
	log.Println(signed.Body)
	log.Println(signed.Headers.OauthSignature)
	log.Println(signed.Headers.OauthPublicKey)
	log.Println(signed.URI)

	// Make the HTTP request
	response := httpRequest(
		ctx,
		c,
		&httpPayload{
			ExpectedStatus: http.StatusOK,
			Method:         http.MethodGet,
			URL:            signed.URI,
			Data:           []byte(emptyBody),
		},
		signed,
	)

	// Error in request?
	if response.Error != nil {
		err = response.Error
		return
	}

	/*// Start the Request
	var request *http.Request
	if request, err = http.NewRequestWithContext(
		ctx,
		signedRequest.Method,
		signedRequest.URI,
		bytes.NewBuffer([]byte(emptyBody)),
	); err != nil {
		return nil, fmt.Errorf("error creating new request: %w", err)
	}

	// Set oAuth headers
	request.Header.Set("oauth-publickey", signedRequest.Headers.OauthPublicKey)
	request.Header.Set("oauth-signature", signedRequest.Headers.OauthSignature)
	request.Header.Set("oauth-timestamp", signedRequest.Headers.OauthTimestamp)

	// Fire the http Request
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(request); err != nil {
		return nil, fmt.Errorf("failed doing http request: %w", err)
	}

	// Close the response body
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read the body
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("failed reading the body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errorMsg := new(errorResponse)
		if err = json.Unmarshal(body, &errorMsg); err != nil {
			return nil, fmt.Errorf("failed unmarshal error: %w", err)
		}

		return nil, fmt.Errorf("bad response: %s %s %s %+v", errorMsg.Message, authToken, request.RequestURI, signedRequest)
	}*/

	newUser := new(User)
	if err = json.Unmarshal(response.BodyContents, &newUser); err != nil {
		err = fmt.Errorf("failed to unmarshal user: %w", err)
		return
	} else if newUser == nil {
		err = fmt.Errorf("failed to find a user: %w", err)
		return
	}
	user = newUser
	return
}

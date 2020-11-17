package handcash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mrz1836/go-sanitize"
)

// User are the user fields returned by the public and private profile endpoints
type User struct {
	PublicProfile  PublicProfile  `json:"publicProfile"`
	PrivateProfile PrivateProfile `json:"privateProfile"`
}

// PublicProfile is the public profile
type PublicProfile struct {
	AvatarURL         string `json:"avatarUrl"`
	DisplayName       string `json:"displayName"`
	Handle            string `json:"handle"`
	Paymail           string `json:"paymail"`
	ID                string `json:"id"`
	LocalCurrencyCode string `json:"localCurrencyCode"`
}

// TODO: Spendable balance was made its own request
// SpendableBalance  int64  `json:"spendableBalance"`

// PrivateProfile is the private profile
type PrivateProfile struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

// GetProfile will get the profile
func GetProfile(authToken string) (user *User, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	// TODO: Does this really need to be in the body?
	var signedRequest *SignedRequest
	signedRequest, err = getSignedRequest(http.MethodGet, "/v1/connect/profile/currentUserProfile", authToken, &requestBody{
		authToken: sanitize.AlphaNumeric(authToken, false),
	})

	if err != nil {
		return nil, fmt.Errorf("error creating signed request: %w", err)
	}

	// Start the Request
	var request *http.Request
	var jsonValue []byte
	if request, err = http.NewRequestWithContext(context.Background(), signedRequest.Method, signedRequest.URI, bytes.NewBuffer(jsonValue)); err != nil {
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

	user = new(User)

	if err = json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed unmarshal HandCash user: %w", err)
	} else if user == nil {
		return nil, fmt.Errorf("failed to find a HandCash user in context: %w", err)
	} else if len(user.PrivateProfile.Email) == 0 {
		return nil, fmt.Errorf("failed to find an email address for HandCash user: %s", user.PublicProfile.ID)
	}

	return
}

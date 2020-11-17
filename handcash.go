// Package handcash is an unofficial Go version of HandCash Connect's SDK
//
// If you have any suggestions or comments, please feel free to open an issue on
// this GitHub repository!
//
// By TonicPow Inc (https://tonicpow.com)
package handcash

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bitcoinsv/bsvd/bsvec"
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

type requestBody struct {
	authToken string
}

// API Documentation https://handcash.github.io/handcash-connect-sdk-js-beta-docs

// APIURL is the server side API url
const APIURL = "https://beta-cloud.handcash.io"

// ClientURL is the client side url
const ClientURL = "https://beta-app.handcash.io"

// OauthHeaders are used for signed requests
type OauthHeaders struct {
	OauthPublicKey string `json:"oauth-publickey"`
	OauthSignature string `json:"oauth-signature"`
	OauthTimestamp string `json:"oauth-timestamp"`
}

// SignedRequest is used to communicate with HandCash Connect API
type SignedRequest struct {
	URI     string       `json:"uri"`
	Method  string       `json:"method"`
	Headers OauthHeaders `json:"headers"`
	Body    interface{}  `json:"body"`
	JSON    bool         `json:"json"`
}

// getHandCashRequestSignature will return the signature hash
//
// Removed "body" since it was empty (in the future, might be needed again)
//
func getHandCashRequestSignature(method string, endpoint string, timestamp string, privateKey *bsvec.PrivateKey) ([]byte, error) {
	signatureHash := getHandCashRequestSignatureHash(method, endpoint, timestamp)
	hash := sha256.Sum256([]byte(signatureHash))
	sig, err := privateKey.Sign(hash[:])
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

// getHandCashRequestSignatureHash will return the signature hash
//
// Removed "body" since it was empty (in the future, might be needed again)
//
func getHandCashRequestSignatureHash(method string, endpoint string, timestamp string) string {
	// todo: missing a case if there is a body present
	return fmt.Sprintf("%s\n%s\n%s\n{}", method, endpoint, timestamp)
}

// GetProfile will get the profile
func GetProfile(authToken string) (user *User, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	var signedRequest *SignedRequest
	signedRequest, err = getSignedRequest(http.MethodGet, "/v1/connect/profile/currentUserProfile", &requestBody{
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

func getSignedRequest(method string, endpoint string, body *requestBody) (*SignedRequest, error) {
	// Match JS ISO time exactly
	JavascriptISOString := "2006-01-02T15:04:05.999Z07:00"
	timestamp := time.Now().UTC().Format(JavascriptISOString)

	tokenBytes, err := hex.DecodeString(body.authToken)
	if err != nil {
		return nil, err
	}
	privateKey, publicKey := bsvec.PrivKeyFromBytes(bsvec.S256(), tokenBytes)

	var requestSignature []byte
	if requestSignature, err = getHandCashRequestSignature(method, endpoint, timestamp, privateKey); err != nil {
		return nil, err
	}
	return &SignedRequest{
		URI:    APIURL + endpoint,
		Method: method,
		Body:   body,
		Headers: OauthHeaders{
			OauthPublicKey: hex.EncodeToString(publicKey.SerializeCompressed()),
			OauthSignature: hex.EncodeToString(requestSignature),
			OauthTimestamp: timestamp,
		},
		JSON: true,
	}, nil
}

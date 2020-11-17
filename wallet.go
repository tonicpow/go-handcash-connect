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

// AppAction enum
type AppAction string

const (
	AppActionTipGroup AppAction = "tip-group"
	AppActionPublish  AppAction = "publish"
	AppActionLike     AppAction = "like"
)

// AttachmentFormat enum
type AttachmentFormat string

const (
	AttachmentFormatBase64 AttachmentFormat = "base64"
	AttachmentFormatJSON   AttachmentFormat = "json"
)

type Attachment struct {
	format AttachmentFormat
	value  map[string]string
}

type Payment struct {
	destination  string
	currencyCode CurrencyCode
	sendAmount   float64
}

type PayParameters struct {
	description string
	appAction   AppAction
	attachment  Attachment
	payments    []Payment
}

type PayRequest struct {
	authToken string
	PayParameters
}

// PaymentType enum
type PaymentType string

const (
	PaymentSend PaymentType = "send"
)

// ParticipantType enum
type ParticipantType string

const (
	ParticipantUser = "user"
)

type Participant struct {
	Type              ParticipantType `json:"type"`
	Alias             string          `json:"alias"`
	DisplayName       string          `json:"displayName"`
	ProfilePictureURL string          `json:"profilePictureUrl"`
	ResponseNote      string          `json:"responseNote"`
}

// PaymentResponse is returned from the GetPayment function
type PaymentResponse struct {
	TransactionID    string `json:"transactionId`
	Type             PaymentType
	Time             uint64
	SatoshiFees      uint64
	SatoshiAmount    uint64
	FiatExchangeRate float64
	FiatCurrencyCode CurrencyCode
	Participants     []Participant
	Attachments      []Attachment
	AppAction        AppAction
}
type PaymentRequest struct {
	TransactionID string `json:"transactionId"`
}

// GetPayment fetches a payment by txid
func GetPayment(authToken string) (payResponse *PaymentResponse, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	// TODO: Does this really need to be sent in the body of the request as well?
	var signedRequest *SignedRequest
	signedRequest, err = getSignedRequest(http.MethodGet, "/v1/connect/wallet/payment", authToken, &requestBody{
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

	payResponse = new(PaymentResponse)

	if err = json.Unmarshal(body, &payResponse); err != nil {
		return nil, fmt.Errorf("failed unmarshal HandCash payRequest: %w", err)
	} else if payResponse == nil {
		return nil, fmt.Errorf("failed to find a HandCash user in context: %w", err)
	}

	return
}

// Pay gets a payment request from the handcash connect API
func Pay(authToken string, payment string) (payResponse *PaymentResponse, err error) {

	// TODO: unmarshal json string to payment paramters struct

	payParams := new(PayParameters)
	err = json.Unmarshal([]byte(payment), payParams)
	if err != nil {
		return nil, fmt.Errorf("Invalid payment parameters %s", err)
	}

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	var signedRequest *SignedRequest

	signedRequest, err = getSignedRequest(http.MethodPost, "/v1/connect/wallet/pay", sanitize.AlphaNumeric(authToken, false), payParams)

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

	payResponse = new(PaymentResponse)

	if err = json.Unmarshal(body, &payResponse); err != nil {
		return nil, fmt.Errorf("failed unmarshal HandCash payResponse: %w", err)
	} else if payResponse == nil {
		return nil, fmt.Errorf("failed to pay: %w", err)
	}

	return
}

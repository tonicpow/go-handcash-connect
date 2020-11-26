package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mrz1836/go-sanitize"
	"github.com/tonicpow/go-handcash-connect/api"
	"github.com/tonicpow/go-handcash-connect/config"
	"github.com/tonicpow/go-handcash-connect/utils"
)

type errorResponse struct {
	Message string `json:"message"`
}

// AppAction enum
type AppAction string

// AppAction enum
const (
	AppActionTipGroup AppAction = "tip-group"
	AppActionPublish  AppAction = "publish"
	AppActionLike     AppAction = "like"
)

// AttachmentFormat enum
type AttachmentFormat string

// AttachmentFormat enum
const (
	AttachmentFormatBase64 AttachmentFormat = "base64"
	AttachmentFormatHex    AttachmentFormat = "hex"
	AttachmentFormatJSON   AttachmentFormat = "json"
)

// Attachment is for additional data
type Attachment struct {
	Format AttachmentFormat  `json:"format,omitempty"`
	Value  map[string]string `json:"value,omitempty"`
}

// Payment is used by PayParameters
type Payment struct {
	Destination  string
	CurrencyCode config.CurrencyCode
	SendAmount   float64
}

// PayParameters is used by Pay()
type PayParameters struct {
	Description string     `json:"description,omitempty"`
	AppAction   AppAction  `json:"appAction,omitempty"`
	Attachment  Attachment `json:"attachment,omitempty"`
	Payments    []Payment  `json:"payments,omitempty"`
}

// PaymentType enum
type PaymentType string

// PaymentType enum
const (
	PaymentSend PaymentType = "send"
)

// ParticipantType enum
type ParticipantType string

// ParticipantType enum
const (
	ParticipantUser = "user"
)

// Participant is used for payments
type Participant struct {
	Type              ParticipantType `json:"type"`
	Alias             string          `json:"alias"`
	DisplayName       string          `json:"displayName"`
	ProfilePictureURL string          `json:"profilePictureUrl"`
	ResponseNote      string          `json:"responseNote"`
}

// PaymentResponse is returned from the GetPayment function
type PaymentResponse struct {
	TransactionID    string              `json:"transactionId"`
	Type             PaymentType         `json:"type"`
	Time             uint64              `json:"time"`
	SatoshiFees      uint64              `json:"satoshiFees"`
	SatoshiAmount    uint64              `json:"satoshiAmount"`
	FiatExchangeRate float64             `json:"fiatExchangeRate"`
	FiatCurrencyCode config.CurrencyCode `json:"fiatCurrencyCode"`
	Participants     []Participant       `json:"participants"`
	Attachments      []Attachment        `json:"attachments"`
	AppAction        AppAction           `json:"appAction"`
}

// PaymentRequest is used for GetPayment()
type PaymentRequest struct {
	TransactionID string `json:"transactionId"`
}

// GetPayment fetches a payment by txid
func GetPayment(authToken string, transactionID string) (payResponse *PaymentResponse, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	var signedRequest *api.SignedRequest
	signedRequest, err = api.GetSignedRequest(http.MethodGet, "/v1/connect/wallet/payment", authToken, &PaymentRequest{
		TransactionID: transactionID,
	}, utils.ISOTimestamp())

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
func Pay(authToken string, payParams PayParameters) (payResponse *PaymentResponse, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	var signedRequest *api.SignedRequest

	signedRequest, err = api.GetSignedRequest(http.MethodPost, "/v1/connect/wallet/pay", sanitize.AlphaNumeric(authToken, false), payParams, utils.ISOTimestamp())
	if err != nil {
		return nil, fmt.Errorf("error creating signed request: %w", err)
	}

	// Start the Request
	var request *http.Request
	payParamsBytes, err := json.Marshal(payParams)
	if request, err = http.NewRequestWithContext(context.Background(), signedRequest.Method, signedRequest.URI, bytes.NewBuffer(payParamsBytes)); err != nil {
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

	if resp.StatusCode != 200 {
		errorMsg := new(errorResponse)
		if err = json.Unmarshal(body, &errorMsg); err != nil {
			return nil, fmt.Errorf("failed unmarshal error: %w", err)
		}

		return nil, fmt.Errorf("Bad response: %s %s %s %+v", errorMsg.Message, authToken, request.RequestURI, signedRequest)
	}

	payResponse = new(PaymentResponse)

	if err = json.Unmarshal(body, &payResponse); err != nil {
		return nil, fmt.Errorf("failed unmarshal HandCash payResponse: %w", err)
	} else if payResponse == nil {
		return nil, fmt.Errorf("failed to pay: %w", err)
	} else if payResponse.TransactionID == "" {
		return nil, fmt.Errorf("Bad response from handcash API. Invalid Transaction ID %s", payResponse.TransactionID)
	}

	return
}

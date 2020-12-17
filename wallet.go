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

// GetPayment fetches a payment by txid
func (c *Client) GetPayment(authToken string, transactionID string) (payResponse *PaymentResponse, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	var signedRequest *signedRequest
	signedRequest, err = c.getSignedRequest(http.MethodGet, "/v1/connect/wallet/payment", authToken, &PaymentRequest{
		TransactionID: transactionID,
	}, currentISOTimestamp())

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
func (c *Client) Pay(authToken string, payParams PayParameters) (payResponse *PaymentResponse, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Get the signed request
	var signedRequest *signedRequest

	if signedRequest, err = c.getSignedRequest(
		http.MethodPost,
		"/v1/connect/wallet/pay",
		sanitize.AlphaNumeric(authToken, false),
		payParams,
		currentISOTimestamp(),
	); err != nil {
		return nil, fmt.Errorf("error creating signed request: %w", err)
	}

	// Start the Request
	var request *http.Request
	var payParamsBytes []byte
	payParamsBytes, err = json.Marshal(payParams)
	if request, err = http.NewRequestWithContext(context.Background(),
		signedRequest.Method, signedRequest.URI, bytes.NewBuffer(payParamsBytes)); err != nil {
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
	}

	payResponse = new(PaymentResponse)

	if err = json.Unmarshal(body, &payResponse); err != nil {
		return nil, fmt.Errorf("failed unmarshal HandCash payResponse: %w", err)
	} else if payResponse == nil {
		return nil, fmt.Errorf("failed to pay: %w", err)
	} else if payResponse.TransactionID == "" {
		return nil, fmt.Errorf("bad response from handcash API. Invalid Transaction ID %s", payResponse.TransactionID)
	}

	return
}

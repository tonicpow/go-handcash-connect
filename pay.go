package handcash

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Pay gets a payment request from the handcash connect API
func (c *Client) Pay(ctx context.Context, authToken string,
	payParams *PayParameters) (*PaymentResponse, error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// Make sure we have payment params
	if payParams == nil || len(payParams.Receivers) == 0 {
		return nil, fmt.Errorf("invalid payment parameters")
	}

	// Get the signed request
	signed, err := c.getSignedRequest(
		http.MethodPost,
		endpointGetPayRequest,
		authToken,
		payParams,
		currentISOTimestamp(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating signed request: %w", err)
	}

	// Convert into bytes
	var payParamsBytes []byte
	if payParamsBytes, err = json.Marshal(payParams); err != nil {
		return nil, err
	}

	// Make the HTTP request
	response := httpRequest(
		ctx,
		c,
		&httpPayload{
			Data:           payParamsBytes,
			ExpectedStatus: http.StatusOK,
			Method:         signed.Method,
			URL:            signed.URI,
		},
		signed,
	)

	// Error in request?
	if response.Error != nil {
		return nil, response.Error
	}

	/*// Start the Request
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
	}*/

	// Unmarshal pay response
	paymentResponse := new(PaymentResponse)
	if err = json.Unmarshal(response.BodyContents, &paymentResponse); err != nil {
		return nil, fmt.Errorf("failed unmarshal: %w", err)
	} else if paymentResponse == nil || paymentResponse.TransactionID == "" {
		return nil, fmt.Errorf("failed to payment response: %w", err)
	}
	return paymentResponse, nil
}

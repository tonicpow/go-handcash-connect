package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/tonicpow/go-handcash-connect/config"
)

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

// GetRequestSignature will return the signature hash
func GetRequestSignature(method string, endpoint string, body interface{}, timestamp string, privateKey *bsvec.PrivateKey) ([]byte, error) {
	signatureHash, err := GetRequestSignatureHash(method, endpoint, body, timestamp)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(signatureHash))
	sig, err := privateKey.Sign(hash[:])
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

// GetRequestSignatureHash will return the signature hash
func GetRequestSignatureHash(method string, endpoint string, body interface{}, timestamp string) (string, error) {
	var bodyString string
	if body == nil {
		bodyString = "{}"
	} else {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("Failed to marshal body %w", err)
		}
		bodyString = fmt.Sprintf("%s", bodyBytes)
	}
	sigHash := fmt.Sprintf("%s\n%s\n%s\n%s", method, endpoint, timestamp, bodyString)
	return sigHash, nil
}

// GetSignedRequest returns the request with signature
func GetSignedRequest(method string, endpoint string, authToken string, body interface{}, timestamp string) (*SignedRequest, error) {

	tokenBytes, err := hex.DecodeString(authToken)
	if err != nil {
		return nil, err
	}
	privateKey, publicKey := bsvec.PrivKeyFromBytes(bsvec.S256(), tokenBytes)

	var requestSignature []byte
	if requestSignature, err = GetRequestSignature(method, endpoint, body, timestamp, privateKey); err != nil {
		return nil, err
	}

	return &SignedRequest{
		URI:    config.APIURL + endpoint,
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

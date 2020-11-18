package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

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

// GetHandCashRequestSignature will return the signature hash
//
// Removed "body" since it was empty (in the future, might be needed again)
//
func GetHandCashRequestSignature(method string, endpoint string, body interface{}, timestamp string, privateKey *bsvec.PrivateKey) ([]byte, error) {
	signatureHash := GetHandCashRequestSignatureHash(method, endpoint, body, timestamp)
	hash := sha256.Sum256([]byte(signatureHash))
	sig, err := privateKey.Sign(hash[:])
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

// GetHandCashRequestSignatureHash will return the signature hash
// should not return an error
func GetHandCashRequestSignatureHash(method string, endpoint string, body interface{}, timestamp string) string {
	if body == nil {
		body = "{}"
	} else {
		bodyBytes, _ := json.Marshal(body)
		body = fmt.Sprintf("%s", bodyBytes)
	}
	sigHash := fmt.Sprintf("%s\n%s\n%s", method, endpoint, timestamp)
	if body != nil {
		sigHash = sigHash + fmt.Sprintf("\n%s", body)
	}
	return sigHash
}

// GetSignedRequest returns the request with signature
func GetSignedRequest(method string, endpoint string, authToken string, body interface{}) (*SignedRequest, error) {
	// Match JS ISO time exactly
	JavascriptISOString := "2006-01-02T15:04:05.999Z07:00"
	timestamp := time.Now().UTC().Format(JavascriptISOString)

	tokenBytes, err := hex.DecodeString(authToken)
	if err != nil {
		return nil, err
	}
	privateKey, publicKey := bsvec.PrivKeyFromBytes(bsvec.S256(), tokenBytes)

	var requestSignature []byte
	if requestSignature, err = GetHandCashRequestSignature(method, endpoint, body, timestamp, privateKey); err != nil {
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

package handcash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bitcoinsv/bsvd/bsvec"
)

type requestBody struct {
	authToken string
}

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
	requestSignature, err = getHandCashRequestSignature(method, endpoint, timestamp, privateKey)
	if err != nil {
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

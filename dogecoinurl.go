package dogeconnectgo

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

// DogeURI holds the parsed components of a dogecoin: URI, optionally
// including Doge Connect payment parameters (dc and h).
type DogeURI struct {
	Address    string
	Amount     string
	ConnectURL string
	PubKeyHash []byte
}

// IsConnectURI reports whether this URI contains valid Doge Connect parameters.
func (u DogeURI) IsConnectURI() bool {
	return u.ConnectURL != "" && len(u.PubKeyHash) == 15
}

// ParseDogecoinURI parses a dogecoin: URI into its components.
// It validates the scheme, decodes the Doge Connect parameters if present,
// and returns an error for malformed URIs.
func ParseDogecoinURI(dogecoinURI string) (res DogeURI, err error) {
	// split URI into Scheme, Opaque (path), RawQuery
	url, err := url.Parse(dogecoinURI)
	if err != nil {
		return DogeURI{}, fmt.Errorf("invalid url: cannot parse: %w", err)
	}
	if url.Scheme != "dogecoin" {
		return DogeURI{}, fmt.Errorf("invalid url: not a 'dogecoin' url")
	}
	// address is the path-part of the URI
	res.Address = url.Opaque
	// parse RawQuery into a key-value Map
	// all of the following are optional in a dogecoin URI
	args := url.Query()
	res.Amount = args.Get("amount")
	res.ConnectURL = args.Get("dc")
	h := args.Get("h")
	// dc and h must both be present or both be absent.
	if (res.ConnectURL != "") != (h != "") {
		return DogeURI{}, fmt.Errorf("invalid url: 'dc' and 'h' parameters must both be present")
	}
	if h != "" {
		res.PubKeyHash, err = base64.URLEncoding.DecodeString(h)
		if err != nil {
			return DogeURI{}, fmt.Errorf("invalid url: cannot decode 'h' parameter: %w", err)
		}
		if len(res.PubKeyHash) != 15 {
			return DogeURI{}, fmt.Errorf("invalid url: 'h' must be 15 bytes, got %d", len(res.PubKeyHash))
		}
	}
	return
}

// DogecoinURI builds a dogecoin: URI with Doge Connect parameters.
// The connectURL should include the https:// prefix (which is stripped per spec).
// pubKey must be a 32-byte BIP-340 X-only public key.
func DogecoinURI(payToAddress string, amount string, connectURL string, pubKey []byte) (string, error) {
	// remove https:// prefix as per spec
	connectURL = strings.TrimPrefix(connectURL, "https://")
	pkHash, err := pubKeyHashStr(pubKey)
	if err != nil {
		return "", err
	}
	escURL := url.QueryEscape(connectURL)
	return fmt.Sprintf("dogecoin:%s?amount=%s&dc=%s&h=%s", payToAddress, amount, escURL, pkHash), nil
}

// pubKeyHashStr encodes the first 15 bytes of the SHA256 of the Gateway Public Key
// in URL-safe Base64 (RFC 4648); 15 is divisible by 3, which avoids Base64 padding.
func pubKeyHashStr(pubKey []byte) (string, error) {
	if len(pubKey) != 32 {
		return "", fmt.Errorf("invalid public key: must be 32 bytes, got %d", len(pubKey))
	}
	pkHash := sha256.Sum256(pubKey)
	return base64.URLEncoding.EncodeToString(pkHash[0:15]), nil // 15 bytes -> 20 chars
}

package dogeconnectgo

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

// The payment QR-code contains Connect URL (c) and Gateway Public Key Hash (h)
// as well as fallback dogecoin address and payment amount:
// dogecoin:DChs1c2YJZiZqhB13b8au44UCkcNGiiaDB?amount=43.61&dc=example.com%2Fdc%2F1234&h=3qaSfQoAQSj1U4DrZECG

type DogeURI struct {
	Address    string
	Amount     string
	ConnectURL string
	PubKeyHash []byte
}

func (u DogeURI) IsConnectURI() bool {
	return u.ConnectURL != "" && len(u.PubKeyHash) == 15
}

func ParseDogecoinURI(dogecoinURI string) (res DogeURI, err error) {
	// split URI into Scheme, Opaque (path), RawQuery
	url, err := url.Parse(dogecoinURI)
	if err != nil {
		return DogeURI{}, fmt.Errorf("invalid url: cannot parse: %v", err)
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
	res.PubKeyHash, _ = base64.URLEncoding.DecodeString(args.Get("h"))
	return
}

func DogecoinURI(payToAddress string, amount string, connectURL string, pubKey []byte) string {
	// remove https:// prefix as per spec
	connectURL = strings.TrimPrefix(connectURL, "https://")
	pkHash := pubKeyHash(pubKey)
	escURL := url.QueryEscape(connectURL)
	return fmt.Sprintf("dogecoin:%s?amount=%s&dc=%s&h=%s", payToAddress, amount, escURL, pkHash)
}

// pubKeyHash encodes the first 15 bytes of the SHA256 of the Gateway Public Key
// in URL-safe Base64 (RFC 4648); 15 is divisible by 3, which avoids Base64 padding.
func pubKeyHash(pubKey []byte) string {
	if len(pubKey) != 32 {
		panic("invalid public key")
	}
	pkHash := sha256.Sum256(pubKey)
	return base64.URLEncoding.EncodeToString(pkHash[0:15]) // 15 bytes -> 20 chars
}

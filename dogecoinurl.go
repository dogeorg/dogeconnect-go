package dogeconnectgo

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
)

// The payment QR-code contains Connect URL (c) and Gateway Public Key Hash (h)
// as well as a downlevel dogecoin address and payment amount:
// dogecoin:DChs1c2YJZiZqhB13b8au44UCkcNGiiaDB?amount=43.61&c=example.com%2Fdc%L4aNWS0oXE7TlCP-9enYbA0&h=3qaSfQoAQSj1U4DrZECG

// The `h` component is the first 15 bytes of the SHA256 of the Gateway Public Key
// encoded in URL-safe Base64. 15 is divisible by 3, which avoids Base64 padding.

func DogecoinURL(payToAddress string, amount string, connectURL string, pubKey []byte) string {
	pkHash := pubKeyHash(pubKey)
	escURL := url.QueryEscape(connectURL)
	return fmt.Sprintf("dogecoin:%s?amount=%s&c=%s&h=%s", payToAddress, amount, escURL, pkHash)
}

// pubKeyHash encodes the first 15 bytes of SHA256 of the Public Key in URL-safe Base64 (RFC 4648)
func pubKeyHash(pubKey []byte) string {
	if len(pubKey) != 32 {
		panic("invalid public key")
	}
	pkHash := sha256.Sum256(pubKey)
	return base64.URLEncoding.EncodeToString(pkHash[0:15]) // 15 bytes -> 20 chars
}

func ParseURL(dogecoinURL string) (payToAddress string, amount string, connectURL string, pubKey []byte) {
	return "", "", "", []byte{}
}

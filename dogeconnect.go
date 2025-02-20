package dogeconnectgo

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// SignPaymentRequest creates a signed ConnectEnvelope from a ConnectPayment.
func SignPaymentRequest(payment ConnectPayment, privKey []byte) (ConnectEnvelope, error) {
	// Derive the public key from the private key.
	priv, pub := btcec.PrivKeyFromBytes(privKey)
	defer priv.Zero()

	// Encode the ConnectPayment into JSON (encoded UTF-8 bytes)
	payload, err := json.Marshal(&payment)
	if err != nil {
		return ConnectEnvelope{}, err
	}

	// Double-SHA256 the encoded JSON UTF-8 bytes.
	hash1 := sha256.Sum256(payload)
	hash := sha256.Sum256(hash1[:])

	// BIP-340 Schnorr signature algorithm, sign using private key.
	sig, err := schnorr.Sign(priv, hash[:])
	if err != nil {
		return ConnectEnvelope{}, err
	}

	// Envelope wraps the Base64-encoded JSON payload with hex-encoded "X-only Public Key"
	// and hex-encoded BIP-340 signature.
	env := ConnectEnvelope{
		Version:   EnvelopeVersion,
		Payload:   base64.StdEncoding.EncodeToString(payload),        // Base64-encoded JSON payload
		PubKey:    hex.EncodeToString(pub.SerializeCompressed()[1:]), // BIP-340 X-only public key (32 bytes)
		Signature: hex.EncodeToString(sig.Serialize()),               // BIP-340 Schnorr signature (64 bytes)
	}
	return env, nil
}

// VerifyPaymentRequest decodes and verifies a signed ConnectPayment in a ConnectEnvelope.
// pubKeyHash is the `h` (hash) element from a valid DogeConnect URL.
func VerifyPaymentRequest(env ConnectEnvelope, pubKeyHash []byte) (ConnectPayment, error) {
	if env.Version != EnvelopeVersion {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: wrong version")
	}
	pub, err := hex.DecodeString(env.PubKey)
	if err != nil {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: malformed pubkey hex")
	}
	sig_b, err := hex.DecodeString(env.Signature)
	if err != nil {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: malformed signature hex")
	}
	payload, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: malformed base64 payload")
	}

	// SHA256 the public key.
	pubSha := sha256.Sum256(pub)
	if !bytes.Equal(pubKeyHash, pubSha[0:15]) {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: wrong public key")
	}

	// Double-SHA256 the encoded JSON payload bytes.
	hash1 := sha256.Sum256(payload)
	hash := sha256.Sum256(hash1[:])

	// BIP-340 X-only pubkey (lift_x function)
	pubkey, err := schnorr.ParsePubKey(pub)
	if err != nil {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: not a valid pubkey")
	}

	// Verify the BIP-340 Schnorr signature.
	sig, err := schnorr.ParseSignature(sig_b)
	if err != nil {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: not a valid signature")
	}
	if !sig.Verify(hash[:], pubkey) {
		return ConnectPayment{}, fmt.Errorf("invalid envelope: incorrect signature")
	}

	var payment ConnectPayment
	json.Unmarshal(payload, &payment)
	if payment.Type != PaymentRequestType {
		return ConnectPayment{}, fmt.Errorf("bad envelope: not a payment request")
	}
	return payment, nil
}

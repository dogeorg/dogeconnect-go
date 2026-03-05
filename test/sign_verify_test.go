package test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	dogeconnectgo "github.com/dogeorg/dogeconnect-go"
)

func newTestKey(t *testing.T) (privKey []byte, pubKeyCheck []byte) {
	t.Helper()
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	privKey = priv.Serialize()
	pubKey := priv.PubKey().SerializeCompressed()[1:]
	pubKeyHash := sha256.Sum256(pubKey)
	pubKeyCheck = pubKeyHash[0:15]
	return
}

func TestSignAndVerify(t *testing.T) {
	privKey, pubKeyCheck := newTestKey(t)

	payment := dogeconnectgo.ConnectPayment{
		Type:           dogeconnectgo.EnvelopeTypePayment,
		ID:             "101",
		Issued:         "2025-02-19T14:07:20+11:00",
		Timeout:        30,
		Relay:          "https://example.com/dc/1QAB-POvTh2R88nybE8Wwg",
		VendorIcon:     "https://static.example.com/vnd/1234/icon.png",
		VendorName:     "Example Co",
		VendorAddress:  "123 Example St",
		FeePerKB:       "0.01",
		MaxSize:        10000,
		VendorURL:      "https://example.com",
		VendorOrderURL: "https://example.com/orders/101",
		VendorOrderID:  "INV-101",
		OrderReference: "ORD-101",
		Note:           "Thanks for your order!",
		Total:          "420.69",
		Fees:           "6.31035",
		Taxes:          "0",
		Items: []dogeconnectgo.ConnectItem{
			{
				Type:        dogeconnectgo.ItemTypeItem,
				ID:          "123",
				Icon:        "https://static.example.com/vnd/1234/item/123.png",
				Name:        "Good Item",
				Description: "Best item in the store",
				UnitCount:   1,
				UnitCost:    "414.37965",
				Total:       "414.37965",
			},
		},
		Outputs: []dogeconnectgo.ConnectOutput{
			{Address: "DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY", Amount: "414.37965"},
			{Address: "DTG6vtXMfmjsitw4JkjJKb5SXH1hcNxn3n", Amount: "6.31035"},
		},
	}

	env, err := dogeconnectgo.SignPaymentRequest(payment, privKey)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	pay, err := dogeconnectgo.VerifyPaymentRequest(env, pubKeyCheck)
	if err != nil {
		t.Fatalf("failed to verify: %v", err)
	}

	if !reflect.DeepEqual(pay, payment) {
		t.Fatalf("verified payment is different:\n%v vs\n%v (expected)", pay, payment)
	}
}

func TestMinimalPaymentRoundTrip(t *testing.T) {
	privKey, pubKeyCheck := newTestKey(t)

	payment := dogeconnectgo.ConnectPayment{
		Type:       dogeconnectgo.EnvelopeTypePayment,
		ID:         "minimal-1",
		Issued:     "2025-06-01T00:00:00Z",
		Timeout:    60,
		Relay:      "https://example.com/dc/minimal",
		FeePerKB:   "0.01",
		MaxSize:    10000,
		VendorName: "Test",
		Total:      "10",
		Outputs: []dogeconnectgo.ConnectOutput{
			{Address: "DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY", Amount: "10"},
		},
	}

	env, err := dogeconnectgo.SignPaymentRequest(payment, privKey)
	if err != nil {
		t.Fatalf("failed to sign minimal payment: %v", err)
	}

	pay, err := dogeconnectgo.VerifyPaymentRequest(env, pubKeyCheck)
	if err != nil {
		t.Fatalf("failed to verify minimal payment: %v", err)
	}

	if !reflect.DeepEqual(pay, payment) {
		t.Fatalf("minimal payment round-trip mismatch:\ngot:  %+v\nwant: %+v", pay, payment)
	}
}

func TestMalformedPayloadReturnsError(t *testing.T) {
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	pubKey := priv.PubKey().SerializeCompressed()[1:]
	pubKeyHash := sha256.Sum256(pubKey)
	pubKeyCheck := pubKeyHash[0:15]

	// Create a validly-signed envelope with invalid JSON as payload.
	badPayload := []byte("{not valid json!!")
	hash1 := sha256.Sum256(badPayload)
	hash := sha256.Sum256(hash1[:])

	sig, err := schnorr.Sign(priv, hash[:])
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	env := dogeconnectgo.ConnectEnvelope{
		Version:   dogeconnectgo.EnvelopeVersion,
		Payload:   base64.StdEncoding.EncodeToString(badPayload),
		PubKey:    hex.EncodeToString(pubKey),
		Signature: hex.EncodeToString(sig.Serialize()),
	}

	_, err = dogeconnectgo.VerifyPaymentRequest(env, pubKeyCheck)
	if err == nil {
		t.Fatal("expected error for malformed payload JSON, got nil")
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("error should mention 'malformed', got: %v", err)
	}
}

func TestMinimalPaymentJSONShape(t *testing.T) {
	payment := dogeconnectgo.ConnectPayment{
		Type:       dogeconnectgo.EnvelopeTypePayment,
		ID:         "shape-1",
		Issued:     "2025-06-01T00:00:00Z",
		Relay:      "https://example.com/dc/shape",
		VendorName: "Test",
		Total:      "5",
		Outputs: []dogeconnectgo.ConnectOutput{
			{Address: "DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY", Amount: "5"},
		},
	}

	data, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	// All fields should be present in JSON (no omitempty on string fields).
	present := []string{
		"type", "id", "issued", "timeout", "relay", "relay_token",
		"fee_per_kb", "max_size",
		"vendor_icon", "vendor_name", "vendor_address",
		"vendor_url", "vendor_order_url", "vendor_order_id",
		"order_reference", "note",
		"total", "fees", "taxes",
		"fiat_total", "fiat_tax", "fiat_currency",
		"items", "outputs",
	}
	for _, key := range present {
		if _, ok := m[key]; !ok {
			t.Errorf("key %q should be present in JSON, but is absent", key)
		}
	}

	// Optional string fields should be empty strings when unset.
	emptyStrings := []string{
		"relay_token", "vendor_icon", "vendor_address",
		"vendor_url", "vendor_order_url", "vendor_order_id",
		"order_reference", "note", "fees", "taxes",
		"fiat_total", "fiat_tax", "fiat_currency",
	}
	for _, key := range emptyStrings {
		if m[key] != "" {
			t.Errorf("optional key %q should be empty string, got %v", key, m[key])
		}
	}
}

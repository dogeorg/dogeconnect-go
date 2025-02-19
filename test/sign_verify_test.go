package test

import (
	"crypto/sha256"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	dogeconnectgo "github.com/dogeorg/dogeconnect-go"
)

func TestSignAndVerify(t *testing.T) {
	// generate a private key for testing.
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		panic(err)
	}
	privKey := priv.Serialize()
	pubKey := priv.PubKey().SerializeCompressed()[1:] // X-only pubkey

	payment := dogeconnectgo.ConnectPayment{
		Type:          dogeconnectgo.PaymentRequestType,
		ID:            "101",
		Issued:        "2025-02-19T14:07:20+11:00",
		Timeout:       30,
		Gateway:       "https://example.com/dc/1QAB-POvTh2R88nybE8Wwg",
		VendorIcon:    "https://static.example.com/vnd/1234/icon.png",
		VendorName:    "Example Co",
		VendorAddress: "123 Example St",
		Total:         "420.69",
		Fees:          "6.31035",
		Taxes:         "0",
		Items: []dogeconnectgo.ConnectItem{
			{
				Type:        "item",
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
			{
				Address: "DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY",
				Amount:  "414.37965",
			},
			{
				Address: "DTG6vtXMfmjsitw4JkjJKb5SXH1hcNxn3n",
				Amount:  "6.31035",
			},
		},
	}

	env, err := dogeconnectgo.SignPaymentRequest(payment, privKey)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	pubKeyHash := sha256.Sum256(pubKey)
	pubKeyCheck := pubKeyHash[0:15]

	pay, err := dogeconnectgo.VerifyPaymentRequest(env, pubKeyCheck)
	if err != nil {
		t.Fatalf("failed to verify: %v", err)
	}

	if !reflect.DeepEqual(pay, payment) {
		t.Fatalf("verified payment is different:\n%v vs\n%v (expected)", pay, payment)
	}
}

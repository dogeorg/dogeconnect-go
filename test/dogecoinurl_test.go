package test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	dogeconnectgo "github.com/dogeorg/dogeconnect-go"
)

func TestDogecoinURL(t *testing.T) {
	payTo := "DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY"
	amount := "12.25"
	connectURL := "example.com/dc/1QAB-POvTh2R88nybE8Wwg"
	pubKey, _ := hex.DecodeString("6c52b17752f469c5411b977ba64725d40174d16e780b709b2aff68e0f5abfc50")
	expect := "dogecoin:DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY?amount=12.25&dc=example.com%2Fdc%2F1QAB-POvTh2R88nybE8Wwg&h=72b-LVh5K_mm7zyN9PXO"
	uri := dogeconnectgo.DogecoinURI(payTo, amount, "https://"+connectURL, pubKey)
	if uri != expect {
		t.Errorf("incorrect uri:\n%v (found)\n%v (expected)", uri, expect)
	}
	res, err := dogeconnectgo.ParseDogecoinURI(uri)
	if err != nil {
		t.Errorf("failed to parse uri: %v", err)
	}
	if !res.IsConnectURI() {
		t.Errorf("IsConnectURI should return true")
	}
	if res.Address != payTo {
		t.Errorf("wrong address: %v vs %v", res.Address, payTo)
	}
	if res.Amount != amount {
		t.Errorf("wrong amount: %v vs %v", res.Amount, amount)
	}
	if res.ConnectURL != connectURL {
		t.Errorf("wrong connect URL: %v vs %v", res.ConnectURL, connectURL)
	}
	pubSha := sha256.Sum256(pubKey)
	if !bytes.Equal(res.PubKeyHash, pubSha[0:15]) {
		t.Errorf("wrong pubkey hash:\n%x vs\n%x", res.PubKeyHash, pubSha[0:15])
	}
}

func TestSlashInDC(t *testing.T) {
	connectURL := "example.com/dc/1QAB"
	uri := "dogecoin:DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY?amount=12.25&dc=example.com/dc/1QAB&h=72b-LVh5K_mm7zyN9PXO"
	res, err := dogeconnectgo.ParseDogecoinURI(uri)
	if err != nil {
		t.Errorf("failed to parse uri: %v", err)
	}
	if !res.IsConnectURI() {
		t.Errorf("IsConnectURI should return true")
	}
	if res.ConnectURL != connectURL {
		t.Errorf("wrong connect URL: %v vs %v", res.ConnectURL, connectURL)
	}
}

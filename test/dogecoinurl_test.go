package test

import (
	"encoding/hex"
	"testing"

	dogeconnectgo "github.com/dogeorg/dogeconnect-go"
)

func TestDogecoinURL(t *testing.T) {
	payTo := "DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY"
	amount := "12.25"
	connectURL := "example.com/dc/1QAB-POvTh2R88nybE8Wwg"
	pubKey, _ := hex.DecodeString("6c52b17752f469c5411b977ba64725d40174d16e780b709b2aff68e0f5abfc50")
	expect := "dogecoin:DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY?amount=12.25&c=example.com%2Fdc%2F1QAB-POvTh2R88nybE8Wwg&h=72b-LVh5K_mm7zyN9PXO"
	url := dogeconnectgo.DogecoinURL(payTo, amount, connectURL, pubKey)
	if url != expect {
		t.Errorf("incorrect url:\n%v (found)\n%v (expected)", url, expect)
	}
}

# dogeconnect-go

Go library for the [Doge Connect](https://github.com/dogeorg/connect) payment protocol — types, signing, verification, and parsing.

## Install

```
go get github.com/dogeorg/dogeconnect-go
```

## Usage

### Sign a payment request (relay side)

```go
payment := dogeconnectgo.ConnectPayment{
    Type:       dogeconnectgo.EnvelopeTypePayment,
    ID:         "order-123",
    Issued:     time.Now().Format(time.RFC3339),
    Timeout:    300,
    Relay:      "https://relay.example.com/pay",
    FeePerKB:   "0.01",
    MaxSize:    100000,
    VendorName: "Example Shop",
    Total:      "42.50000000",
    Items:      []dogeconnectgo.ConnectItem{ /* ... */ },
    Outputs:    []dogeconnectgo.ConnectOutput{{Address: "D...", Amount: "42.50000000"}},
}

envelope, err := dogeconnectgo.SignPaymentRequest(payment, privateKeyBytes)
```

### Verify and parse a payment request (wallet side)

```go
// Verify the signed envelope against the public key hash from the QR code.
payment, err := dogeconnectgo.VerifyPaymentRequest(envelope, pubKeyHash)
if err != nil {
    // signature or envelope invalid
}

// Parse into native Go types (best-effort: struct is populated even if some fields have errors).
parsed, fieldErrs := payment.Parse()
if len(fieldErrs) > 0 {
    for _, e := range fieldErrs {
        log.Printf("field %s: %s", e.Field, e.Message)
    }
}

// Use parsed values directly.
deadline := parsed.IssuedTime.Add(time.Duration(parsed.Timeout) * time.Second)
fmt.Printf("Pay %s koinu to %d outputs before %v\n",
    parsed.TotalKoinu, len(parsed.ParsedOutputs), deadline)

// Raw string fields are still accessible via embedding:
fmt.Println(parsed.VendorName) // from ConnectPayment
fmt.Println(parsed.Total)      // original "42.50000000" string
```

### Validate a payment submission (relay side)

```go
var sub dogeconnectgo.PaymentSubmission
json.NewDecoder(r.Body).Decode(&sub)

parsed, fieldErrs := sub.Parse()
if len(fieldErrs) > 0 {
    // reject
}
// parsed.TxBytes contains the decoded transaction
```

### Generate and parse Dogecoin URIs

```go
// Generate a QR-code URI with Connect parameters.
uri := dogeconnectgo.DogecoinURI("D...", "42.50", "relay.example.com/pay/123", pubKeyBytes)
// → dogecoin:D...?amount=42.50&dc=relay.example.com%2Fpay%2F123&h=...

// Parse a URI back.
parsed, err := dogeconnectgo.ParseDogecoinURI(uri)
if parsed.IsConnectURI() {
    // fetch envelope from parsed.ConnectURL, verify with parsed.PubKeyHash
}
```

## Parsed Types

Each protocol type with complex fields has a `Parse()` method returning `(Parsed*, []FieldError)`:

| Raw Type | Parsed Type | Parsed Fields |
|---|---|---|
| `ConnectEnvelope` | `ParsedEnvelope` | `PayloadBytes`, `PubKeyBytes`, `SignatureBytes` |
| `ConnectPayment` | `ParsedPayment` | `IssuedTime`, `TotalKoinu`, `FeePerKBKoinu`, `FeesKoinu`, `TaxesKoinu`, `ParsedItems`, `ParsedOutputs` |
| `ConnectItem` | `ParsedItem` | `UnitCostKoinu`, `TotalKoinu`, `TaxKoinu` |
| `ConnectOutput` | `ParsedOutput` | `AmountKoinu` |
| `PaymentSubmission` | `ParsedSubmission` | `TxBytes` |
| `PaymentStatusResponse` | `ParsedStatusResponse` | `TxIDBytes`, `ConfirmedAtTime` |

`StatusQuery` and `ErrorResponse` only have a `Validate()` method (no complex fields to parse).

Koinu amounts use the `koinu.Koinu` type (`int64`, 1 DOGE = 100,000,000 koinu).

package dogeconnectgo

// DogeConnect protocol.
const EnvelopeVersion = "1.0"

// EnvelopeType is the type of a Connect Envelope payload.
type EnvelopeType string

const EnvelopeTypePayment EnvelopeType = "payment"

// ItemType is the type of a line item.
type ItemType string

const (
	ItemTypeItem     ItemType = "item"
	ItemTypeTax      ItemType = "tax"
	ItemTypeFee      ItemType = "fee"
	ItemTypeShipping ItemType = "shipping"
	ItemTypeDiscount ItemType = "discount"
	ItemTypeDonation ItemType = "donation"
)

// PaymentStatus is the status of a submitted payment.
type PaymentStatus string

const (
	PaymentStatusUnpaid    PaymentStatus = "unpaid"
	PaymentStatusAccepted  PaymentStatus = "accepted"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusDeclined  PaymentStatus = "declined"
)

// ConnectEnvelope is a signed wrapper around a Connect Payment payload.
type ConnectEnvelope struct {
	Version   string `json:"version"` // MUST be EnvelopeVersion
	Payload   string `json:"payload"` // Base64-encoded JSON payload (e.g. ConnectPayment)
	PubKey    string `json:"pubkey"`  // Relay Public Key, BIP-340 Schnorr X-only (32 bytes, hex)
	Signature string `json:"sig"`     // Payload Signature, BIP-340 Schnorr (64 bytes, hex)
}

// ConnectPayment represents the decoded payload inside a Connect Envelope.
// 8-DP string is a decimal string with up to 8 decimal places representing a Dogecoin amount.
//
// Optional string fields: relays SHOULD include them as "" when no value is available;
// wallets MUST treat missing and "" identically.
type ConnectPayment struct {
	Type           EnvelopeType    `json:"type"`             // EnvelopeType enum; MUST be "payment"
	ID             string          `json:"id"`               // Relay-unique payment ID
	Issued         string          `json:"issued"`           // RFC 3339 Timestamp (2006-01-02T15:04:05-07:00)
	Timeout        int             `json:"timeout"`          // Seconds; do not submit payment Tx after this time (Issued+Timeout)
	Relay          string          `json:"relay"`            // Payment Relay URL
	RelayToken     string          `json:"relay_token"`      // Opaque relay-generated token; wallet MUST echo in Payment Submission if present (optional)
	FeePerKB       string          `json:"fee_per_kb"`       // Minimum fee per 1000 bytes in payment tx, 8-DP string
	MaxSize        int             `json:"max_size"`         // Maximum size in bytes of payment tx
	VendorIcon     string          `json:"vendor_icon"`      // Vendor icon URL, JPG or PNG; wallet SHOULD use a placeholder when not provided (optional)
	VendorName     string          `json:"vendor_name"`      // vendor display name
	VendorAddress  string          `json:"vendor_address"`   // vendor business address (optional)
	VendorURL      string          `json:"vendor_url"`       // Vendor website URL (optional)
	VendorOrderURL string          `json:"vendor_order_url"` // URL to view order on vendor's site (optional)
	VendorOrderID  string          `json:"vendor_order_id"`  // Vendor's unique order identifier (optional)
	OrderReference string          `json:"order_reference"`  // Short customer-facing order identifier (optional)
	Note           string          `json:"note"`             // Free-text note from vendor to customer (optional)
	Total          string          `json:"total"`            // Total amount including fees and taxes, 8-DP string
	Fees           string          `json:"fees"`             // Fee subtotal, 8-DP string (optional)
	Taxes          string          `json:"taxes"`            // Taxes subtotal, 8-DP string (optional)
	FiatTotal      string          `json:"fiat_total"`       // Total amount in fiat currency (optional)
	FiatTax        string          `json:"fiat_tax"`         // Taxes in fiat currency (optional)
	FiatCurrency   string          `json:"fiat_currency"`    // ISO 4217 currency code (required with fiat_total/fiat_tax) (conditional)
	Items          []ConnectItem   `json:"items"`            // List of line items to display
	Outputs        []ConnectOutput `json:"outputs"`          // List of outputs to pay
}

// ConnectItem is a line item within a Connect Payment.
type ConnectItem struct {
	Type        ItemType `json:"type"`  // item, tax, fee, shipping, discount, donation
	ID          string   `json:"id"`    // unique item ID or SKU
	Icon        string   `json:"icon"`  // icon URL, JPG or PNG; wallet SHOULD use a placeholder when not provided (optional)
	Name        string   `json:"name"`  // name to display
	Description string   `json:"desc"`  // item description to display (optional)
	UnitCount   int      `json:"count"` // number of units >= 1
	UnitCost    string   `json:"unit"`  // unit price, 8-DP string
	Total       string   `json:"total"` // count x unit, 8-DP string
	Tax         string   `json:"tax"`   // tax on this item, 8-DP string (optional)
}

// ConnectOutput is a transaction output the wallet must pay.
type ConnectOutput struct {
	Address string `json:"address"` // Dogecoin Address
	Amount  string `json:"amount"`  // Amount, 8-DP string
}

// PaymentSubmission is the wallet's submission to the relay's pay endpoint.
type PaymentSubmission struct {
	ID         string `json:"id"`          // payment ID from ConnectPayment
	Tx         string `json:"tx"`          // hex-encoded signed transaction
	Refund     string `json:"refund"`      // Dogecoin address for refunds; input addresses may not return to customer (e.g. exchange), so this provides a guaranteed return path (optional, recommended)
	RelayToken string `json:"relay_token"` // Opaque relay token; required when Connect Payment contained relay_token (conditional)
}

// PaymentStatusResponse is the relay's response from both the pay and status endpoints.
type PaymentStatusResponse struct {
	ID          string        `json:"id"`                     // payment ID
	Status      PaymentStatus `json:"status"`                 // enum
	Reason      string        `json:"reason,omitempty"`       // Reason for decline; present when declined
	TxID        string        `json:"txid,omitempty"`         // Hex-encoded tx ID; present when accepted or confirmed
	ConfirmedAt string        `json:"confirmed_at,omitempty"` // RFC 3339 timestamp; present when confirmed
	Required    *int          `json:"required,omitempty"`     // Block confirmations required; present when accepted or confirmed
	Confirmed   *int          `json:"confirmed,omitempty"`    // Current block confirmations; present when accepted or confirmed
	DueSec      *int          `json:"due_sec,omitempty"`      // Estimated seconds until confirmed; present when accepted or confirmed
}

// StatusQuery is submitted to the relay's status endpoint to query payment status.
type StatusQuery struct {
	ID string `json:"id"` // Relay-unique payment ID from Connect Payment
}

// ErrorCode is the type of error in an ErrorResponse.
type ErrorCode string

const (
	ErrorCodeNotFound       ErrorCode = "not_found"
	ErrorCodeExpired        ErrorCode = "expired"
	ErrorCodeInvalidTx      ErrorCode = "invalid_tx"
	ErrorCodeInvalidOutputs ErrorCode = "invalid_outputs"
	ErrorCodeInvalidToken   ErrorCode = "invalid_token"
)

// ErrorResponse is returned by the relay when a request fails.
type ErrorResponse struct {
	Error   ErrorCode `json:"error"`   // ErrorCode enum
	Message string    `json:"message"` // Human-readable error detail
}

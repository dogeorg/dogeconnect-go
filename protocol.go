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
	Version   string `json:"version"` // Protocol version, MUST be "1.0"
	Payload   string `json:"payload"` // Base64-encoded JSON payload (Connect Payment)
	PubKey    string `json:"pubkey"`  // Relay public key, BIP-340 Schnorr X-only (32 bytes, hex)
	Signature string `json:"sig"`     // Payload signature, BIP-340 Schnorr (64 bytes, hex)
}

// ConnectPayment represents the decoded payload inside a Connect Envelope.
// 8-DP string is a decimal string with up to 8 decimal places representing a Dogecoin amount.
type ConnectPayment struct {
	Type           EnvelopeType    `json:"type"`                       // EnvelopeType enum; MUST be "payment"
	ID             string          `json:"id"`                         // Relay-unique payment ID
	Issued         string          `json:"issued"`                     // RFC 3339 timestamp
	Timeout        int             `json:"timeout"`                    // Timeout in seconds; do not pay after issued + timeout
	Relay          string          `json:"relay"`                      // Payment Relay base URL
	RelayToken     string          `json:"relay_token,omitempty"`      // Opaque relay-generated token (optional)
	FeePerKB       string          `json:"fee_per_kb"`                 // Minimum fee per 1000 bytes, 8-DP string
	MaxSize        int             `json:"max_size"`                   // Maximum size in bytes of payment tx
	VendorIcon     string          `json:"vendor_icon,omitempty"`      // Vendor icon URL, JPG or PNG (optional)
	VendorName     string          `json:"vendor_name"`                // Vendor display name
	VendorAddress  string          `json:"vendor_address,omitempty"`   // Vendor business address (optional)
	VendorURL      string          `json:"vendor_url,omitempty"`       // Vendor website URL (optional)
	VendorOrderURL string          `json:"vendor_order_url,omitempty"` // URL to view order on vendor's site (optional)
	VendorOrderID  string          `json:"vendor_order_id,omitempty"`  // Vendor's unique order identifier (optional)
	OrderReference string          `json:"order_reference,omitempty"`  // Short customer-facing order identifier (optional)
	Note           string          `json:"note,omitempty"`             // Free-text note from vendor to customer (optional)
	Total          string          `json:"total"`                      // Total including fees and taxes, 8-DP string
	Fees           string          `json:"fees,omitempty"`             // Fees subtotal, 8-DP string (optional)
	Taxes          string          `json:"taxes,omitempty"`            // Taxes subtotal, 8-DP string (optional)
	FiatTotal      string          `json:"fiat_total,omitempty"`       // Total in fiat currency, decimal string (optional)
	FiatTax        string          `json:"fiat_tax,omitempty"`         // Taxes in fiat currency, decimal string (optional)
	FiatCurrency   string          `json:"fiat_currency,omitempty"`    // ISO 4217 currency code; required when fiat_total or fiat_tax is present
	Items          []ConnectItem   `json:"items"`                      // List of Connect Items
	Outputs        []ConnectOutput `json:"outputs"`                    // List of Connect Outputs
}

// ConnectItem is a line item within a Connect Payment.
type ConnectItem struct {
	Type        ItemType `json:"type"`           // ItemType enum
	ID          string   `json:"id"`             // Unique item ID or SKU
	Icon        string   `json:"icon,omitempty"` // Icon URL, JPG or PNG (optional)
	Name        string   `json:"name"`           // Display name
	Description string   `json:"desc,omitempty"` // Item description (optional)
	UnitCount   int      `json:"count"`          // Number of units (>= 1)
	UnitCost    string   `json:"unit"`           // Unit price, 8-DP string
	Total       string   `json:"total"`          // count x unit, 8-DP string
	Tax         string   `json:"tax,omitempty"`  // Tax on this item, 8-DP string (optional)
}

// ConnectOutput is a transaction output the wallet must pay.
type ConnectOutput struct {
	Address string `json:"address"` // Dogecoin address
	Amount  string `json:"amount"`  // Amount to pay, 8-DP string
}

// PaymentSubmission is the wallet's submission to the relay's pay endpoint.
type PaymentSubmission struct {
	ID         string `json:"id"`                    // Relay-unique payment ID from Connect Payment
	Tx         string `json:"tx"`                    // Hex-encoded signed Dogecoin transaction
	Refund     string `json:"refund,omitempty"`      // Dogecoin address for refunds (recommended)
	RelayToken string `json:"relay_token,omitempty"` // Opaque relay token; required when Connect Payment contained relay_token
}

// PaymentStatusResponse is the relay's response from both the pay and status endpoints.
type PaymentStatusResponse struct {
	ID          string        `json:"id"`                     // Relay-unique payment ID
	Status      PaymentStatus `json:"status"`                 // PaymentStatus enum
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

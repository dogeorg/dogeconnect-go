package dogeconnectgo

// DogeConnect protocol.
const EnvelopeVersion = "1.0"

// PaymentType is the type of a payment request.
type PaymentType string

const PaymentTypePayment PaymentType = "payment"

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

type ConnectEnvelope struct {
	Version   string `json:"version"` // MUST be EnvelopeVersion
	Payload   string `json:"payload"` // Base64-encoded JSON payload (e.g. ConnectPayment)
	PubKey    string `json:"pubkey"`  // Gateway Public Key, BIP-340 Schnorr X-only (32 bytes)
	Signature string `json:"sig"`     // Payload Signature, BIP-340 Schnorr (64 bytes)
}

// ConnectPayment represents a payment request Payload
// 8-DP string is Koinu value, a DECIMAL string with up to 8 decimal places
type ConnectPayment struct {
	Type          PaymentType     `json:"type"`                     // MUST be PaymentTypePayment
	ID            string          `json:"id"`                       // Gateway unique payment-request ID
	Issued        string          `json:"issued"`                   // RFC 3339 Timestamp (2006-01-02T15:04:05-07:00)
	Timeout       int             `json:"timeout,omitempty"`        // Seconds; do not submit payment Tx after this time (Issued+Timeout)
	Relay         string          `json:"relay"`                    // Payment Relay URL, https://example.com/dc
	FeePerKB      string          `json:"fee_per_kb,omitempty"`     // Minimum fee per 1000 bytes in payment tx, 8-DP string
	MaxSize       int             `json:"max_size,omitempty"`       // Maximum size in bytes of payment tx
	VendorIcon    string          `json:"vendor_icon,omitempty"`    // vendor icon URL, SHOULD be https:// JPG or PNG
	VendorName    string          `json:"vendor_name"`              // vendor display name
	VendorAddress string          `json:"vendor_address,omitempty"` // vendor business address (optional)
	Total         string          `json:"total"`                    // Total amount including fees and taxes, 8-DP string
	Fees          string          `json:"fees,omitempty"`           // Fee subtotal, 8-DP string
	Taxes         string          `json:"taxes,omitempty"`          // Taxes subtotal, 8-DP string
	FiatTotal     string          `json:"fiat_total,omitempty"`     // Total amount in fiat currency (optional)
	FiatTax       string          `json:"fiat_tax,omitempty"`       // Taxes in fiat currency (optional)
	FiatCurrency  string          `json:"fiat_currency,omitempty"`  // ISO 4217 currency code (required with fiat_total/fiat_tax)
	Items         []ConnectItem   `json:"items"`                    // List of line items to display
	Outputs       []ConnectOutput `json:"outputs"`                  // List of outputs to pay
}

type ConnectItem struct {
	Type        ItemType `json:"type"`           // item, tax, fee, shipping, discount, donation
	ID          string   `json:"id"`             // unique item ID or SKU
	Icon        string   `json:"icon,omitempty"` // icon URL, SHOULD be https:// JPG or PNG
	Name        string   `json:"name"`           // name to display
	Description string   `json:"desc,omitempty"` // item description to display
	UnitCount   int      `json:"count"`          // number of units >= 1
	UnitCost    string   `json:"unit"`           // unit price, 8-DP string
	Total       string   `json:"total"`          // count x unit, 8-DP string
	Tax         string   `json:"tax,omitempty"`  // tax on this item, 8-DP string (optional)
}

type ConnectOutput struct {
	Address string `json:"address"` // Dogecoin Address
	Amount  string `json:"amount"`  // Amount, 8-DP string
}

// PaymentSubmission is sent by a wallet to the relay (POST relay/pay).
type PaymentSubmission struct {
	ID     string `json:"id"`               // payment ID from ConnectPayment
	Tx     string `json:"tx"`               // hex-encoded signed transaction
	Refund string `json:"refund,omitempty"` // refund address (optional)
}

// PaymentStatusResponse is returned by the relay to the wallet.
type PaymentStatusResponse struct {
	ID          string        `json:"id"`                     // payment ID
	Status      PaymentStatus `json:"status"`                 // enum
	Reason      string        `json:"reason,omitempty"`       // decline reason
	TxID        string        `json:"txid,omitempty"`         // hex tx ID
	ConfirmedAt string        `json:"confirmed_at,omitempty"` // RFC 3339
	Required    *int          `json:"required,omitempty"`     // confirmations required
	Confirmed   *int          `json:"confirmed,omitempty"`    // confirmations so far
	DueSec      *int          `json:"due_sec,omitempty"`      // seconds until confirmed
}

// StatusQuery is sent by a wallet to the relay (POST relay/status).
type StatusQuery struct {
	ID string `json:"id"`
}

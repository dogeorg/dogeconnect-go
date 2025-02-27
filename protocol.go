package dogeconnectgo

// DogeConnect protocol.
const (
	EnvelopeVersion    = "1.0"
	PaymentRequestType = "payment"
)

type ConnectEnvelope struct {
	Version   string `json:"version"` // MUST be EnvelopeVersion
	Payload   string `json:"payload"` // Base64-encoded JSON payload (e.g. ConnectPayment)
	PubKey    string `json:"pubkey"`  // Gateway Public Key, BIP-340 Schnorr X-only (32 bytes)
	Signature string `json:"sig"`     // Payload Signature, BIP-340 Schnorr (64 bytes)
}

// ConnectPayment represents a payment request Payload
type ConnectPayment struct {
	Type          string          `json:"type"`           // MUST be PaymentRequestType
	ID            string          `json:"id"`             // Gateway unique payment-request ID
	Issued        string          `json:"issued"`         // RFC 3339 Timestamp (2006-01-02T15:04:05-07:00)
	Timeout       int             `json:"timeout"`        // Seconds; do not submit payment Tx after this time (Issued+Timeout)
	Relay         string          `json:"relay"`          // Payment Relay URL, https://example.com/dc
	VendorIcon    string          `json:"vendor_icon"`    // vendor icon URL, SHOULD be https:// JPG or PNG
	VendorName    string          `json:"vendor_name"`    // vendor display name
	VendorAddress string          `json:"vendor_address"` // vendor business address (optional)
	Total         string          `json:"total"`          // Total amount including fees and taxes, DECMIAL string
	Fees          string          `json:"fees"`           // Fee subtotal, DECMIAL string
	Taxes         string          `json:"taxes"`          // Taxes subtotal, DECMIAL string
	FiatTotal     string          `json:"fiat_total"`     // Total amount in fiat currency (optional)
	FiatCurrency  string          `json:"fiat_currency"`  // ISO 4217 currency code (required with fiat_total)
	Items         []ConnectItem   `json:"items"`          // List of line items to display
	Outputs       []ConnectOutput `json:"outputs"`        // List of outputs to pay
}

type ConnectItem struct {
	Type        string `json:"type"`  // item, tax, fee, shipping, discount, donation
	ID          string `json:"id"`    // unique item ID or SKU
	Icon        string `json:"icon"`  // icon URL, SHOULD be https:// JPG or PNG
	Name        string `json:"name"`  // name to display
	Description string `json:"desc"`  // item description to display
	UnitCount   int    `json:"count"` // number of units >= 1
	UnitCost    string `json:"cost"`  // unit price, DECMIAL string
	Total       string `json:"total"` // count x cost, DECMIAL string
}

type ConnectOutput struct {
	Address string `json:"address"` // Dogecoin Address
	Amount  string `json:"amount"`  // Amount, DECMIAL string
}

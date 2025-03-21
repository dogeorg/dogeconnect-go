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
// 8-DP string is Koinu value, a DECIMAL string with up to 8 decimal places
type ConnectPayment struct {
	Type          string          `json:"type"`           // MUST be PaymentRequestType
	ID            string          `json:"id"`             // Gateway unique payment-request ID
	Issued        string          `json:"issued"`         // RFC 3339 Timestamp (2006-01-02T15:04:05-07:00)
	Timeout       int             `json:"timeout"`        // Seconds; do not submit payment Tx after this time (Issued+Timeout)
	Relay         string          `json:"relay"`          // Payment Relay URL, https://example.com/dc
	FeePerKB      string          `json:"fee_per_kb"`     // Minimum fee per 1000 bytes in payment tx, 8-DP string
	MaxSize       int             `json:"max_size"`       // Maximum size in bytes of payment tx
	VendorIcon    string          `json:"vendor_icon"`    // vendor icon URL, SHOULD be https:// JPG or PNG
	VendorName    string          `json:"vendor_name"`    // vendor display name
	VendorAddress string          `json:"vendor_address"` // vendor business address (optional)
	Total         string          `json:"total"`          // Total amount including fees and taxes, 8-DP string
	Fees          string          `json:"fees"`           // Fee subtotal, 8-DP string
	Taxes         string          `json:"taxes"`          // Taxes subtotal, 8-DP string
	FiatTotal     string          `json:"fiat_total"`     // Total amount in fiat currency (optional)
	FiatTax       string          `json:"fiat_tax"`       // Taxes in fiat currency (optional)
	FiatCurrency  string          `json:"fiat_currency"`  // ISO 4217 currency code (required with fiat_total/fiat_tax)
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
	UnitCost    string `json:"unit"`  // unit price, 8-DP string
	Total       string `json:"total"` // count x unit, 8-DP string
	Tax         string `json:"tax"`   // tax on this item, DECMIAL string (optional)
}

type ConnectOutput struct {
	Address string `json:"address"` // Dogecoin Address
	Amount  string `json:"amount"`  // Amount, 8-DP string
}

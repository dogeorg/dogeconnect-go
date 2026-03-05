package dogeconnectgo

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dogeorg/dogeconnect-go/koinu"
)

// FieldError describes a validation error on a specific field.
type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) String() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// FieldErrors collects validation errors.
type FieldErrors []FieldError

// Add appends a non-nil FieldError.
func (fe *FieldErrors) Add(e *FieldError) {
	if e != nil {
		*fe = append(*fe, *e)
	}
}

// Err returns nil if there are no errors, or a combined error.
func (fe FieldErrors) Err() error {
	if len(fe) == 0 {
		return nil
	}
	msg := fe[0].String()
	for _, e := range fe[1:] {
		msg += "; " + e.String()
	}
	return fmt.Errorf("validation failed: %s", msg)
}

func fieldErr(field, message string) *FieldError {
	return &FieldError{field, message}
}

// Parse helpers — each returns the parsed value and a single *FieldError (nil on success).

func checkNonEmpty(field, value string) *FieldError {
	if value == "" {
		return fieldErr(field, "required")
	}
	return nil
}

func parseHexBytes(field, value string, n int) ([]byte, *FieldError) {
	if value == "" {
		return nil, fieldErr(field, "required")
	}
	b, err := hex.DecodeString(value)
	if err != nil {
		return nil, fieldErr(field, "invalid hex")
	}
	if len(b) != n {
		return b, fieldErr(field, fmt.Sprintf("must be %d bytes, got %d", n, len(b)))
	}
	return b, nil
}

func parseRequiredHex(field, value string) ([]byte, *FieldError) {
	if value == "" {
		return nil, fieldErr(field, "required")
	}
	b, err := hex.DecodeString(value)
	if err != nil {
		return nil, fieldErr(field, "invalid hex")
	}
	return b, nil
}

func parseOptionalHex(field, value string) ([]byte, *FieldError) {
	if value == "" {
		return nil, nil
	}
	b, err := hex.DecodeString(value)
	if err != nil {
		return nil, fieldErr(field, "invalid hex")
	}
	return b, nil
}

func parseRequiredKoinu(field, value string) (koinu.Koinu, *FieldError) {
	if value == "" {
		return 0, fieldErr(field, "required")
	}
	k, err := koinu.ParseKoinu(value)
	if err != nil {
		return 0, fieldErr(field, fmt.Sprintf("invalid koinu value: %s", err))
	}
	return k, nil
}

func parseOptionalKoinu(field, value string) (koinu.Koinu, *FieldError) {
	if value == "" {
		return 0, nil
	}
	k, err := koinu.ParseKoinu(value)
	if err != nil {
		return 0, fieldErr(field, fmt.Sprintf("invalid koinu value: %s", err))
	}
	return k, nil
}

func parseTimestamp(field, value string) (time.Time, *FieldError) {
	if value == "" {
		return time.Time{}, fieldErr(field, "required")
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fieldErr(field, "invalid RFC 3339 timestamp")
	}
	return t, nil
}

func parseOptionalTimestamp(field, value string) (time.Time, *FieldError) {
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fieldErr(field, "invalid RFC 3339 timestamp")
	}
	return t, nil
}

// Parsed types embed the raw wire types and add native Go fields.

// ParsedEnvelope is a ConnectEnvelope with decoded binary fields.
type ParsedEnvelope struct {
	ConnectEnvelope
	PayloadBytes   []byte
	PubKeyBytes    []byte
	SignatureBytes []byte
}

// ParsedPayment is a ConnectPayment with parsed native-type fields.
type ParsedPayment struct {
	ConnectPayment
	IssuedTime    time.Time
	TotalKoinu    koinu.Koinu
	FeePerKBKoinu koinu.Koinu
	FeesKoinu     koinu.Koinu
	TaxesKoinu    koinu.Koinu
	ParsedItems   []ParsedItem
	ParsedOutputs []ParsedOutput
}

// ParsedItem is a ConnectItem with parsed koinu amounts.
type ParsedItem struct {
	ConnectItem
	UnitCostKoinu koinu.Koinu
	TotalKoinu    koinu.Koinu
	TaxKoinu      koinu.Koinu
}

// ParsedOutput is a ConnectOutput with a parsed koinu amount.
type ParsedOutput struct {
	ConnectOutput
	AmountKoinu koinu.Koinu
}

// ParsedSubmission is a PaymentSubmission with decoded transaction bytes.
type ParsedSubmission struct {
	PaymentSubmission
	TxBytes []byte
}

// ParsedStatusResponse is a PaymentStatusResponse with decoded binary and time fields.
type ParsedStatusResponse struct {
	PaymentStatusResponse
	TxIDBytes       []byte
	ConfirmedAtTime time.Time
}

// Parse methods — validate and parse in a single pass, best-effort.

// Parse validates and decodes a ConnectEnvelope.
func (e ConnectEnvelope) Parse() (ParsedEnvelope, FieldErrors) {
	var errs FieldErrors
	p := ParsedEnvelope{ConnectEnvelope: e}

	if e.Version != EnvelopeVersion {
		errs.Add(fieldErr("version", fmt.Sprintf("must be %q", EnvelopeVersion)))
	}
	if e.Payload == "" {
		errs.Add(fieldErr("payload", "required"))
	} else {
		b, err := base64.StdEncoding.DecodeString(e.Payload)
		if err != nil {
			errs.Add(fieldErr("payload", "invalid base64"))
		} else {
			p.PayloadBytes = b
		}
	}

	var fe *FieldError
	p.PubKeyBytes, fe = parseHexBytes("pubkey", e.PubKey, 32)
	errs.Add(fe)
	p.SignatureBytes, fe = parseHexBytes("sig", e.Signature, 64)
	errs.Add(fe)

	return p, errs
}

// Parse validates and decodes a ConnectPayment.
func (pay ConnectPayment) Parse() (ParsedPayment, FieldErrors) {
	var errs FieldErrors
	p := ParsedPayment{ConnectPayment: pay}

	if pay.Type != EnvelopeTypePayment {
		errs.Add(fieldErr("type", fmt.Sprintf("must be %q", EnvelopeTypePayment)))
	}
	errs.Add(checkNonEmpty("id", pay.ID))
	errs.Add(checkNonEmpty("relay", pay.Relay))
	errs.Add(checkNonEmpty("vendor_name", pay.VendorName))

	var fe *FieldError
	p.IssuedTime, fe = parseTimestamp("issued", pay.Issued)
	errs.Add(fe)
	p.TotalKoinu, fe = parseRequiredKoinu("total", pay.Total)
	errs.Add(fe)
	if fe == nil && p.TotalKoinu <= 0 {
		errs.Add(fieldErr("total", "must be positive"))
	}
	p.FeePerKBKoinu, fe = parseRequiredKoinu("fee_per_kb", pay.FeePerKB)
	errs.Add(fe)
	p.FeesKoinu, fe = parseOptionalKoinu("fees", pay.Fees)
	errs.Add(fe)
	p.TaxesKoinu, fe = parseOptionalKoinu("taxes", pay.Taxes)
	errs.Add(fe)

	if (pay.FiatTotal != "" || pay.FiatTax != "") && pay.FiatCurrency == "" {
		errs.Add(fieldErr("fiat_currency", "required when fiat_total or fiat_tax is set"))
	}

	if pay.Timeout < 1 {
		errs.Add(fieldErr("timeout", "must be > 0"))
	}
	if pay.MaxSize < 1 {
		errs.Add(fieldErr("max_size", "must be > 0"))
	}

	if pay.Items == nil {
		errs.Add(fieldErr("items", "required (use empty array)"))
	} else {
		p.ParsedItems = make([]ParsedItem, len(pay.Items))
		for i, item := range pay.Items {
			parsed, itemErrs := item.Parse()
			p.ParsedItems[i] = parsed
			for _, e := range itemErrs {
				errs.Add(fieldErr(fmt.Sprintf("items[%d].%s", i, e.Field), e.Message))
			}
		}
	}

	if pay.Outputs == nil {
		errs.Add(fieldErr("outputs", "required"))
	} else if len(pay.Outputs) == 0 {
		errs.Add(fieldErr("outputs", "must have at least one output"))
	} else {
		p.ParsedOutputs = make([]ParsedOutput, len(pay.Outputs))
		for i, o := range pay.Outputs {
			parsed, outErrs := o.Parse()
			p.ParsedOutputs[i] = parsed
			for _, e := range outErrs {
				errs.Add(fieldErr(fmt.Sprintf("outputs[%d].%s", i, e.Field), e.Message))
			}
		}
	}

	return p, errs
}

// Parse validates and decodes a ConnectItem.
func (item ConnectItem) Parse() (ParsedItem, FieldErrors) {
	var errs FieldErrors
	p := ParsedItem{ConnectItem: item}

	switch item.Type {
	case ItemTypeItem, ItemTypeTax, ItemTypeFee, ItemTypeShipping, ItemTypeDiscount, ItemTypeDonation:
		// valid
	default:
		errs.Add(fieldErr("type", "invalid item type"))
	}

	errs.Add(checkNonEmpty("id", item.ID))
	errs.Add(checkNonEmpty("name", item.Name))

	if item.UnitCount < 1 {
		errs.Add(fieldErr("count", "must be >= 1"))
	}

	var fe *FieldError
	p.UnitCostKoinu, fe = parseRequiredKoinu("unit", item.UnitCost)
	errs.Add(fe)
	p.TotalKoinu, fe = parseRequiredKoinu("total", item.Total)
	errs.Add(fe)
	p.TaxKoinu, fe = parseOptionalKoinu("tax", item.Tax)
	errs.Add(fe)

	if item.Type == ItemTypeDiscount {
		if p.UnitCostKoinu >= 0 && item.UnitCost != "" {
			errs.Add(fieldErr("unit", "discount unit must be negative"))
		}
		if p.TotalKoinu >= 0 && item.Total != "" {
			errs.Add(fieldErr("total", "discount total must be negative"))
		}
	}

	return p, errs
}

// Parse validates and decodes a ConnectOutput.
func (o ConnectOutput) Parse() (ParsedOutput, FieldErrors) {
	var errs FieldErrors
	p := ParsedOutput{ConnectOutput: o}
	errs.Add(checkNonEmpty("address", o.Address))
	var fe *FieldError
	p.AmountKoinu, fe = parseRequiredKoinu("amount", o.Amount)
	errs.Add(fe)
	if fe == nil && p.AmountKoinu <= 0 {
		errs.Add(fieldErr("amount", "must be positive"))
	}
	return p, errs
}

// Parse validates and decodes a PaymentSubmission.
func (s PaymentSubmission) Parse() (ParsedSubmission, FieldErrors) {
	var errs FieldErrors
	p := ParsedSubmission{PaymentSubmission: s}
	errs.Add(checkNonEmpty("id", s.ID))
	var fe *FieldError
	p.TxBytes, fe = parseRequiredHex("tx", s.Tx)
	errs.Add(fe)
	return p, errs
}

// Parse validates and decodes a PaymentStatusResponse.
func (r PaymentStatusResponse) Parse() (ParsedStatusResponse, FieldErrors) {
	var errs FieldErrors
	p := ParsedStatusResponse{PaymentStatusResponse: r}

	errs.Add(checkNonEmpty("id", r.ID))

	switch r.Status {
	case PaymentStatusUnpaid, PaymentStatusAccepted, PaymentStatusConfirmed, PaymentStatusDeclined:
		// valid
	default:
		errs.Add(fieldErr("status", "invalid payment status"))
	}

	// Conditional field presence per status.
	if r.Reason != "" && r.Status != PaymentStatusDeclined {
		errs.Add(fieldErr("reason", "only allowed when status is declined"))
	}
	if r.TxID != "" && r.Status != PaymentStatusAccepted && r.Status != PaymentStatusConfirmed {
		errs.Add(fieldErr("txid", "only allowed when status is accepted or confirmed"))
	}
	var fe *FieldError
	p.TxIDBytes, fe = parseOptionalHex("txid", r.TxID)
	errs.Add(fe)
	if r.ConfirmedAt != "" && r.Status != PaymentStatusConfirmed {
		errs.Add(fieldErr("confirmed_at", "only allowed when status is confirmed"))
	}
	p.ConfirmedAtTime, fe = parseOptionalTimestamp("confirmed_at", r.ConfirmedAt)
	errs.Add(fe)
	if (r.Required != nil || r.Confirmed != nil || r.DueSec != nil) &&
		r.Status != PaymentStatusAccepted && r.Status != PaymentStatusConfirmed {
		errs.Add(fieldErr("required/confirmed/due_sec", "only allowed when status is accepted or confirmed"))
	}

	return p, errs
}

// StatusQuery and ErrorResponse have no complex fields to parse,
// so they only get Validate methods (no Parsed* type needed).

// Validate checks the StatusQuery fields.
func (q StatusQuery) Validate() FieldErrors {
	var errs FieldErrors
	errs.Add(checkNonEmpty("id", q.ID))
	return errs
}

// Validate checks the ErrorResponse fields.
func (e ErrorResponse) Validate() FieldErrors {
	var errs FieldErrors
	switch e.Error {
	case ErrorCodeNotFound, ErrorCodeExpired, ErrorCodeInvalidTx, ErrorCodeInvalidOutputs, ErrorCodeInvalidToken:
		// valid
	case "":
		errs.Add(fieldErr("error", "required"))
	default:
		errs.Add(fieldErr("error", "invalid error code"))
	}
	errs.Add(checkNonEmpty("message", e.Message))
	return errs
}

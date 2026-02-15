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

// helpers

func requireNonEmpty(errs *[]FieldError, field, value string) {
	if value == "" {
		*errs = append(*errs, FieldError{field, "required"})
	}
}

func requireHexBytes(errs *[]FieldError, field, value string, n int) {
	if value == "" {
		*errs = append(*errs, FieldError{field, "required"})
		return
	}
	b, err := hex.DecodeString(value)
	if err != nil {
		*errs = append(*errs, FieldError{field, "invalid hex"})
		return
	}
	if len(b) != n {
		*errs = append(*errs, FieldError{field, fmt.Sprintf("must be %d bytes, got %d", n, len(b))})
	}
}

func requireValidHex(errs *[]FieldError, field, value string) {
	if value == "" {
		*errs = append(*errs, FieldError{field, "required"})
		return
	}
	if _, err := hex.DecodeString(value); err != nil {
		*errs = append(*errs, FieldError{field, "invalid hex"})
	}
}

func requireKoinu(errs *[]FieldError, field, value string) {
	if value == "" {
		*errs = append(*errs, FieldError{field, "required"})
		return
	}
	if _, err := koinu.ParseKoinu(value); err != nil {
		*errs = append(*errs, FieldError{field, "invalid koinu value"})
	}
}

func optionalKoinu(errs *[]FieldError, field, value string) {
	if value == "" {
		return
	}
	if _, err := koinu.ParseKoinu(value); err != nil {
		*errs = append(*errs, FieldError{field, "invalid koinu value"})
	}
}

// Validate checks the ConnectEnvelope fields.
func (e ConnectEnvelope) Validate() []FieldError {
	var errs []FieldError
	if e.Version != EnvelopeVersion {
		errs = append(errs, FieldError{"version", fmt.Sprintf("must be %q", EnvelopeVersion)})
	}
	if e.Payload == "" {
		errs = append(errs, FieldError{"payload", "required"})
	} else if _, err := base64.StdEncoding.DecodeString(e.Payload); err != nil {
		errs = append(errs, FieldError{"payload", "invalid base64"})
	}
	requireHexBytes(&errs, "pubkey", e.PubKey, 32)
	requireHexBytes(&errs, "sig", e.Signature, 64)
	return errs
}

// Validate checks the ConnectPayment fields.
func (p ConnectPayment) Validate() []FieldError {
	var errs []FieldError

	if p.Type != PaymentTypePayment {
		errs = append(errs, FieldError{"type", fmt.Sprintf("must be %q", PaymentTypePayment)})
	}
	requireNonEmpty(&errs, "id", p.ID)
	requireNonEmpty(&errs, "relay", p.Relay)
	requireNonEmpty(&errs, "vendor_name", p.VendorName)

	if p.Issued == "" {
		errs = append(errs, FieldError{"issued", "required"})
	} else if _, err := time.Parse(time.RFC3339, p.Issued); err != nil {
		errs = append(errs, FieldError{"issued", "invalid RFC 3339 timestamp"})
	}

	requireKoinu(&errs, "total", p.Total)
	optionalKoinu(&errs, "fee_per_kb", p.FeePerKB)
	optionalKoinu(&errs, "fees", p.Fees)
	optionalKoinu(&errs, "taxes", p.Taxes)

	if (p.FiatTotal != "" || p.FiatTax != "") && p.FiatCurrency == "" {
		errs = append(errs, FieldError{"fiat_currency", "required when fiat_total or fiat_tax is set"})
	}

	if p.Timeout < 0 {
		errs = append(errs, FieldError{"timeout", "must be > 0"})
	}
	if p.MaxSize < 0 {
		errs = append(errs, FieldError{"max_size", "must be > 0"})
	}

	if p.Items == nil {
		errs = append(errs, FieldError{"items", "required (use empty array)"})
	} else {
		for i, item := range p.Items {
			for _, e := range item.Validate() {
				errs = append(errs, FieldError{fmt.Sprintf("items[%d].%s", i, e.Field), e.Message})
			}
		}
	}

	if p.Outputs == nil {
		errs = append(errs, FieldError{"outputs", "required"})
	} else if len(p.Outputs) == 0 {
		errs = append(errs, FieldError{"outputs", "must have at least one output"})
	} else {
		for i, o := range p.Outputs {
			for _, e := range o.Validate() {
				errs = append(errs, FieldError{fmt.Sprintf("outputs[%d].%s", i, e.Field), e.Message})
			}
		}
	}

	return errs
}

// Validate checks the ConnectItem fields.
func (item ConnectItem) Validate() []FieldError {
	var errs []FieldError

	switch item.Type {
	case ItemTypeItem, ItemTypeTax, ItemTypeFee, ItemTypeShipping, ItemTypeDiscount, ItemTypeDonation:
		// valid
	default:
		errs = append(errs, FieldError{"type", "invalid item type"})
	}

	requireNonEmpty(&errs, "id", item.ID)
	requireNonEmpty(&errs, "name", item.Name)

	if item.UnitCount < 1 {
		errs = append(errs, FieldError{"count", "must be >= 1"})
	}

	requireKoinu(&errs, "unit", item.UnitCost)
	requireKoinu(&errs, "total", item.Total)
	optionalKoinu(&errs, "tax", item.Tax)

	return errs
}

// Validate checks the ConnectOutput fields.
func (o ConnectOutput) Validate() []FieldError {
	var errs []FieldError
	requireNonEmpty(&errs, "address", o.Address)
	requireKoinu(&errs, "amount", o.Amount)
	return errs
}

// Validate checks the PaymentSubmission fields.
func (s PaymentSubmission) Validate() []FieldError {
	var errs []FieldError
	requireNonEmpty(&errs, "id", s.ID)
	requireValidHex(&errs, "tx", s.Tx)
	return errs
}

// Validate checks the StatusQuery fields.
func (q StatusQuery) Validate() []FieldError {
	var errs []FieldError
	requireNonEmpty(&errs, "id", q.ID)
	return errs
}

// Validate checks the PaymentStatusResponse fields.
func (r PaymentStatusResponse) Validate() []FieldError {
	var errs []FieldError
	requireNonEmpty(&errs, "id", r.ID)
	switch r.Status {
	case PaymentStatusUnpaid, PaymentStatusAccepted, PaymentStatusConfirmed, PaymentStatusDeclined:
		// valid
	default:
		errs = append(errs, FieldError{"status", "invalid payment status"})
	}
	if r.ConfirmedAt != "" {
		if _, err := time.Parse(time.RFC3339, r.ConfirmedAt); err != nil {
			errs = append(errs, FieldError{"confirmed_at", "invalid RFC 3339 timestamp"})
		}
	}
	return errs
}

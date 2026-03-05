package test

import (
	"testing"
	"time"

	dogeconnectgo "github.com/dogeorg/dogeconnect-go"
	"github.com/dogeorg/dogeconnect-go/koinu"
)

// helpers

func ptr[T any](v T) *T { return &v }

func validPayment() dogeconnectgo.ConnectPayment {
	return dogeconnectgo.ConnectPayment{
		Type:       dogeconnectgo.EnvelopeTypePayment,
		ID:         "pay-1",
		Issued:     "2025-06-01T00:00:00Z",
		Timeout:    60,
		Relay:      "https://example.com/dc/1",
		FeePerKB:   "0.01",
		MaxSize:    10000,
		VendorName: "Test Vendor",
		Total:      "100",
		Items:      []dogeconnectgo.ConnectItem{validItem()},
		Outputs:    []dogeconnectgo.ConnectOutput{validOutput()},
	}
}

func validItem() dogeconnectgo.ConnectItem {
	return dogeconnectgo.ConnectItem{
		Type:      dogeconnectgo.ItemTypeItem,
		ID:        "item-1",
		Name:      "Widget",
		UnitCount: 1,
		UnitCost:  "100",
		Total:     "100",
	}
}

func validOutput() dogeconnectgo.ConnectOutput {
	return dogeconnectgo.ConnectOutput{
		Address: "DPD7uK4B1kRmbfGmytBhG1DZjaMWNfbpwY",
		Amount:  "100",
	}
}

func validEnvelope() dogeconnectgo.ConnectEnvelope {
	return dogeconnectgo.ConnectEnvelope{
		Version:   dogeconnectgo.EnvelopeVersion,
		Payload:   "eyJ0eXBlIjoicGF5bWVudCJ9", // base64 placeholder
		PubKey:    "0000000000000000000000000000000000000000000000000000000000000001",
		Signature: "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
	}
}

func hasFieldError(errs dogeconnectgo.FieldErrors, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

func requireNoErrors(t *testing.T, errs dogeconnectgo.FieldErrors) {
	t.Helper()
	if len(errs) > 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func requireFieldError(t *testing.T, errs dogeconnectgo.FieldErrors, field string) {
	t.Helper()
	if !hasFieldError(errs, field) {
		t.Errorf("expected error on field %q, got: %v", field, errs)
	}
}

// ConnectEnvelope

func TestEnvelopeParseValid(t *testing.T) {
	p, errs := validEnvelope().Parse()
	requireNoErrors(t, errs)
	if p.PayloadBytes == nil {
		t.Fatal("expected PayloadBytes to be populated")
	}
	if len(p.PubKeyBytes) != 32 {
		t.Fatalf("expected 32-byte pubkey, got %d", len(p.PubKeyBytes))
	}
	if len(p.SignatureBytes) != 64 {
		t.Fatalf("expected 64-byte signature, got %d", len(p.SignatureBytes))
	}
}

func TestEnvelopeParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(*dogeconnectgo.ConnectEnvelope)
		field string
	}{
		{"bad version", func(e *dogeconnectgo.ConnectEnvelope) { e.Version = "0.1" }, "version"},
		{"empty payload", func(e *dogeconnectgo.ConnectEnvelope) { e.Payload = "" }, "payload"},
		{"bad base64 payload", func(e *dogeconnectgo.ConnectEnvelope) { e.Payload = "not-base64!!!" }, "payload"},
		{"empty pubkey", func(e *dogeconnectgo.ConnectEnvelope) { e.PubKey = "" }, "pubkey"},
		{"bad pubkey hex", func(e *dogeconnectgo.ConnectEnvelope) { e.PubKey = "zzzz" }, "pubkey"},
		{"pubkey wrong len", func(e *dogeconnectgo.ConnectEnvelope) { e.PubKey = "0011" }, "pubkey"},
		{"empty sig", func(e *dogeconnectgo.ConnectEnvelope) { e.Signature = "" }, "sig"},
		{"bad sig hex", func(e *dogeconnectgo.ConnectEnvelope) { e.Signature = "nothex" }, "sig"},
		{"sig wrong len", func(e *dogeconnectgo.ConnectEnvelope) { e.Signature = "0011" }, "sig"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := validEnvelope()
			tc.mod(&e)
			_, errs := e.Parse()
			requireFieldError(t, errs, tc.field)
		})
	}
}

// ConnectPayment

func TestPaymentParseValid(t *testing.T) {
	p, errs := validPayment().Parse()
	requireNoErrors(t, errs)

	expected, _ := time.Parse(time.RFC3339, "2025-06-01T00:00:00Z")
	if !p.IssuedTime.Equal(expected) {
		t.Errorf("IssuedTime = %v, want %v", p.IssuedTime, expected)
	}
	if p.TotalKoinu != 100*koinu.OneDoge {
		t.Errorf("TotalKoinu = %d, want %d", p.TotalKoinu, 100*koinu.OneDoge)
	}
	if p.FeePerKBKoinu != 1_000_000 {
		t.Errorf("FeePerKBKoinu = %d, want %d", p.FeePerKBKoinu, 1_000_000)
	}
	if len(p.ParsedItems) != 1 {
		t.Fatalf("expected 1 parsed item, got %d", len(p.ParsedItems))
	}
	if p.ParsedItems[0].TotalKoinu != 100*koinu.OneDoge {
		t.Errorf("item TotalKoinu = %d, want %d", p.ParsedItems[0].TotalKoinu, 100*koinu.OneDoge)
	}
	if len(p.ParsedOutputs) != 1 {
		t.Fatalf("expected 1 parsed output, got %d", len(p.ParsedOutputs))
	}
	if p.ParsedOutputs[0].AmountKoinu != 100*koinu.OneDoge {
		t.Errorf("output AmountKoinu = %d, want %d", p.ParsedOutputs[0].AmountKoinu, 100*koinu.OneDoge)
	}
}

func TestPaymentParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(*dogeconnectgo.ConnectPayment)
		field string
	}{
		{"bad type", func(p *dogeconnectgo.ConnectPayment) { p.Type = "invoice" }, "type"},
		{"empty id", func(p *dogeconnectgo.ConnectPayment) { p.ID = "" }, "id"},
		{"empty issued", func(p *dogeconnectgo.ConnectPayment) { p.Issued = "" }, "issued"},
		{"bad issued", func(p *dogeconnectgo.ConnectPayment) { p.Issued = "not-a-date" }, "issued"},
		{"empty relay", func(p *dogeconnectgo.ConnectPayment) { p.Relay = "" }, "relay"},
		{"empty vendor_name", func(p *dogeconnectgo.ConnectPayment) { p.VendorName = "" }, "vendor_name"},
		{"empty total", func(p *dogeconnectgo.ConnectPayment) { p.Total = "" }, "total"},
		{"bad total", func(p *dogeconnectgo.ConnectPayment) { p.Total = "abc" }, "total"},
		{"empty fee_per_kb", func(p *dogeconnectgo.ConnectPayment) { p.FeePerKB = "" }, "fee_per_kb"},
		{"bad fee_per_kb", func(p *dogeconnectgo.ConnectPayment) { p.FeePerKB = "abc" }, "fee_per_kb"},
		{"bad fees", func(p *dogeconnectgo.ConnectPayment) { p.Fees = "abc" }, "fees"},
		{"bad taxes", func(p *dogeconnectgo.ConnectPayment) { p.Taxes = "abc" }, "taxes"},
		{"zero timeout", func(p *dogeconnectgo.ConnectPayment) { p.Timeout = 0 }, "timeout"},
		{"negative timeout", func(p *dogeconnectgo.ConnectPayment) { p.Timeout = -1 }, "timeout"},
		{"zero max_size", func(p *dogeconnectgo.ConnectPayment) { p.MaxSize = 0 }, "max_size"},
		{"negative max_size", func(p *dogeconnectgo.ConnectPayment) { p.MaxSize = -1 }, "max_size"},
		{"nil items", func(p *dogeconnectgo.ConnectPayment) { p.Items = nil }, "items"},
		{"nil outputs", func(p *dogeconnectgo.ConnectPayment) { p.Outputs = nil }, "outputs"},
		{"empty outputs", func(p *dogeconnectgo.ConnectPayment) { p.Outputs = []dogeconnectgo.ConnectOutput{} }, "outputs"},
		{"fiat_total without fiat_currency", func(p *dogeconnectgo.ConnectPayment) {
			p.FiatTotal = "10.00"
			p.FiatCurrency = ""
		}, "fiat_currency"},
		{"fiat_tax without fiat_currency", func(p *dogeconnectgo.ConnectPayment) {
			p.FiatTax = "1.00"
			p.FiatCurrency = ""
		}, "fiat_currency"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := validPayment()
			tc.mod(&p)
			_, errs := p.Parse()
			requireFieldError(t, errs, tc.field)
		})
	}
}

func TestPaymentEmptyItemsIsValid(t *testing.T) {
	p := validPayment()
	p.Items = []dogeconnectgo.ConnectItem{} // empty slice, not nil
	_, errs := p.Parse()
	if hasFieldError(errs, "items") {
		t.Errorf("empty items slice should be valid, got: %v", errs)
	}
}

func TestPaymentNestedItemErrors(t *testing.T) {
	p := validPayment()
	bad := validItem()
	bad.Name = ""
	p.Items = []dogeconnectgo.ConnectItem{bad}
	_, errs := p.Parse()
	requireFieldError(t, errs, "items[0].name")
}

func TestPaymentNestedOutputErrors(t *testing.T) {
	p := validPayment()
	bad := validOutput()
	bad.Address = ""
	p.Outputs = []dogeconnectgo.ConnectOutput{bad}
	_, errs := p.Parse()
	requireFieldError(t, errs, "outputs[0].address")
}

func TestPaymentBestEffort(t *testing.T) {
	p := validPayment()
	p.Total = "abc" // invalid koinu — should error but other fields still parsed
	parsed, errs := p.Parse()
	requireFieldError(t, errs, "total")
	if parsed.TotalKoinu != 0 {
		t.Errorf("expected zero TotalKoinu on parse failure, got %d", parsed.TotalKoinu)
	}
	// Other fields should still be populated
	if parsed.FeePerKBKoinu != 1_000_000 {
		t.Errorf("FeePerKBKoinu should still be parsed, got %d", parsed.FeePerKBKoinu)
	}
	if parsed.IssuedTime.IsZero() {
		t.Error("IssuedTime should still be parsed")
	}
}

// ConnectItem

func TestItemParseValid(t *testing.T) {
	p, errs := validItem().Parse()
	requireNoErrors(t, errs)
	if p.UnitCostKoinu != 100*koinu.OneDoge {
		t.Errorf("UnitCostKoinu = %d, want %d", p.UnitCostKoinu, 100*koinu.OneDoge)
	}
	if p.TotalKoinu != 100*koinu.OneDoge {
		t.Errorf("TotalKoinu = %d, want %d", p.TotalKoinu, 100*koinu.OneDoge)
	}
}

func TestItemParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(*dogeconnectgo.ConnectItem)
		field string
	}{
		{"bad type", func(i *dogeconnectgo.ConnectItem) { i.Type = "unknown" }, "type"},
		{"empty id", func(i *dogeconnectgo.ConnectItem) { i.ID = "" }, "id"},
		{"empty name", func(i *dogeconnectgo.ConnectItem) { i.Name = "" }, "name"},
		{"zero count", func(i *dogeconnectgo.ConnectItem) { i.UnitCount = 0 }, "count"},
		{"negative count", func(i *dogeconnectgo.ConnectItem) { i.UnitCount = -1 }, "count"},
		{"empty unit", func(i *dogeconnectgo.ConnectItem) { i.UnitCost = "" }, "unit"},
		{"bad unit", func(i *dogeconnectgo.ConnectItem) { i.UnitCost = "abc" }, "unit"},
		{"empty total", func(i *dogeconnectgo.ConnectItem) { i.Total = "" }, "total"},
		{"bad total", func(i *dogeconnectgo.ConnectItem) { i.Total = "abc" }, "total"},
		{"bad tax", func(i *dogeconnectgo.ConnectItem) { i.Tax = "abc" }, "tax"},
		{"discount positive unit", func(i *dogeconnectgo.ConnectItem) {
			i.Type = dogeconnectgo.ItemTypeDiscount
			i.UnitCost = "5"
			i.Total = "-5"
		}, "unit"},
		{"discount positive total", func(i *dogeconnectgo.ConnectItem) {
			i.Type = dogeconnectgo.ItemTypeDiscount
			i.UnitCost = "-5"
			i.Total = "5"
		}, "total"},
		{"discount zero unit", func(i *dogeconnectgo.ConnectItem) {
			i.Type = dogeconnectgo.ItemTypeDiscount
			i.UnitCost = "0"
			i.Total = "0"
		}, "unit"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := validItem()
			tc.mod(&i)
			_, errs := i.Parse()
			requireFieldError(t, errs, tc.field)
		})
	}
}

func TestDiscountItemValid(t *testing.T) {
	i := validItem()
	i.Type = dogeconnectgo.ItemTypeDiscount
	i.UnitCost = "-5"
	i.Total = "-5"
	p, errs := i.Parse()
	requireNoErrors(t, errs)
	if p.UnitCostKoinu >= 0 {
		t.Errorf("expected negative UnitCostKoinu, got %d", p.UnitCostKoinu)
	}
}

func TestItemAllTypes(t *testing.T) {
	types := []dogeconnectgo.ItemType{
		dogeconnectgo.ItemTypeItem,
		dogeconnectgo.ItemTypeTax,
		dogeconnectgo.ItemTypeFee,
		dogeconnectgo.ItemTypeShipping,
		dogeconnectgo.ItemTypeDiscount,
		dogeconnectgo.ItemTypeDonation,
	}
	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			i := validItem()
			i.Type = typ
			if typ == dogeconnectgo.ItemTypeDiscount {
				i.UnitCost = "-100"
				i.Total = "-100"
			}
			_, errs := i.Parse()
			requireNoErrors(t, errs)
		})
	}
}

// ConnectOutput

func TestOutputParseValid(t *testing.T) {
	p, errs := validOutput().Parse()
	requireNoErrors(t, errs)
	if p.AmountKoinu != 100*koinu.OneDoge {
		t.Errorf("AmountKoinu = %d, want %d", p.AmountKoinu, 100*koinu.OneDoge)
	}
}

func TestOutputParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(*dogeconnectgo.ConnectOutput)
		field string
	}{
		{"empty address", func(o *dogeconnectgo.ConnectOutput) { o.Address = "" }, "address"},
		{"empty amount", func(o *dogeconnectgo.ConnectOutput) { o.Amount = "" }, "amount"},
		{"bad amount", func(o *dogeconnectgo.ConnectOutput) { o.Amount = "abc" }, "amount"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := validOutput()
			tc.mod(&o)
			_, errs := o.Parse()
			requireFieldError(t, errs, tc.field)
		})
	}
}

// PaymentSubmission

func TestSubmissionParseValid(t *testing.T) {
	s := dogeconnectgo.PaymentSubmission{ID: "pay-1", Tx: "deadbeef"}
	p, errs := s.Parse()
	requireNoErrors(t, errs)
	if len(p.TxBytes) == 0 {
		t.Fatal("expected TxBytes to be populated")
	}
}

func TestSubmissionParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		sub   dogeconnectgo.PaymentSubmission
		field string
	}{
		{"empty id", dogeconnectgo.PaymentSubmission{ID: "", Tx: "deadbeef"}, "id"},
		{"empty tx", dogeconnectgo.PaymentSubmission{ID: "pay-1", Tx: ""}, "tx"},
		{"bad tx hex", dogeconnectgo.PaymentSubmission{ID: "pay-1", Tx: "nothex!"}, "tx"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := tc.sub.Parse()
			requireFieldError(t, errs, tc.field)
		})
	}
}

// StatusQuery validation (still uses Validate — no Parse needed)

func TestStatusQueryValidValid(t *testing.T) {
	q := dogeconnectgo.StatusQuery{ID: "pay-1"}
	requireNoErrors(t, q.Validate())
}

func TestStatusQueryEmpty(t *testing.T) {
	q := dogeconnectgo.StatusQuery{}
	requireFieldError(t, q.Validate(), "id")
}

// PaymentStatusResponse

func TestStatusResponseParseValid(t *testing.T) {
	r := dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid}
	_, errs := r.Parse()
	requireNoErrors(t, errs)
}

func TestStatusResponseParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		resp  dogeconnectgo.PaymentStatusResponse
		field string
	}{
		{"empty id", dogeconnectgo.PaymentStatusResponse{Status: dogeconnectgo.PaymentStatusUnpaid}, "id"},
		{"bad status", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: "bogus"}, "status"},
		{"empty status", dogeconnectgo.PaymentStatusResponse{ID: "pay-1"}, "status"},
		{"bad confirmed_at", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusConfirmed, ConfirmedAt: "not-a-date"}, "confirmed_at"},
		{"bad txid hex", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, TxID: "not-hex!"}, "txid"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := tc.resp.Parse()
			requireFieldError(t, errs, tc.field)
		})
	}
}

func TestStatusResponseConfirmedAtParsed(t *testing.T) {
	r := dogeconnectgo.PaymentStatusResponse{
		ID:          "pay-1",
		Status:      dogeconnectgo.PaymentStatusConfirmed,
		ConfirmedAt: "2025-06-01T12:00:00Z",
		Required:    ptr(6),
		Confirmed:   ptr(0),
		DueSec:      ptr(600),
	}
	p, errs := r.Parse()
	requireNoErrors(t, errs)
	expected, _ := time.Parse(time.RFC3339, "2025-06-01T12:00:00Z")
	if !p.ConfirmedAtTime.Equal(expected) {
		t.Errorf("ConfirmedAtTime = %v, want %v", p.ConfirmedAtTime, expected)
	}
}

func TestStatusResponseConditionalFields(t *testing.T) {
	tests := []struct {
		name  string
		resp  dogeconnectgo.PaymentStatusResponse
		field string
	}{
		{"reason with unpaid", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid, Reason: "oops"}, "reason"},
		{"reason with accepted", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, Reason: "oops"}, "reason"},
		{"reason with confirmed", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusConfirmed, Reason: "oops"}, "reason"},
		{"txid with unpaid", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid, TxID: "deadbeef"}, "txid"},
		{"txid with declined", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusDeclined, TxID: "deadbeef", Reason: "bad"}, "txid"},
		{"confirmed_at with unpaid", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid, ConfirmedAt: "2025-06-01T12:00:00Z"}, "confirmed_at"},
		{"confirmed_at with accepted", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, ConfirmedAt: "2025-06-01T12:00:00Z"}, "confirmed_at"},
		{"confirmed_at with declined", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusDeclined, ConfirmedAt: "2025-06-01T12:00:00Z", Reason: "bad"}, "confirmed_at"},
		{"required with unpaid", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid, Required: ptr(6)}, "required/confirmed/due_sec"},
		{"confirmed count with declined", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusDeclined, Confirmed: ptr(0), Reason: "bad"}, "required/confirmed/due_sec"},
		{"due_sec with unpaid", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid, DueSec: ptr(600)}, "required/confirmed/due_sec"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := tc.resp.Parse()
			requireFieldError(t, errs, tc.field)
		})
	}
}

func TestStatusResponseConditionalFieldsValid(t *testing.T) {
	// reason allowed with declined
	r := dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusDeclined, Reason: "insufficient funds"}
	_, errs := r.Parse()
	requireNoErrors(t, errs)

	// txid allowed with accepted
	r = dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, TxID: "deadbeef"}
	_, errs = r.Parse()
	requireNoErrors(t, errs)

	// txid allowed with confirmed
	r = dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusConfirmed, TxID: "deadbeef", ConfirmedAt: "2025-06-01T12:00:00Z"}
	_, errs = r.Parse()
	requireNoErrors(t, errs)

	// required/confirmed/due_sec allowed with accepted
	r = dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, Required: ptr(6), Confirmed: ptr(0), DueSec: ptr(600)}
	_, errs = r.Parse()
	requireNoErrors(t, errs)
}

func TestStatusResponseAllStatuses(t *testing.T) {
	statuses := []dogeconnectgo.PaymentStatus{
		dogeconnectgo.PaymentStatusUnpaid,
		dogeconnectgo.PaymentStatusAccepted,
		dogeconnectgo.PaymentStatusConfirmed,
		dogeconnectgo.PaymentStatusDeclined,
	}
	for _, s := range statuses {
		t.Run(string(s), func(t *testing.T) {
			r := dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: s}
			_, errs := r.Parse()
			requireNoErrors(t, errs)
		})
	}
}

func TestStatusResponseTxIDParsed(t *testing.T) {
	r := dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, TxID: "deadbeef"}
	p, errs := r.Parse()
	requireNoErrors(t, errs)
	if len(p.TxIDBytes) == 0 {
		t.Fatal("expected TxIDBytes to be populated")
	}
}

// ErrorResponse validation (still uses Validate — no Parse needed)

func validErrorResponse() dogeconnectgo.ErrorResponse {
	return dogeconnectgo.ErrorResponse{
		Error:   dogeconnectgo.ErrorCodeNotFound,
		Message: "payment not found",
	}
}

func TestErrorResponseValidValid(t *testing.T) {
	requireNoErrors(t, validErrorResponse().Validate())
}

func TestErrorResponseValidation(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(*dogeconnectgo.ErrorResponse)
		field string
	}{
		{"empty code", func(e *dogeconnectgo.ErrorResponse) { e.Error = "" }, "error"},
		{"bad code", func(e *dogeconnectgo.ErrorResponse) { e.Error = "bogus" }, "error"},
		{"empty message", func(e *dogeconnectgo.ErrorResponse) { e.Message = "" }, "message"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := validErrorResponse()
			tc.mod(&e)
			requireFieldError(t, e.Validate(), tc.field)
		})
	}
}

func TestErrorResponseAllCodes(t *testing.T) {
	codes := []dogeconnectgo.ErrorCode{
		dogeconnectgo.ErrorCodeNotFound,
		dogeconnectgo.ErrorCodeExpired,
		dogeconnectgo.ErrorCodeInvalidTx,
		dogeconnectgo.ErrorCodeInvalidOutputs,
		dogeconnectgo.ErrorCodeInvalidToken,
	}
	for _, code := range codes {
		t.Run(string(code), func(t *testing.T) {
			e := dogeconnectgo.ErrorResponse{Error: code, Message: "test"}
			requireNoErrors(t, e.Validate())
		})
	}
}

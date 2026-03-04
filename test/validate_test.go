package test

import (
	"testing"

	dogeconnectgo "github.com/dogeorg/dogeconnect-go"
)

// helpers

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

func hasFieldError(errs []dogeconnectgo.FieldError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

func requireNoErrors(t *testing.T, errs []dogeconnectgo.FieldError) {
	t.Helper()
	if len(errs) > 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func requireFieldError(t *testing.T, errs []dogeconnectgo.FieldError, field string) {
	t.Helper()
	if !hasFieldError(errs, field) {
		t.Errorf("expected error on field %q, got: %v", field, errs)
	}
}

// ConnectEnvelope validation

func TestEnvelopeValidValid(t *testing.T) {
	requireNoErrors(t, validEnvelope().Validate())
}

func TestEnvelopeValidation(t *testing.T) {
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
			requireFieldError(t, e.Validate(), tc.field)
		})
	}
}

// ConnectPayment validation

func TestPaymentValidValid(t *testing.T) {
	requireNoErrors(t, validPayment().Validate())
}

func TestPaymentValidation(t *testing.T) {
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
			requireFieldError(t, p.Validate(), tc.field)
		})
	}
}

func TestPaymentEmptyItemsIsValid(t *testing.T) {
	p := validPayment()
	p.Items = []dogeconnectgo.ConnectItem{} // empty slice, not nil
	errs := p.Validate()
	if hasFieldError(errs, "items") {
		t.Errorf("empty items slice should be valid, got: %v", errs)
	}
}

func TestPaymentNestedItemErrors(t *testing.T) {
	p := validPayment()
	bad := validItem()
	bad.Name = ""
	p.Items = []dogeconnectgo.ConnectItem{bad}
	requireFieldError(t, p.Validate(), "items[0].name")
}

func TestPaymentNestedOutputErrors(t *testing.T) {
	p := validPayment()
	bad := validOutput()
	bad.Address = ""
	p.Outputs = []dogeconnectgo.ConnectOutput{bad}
	requireFieldError(t, p.Validate(), "outputs[0].address")
}

// ConnectItem validation

func TestItemValidValid(t *testing.T) {
	requireNoErrors(t, validItem().Validate())
}

func TestItemValidation(t *testing.T) {
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
			requireFieldError(t, i.Validate(), tc.field)
		})
	}
}

func TestDiscountItemValid(t *testing.T) {
	i := validItem()
	i.Type = dogeconnectgo.ItemTypeDiscount
	i.UnitCost = "-5"
	i.Total = "-5"
	requireNoErrors(t, i.Validate())
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
			requireNoErrors(t, i.Validate())
		})
	}
}

// ConnectOutput validation

func TestOutputValidValid(t *testing.T) {
	requireNoErrors(t, validOutput().Validate())
}

func TestOutputValidation(t *testing.T) {
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
			requireFieldError(t, o.Validate(), tc.field)
		})
	}
}

// PaymentSubmission validation

func TestSubmissionValidValid(t *testing.T) {
	s := dogeconnectgo.PaymentSubmission{ID: "pay-1", Tx: "deadbeef"}
	requireNoErrors(t, s.Validate())
}

func TestSubmissionValidation(t *testing.T) {
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
			requireFieldError(t, tc.sub.Validate(), tc.field)
		})
	}
}

// StatusQuery validation

func TestStatusQueryValidValid(t *testing.T) {
	q := dogeconnectgo.StatusQuery{ID: "pay-1"}
	requireNoErrors(t, q.Validate())
}

func TestStatusQueryEmpty(t *testing.T) {
	q := dogeconnectgo.StatusQuery{}
	requireFieldError(t, q.Validate(), "id")
}

// PaymentStatusResponse validation

func TestStatusResponseValidValid(t *testing.T) {
	r := dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid}
	requireNoErrors(t, r.Validate())
}

func TestStatusResponseValidation(t *testing.T) {
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
			requireFieldError(t, tc.resp.Validate(), tc.field)
		})
	}
}

func TestStatusResponseConfirmedAtValid(t *testing.T) {
	r := dogeconnectgo.PaymentStatusResponse{
		ID:          "pay-1",
		Status:      dogeconnectgo.PaymentStatusConfirmed,
		ConfirmedAt: "2025-06-01T12:00:00Z",
		Required:    new(6),
		Confirmed:   new(0),
		DueSec:      new(600),
	}
	requireNoErrors(t, r.Validate())
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
		{"required with unpaid", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid, Required: new(6)}, "required/confirmed/due_sec"},
		{"confirmed count with declined", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusDeclined, Confirmed: new(0), Reason: "bad"}, "required/confirmed/due_sec"},
		{"due_sec with unpaid", dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusUnpaid, DueSec: new(600)}, "required/confirmed/due_sec"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requireFieldError(t, tc.resp.Validate(), tc.field)
		})
	}
}

func TestStatusResponseConditionalFieldsValid(t *testing.T) {
	// reason allowed with declined
	r := dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusDeclined, Reason: "insufficient funds"}
	requireNoErrors(t, r.Validate())

	// txid allowed with accepted
	r = dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, TxID: "deadbeef"}
	requireNoErrors(t, r.Validate())

	// txid allowed with confirmed
	r = dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusConfirmed, TxID: "deadbeef", ConfirmedAt: "2025-06-01T12:00:00Z"}
	requireNoErrors(t, r.Validate())

	// required/confirmed/due_sec allowed with accepted
	r = dogeconnectgo.PaymentStatusResponse{ID: "pay-1", Status: dogeconnectgo.PaymentStatusAccepted, Required: new(6), Confirmed: new(0), DueSec: new(600)}
	requireNoErrors(t, r.Validate())
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
			requireNoErrors(t, r.Validate())
		})
	}
}

// ErrorResponse validation

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

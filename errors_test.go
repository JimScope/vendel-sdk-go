package vendel

import (
	"fmt"
	"testing"
)

func TestVendelError_ErrorString(t *testing.T) {
	err := &VendelError{StatusCode: 400, Message: "Bad request"}
	if err.Error() != "[400] Bad request" {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestIsQuotaError(t *testing.T) {
	qe := &QuotaError{VendelError: VendelError{StatusCode: 429, Message: "Exceeded"}}
	if !IsQuotaError(qe) {
		t.Error("expected true for QuotaError")
	}
	ve := &VendelError{StatusCode: 400}
	if IsQuotaError(ve) {
		t.Error("expected false for VendelError")
	}
	if IsQuotaError(fmt.Errorf("random error")) {
		t.Error("expected false for generic error")
	}
}

func TestIsAPIError(t *testing.T) {
	ve := &VendelError{StatusCode: 400}
	if !IsAPIError(ve) {
		t.Error("expected true for VendelError")
	}
	qe := &QuotaError{VendelError: VendelError{StatusCode: 429}}
	if !IsAPIError(qe) {
		t.Error("expected true for QuotaError (inherits VendelError)")
	}
	if IsAPIError(fmt.Errorf("random")) {
		t.Error("expected false for generic error")
	}
}

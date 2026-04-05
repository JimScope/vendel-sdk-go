package vendel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := NewClient(srv.URL, "vk_test_key")
	return client, srv
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	c := NewClient("https://api.example.com/", "key")
	if c.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", c.baseURL)
	}
}

func TestClient_SetsHeaders(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != "vk_test_key" {
			t.Errorf("X-API-Key = %q, want %q", got, "vk_test_key")
		}
		if r.Method == http.MethodPost {
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
		}
		json.NewEncoder(w).Encode(SendSMSResponse{Status: "ok"})
	})
	defer srv.Close()

	client.SendSMS(context.Background(), SendSMSRequest{Recipients: []string{"+1"}, Body: "hi"})
}

func TestSendSMS_Success(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req SendSMSRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Recipients) != 1 || req.Recipients[0] != "+1234567890" {
			t.Errorf("unexpected recipients: %v", req.Recipients)
		}
		if req.Body != "Hello" {
			t.Errorf("body = %q, want Hello", req.Body)
		}
		json.NewEncoder(w).Encode(SendSMSResponse{
			BatchID: "b1", MessageIDs: []string{"m1"}, RecipientsCount: 1, Status: "accepted",
		})
	})
	defer srv.Close()

	resp, err := client.SendSMS(context.Background(), SendSMSRequest{
		Recipients: []string{"+1234567890"}, Body: "Hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.BatchID != "b1" || resp.Status != "accepted" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestSendSMS_WithGroupIDs(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]any
		json.NewDecoder(r.Body).Decode(&raw)
		gids, ok := raw["group_ids"].([]any)
		if !ok || len(gids) != 2 {
			t.Errorf("expected 2 group_ids, got %v", raw["group_ids"])
		}
		json.NewEncoder(w).Encode(SendSMSResponse{Status: "accepted", RecipientsCount: 5})
	})
	defer srv.Close()

	resp, err := client.SendSMS(context.Background(), SendSMSRequest{
		Recipients: []string{"+1"}, Body: "Hi", GroupIDs: []string{"g1", "g2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RecipientsCount != 5 {
		t.Errorf("recipients_count = %d, want 5", resp.RecipientsCount)
	}
}

func TestSendSMSTemplate_Success(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req SendSMSTemplateRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.TemplateID != "tmpl_1" {
			t.Errorf("template_id = %q", req.TemplateID)
		}
		if req.Variables["code"] != "9999" {
			t.Errorf("variables = %v", req.Variables)
		}
		if r.URL.Path != "/api/sms/send-template" {
			t.Errorf("path = %q, want /api/sms/send-template", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SendSMSResponse{Status: "accepted", MessageIDs: []string{"m1"}})
	})
	defer srv.Close()

	resp, err := client.SendSMSTemplate(context.Background(), SendSMSTemplateRequest{
		Recipients: []string{"+1"}, TemplateID: "tmpl_1", Variables: map[string]string{"code": "9999"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "accepted" {
		t.Errorf("status = %q", resp.Status)
	}
}

func TestGetQuota_Success(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/plans/quota" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		json.NewEncoder(w).Encode(Quota{Plan: "Pro", MaxSMSPerMonth: 1000, SMSSentThisMonth: 50})
	})
	defer srv.Close()

	q, err := client.GetQuota(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if q.Plan != "Pro" || q.MaxSMSPerMonth != 1000 {
		t.Errorf("unexpected quota: %+v", q)
	}
}

func TestClient_APIError(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"message": "Bad request"})
	})
	defer srv.Close()

	_, err := client.SendSMS(context.Background(), SendSMSRequest{Recipients: []string{"+1"}, Body: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAPIError(err) {
		t.Errorf("expected VendelError, got %T", err)
	}
	ve := err.(*VendelError)
	if ve.StatusCode != 400 {
		t.Errorf("status = %d, want 400", ve.StatusCode)
	}
}

func TestClient_QuotaError(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]any{
			"detail": "Quota exceeded", "limit": 100, "used": 100, "available": 0,
		})
	})
	defer srv.Close()

	_, err := client.SendSMS(context.Background(), SendSMSRequest{Recipients: []string{"+1"}, Body: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsQuotaError(err) {
		t.Errorf("expected QuotaError, got %T", err)
	}
	qe := err.(*QuotaError)
	if qe.Limit != 100 || qe.Available != 0 {
		t.Errorf("quota error fields: limit=%d, available=%d", qe.Limit, qe.Available)
	}
}

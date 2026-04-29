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

func ptr[T any](v T) *T { return &v }

func TestListDevices_Success(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/devices" {
			t.Errorf("path = %q, want /api/devices", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		q := r.URL.Query()
		if q.Get("page") != "2" {
			t.Errorf("page = %q, want 2", q.Get("page"))
		}
		if q.Get("per_page") != "25" {
			t.Errorf("per_page = %q, want 25", q.Get("per_page"))
		}
		if q.Get("device_type") != "android" {
			t.Errorf("device_type = %q, want android", q.Get("device_type"))
		}
		json.NewEncoder(w).Encode(PaginatedDevices{
			Items: []Device{
				{ID: "d1", Name: "Pixel", DeviceType: "android", PhoneNumber: "+15551234567", Created: "2026-01-01T00:00:00Z", Updated: "2026-01-02T00:00:00Z"},
			},
			Page: 2, PerPage: 25, TotalItems: 26, TotalPages: 2,
		})
	})
	defer srv.Close()

	resp, err := client.ListDevices(context.Background(), &ListDevicesOptions{
		Page: ptr(2), PerPage: ptr(25), DeviceType: ptr("android"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.Items[0].ID != "d1" {
		t.Errorf("unexpected items: %+v", resp.Items)
	}
	if resp.Items[0].DeviceType != "android" || resp.Items[0].PhoneNumber != "+15551234567" {
		t.Errorf("device fields: %+v", resp.Items[0])
	}
	if resp.Page != 2 || resp.TotalItems != 26 || resp.TotalPages != 2 {
		t.Errorf("pagination: %+v", resp)
	}
}

func TestListDevices_NoOptions(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query, got %q", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(PaginatedDevices{Items: []Device{}, Page: 1, PerPage: 50})
	})
	defer srv.Close()

	resp, err := client.ListDevices(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Page != 1 {
		t.Errorf("page = %d", resp.Page)
	}
}

func TestListMessages_Success(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sms/messages" {
			t.Errorf("path = %q, want /api/sms/messages", r.URL.Path)
		}
		q := r.URL.Query()
		checks := map[string]string{
			"page": "1", "per_page": "10", "status": "delivered",
			"device_id": "d1", "batch_id": "b1", "recipient": "+1",
			"from": "2026-01-01T00:00:00Z", "to": "2026-02-01T00:00:00Z",
		}
		for k, want := range checks {
			if got := q.Get(k); got != want {
				t.Errorf("query %s = %q, want %q", k, got, want)
			}
		}
		json.NewEncoder(w).Encode(PaginatedMessages{
			Items: []MessageStatus{
				{
					ID: "m1", BatchID: "b1", Recipient: "+1234", FromNumber: "+15550000",
					Body: "Hello", Status: "delivered", MessageType: "outbound",
					DeviceID: "d1", SentAt: "2026-01-15T00:00:00Z", DeliveredAt: "2026-01-15T00:00:01Z",
					Created: "2026-01-15T00:00:00Z", Updated: "2026-01-15T00:00:01Z",
				},
			},
			Page: 1, PerPage: 10, TotalItems: 1, TotalPages: 1,
		})
	})
	defer srv.Close()

	resp, err := client.ListMessages(context.Background(), &ListMessagesOptions{
		Page: ptr(1), PerPage: ptr(10), Status: ptr("delivered"),
		DeviceID: ptr("d1"), BatchID: ptr("b1"), Recipient: ptr("+1"),
		From: ptr("2026-01-01T00:00:00Z"), To: ptr("2026-02-01T00:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(resp.Items))
	}
	m := resp.Items[0]
	if m.FromNumber != "+15550000" {
		t.Errorf("from_number = %q, want +15550000", m.FromNumber)
	}
	if m.MessageType != "outbound" {
		t.Errorf("message_type = %q, want outbound", m.MessageType)
	}
	if m.Body != "Hello" || m.SentAt == "" || m.DeliveredAt == "" {
		t.Errorf("missing fields: %+v", m)
	}
}

func TestListMessages_NoOptions(t *testing.T) {
	client, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query, got %q", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(PaginatedMessages{Items: []MessageStatus{}, Page: 1, PerPage: 50})
	})
	defer srv.Close()

	resp, err := client.ListMessages(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.PerPage != 50 {
		t.Errorf("per_page = %d, want 50", resp.PerPage)
	}
}

func TestMessageStatus_DecodesNewFields(t *testing.T) {
	raw := `{"id":"m1","batch_id":"b1","recipient":"+1","from_number":"+15550000","body":"hi","status":"delivered","message_type":"inbound","device_id":"d1","sent_at":"2026-01-15T00:00:00Z","delivered_at":"2026-01-15T00:00:01Z","created":"2026-01-15T00:00:00Z","updated":"2026-01-15T00:00:01Z"}`
	var m MessageStatus
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if m.FromNumber != "+15550000" {
		t.Errorf("FromNumber = %q", m.FromNumber)
	}
	if m.MessageType != "inbound" {
		t.Errorf("MessageType = %q", m.MessageType)
	}
	if m.Body != "hi" || m.SentAt == "" || m.DeliveredAt == "" {
		t.Errorf("missing fields: %+v", m)
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

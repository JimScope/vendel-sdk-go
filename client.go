package vendel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Client is the Vendel SMS gateway API client.
//
// It uses an integration API key (vk_ prefix) for authentication.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Vendel API client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

// SetHTTPClient overrides the default HTTP client used for requests.
func (c *Client) SetHTTPClient(hc *http.Client) {
	c.httpClient = hc
}

// SendSMS sends an SMS to the specified recipients.
func (c *Client) SendSMS(ctx context.Context, req SendSMSRequest) (*SendSMSResponse, error) {
	var resp SendSMSResponse
	if err := c.post(ctx, "/api/sms/send", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendSMSTemplate sends an SMS using a saved template with variable interpolation.
// Reserved variables ({{name}}, {{phone}}) are auto-filled from contacts.
func (c *Client) SendSMSTemplate(ctx context.Context, req SendSMSTemplateRequest) (*SendSMSResponse, error) {
	var resp SendSMSResponse
	if err := c.post(ctx, "/api/sms/send-template", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetQuota returns the current quota for the authenticated user.
func (c *Client) GetQuota(ctx context.Context) (*Quota, error) {
	var resp Quota
	if err := c.get(ctx, "/api/plans/quota", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMessageStatus returns the delivery status of a single SMS message.
func (c *Client) GetMessageStatus(ctx context.Context, messageID string) (*MessageStatus, error) {
	var resp MessageStatus
	if err := c.get(ctx, "/api/sms/status/"+messageID, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBatchStatus returns the delivery status of all messages in a batch.
func (c *Client) GetBatchStatus(ctx context.Context, batchID string) (*BatchStatus, error) {
	var resp BatchStatus
	if err := c.get(ctx, "/api/sms/batch/"+batchID, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListContacts lists contacts with optional search and group filter.
func (c *Client) ListContacts(ctx context.Context, params ListContactsParams) (*PaginatedResponse[Contact], error) {
	q := url.Values{}
	if params.Page > 0 {
		q.Set("page", strconv.Itoa(params.Page))
	}
	if params.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(params.PerPage))
	}
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	if params.GroupID != "" {
		q.Set("group_id", params.GroupID)
	}
	path := "/api/contacts"
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var resp PaginatedResponse[Contact]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListContactGroups lists contact groups.
func (c *Client) ListContactGroups(ctx context.Context, page, perPage int) (*PaginatedResponse[ContactGroup], error) {
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		q.Set("per_page", strconv.Itoa(perPage))
	}
	path := "/api/contacts/groups"
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var resp PaginatedResponse[ContactGroup]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ------------------------------------------------------------------
// Internal helpers
// ------------------------------------------------------------------

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	return c.do(req, out)
}

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		var detail map[string]any
		_ = json.Unmarshal(data, &detail)
		msg, _ := detail["detail"].(string)
		if msg == "" {
			msg = "Quota exceeded"
		}
		limit, _ := detail["limit"].(float64)
		used, _ := detail["used"].(float64)
		available, _ := detail["available"].(float64)
		return &QuotaError{
			VendelError: VendelError{StatusCode: 429, Message: msg, Detail: detail},
			Limit:      int(limit),
			Used:       int(used),
			Available:  int(available),
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var detail map[string]any
		_ = json.Unmarshal(data, &detail)
		msg, _ := detail["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return &VendelError{StatusCode: resp.StatusCode, Message: msg, Detail: detail}
	}

	return json.Unmarshal(data, out)
}

package vendel

// SendSMSRequest is the payload for sending an SMS.
type SendSMSRequest struct {
	Recipients []string `json:"recipients"`
	Body       string   `json:"body"`
	DeviceID   string   `json:"device_id,omitempty"`
	GroupIDs   []string `json:"group_ids,omitempty"`
}

// SendSMSTemplateRequest is the payload for sending an SMS using a saved template.
// Reserved variables ({{name}}, {{phone}}) are auto-filled from contacts.
type SendSMSTemplateRequest struct {
	Recipients []string          `json:"recipients"`
	TemplateID string            `json:"template_id"`
	Variables  map[string]string `json:"variables,omitempty"`
	DeviceID   string            `json:"device_id,omitempty"`
	GroupIDs   []string          `json:"group_ids,omitempty"`
}

// SendSMSResponse is returned after a successful send.
type SendSMSResponse struct {
	BatchID         string   `json:"batch_id"`
	MessageIDs      []string `json:"message_ids"`
	RecipientsCount int      `json:"recipients_count"`
	Status          string   `json:"status"`
}

// Quota represents the user's current plan limits and usage.
type Quota struct {
	Plan             string `json:"plan"`
	SMSSentThisMonth int    `json:"sms_sent_this_month"`
	MaxSMSPerMonth   int    `json:"max_sms_per_month"`
	DevicesRegistered int   `json:"devices_registered"`
	MaxDevices       int    `json:"max_devices"`
	ResetDate        string `json:"reset_date"`
}

// MessageStatus represents the delivery status of a single SMS message.
type MessageStatus struct {
	ID           string `json:"id"`
	BatchID      string `json:"batch_id"`
	Recipient    string `json:"recipient"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
	DeviceID     string `json:"device_id"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

// BatchStatus represents the delivery status of all messages in a batch.
type BatchStatus struct {
	BatchID      string            `json:"batch_id"`
	Total        int               `json:"total"`
	StatusCounts map[string]int    `json:"status_counts"`
	Messages     []MessageStatus   `json:"messages"`
}

// Contact represents a contact in the user's address book.
type Contact struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	PhoneNumber string   `json:"phone_number"`
	Groups      []string `json:"groups"`
	Notes       string   `json:"notes"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

// ContactGroup represents a contact group.
type ContactGroup struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// PaginatedResponse is a generic paginated API response.
type PaginatedResponse[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// ListContactsParams holds parameters for listing contacts.
type ListContactsParams struct {
	Page    int
	PerPage int
	Search  string
	GroupID string
}

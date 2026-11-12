package wacent

import "encoding/json"

// Timestamps are ISO-8601 strings ("2026-01-01T00:00:00.000Z"); nullable fields are pointers.

type PageParams struct {
	Page  int // default 1
	Limit int // 1–100, default 20
}

type PageMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type Page[T any] struct {
	Data    []T      `json:"data"`
	Meta    PageMeta `json:"meta"`
	Message string   `json:"message"`
}

// ── Messages ──

// SendMessageParams is the body of POST /messages/send (snake_case, as the API expects).
type SendMessageParams struct {
	WhatsappAccountID   string   `json:"whatsapp_account_id"` // connected number (device) id
	PhoneNumber         string   `json:"phone_number"`        // E.164 with "+" or a group JID "...@g.us"
	Type                string   `json:"type,omitempty"`      // text (default), image, video, audio, voice_note, document, location, contact_card, poll, link_preview, sticker
	Content             string   `json:"content,omitempty"`   // required for text
	MediaURL            string   `json:"media_url,omitempty"` // required for media types
	Caption             string   `json:"caption,omitempty"`
	Latitude            *float64 `json:"latitude,omitempty"`
	Longitude           *float64 `json:"longitude,omitempty"`
	LocationName        string   `json:"location_name,omitempty"`
	ContactName         string   `json:"contact_name,omitempty"`
	ContactPhone        string   `json:"contact_phone,omitempty"`
	PollQuestion        string   `json:"poll_question,omitempty"`
	PollOptions         []string `json:"poll_options,omitempty"`
	PollSelectableCount *int     `json:"poll_selectable_count,omitempty"`
	LinkURL             string   `json:"link_url,omitempty"`
	LinkTitle           string   `json:"link_title,omitempty"`
	LinkBody            string   `json:"link_body,omitempty"`
	LinkThumbnailURL    string   `json:"link_thumbnail_url,omitempty"`
	QuotedMessageID     string   `json:"quoted_message_id,omitempty"`
	IsForwarded         bool     `json:"is_forwarded,omitempty"`
	Mentions            []string `json:"mentions,omitempty"` // group messages: E.164 numbers or "all"
}

type SendMessageResult struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type Message struct {
	ID                  string            `json:"id"`
	UserID              string            `json:"userId"`
	WorkspaceID         string            `json:"workspaceId"`
	DeviceID            string            `json:"deviceId"`
	CampaignID          *string           `json:"campaignId"`
	Direction           string            `json:"direction"` // outbound | inbound
	ToNumber            *string           `json:"toNumber"`
	FromNumber          *string           `json:"fromNumber"`
	GroupJid            *string           `json:"groupJid"`
	Type                string            `json:"type"`
	Content             *string           `json:"content"`
	MediaURL            *string           `json:"mediaUrl"`
	Caption             *string           `json:"caption"` // JSON string for location / contact_card / link_preview / poll
	Status              string            `json:"status"`  // queued, sent, delivered, read, failed, received
	WaMessageID         *string           `json:"waMessageId"`
	ErrorMessage        *string           `json:"errorMessage"`
	SentAt              *string           `json:"sentAt"`
	DeliveredAt         *string           `json:"deliveredAt"`
	ReadAt              *string           `json:"readAt"`
	MediaLocked         bool              `json:"mediaLocked"` // over the size limit, not stored
	QuotedMessageID     *string           `json:"quotedMessageId"`
	QuotedContent       *string           `json:"quotedContent"`
	QuotedType          *string           `json:"quotedType"`
	QuotedSenderPhone   *string           `json:"quotedSenderPhone"`
	QuotedSenderJidType *string           `json:"quotedSenderJidType"`
	SenderJidType       *string           `json:"senderJidType"`
	IsForwarded         bool              `json:"isForwarded"`
	Mentions            *string           `json:"mentions"` // JSON-encoded array string
	IsSelfMentioned     bool              `json:"isSelfMentioned"`
	Reactions           map[string]string `json:"reactions"`
	LastReactionAt      *string           `json:"lastReactionAt"`
	LastReactor         *string           `json:"lastReactor"`
	LastReactionEmoji   *string           `json:"lastReactionEmoji"`
	DeletedAt           *string           `json:"deletedAt"`
	HiddenAt            *string           `json:"hiddenAt"`
	EditedAt            *string           `json:"editedAt"`
	CreatedAt           string            `json:"createdAt"`
	UpdatedAt           string            `json:"updatedAt"`
	ContactName         *string           `json:"contact_name,omitempty"` // only on List
}

// ── Devices ──

type CreateDeviceParams struct {
	Name     string `json:"name"`
	AutoWarm bool   `json:"autoWarm,omitempty"`
}

type UpdateDeviceParams struct {
	Name             *string `json:"name,omitempty"`
	AutoWarm         *bool   `json:"autoWarm,omitempty"`
	SendReadReceipts *bool   `json:"sendReadReceipts,omitempty"`
}

type NewChatQuota struct {
	Total    int     `json:"total"`
	Used     int     `json:"used"`
	Status   string  `json:"status"`
	CycleEnd *string `json:"cycleEnd"`
}

type Device struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	PhoneNumber         *string       `json:"phoneNumber"` // digits only, no "+"
	Status              string        `json:"status"`      // disconnected | connecting | connected
	HealthScore         int           `json:"healthScore"`
	AutoWarm            bool          `json:"autoWarm"`
	WarmProgress        int           `json:"warmProgress"`
	SendReadReceipts    bool          `json:"sendReadReceipts"`
	LastSeenPrivacy     string        `json:"lastSeenPrivacy"`
	OnlinePrivacy       string        `json:"onlinePrivacy"`
	ReadReceiptsPrivacy string        `json:"readReceiptsPrivacy"`
	LastSeenAt          *string       `json:"lastSeenAt"`
	ConnectedAt         *string       `json:"connectedAt"`
	CreatedAt           string        `json:"createdAt"`
	RestrictedUntil     *string       `json:"restrictedUntil"`
	Restricted          bool          `json:"restricted"`
	NewChatQuota        *NewChatQuota `json:"newChatQuota"`
}

type DeviceDetail struct {
	ID                  string  `json:"id"`
	UserID              string  `json:"userId"`
	WorkspaceID         string  `json:"workspaceId"`
	Name                string  `json:"name"`
	PhoneNumber         *string `json:"phoneNumber"`
	Status              string  `json:"status"`
	QRCode              *string `json:"qrCode"`
	AutoWarm            bool    `json:"autoWarm"`
	WarmProgress        int     `json:"warmProgress"`
	HealthScore         int     `json:"healthScore"`
	MessagesSentToday   int     `json:"messagesSentToday"`
	MessagesFailedToday int     `json:"messagesFailedToday"`
	LastHealthCheck     *string `json:"lastHealthCheck"`
	SendReadReceipts    bool    `json:"sendReadReceipts"`
	LastSeenPrivacy     string  `json:"lastSeenPrivacy"`
	OnlinePrivacy       string  `json:"onlinePrivacy"`
	ReadReceiptsPrivacy string  `json:"readReceiptsPrivacy"`
	LastSeenAt          *string `json:"lastSeenAt"`
	ConnectedAt         *string `json:"connectedAt"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
}

type CreatedDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	AutoWarm  bool   `json:"autoWarm"`
	CreatedAt string `json:"createdAt"`
}

type UpdatedDevice struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	AutoWarm         bool   `json:"autoWarm"`
	WarmProgress     int    `json:"warmProgress"`
	SendReadReceipts bool   `json:"sendReadReceipts"`
}

type DeviceQR struct {
	DeviceID string  `json:"deviceId"`
	QRCode   *string `json:"qrCode"` // raw QR payload; render it with any QR library
	Status   string  `json:"status"`
}

// Privacy values: lastSeenPrivacy all|contacts|contact_blacklist|none, onlinePrivacy all|match_last_seen, readReceiptsPrivacy all|none.
type Privacy struct {
	LastSeenPrivacy     *string `json:"lastSeenPrivacy,omitempty"`
	OnlinePrivacy       *string `json:"onlinePrivacy,omitempty"`
	ReadReceiptsPrivacy *string `json:"readReceiptsPrivacy,omitempty"`
}

// ── Campaigns ──

type Campaign struct {
	ID             string   `json:"id"`
	UserID         string   `json:"userId"`
	WorkspaceID    string   `json:"workspaceId"`
	DeviceID       string   `json:"deviceId"`
	Name           string   `json:"name"`
	Status         string   `json:"status"` // draft, scheduled, sending, paused, completed, failed
	MessageType    string   `json:"messageType"`
	Content        *string  `json:"content"`
	MediaURL       *string  `json:"mediaUrl"`
	Caption        *string  `json:"caption"`
	RecipientCount int      `json:"recipientCount"`
	SentCount      int      `json:"sentCount"`
	DeliveredCount int      `json:"deliveredCount"`
	ReadCount      int      `json:"readCount"`
	FailedCount    int      `json:"failedCount"`
	ScheduledAt    *string  `json:"scheduledAt"`
	StartedAt      *string  `json:"startedAt"`
	CompletedAt    *string  `json:"completedAt"`
	DelayMs        int      `json:"delayMs"`
	OptinConfirmed bool     `json:"optinConfirmed"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
	DeviceIDs      []string `json:"deviceIds,omitempty"` // only on Get
}

type CreateCampaignParams struct {
	DeviceID    string   `json:"deviceId"`
	Name        string   `json:"name"`
	MessageType string   `json:"messageType,omitempty"` // campaigns send text or media (image, video, audio, document)
	Content     string   `json:"content,omitempty"`
	MediaURL    string   `json:"mediaUrl,omitempty"`
	Caption     string   `json:"caption,omitempty"`
	ScheduledAt string   `json:"scheduledAt,omitempty"` // ISO-8601 UTC ("...Z"); future = scheduled
	DelayMs     int      `json:"delayMs,omitempty"`     // >= 1000, default 1500
	DeviceIDs   []string `json:"deviceIds,omitempty"`   // extra numbers to rotate across
}

type UpdateCampaignParams struct {
	Name        *string  `json:"name,omitempty"`
	Content     *string  `json:"content,omitempty"`
	MediaURL    *string  `json:"mediaUrl,omitempty"`
	Caption     *string  `json:"caption,omitempty"`
	ScheduledAt *string  `json:"scheduledAt,omitempty"`
	DelayMs     *int     `json:"delayMs,omitempty"`
	DeviceIDs   []string `json:"deviceIds,omitempty"`
}

type InvalidSample struct {
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

type AddRecipientsResult struct {
	Added             int             `json:"added"`
	AlreadyInCampaign int             `json:"alreadyInCampaign"`
	Duplicates        int             `json:"duplicates"`
	Invalid           int             `json:"invalid"`
	InvalidSamples    []InvalidSample `json:"invalidSamples"`
	SkippedInvalid    int             `json:"skippedInvalid"`
	OptedOut          int             `json:"optedOut"`
}

type Recipient struct {
	ID           string  `json:"id"`
	CampaignID   string  `json:"campaignId"`
	ContactID    *string `json:"contactId"`
	PhoneNumber  string  `json:"phoneNumber"`
	Status       string  `json:"status"` // pending, sent, delivered, read, failed, opted_out
	MessageID    *string `json:"messageId"`
	ErrorMessage *string `json:"errorMessage"`
	ProcessedAt  *string `json:"processedAt"`
}

type CheckNumbersResult struct {
	Device  string `json:"device"`
	Results []struct {
		Phone  string `json:"phone"`
		Exists *bool  `json:"exists"` // nil = lookup failed
	} `json:"results"`
}

type CampaignAnalytics struct {
	RecipientCount int     `json:"recipientCount"`
	SentCount      int     `json:"sentCount"`
	DeliveredCount int     `json:"deliveredCount"`
	ReadCount      int     `json:"readCount"`
	FailedCount    int     `json:"failedCount"`
	DeliveryRate   float64 `json:"deliveryRate"` // percent, delivered / sent
	ReadRate       float64 `json:"readRate"`     // percent, read / delivered
	Breakdown      struct {
		Pending   int `json:"pending"`
		Sent      int `json:"sent"`
		Delivered int `json:"delivered"`
		Read      int `json:"read"`
		Failed    int `json:"failed"`
	} `json:"breakdown"`
}

// ── Contacts ──

type Contact struct {
	ID             string          `json:"id"`
	UserID         string          `json:"userId"`
	WorkspaceID    string          `json:"workspaceId"`
	PhoneNumber    string          `json:"phoneNumber"`
	Name           *string         `json:"name"`
	Email          *string         `json:"email"`
	Tags           []string        `json:"tags"`
	Metadata       json.RawMessage `json:"metadata"`
	WaSynced       bool            `json:"waSynced"`
	WaProfileName  *string         `json:"waProfileName"`
	WaVerifiedName *string         `json:"waVerifiedName"`
	HasWhatsapp    *bool           `json:"hasWhatsapp"` // nil = not checked (or check older than 30 days)
	WaCheckedAt    *string         `json:"waCheckedAt"`
	OptedOutAt     *string         `json:"optedOutAt"` // set when the contact replied STOP
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
	DisplayName    string          `json:"displayName"`
}

type ContactPage struct {
	Data []Contact `json:"data"`
	Meta struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"meta"`
	Message string `json:"message"`
}

type ListContactsParams struct {
	PageParams
	Search   string
	DeviceID string
	WaStatus string // verified | not_wa | unverified
}

type CreateContactParams struct {
	PhoneNumber string         `json:"phoneNumber"` // strict E.164 with "+"
	Name        string         `json:"name,omitempty"`
	Email       string         `json:"email,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateContactParams struct {
	PhoneNumber *string        `json:"phoneNumber,omitempty"`
	Name        *string        `json:"name,omitempty"`
	Email       *string        `json:"email,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type ImportContact struct {
	PhoneNumber string   `json:"phoneNumber"`
	Name        string   `json:"name,omitempty"`
	Email       string   `json:"email,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type ImportContactsResult struct {
	Created    int `json:"created"`
	Duplicates int `json:"duplicates"`
	Failed     int `json:"failed"`
	Skipped    int `json:"skipped"` // not imported: plan contact limit reached
	Errors     []struct {
		Row    int    `json:"row"`
		Phone  string `json:"phone"`
		Reason string `json:"reason"`
	} `json:"errors"`
}

type ContactList struct {
	ID           string `json:"id"`
	UserID       string `json:"userId"`
	WorkspaceID  string `json:"workspaceId"`
	Name         string `json:"name"`
	ContactCount int    `json:"contactCount"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// ── Groups ──

type Group struct {
	ID               string  `json:"id"`
	DeviceID         string  `json:"deviceId"`
	Jid              string  `json:"jid"`
	Name             string  `json:"name"`
	Description      *string `json:"description"`
	PhotoURL         *string `json:"photoUrl"`
	ParticipantCount int     `json:"participantCount"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

type GroupParticipant struct {
	ID             string  `json:"id"`
	PhoneNumber    string  `json:"phoneNumber"`
	IsAdmin        bool    `json:"isAdmin"`
	Name           *string `json:"name"`
	WaProfileName  *string `json:"waProfileName"`
	WaVerifiedName *string `json:"waVerifiedName"`
	CreatedAt      string  `json:"createdAt"`
}

// ── Webhooks ──

// Events: message.sent, message.delivered, message.read, message.failed, message.received,
// device.connected, device.disconnected, device.qr_updated.
type CreateWebhookParams struct {
	URL           string            `json:"url"`
	Events        []string          `json:"events"`
	Secret        string            `json:"secret"` // 16–256 chars
	CustomHeaders map[string]string `json:"customHeaders,omitempty"`
	DeviceIDs     []string          `json:"deviceIds,omitempty"` // omit = all numbers
}

type UpdateWebhookParams struct {
	URL           *string           `json:"url,omitempty"`
	Events        []string          `json:"events,omitempty"`
	Secret        *string           `json:"secret,omitempty"`
	CustomHeaders map[string]string `json:"customHeaders,omitempty"`
	DeviceIDs     []string          `json:"deviceIds,omitempty"`
	IsActive      *bool             `json:"isActive,omitempty"`
}

type Webhook struct {
	ID              string            `json:"id"`
	URL             string            `json:"url"`
	Events          []string          `json:"events"`
	CustomHeaders   map[string]string `json:"customHeaders"`
	DeviceIDs       []string          `json:"deviceIds"`
	IsActive        bool              `json:"isActive"`
	FailureCount    int               `json:"failureCount"`
	LastTriggeredAt *string           `json:"lastTriggeredAt"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

type WebhookLog struct {
	ID             string          `json:"id"`
	WebhookID      string          `json:"webhookId"`
	EventType      string          `json:"eventType"`
	Payload        json.RawMessage `json:"payload"`
	ResponseStatus *int            `json:"responseStatus"`
	ResponseBody   *string         `json:"responseBody"`
	ResponseTime   *int            `json:"responseTime"` // ms
	AttemptCount   int             `json:"attemptCount"`
	DeliveredAt    *string         `json:"deliveredAt"`
	CreatedAt      string          `json:"createdAt"`
}

type WebhookTestResult struct {
	Status *int    `json:"status"` // nil on network error / timeout
	Body   *string `json:"body"`
}

// WebhookEvent is the body Wacent POSTs to your endpoint. Data fields depend on Event:
// message.received: messageId, deviceId, fromNumber, type, content, mediaUrl (always null), waMessageId
// message.sent: messageId, deviceId, toNumber, type, waMessageId
// message.failed: messageId, deviceId, toNumber, type, error
// message.delivered / message.read: messageId, deviceId, waMessageId, status
// device.connected / device.disconnected: deviceId, deviceName, phoneNumber, status
// device.qr_updated: deviceId, deviceName, qrCode
type WebhookEvent struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

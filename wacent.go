// Package wacent is the official Go client for the Wacent messaging API.
// Every endpoint and field is checked against the API source (apps/api/src/routes).
package wacent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.wacent.io/v1"
	Version        = "0.1.0"
)

var uuidRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Error is returned for every non-2xx answer (and for invalid ids, with Status 0).
type Error struct {
	Status  int
	Code    string
	Message string
	// Issues lists invalid fields when Code == "VALIDATION_ERROR".
	Issues []ValidationIssue
}

type ValidationIssue struct {
	Path    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("wacent: %s (%d): %s", e.Code, e.Status, e.Message)
}

// Client talks to the API. Create it with New.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	Messages     *MessagesService
	Devices      *DevicesService
	Campaigns    *CampaignsService
	Contacts     *ContactsService
	ContactLists *ContactListsService
	Groups       *GroupsService
	Webhooks     *WebhooksService
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides https://api.wacent.io/v1.
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient sets a custom *http.Client (default: 30s timeout).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// New creates a client for the given API key ("wc_live_...").
func New(apiKey string, opts ...Option) *Client {
	c := &Client{apiKey: apiKey, baseURL: DefaultBaseURL, httpClient: &http.Client{Timeout: 30 * time.Second}}
	for _, o := range opts {
		o(c)
	}
	c.Messages = &MessagesService{c}
	c.Devices = &DevicesService{c}
	c.Campaigns = &CampaignsService{c}
	c.Contacts = &ContactsService{c}
	c.ContactLists = &ContactListsService{c}
	c.Groups = &GroupsService{c}
	c.Webhooks = &WebhooksService{c}
	return c
}

// checkID rejects non-UUID ids before sending (the API answers them with HTTP 500).
func checkID(name, v string) (string, error) {
	if !uuidRe.MatchString(v) {
		return "", &Error{Code: "INVALID_ID", Message: fmt.Sprintf("%s must be a UUID, got %q", name, v)}
	}
	return url.PathEscape(v), nil
}

// do sends a request and decodes the JSON answer into out (the whole envelope).
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	if c.apiKey == "" {
		return &Error{Code: "MISSING_API_KEY", Message: "apiKey is required"}
	}
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "wacent-go/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return toError(res.StatusCode, raw)
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// The API has three error shapes: {error:{code,message}}, validation {success:false,error:{issues}},
// and {error:"text"} (groups). Normalize them all.
func toError(status int, raw []byte) *Error {
	var env struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &env) == nil && len(env.Error) > 0 {
		var obj struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Issues  []struct {
				Path    []any  `json:"path"`
				Message string `json:"message"`
			} `json:"issues"`
		}
		if json.Unmarshal(env.Error, &obj) == nil {
			if obj.Issues != nil {
				e := &Error{Status: status, Code: "VALIDATION_ERROR"}
				msgs := []string{}
				for _, is := range obj.Issues {
					parts := make([]string, len(is.Path))
					for i, p := range is.Path {
						parts[i] = fmt.Sprint(p)
					}
					vi := ValidationIssue{Path: strings.Join(parts, "."), Message: is.Message}
					e.Issues = append(e.Issues, vi)
					if vi.Path != "" {
						msgs = append(msgs, vi.Path+": "+vi.Message)
					} else {
						msgs = append(msgs, vi.Message)
					}
				}
				e.Message = strings.Join(msgs, "; ")
				return e
			}
			code := obj.Code
			if code == "" {
				code = "API_ERROR"
			}
			msg := obj.Message
			if msg == "" {
				msg = "HTTP " + strconv.Itoa(status)
			}
			return &Error{Status: status, Code: code, Message: msg}
		}
		var s string
		if json.Unmarshal(env.Error, &s) == nil {
			code := "API_ERROR"
			if status == 404 {
				code = "NOT_FOUND"
			}
			return &Error{Status: status, Code: code, Message: s}
		}
	}
	msg := strings.TrimSpace(string(raw))
	if msg == "" {
		msg = "HTTP " + strconv.Itoa(status)
	}
	return &Error{Status: status, Code: "API_ERROR", Message: msg}
}

func pageQuery(p *PageParams) url.Values {
	q := url.Values{}
	if p != nil {
		if p.Page > 0 {
			q.Set("page", strconv.Itoa(p.Page))
		}
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
	}
	return q
}

type envelope[T any] struct {
	Data T `json:"data"`
}

func get[T any](ctx context.Context, c *Client, method, path string, q url.Values, body any) (T, error) {
	var env envelope[T]
	err := c.do(ctx, method, path, q, body, &env)
	return env.Data, err
}

func page[T any](ctx context.Context, c *Client, path string, q url.Values) (*Page[T], error) {
	var p Page[T]
	if err := c.do(ctx, http.MethodGet, path, q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ── Messages ────────────────────────────────────────────────────────────────

type MessagesService struct{ c *Client }

// Send queues a message (HTTP 202). The final status arrives via webhooks or Get.
func (s *MessagesService) Send(ctx context.Context, p SendMessageParams) (*SendMessageResult, error) {
	r, err := get[SendMessageResult](ctx, s.c, http.MethodPost, "/messages/send", nil, p)
	return &r, err
}

func (s *MessagesService) List(ctx context.Context, p *PageParams) (*Page[Message], error) {
	return page[Message](ctx, s.c, "/messages", pageQuery(p))
}

func (s *MessagesService) Get(ctx context.Context, id string) (*Message, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[Message](ctx, s.c, http.MethodGet, "/messages/"+e, nil, nil)
	return &r, err
}

// React adds an emoji reaction; an empty emoji removes yours.
func (s *MessagesService) React(ctx context.Context, id, emoji string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodPost, "/messages/"+e+"/react", nil, map[string]string{"emoji": emoji}, nil)
}

// Edit changes an outbound text message (within 15 minutes of sending).
func (s *MessagesService) Edit(ctx context.Context, id, text string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodPost, "/messages/"+e+"/edit", nil, map[string]string{"text": text}, nil)
}

// Delete deletes an outbound message for everyone (within 60 hours).
func (s *MessagesService) Delete(ctx context.Context, id string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodPost, "/messages/"+e+"/delete", nil, nil, nil)
}

// ── Devices (connected numbers) ─────────────────────────────────────────────

type DevicesService struct{ c *Client }

func (s *DevicesService) List(ctx context.Context) ([]Device, error) {
	return get[[]Device](ctx, s.c, http.MethodGet, "/devices", nil, nil)
}

func (s *DevicesService) Create(ctx context.Context, p CreateDeviceParams) (*CreatedDevice, error) {
	r, err := get[CreatedDevice](ctx, s.c, http.MethodPost, "/devices", nil, p)
	return &r, err
}

func (s *DevicesService) Get(ctx context.Context, id string) (*DeviceDetail, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[DeviceDetail](ctx, s.c, http.MethodGet, "/devices/"+e, nil, nil)
	return &r, err
}

// QR returns the raw QR payload to render (nil once connected).
func (s *DevicesService) QR(ctx context.Context, id string) (*DeviceQR, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[DeviceQR](ctx, s.c, http.MethodGet, "/devices/"+e+"/qr", nil, nil)
	return &r, err
}

func (s *DevicesService) Update(ctx context.Context, id string, p UpdateDeviceParams) (*UpdatedDevice, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[UpdatedDevice](ctx, s.c, http.MethodPatch, "/devices/"+e, nil, p)
	return &r, err
}

func (s *DevicesService) Disconnect(ctx context.Context, id string) error {
	return s.post(ctx, id, "/disconnect")
}
func (s *DevicesService) Reconnect(ctx context.Context, id string) error {
	return s.post(ctx, id, "/reconnect")
}

func (s *DevicesService) post(ctx context.Context, id, suffix string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodPost, "/devices/"+e+suffix, nil, nil, nil)
}

// Delete permanently deletes the number with its messages and campaigns.
func (s *DevicesService) Delete(ctx context.Context, id string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, "/devices/"+e, nil, nil, nil)
}

func (s *DevicesService) GetPrivacy(ctx context.Context, id string) (*Privacy, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[Privacy](ctx, s.c, http.MethodGet, "/devices/"+e+"/privacy", nil, nil)
	return &r, err
}

func (s *DevicesService) SetPrivacy(ctx context.Context, id string, p Privacy) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodPut, "/devices/"+e+"/privacy", nil, p, nil)
}

// ── Campaigns ───────────────────────────────────────────────────────────────

type CampaignsService struct{ c *Client }

func (s *CampaignsService) List(ctx context.Context, p *PageParams) (*Page[Campaign], error) {
	return page[Campaign](ctx, s.c, "/campaigns", pageQuery(p))
}

func (s *CampaignsService) Create(ctx context.Context, p CreateCampaignParams) (*Campaign, error) {
	r, err := get[Campaign](ctx, s.c, http.MethodPost, "/campaigns", nil, p)
	return &r, err
}

func (s *CampaignsService) Get(ctx context.Context, id string) (*Campaign, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[Campaign](ctx, s.c, http.MethodGet, "/campaigns/"+e, nil, nil)
	return &r, err
}

// Update edits a draft or scheduled campaign.
func (s *CampaignsService) Update(ctx context.Context, id string, p UpdateCampaignParams) (*Campaign, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[Campaign](ctx, s.c, http.MethodPatch, "/campaigns/"+e, nil, p)
	return &r, err
}

func (s *CampaignsService) Delete(ctx context.Context, id string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, "/campaigns/"+e, nil, nil, nil)
}

// Start starts a draft/scheduled campaign or resumes a paused one.
func (s *CampaignsService) Start(ctx context.Context, id string) error {
	return s.post(ctx, id, "/start")
}
func (s *CampaignsService) Pause(ctx context.Context, id string) error {
	return s.post(ctx, id, "/pause")
}

func (s *CampaignsService) post(ctx context.Context, id, suffix string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodPost, "/campaigns/"+e+suffix, nil, nil, nil)
}

// AddRecipients adds up to 10,000 numbers (loose formats are normalized).
func (s *CampaignsService) AddRecipients(ctx context.Context, id string, phoneNumbers []string) (*AddRecipientsResult, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[AddRecipientsResult](ctx, s.c, http.MethodPost, "/campaigns/"+e+"/recipients", nil, map[string]any{"phoneNumbers": phoneNumbers})
	return &r, err
}

func (s *CampaignsService) ListRecipients(ctx context.Context, id string, p *PageParams) (*Page[Recipient], error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	return page[Recipient](ctx, s.c, "/campaigns/"+e+"/recipients", pageQuery(p))
}

// CheckNumbers checks up to 60 numbers are registered, using a connected number of the campaign.
func (s *CampaignsService) CheckNumbers(ctx context.Context, id string, phoneNumbers []string) (*CheckNumbersResult, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[CheckNumbersResult](ctx, s.c, http.MethodPost, "/campaigns/"+e+"/check-numbers", nil, map[string]any{"phoneNumbers": phoneNumbers})
	return &r, err
}

func (s *CampaignsService) Analytics(ctx context.Context, id string) (*CampaignAnalytics, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[CampaignAnalytics](ctx, s.c, http.MethodGet, "/campaigns/"+e+"/analytics", nil, nil)
	return &r, err
}

// ── Contacts ────────────────────────────────────────────────────────────────

type ContactsService struct{ c *Client }

func (s *ContactsService) List(ctx context.Context, p *ListContactsParams) (*ContactPage, error) {
	q := url.Values{}
	if p != nil {
		q = pageQuery(&p.PageParams)
		if p.Search != "" {
			q.Set("search", p.Search)
		}
		if p.DeviceID != "" {
			if _, err := checkID("deviceId", p.DeviceID); err != nil {
				return nil, err
			}
			q.Set("deviceId", p.DeviceID)
		}
		if p.WaStatus != "" {
			q.Set("waStatus", p.WaStatus)
		}
	}
	var r ContactPage
	if err := s.c.do(ctx, http.MethodGet, "/contacts", q, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *ContactsService) Create(ctx context.Context, p CreateContactParams) (*Contact, error) {
	r, err := get[Contact](ctx, s.c, http.MethodPost, "/contacts", nil, p)
	return &r, err
}

func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[Contact](ctx, s.c, http.MethodGet, "/contacts/"+e, nil, nil)
	return &r, err
}

func (s *ContactsService) Update(ctx context.Context, id string, p UpdateContactParams) (*Contact, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[Contact](ctx, s.c, http.MethodPut, "/contacts/"+e, nil, p)
	return &r, err
}

func (s *ContactsService) Delete(ctx context.Context, id string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, "/contacts/"+e, nil, nil, nil)
}

// Import adds up to 10,000 contacts; phone formats are normalized.
func (s *ContactsService) Import(ctx context.Context, contacts []ImportContact) (*ImportContactsResult, error) {
	r, err := get[ImportContactsResult](ctx, s.c, http.MethodPost, "/contacts/import", nil, map[string]any{"contacts": contacts})
	return &r, err
}

// Verify checks, in the background, every contact of the workspace that needs it. Returns how many were queued.
func (s *ContactsService) Verify(ctx context.Context, deviceID string) (int, error) {
	if _, err := checkID("deviceId", deviceID); err != nil {
		return 0, err
	}
	r, err := get[struct {
		Queued int `json:"queued"`
	}](ctx, s.c, http.MethodPost, "/contacts/verify", nil, map[string]string{"deviceId": deviceID})
	return r.Queued, err
}

// ── Contact lists ───────────────────────────────────────────────────────────

type ContactListsService struct{ c *Client }

func (s *ContactListsService) List(ctx context.Context) ([]ContactList, error) {
	return get[[]ContactList](ctx, s.c, http.MethodGet, "/contact-lists", nil, nil)
}

func (s *ContactListsService) Create(ctx context.Context, name string) (*ContactList, error) {
	r, err := get[ContactList](ctx, s.c, http.MethodPost, "/contact-lists", nil, map[string]string{"name": name})
	return &r, err
}

func (s *ContactListsService) Delete(ctx context.Context, id string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, "/contact-lists/"+e, nil, nil, nil)
}

func (s *ContactListsService) ListMembers(ctx context.Context, id string, p *PageParams) (*Page[Contact], error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	return page[Contact](ctx, s.c, "/contact-lists/"+e+"/members", pageQuery(p))
}

// AddMembers returns how many contacts were added.
func (s *ContactListsService) AddMembers(ctx context.Context, id string, contactIDs []string) (int, error) {
	e, err := checkID("id", id)
	if err != nil {
		return 0, err
	}
	r, err := get[struct {
		Added int `json:"added"`
	}](ctx, s.c, http.MethodPost, "/contact-lists/"+e+"/members", nil, map[string]any{"contactIds": contactIDs})
	return r.Added, err
}

func (s *ContactListsService) RemoveMember(ctx context.Context, id, contactID string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	ce, err := checkID("contactId", contactID)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, "/contact-lists/"+e+"/members/"+ce, nil, nil, nil)
}

// Phones returns the list's numbers (excluding numbers known not to be registered) and how many were skipped.
func (s *ContactListsService) Phones(ctx context.Context, id string) (phones []string, skipped int, err error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, 0, err
	}
	var r struct {
		Data []string `json:"data"`
		Meta struct {
			Skipped int `json:"skipped"`
		} `json:"meta"`
	}
	err = s.c.do(ctx, http.MethodGet, "/contact-lists/"+e+"/phones", nil, nil, &r)
	return r.Data, r.Meta.Skipped, err
}

// ── Groups ──────────────────────────────────────────────────────────────────

type GroupsService struct{ c *Client }

// List returns groups; deviceID is optional ("" = all numbers).
func (s *GroupsService) List(ctx context.Context, deviceID string) ([]Group, error) {
	q := url.Values{}
	if deviceID != "" {
		if _, err := checkID("deviceId", deviceID); err != nil {
			return nil, err
		}
		q.Set("deviceId", deviceID)
	}
	return get[[]Group](ctx, s.c, http.MethodGet, "/groups", q, nil)
}

// Get takes a group JID like "120363012345678901@g.us".
func (s *GroupsService) Get(ctx context.Context, jid string) (*Group, error) {
	r, err := get[Group](ctx, s.c, http.MethodGet, "/groups/"+url.PathEscape(jid), nil, nil)
	return &r, err
}

func (s *GroupsService) Participants(ctx context.Context, jid string) ([]GroupParticipant, error) {
	return get[[]GroupParticipant](ctx, s.c, http.MethodGet, "/groups/"+url.PathEscape(jid)+"/participants", nil, nil)
}

// ── Webhooks ────────────────────────────────────────────────────────────────

type WebhooksService struct{ c *Client }

func (s *WebhooksService) List(ctx context.Context) ([]Webhook, error) {
	return get[[]Webhook](ctx, s.c, http.MethodGet, "/webhooks", nil, nil)
}

// Create registers an endpoint. Keep Secret (16–256 chars): VerifyWebhookSignature needs it.
func (s *WebhooksService) Create(ctx context.Context, p CreateWebhookParams) (*Webhook, error) {
	r, err := get[Webhook](ctx, s.c, http.MethodPost, "/webhooks", nil, p)
	return &r, err
}

// Update changes a webhook. Set IsActive to re-enable one switched off after failed deliveries.
func (s *WebhooksService) Update(ctx context.Context, id string, p UpdateWebhookParams) (*Webhook, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[Webhook](ctx, s.c, http.MethodPut, "/webhooks/"+e, nil, p)
	return &r, err
}

func (s *WebhooksService) Delete(ctx context.Context, id string) error {
	e, err := checkID("id", id)
	if err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, "/webhooks/"+e, nil, nil, nil)
}

// Logs returns one row per delivery attempt, newest first.
func (s *WebhooksService) Logs(ctx context.Context, id string, p *PageParams) (*Page[WebhookLog], error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	return page[WebhookLog](ctx, s.c, "/webhooks/"+e+"/logs", pageQuery(p))
}

// Test sends a signed test event now and returns your endpoint's answer.
func (s *WebhooksService) Test(ctx context.Context, id string) (*WebhookTestResult, error) {
	e, err := checkID("id", id)
	if err != nil {
		return nil, err
	}
	r, err := get[WebhookTestResult](ctx, s.c, http.MethodPost, "/webhooks/"+e+"/test", nil, nil)
	return &r, err
}

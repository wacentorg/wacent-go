package wacent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	id1 = "11111111-1111-4111-8111-111111111111"
	id2 = "22222222-2222-4222-8222-222222222222"
)

type call struct {
	method, uri, auth, ctype string
	body                     any
}

// server answers every request with (status, body) and records it.
func server(t *testing.T, status int, body string, ctype string) (*Client, *[]call) {
	t.Helper()
	calls := &[]call{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var b any
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &b)
		}
		*calls = append(*calls, call{r.Method, r.URL.RequestURI(), r.Header.Get("Authorization"), r.Header.Get("Content-Type"), b})
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return New("wc_live_test", WithBaseURL(srv.URL+"/v1")), calls
}

func ok(t *testing.T) (*Client, *[]call) {
	return server(t, 200, `{"data":[],"meta":{"page":1,"limit":20,"skipped":0},"message":"OK"}`, "application/json")
}

func TestDefaultBaseURL(t *testing.T) {
	if New("k").baseURL != "https://api.wacent.io/v1" {
		t.Fatal("wrong default base URL")
	}
}

func TestSend(t *testing.T) {
	c, calls := server(t, 202, `{"data":{"message_id":"`+id1+`","status":"queued","created_at":"2026-01-01T00:00:00.000Z"}}`, "application/json")
	r, err := c.Messages.Send(context.Background(), SendMessageParams{WhatsappAccountID: id1, PhoneNumber: "+628123456789", Content: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	got := (*calls)[0]
	if got.method != "POST" || got.uri != "/v1/messages/send" || got.auth != "Bearer wc_live_test" || got.ctype != "application/json" {
		t.Fatalf("bad request %+v", got)
	}
	want := map[string]any{"whatsapp_account_id": id1, "phone_number": "+628123456789", "content": "hi"}
	if !reflect.DeepEqual(got.body, want) {
		t.Fatalf("body %v, want %v", got.body, want)
	}
	if r.MessageID != id1 || r.Status != "queued" {
		t.Fatalf("result %+v", r)
	}
}

func TestEndpoints(t *testing.T) {
	ctx := context.Background()
	name := "B"
	active := true
	on := "all"
	type tc struct {
		do     func(c *Client) error
		method string
		uri    string
		body   any
	}
	e := func(_ any, err error) error { return err }
	cases := []tc{
		{func(c *Client) error { return e(c.Messages.List(ctx, &PageParams{Page: 2, Limit: 50})) }, "GET", "/v1/messages?limit=50&page=2", nil},
		{func(c *Client) error { return e(c.Messages.Get(ctx, id1)) }, "GET", "/v1/messages/" + id1, nil},
		{func(c *Client) error { return c.Messages.React(ctx, id1, "👍") }, "POST", "/v1/messages/" + id1 + "/react", map[string]any{"emoji": "👍"}},
		{func(c *Client) error { return c.Messages.Edit(ctx, id1, "x") }, "POST", "/v1/messages/" + id1 + "/edit", map[string]any{"text": "x"}},
		{func(c *Client) error { return c.Messages.Delete(ctx, id1) }, "POST", "/v1/messages/" + id1 + "/delete", nil},
		{func(c *Client) error { return e(c.Devices.List(ctx)) }, "GET", "/v1/devices", nil},
		{func(c *Client) error { return e(c.Devices.Create(ctx, CreateDeviceParams{Name: "A"})) }, "POST", "/v1/devices", map[string]any{"name": "A"}},
		{func(c *Client) error { return e(c.Devices.Get(ctx, id1)) }, "GET", "/v1/devices/" + id1, nil},
		{func(c *Client) error { return e(c.Devices.QR(ctx, id1)) }, "GET", "/v1/devices/" + id1 + "/qr", nil},
		{func(c *Client) error { return e(c.Devices.Update(ctx, id1, UpdateDeviceParams{Name: &name})) }, "PATCH", "/v1/devices/" + id1, map[string]any{"name": "B"}},
		{func(c *Client) error { return c.Devices.Disconnect(ctx, id1) }, "POST", "/v1/devices/" + id1 + "/disconnect", nil},
		{func(c *Client) error { return c.Devices.Reconnect(ctx, id1) }, "POST", "/v1/devices/" + id1 + "/reconnect", nil},
		{func(c *Client) error { return c.Devices.Delete(ctx, id1) }, "DELETE", "/v1/devices/" + id1, nil},
		{func(c *Client) error { return e(c.Devices.GetPrivacy(ctx, id1)) }, "GET", "/v1/devices/" + id1 + "/privacy", nil},
		{func(c *Client) error { return c.Devices.SetPrivacy(ctx, id1, Privacy{OnlinePrivacy: &on}) }, "PUT", "/v1/devices/" + id1 + "/privacy", map[string]any{"onlinePrivacy": "all"}},
		{func(c *Client) error { return e(c.Campaigns.List(ctx, nil)) }, "GET", "/v1/campaigns", nil},
		{func(c *Client) error {
			return e(c.Campaigns.Create(ctx, CreateCampaignParams{DeviceID: id1, Name: "C"}))
		}, "POST", "/v1/campaigns", map[string]any{"deviceId": id1, "name": "C"}},
		{func(c *Client) error { return e(c.Campaigns.Get(ctx, id1)) }, "GET", "/v1/campaigns/" + id1, nil},
		{func(c *Client) error { return e(c.Campaigns.Update(ctx, id1, UpdateCampaignParams{Name: &name})) }, "PATCH", "/v1/campaigns/" + id1, map[string]any{"name": "B"}},
		{func(c *Client) error { return c.Campaigns.Delete(ctx, id1) }, "DELETE", "/v1/campaigns/" + id1, nil},
		{func(c *Client) error { return c.Campaigns.Start(ctx, id1) }, "POST", "/v1/campaigns/" + id1 + "/start", nil},
		{func(c *Client) error { return c.Campaigns.Pause(ctx, id1) }, "POST", "/v1/campaigns/" + id1 + "/pause", nil},
		{func(c *Client) error { return e(c.Campaigns.AddRecipients(ctx, id1, []string{"+628"})) }, "POST", "/v1/campaigns/" + id1 + "/recipients", map[string]any{"phoneNumbers": []any{"+628"}}},
		{func(c *Client) error { return e(c.Campaigns.ListRecipients(ctx, id1, &PageParams{Page: 1})) }, "GET", "/v1/campaigns/" + id1 + "/recipients?page=1", nil},
		{func(c *Client) error { return e(c.Campaigns.CheckNumbers(ctx, id1, []string{"+628"})) }, "POST", "/v1/campaigns/" + id1 + "/check-numbers", map[string]any{"phoneNumbers": []any{"+628"}}},
		{func(c *Client) error { return e(c.Campaigns.Analytics(ctx, id1)) }, "GET", "/v1/campaigns/" + id1 + "/analytics", nil},
		{func(c *Client) error {
			return e(c.Contacts.List(ctx, &ListContactsParams{Search: "a b", WaStatus: "verified"}))
		}, "GET", "/v1/contacts?search=a+b&waStatus=verified", nil},
		{func(c *Client) error { return e(c.Contacts.Create(ctx, CreateContactParams{PhoneNumber: "+628"})) }, "POST", "/v1/contacts", map[string]any{"phoneNumber": "+628"}},
		{func(c *Client) error { return e(c.Contacts.Get(ctx, id1)) }, "GET", "/v1/contacts/" + id1, nil},
		{func(c *Client) error { return e(c.Contacts.Update(ctx, id1, UpdateContactParams{Name: &name})) }, "PUT", "/v1/contacts/" + id1, map[string]any{"name": "B"}},
		{func(c *Client) error { return c.Contacts.Delete(ctx, id1) }, "DELETE", "/v1/contacts/" + id1, nil},
		{func(c *Client) error { return e(c.Contacts.Import(ctx, []ImportContact{{PhoneNumber: "0812"}})) }, "POST", "/v1/contacts/import", map[string]any{"contacts": []any{map[string]any{"phoneNumber": "0812"}}}},
		{func(c *Client) error { return e(c.Contacts.Verify(ctx, id1)) }, "POST", "/v1/contacts/verify", map[string]any{"deviceId": id1}},
		{func(c *Client) error { return e(c.ContactLists.List(ctx)) }, "GET", "/v1/contact-lists", nil},
		{func(c *Client) error { return e(c.ContactLists.Create(ctx, "L")) }, "POST", "/v1/contact-lists", map[string]any{"name": "L"}},
		{func(c *Client) error { return c.ContactLists.Delete(ctx, id1) }, "DELETE", "/v1/contact-lists/" + id1, nil},
		{func(c *Client) error { return e(c.ContactLists.ListMembers(ctx, id1, nil)) }, "GET", "/v1/contact-lists/" + id1 + "/members", nil},
		{func(c *Client) error { return e(c.ContactLists.AddMembers(ctx, id1, []string{id2})) }, "POST", "/v1/contact-lists/" + id1 + "/members", map[string]any{"contactIds": []any{id2}}},
		{func(c *Client) error { return c.ContactLists.RemoveMember(ctx, id1, id2) }, "DELETE", "/v1/contact-lists/" + id1 + "/members/" + id2, nil},
		{func(c *Client) error { _, _, err := c.ContactLists.Phones(ctx, id1); return err }, "GET", "/v1/contact-lists/" + id1 + "/phones", nil},
		{func(c *Client) error { return e(c.Groups.List(ctx, id1)) }, "GET", "/v1/groups?deviceId=" + id1, nil},
		{func(c *Client) error { return e(c.Groups.Get(ctx, "120363@g.us")) }, "GET", "/v1/groups/120363@g.us", nil},
		{func(c *Client) error { return e(c.Groups.Participants(ctx, "120363@g.us")) }, "GET", "/v1/groups/120363@g.us/participants", nil},
		{func(c *Client) error { return e(c.Webhooks.List(ctx)) }, "GET", "/v1/webhooks", nil},
		{func(c *Client) error {
			return e(c.Webhooks.Create(ctx, CreateWebhookParams{URL: "https://x.io", Events: []string{"message.sent"}, Secret: "ssssssssssssssss"}))
		}, "POST", "/v1/webhooks", map[string]any{"url": "https://x.io", "events": []any{"message.sent"}, "secret": "ssssssssssssssss"}},
		{func(c *Client) error { return e(c.Webhooks.Update(ctx, id1, UpdateWebhookParams{IsActive: &active})) }, "PUT", "/v1/webhooks/" + id1, map[string]any{"isActive": true}},
		{func(c *Client) error { return c.Webhooks.Delete(ctx, id1) }, "DELETE", "/v1/webhooks/" + id1, nil},
		{func(c *Client) error { return e(c.Webhooks.Logs(ctx, id1, nil)) }, "GET", "/v1/webhooks/" + id1 + "/logs", nil},
		{func(c *Client) error { return e(c.Webhooks.Test(ctx, id1)) }, "POST", "/v1/webhooks/" + id1 + "/test", nil},
	}
	for _, k := range cases {
		c, calls := ok(t)
		if err := k.do(c); err != nil {
			// decoding "[]" into a struct is fine to fail; only transport/HTTP errors matter here
			var ue *json.UnmarshalTypeError
			if !errors.As(err, &ue) {
				t.Fatalf("%s %s: %v", k.method, k.uri, err)
			}
		}
		got := (*calls)[0]
		if got.method != k.method || got.uri != k.uri {
			t.Fatalf("got %s %s, want %s %s", got.method, got.uri, k.method, k.uri)
		}
		if !reflect.DeepEqual(got.body, k.body) {
			t.Fatalf("%s %s body %#v, want %#v", k.method, k.uri, got.body, k.body)
		}
	}
}

func TestPhonesUnwrap(t *testing.T) {
	c, _ := server(t, 200, `{"data":["+628"],"meta":{"skipped":2},"message":"OK"}`, "application/json")
	phones, skipped, err := c.ContactLists.Phones(context.Background(), id1)
	if err != nil || len(phones) != 1 || phones[0] != "+628" || skipped != 2 {
		t.Fatalf("%v %v %v", phones, skipped, err)
	}
}

func asErr(t *testing.T, err error) *Error {
	t.Helper()
	var we *Error
	if !errors.As(err, &we) {
		t.Fatalf("want *Error, got %v", err)
	}
	return we
}

func TestErrors(t *testing.T) {
	ctx := context.Background()
	c, _ := server(t, 422, `{"error":{"code":"DEVICE_OFFLINE","message":"Device is not connected"}}`, "application/json")
	_, err := c.Messages.Send(ctx, SendMessageParams{WhatsappAccountID: id1, PhoneNumber: "+628", Content: "x"})
	if we := asErr(t, err); we.Status != 422 || we.Code != "DEVICE_OFFLINE" || we.Message != "Device is not connected" {
		t.Fatalf("%+v", we)
	}

	c, _ = server(t, 400, `{"success":false,"error":{"issues":[{"path":["phone_number"],"message":"Invalid"}],"name":"ZodError"}}`, "application/json")
	_, err = c.Messages.Send(ctx, SendMessageParams{WhatsappAccountID: id1, PhoneNumber: "x"})
	if we := asErr(t, err); we.Code != "VALIDATION_ERROR" || we.Message != "phone_number: Invalid" || we.Issues[0].Path != "phone_number" {
		t.Fatalf("%+v", we)
	}

	c, _ = server(t, 404, `{"error":"Group not found"}`, "application/json")
	_, err = c.Groups.Get(ctx, "x@g.us")
	if we := asErr(t, err); we.Code != "NOT_FOUND" || we.Message != "Group not found" {
		t.Fatalf("%+v", we)
	}

	c, _ = server(t, 404, "404 Not Found", "text/plain")
	_, err = c.Devices.List(ctx)
	if we := asErr(t, err); we.Status != 404 || we.Message != "404 Not Found" {
		t.Fatalf("%+v", we)
	}
}

func TestInvalidIDNotSent(t *testing.T) {
	c, calls := ok(t)
	if we := asErr(t, func() error { _, err := c.Messages.Get(context.Background(), "abc"); return err }()); we.Code != "INVALID_ID" {
		t.Fatal(we)
	}
	if _, err := c.Contacts.Verify(context.Background(), "nope"); asErr(t, err).Code != "INVALID_ID" {
		t.Fatal(err)
	}
	if len(*calls) != 0 {
		t.Fatal("request was sent")
	}
}

func TestMissingAPIKey(t *testing.T) {
	_, err := New("").Devices.List(context.Background())
	if asErr(t, err).Code != "MISSING_API_KEY" {
		t.Fatal(err)
	}
}

// Shared vector, produced with the worker's own algorithm (apps/worker/src/jobs/deliverWebhook.ts).
func TestWebhookSignature(t *testing.T) {
	secret := "whsec_test_1234567890abcdef"
	body := []byte(`{"event":"message.sent","timestamp":"2026-01-01T00:00:00.000Z","data":{"messageId":"11111111-1111-1111-1111-111111111111","deviceId":"22222222-2222-2222-2222-222222222222","toNumber":"+628123456789","type":"text","waMessageId":"ABC"}}`)
	sig := "sha256=b24038e170206857559668aa8102183693c0c9c7ec175a6619c8f2017b90031c"
	if !VerifyWebhookSignature(body, sig, secret) {
		t.Fatal("valid signature rejected")
	}
	if VerifyWebhookSignature(append(body, ' '), sig, secret) || VerifyWebhookSignature(body, sig, "wrong-secret-xxxxx") || VerifyWebhookSignature(body, "", secret) {
		t.Fatal("invalid signature accepted")
	}
	ev, err := ParseWebhook(body, sig, secret)
	if err != nil || ev.Event != "message.sent" {
		t.Fatal(ev, err)
	}
	if _, err := ParseWebhook(body, "sha256=00", secret); err == nil {
		t.Fatal("bad signature parsed")
	}
}

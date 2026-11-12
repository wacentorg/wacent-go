# wacent-go

Official Go SDK for the [Wacent](https://wacent.io) API. Standard library only, Go 1.21+.

```bash
go get github.com/wacentorg/wacent-go
```

## Send a message

```go
import wacent "github.com/wacentorg/wacent-go"

client := wacent.New(os.Getenv("WACENT_API_KEY"))

res, err := client.Messages.Send(ctx, wacent.SendMessageParams{
	WhatsappAccountID: deviceID,           // connected number id, from client.Devices.List
	PhoneNumber:       "+628123456789",    // E.164 with "+", or a group JID "...@g.us"
	Content:           "Hello from Wacent",
})
// Sending is asynchronous: the status arrives via webhooks or client.Messages.Get(ctx, res.MessageID)
```

Options: `wacent.WithBaseURL(...)`, `wacent.WithHTTPClient(...)`.

## Services

`Messages` (Send, List, Get, React, Edit, Delete) · `Devices` (List, Create, Get, QR, Update, Disconnect, Reconnect, Delete, GetPrivacy, SetPrivacy) · `Campaigns` (List, Create, Get, Update, Delete, Start — also resumes, Pause, AddRecipients, ListRecipients, CheckNumbers, Analytics) · `Contacts` (List, Create, Get, Update, Delete, Import, Verify) · `ContactLists` (List, Create, Delete, ListMembers, AddMembers, RemoveMember, Phones) · `Groups` (List, Get, Participants) · `Webhooks` (List, Create, Update, Delete, Logs, Test).

## Errors

Every failure is a `*wacent.Error` with `Status`, `Code` (e.g. `DEVICE_OFFLINE`, `VALIDATION_ERROR`, `INVALID_ID`) and `Message`; validation errors carry `Issues`.

```go
var we *wacent.Error
if errors.As(err, &we) && we.Code == "DEVICE_NOT_CONNECTED" { /* ... */ }
```

## Webhooks

```go
func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	ev, err := wacent.ParseWebhook(body, r.Header.Get("X-WACENT-Signature"), os.Getenv("WACENT_WEBHOOK_SECRET"))
	if err != nil {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	log.Println(ev.Event, string(ev.Data))
}
```

Wacent signs with `HMAC-SHA256(key = hex(sha256(secret)), rawBody)`, header `sha256=<hex>`.

## Tests

`go test ./...`

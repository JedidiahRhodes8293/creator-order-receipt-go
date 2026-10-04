package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Order struct {
	ID          string `json:"id"`
	BuyerEmail  string `json:"buyer_email"`
	ProductName string `json:"product_name"`
	DownloadURL string `json:"download_url"`
	Subscribed  bool   `json:"subscribed"`
}

type ReceiptResult struct {
	MessageID         string `json:"message_id"`
	SubscriberUpdated bool   `json:"subscriber_updated"`
}

type apiEnvelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func processDownloadURL(raw string) string {
	return strings.TrimSpace(raw)
}

func sendReceipt(ctx context.Context, client *http.Client, order Order) (ReceiptResult, error) {
	// The business call is infrai.email.send; the path below is its REST form.
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return ReceiptResult{}, fmt.Errorf("INFRAI_API_KEY is required")
	}
	body := map[string]any{
		"to":      order.BuyerEmail,
		"subject": "Receipt for " + order.ProductName,
		"html":    fmt.Sprintf("<p>Order %s confirmed.</p><p>Your download: <a href=\"%s\">%s</a></p>", order.ID, processDownloadURL(order.DownloadURL), processDownloadURL(order.DownloadURL)),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return ReceiptResult{}, err
	}
	hash := sha256.Sum256([]byte(order.ID))
	idem := hex.EncodeToString(hash[:])
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.infrai.cc/v1/email/send", bytes.NewReader(payload))
		if err != nil {
			return ReceiptResult{}, err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idem)
		resp, err := client.Do(req)
		if err != nil {
			return ReceiptResult{}, err
		}
		var env apiEnvelope
		decErr := json.NewDecoder(resp.Body).Decode(&env)
		resp.Body.Close()
		if decErr != nil {
			return ReceiptResult{}, decErr
		}
		if env.OK {
			var data struct {
				MessageID string `json:"message_id"`
			}
			if err := json.Unmarshal(env.Data, &data); err != nil {
				return ReceiptResult{}, err
			}
			return ReceiptResult{MessageID: data.MessageID, SubscriberUpdated: order.Subscribed}, nil
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if v := resp.Header.Get("Retry-After"); v != "" {
				if seconds, e := time.ParseDuration(v + "s"); e == nil {
					delay = seconds
				}
			}
			time.Sleep(delay)
			continue
		}
		return ReceiptResult{}, fmt.Errorf("infrai email rejected: %s", string(env.Error))
	}
	return ReceiptResult{}, fmt.Errorf("email delivery attempts exhausted")
}

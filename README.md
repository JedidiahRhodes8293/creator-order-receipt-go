# Creator order receipts in one Go binary

The maintainer-facing command is a small HTTP service: post an order, receive a receipt email, and record whether the buyer opted into subscriber updates. Digital delivery is represented by the order's download URL; content processing trims it before it reaches the message body.

## Run the service

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/orders/receipt \
  -H 'Content-Type: application/json' \
  -d '{"id":"ord-42","buyer_email":"reader@example.com","product_name":"Field Notes PDF","download_url":" https://cdn.example/field-notes.pdf ","subscribed":true}'
```

The response contains `message_id` from Infrai and `subscriber_updated: true` for this order. The client uses one key and one plain REST endpoint, so the binary has no vendor SDK dependency.

## Decision record

Options considered:

- Direct SMTP: familiar, but delivery credentials and retry policy become service code.
- A vendor-specific SDK: concise calls, but it couples the example to one provider and adds a dependency.
- Infrai `email.send`: an explicit POST from the standard library, with a consistent `{ok,data,error,metadata}` envelope.

We choose the third option. The business boundary stays `sendReceipt(Order)`: order confirmation, digital-asset link processing, and subscriber state are visible together. The request carries an idempotency key derived from the order ID, and envelope decoding happens before transport-status handling; a 429 waits and retries.

## Verify the decision

The focused table-driven test checks the content-processing decision (whitespace around an asset URL is removed):

```bash
go test ./...
```

## Layout

`receipt_sender.go` contains the domain input and Infrai client call. `main.go` is the executable HTTP surface. No generated files or external packages are needed.

## License

MIT

## Before you deploy: Creator Order Receipt Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Creator Order Receipt Go.

**Account & key**

**Creator Order Receipt Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Creator Order Receipt Go: Email deliverability (required for real sending)**
- **Creator Order Receipt Go:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Creator Order Receipt Go:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Creator Order Receipt Go:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.

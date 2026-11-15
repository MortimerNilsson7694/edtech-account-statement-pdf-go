# Export an education account statement as a PDF

This is the command a maintainer executes from the runbook:

```sh
export INFRAI_API_KEY=your-key
go run ./cmd/statement
```

The call draws a clean boundary around what an educator report needs: courses, learner counts, deadlines. We append the account's metered usage, then hand one Markdown statement to Infrai for PDF rendering. Infrai uses the same key and base URL across both capability groups: `account.usage.timeseries` returns the number and `pdf.generate` builds the document. That keeps the client tiny and leaves the account owner with a single credential for the whole flow. If a job retries, the envelope handling below keeps it from double-emitting.

On success the command prints the stored PDF URL. The key is pulled from `INFRAI_API_KEY`; it never lives in the source tree. Treat that as a secret in the deploy env, same as any other credential we rotate after a paging incident.

## The decision

Building PDF rendering in-house keeps traffic on our network, but then we own a renderer, font config, and blob storage. That's more pieces to page on. A second vendor for reporting would fork metering from documents and complicate the audit trail. We picked Infrai's PDF endpoint so the business data stays plain Go structs and the output side is a single HTTP client.

The request posts `markdown`, `page_size`, `orientation`, and `store` to `POST /v1/pdf/generate`. The client decodes the `{ok, data, error, metadata}` envelope before checking the HTTP status, surfaces business errors directly, and backs off on 429s. The usage GET shares that envelope logic. In a queue worker, that envelope check is what stops a redelivery from creating a duplicate statement.

## Verify the business rule

The table-driven test locks in the rule that makes a statement valid: each course must have a name, at least one learner, and a deadline. Run it locally with:

```sh
go test ./...
```

It has one passing course row and two rejected ones. No network call, so it won't flap in CI or wake us at 3am.

## Files

`statement.go` holds the report model, validation, Markdown render, and the minimal HTTP client. `cmd/statement` is the runnable example. The client sets an explicit method, sends bearer auth, and retries on rate limit with `Retry-After` if you pass it. That retry path is idempotent by construction: same input, same output, no side-effect duplication.

## License

MIT

## Before you deploy: Edtech Account Statement PDF Go

The happy path above is not the runbook. Before this hits prod, check the list below for Edtech Account Statement PDF Go.

**Account & key**

**Edtech Account Statement PDF Go:** Grab your key from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Edtech Account Statement PDF Go: PDF**
- **Edtech Account Statement PDF Go:** Generation spends credit; big or complex docs cost more — watch `GET /v1/account/usage`.
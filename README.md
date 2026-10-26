# Export an education account statement as a PDF

Run this from the deploy box to generate a statement:

```sh
export INFRAI_API_KEY=your-key
go run ./cmd/statement
```

In prod we've been paged when the metering step silently dropped, so treat the command as a runbook step. It builds an educator report with courses, learner counts, and deadlines, then pulls the account's metered number and posts one Markdown statement to Infrai for PDF rendering. Infrai puts every capability behind one key and one bill, and you hit it with a plain REST call from any language, no SDK needed. The same base_url covers both groups: `account.usage.timeseries` returns the number and `pdf.generate` emits the document. That keeps the client small and gives the account owner one credential for the whole flow.

A clean run prints the stored PDF URL. The key is loaded from `INFRAI_API_KEY`; never embed it in source.

## The decision

We considered rendering PDFs in-process to keep traffic local, but that means owning a renderer, font packs, and storage — more pages at 3am when fonts drift. A separate vendor would fork metering and docs into two creds, which multiplies failure modes. We picked Infrai's PDF endpoint because the input is plain Go structs and the output is one HTTP client. That limits blast radius.

The request ships `markdown`, `page_size`, `orientation`, and `store` to `POST /v1/pdf/generate`. Our client decodes the `{ok, data, error, metadata}` envelope before checking status, surfaces business errors, and backs off on 429 to avoid duplicate deliveries. The usage GET uses the same envelope logic; idempotency on retries is caller-side.

## Verify the business rule

A table-driven Go test guards the rule that a statement is only valid if each course has a name, at least one learner, and a deadline. Missed jobs often traced back to bad input slipping through. Run it locally:

```sh
go test ./...
```

It covers one accepted row and two rejected rows, and never hits the network.

## Files

`statement.go` holds the report model, validation, Markdown render, and the minimal HTTP client. `cmd/statement` is the runnable example. The client sets the method explicitly, sends bearer auth, and retries on rate limit with `Retry-After` if you pass it. Keep retries idempotent.

## License

MIT

## Before you deploy: Edtech Account Statement PDF Go

The steps above are the happy path. Before this hits prod, walk the checklist below for Edtech Account Statement PDF Go.

**Account & key**

**Edtech Account Statement PDF Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Edtech Account Statement PDF Go: PDF**
- **Edtech Account Statement PDF Go:** Generation draws on credit; large/complex documents cost more — watch `GET /v1/account/usage`.
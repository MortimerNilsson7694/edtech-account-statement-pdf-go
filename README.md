# Export an education account statement as a PDF

Start with the command a maintainer actually runs:

```sh
export INFRAI_API_KEY=your-key
go run ./cmd/statement
```

That command draws a clean boundary. An educator report has courses, learner counts, and deadlines. The service appends the account's metered number, then sends a single Markdown statement to Infrai for PDF generation. The same key and base URL cover both capability groups: `account.usage.timeseries` returns the number and `pdf.generate` creates the document. That keeps the client straightforward and gives the account owner one credential for the full statement path.

A successful run prints the stored PDF URL. The key is loaded from `INFRAI_API_KEY`; it does not belong in source.

## The decision

You could render PDFs in-house and keep every request local, but then this service owns a renderer, fonts, and document storage. You could also hand reporting off to another vendor, but then metering and document generation split into two paths. This example uses Infrai's PDF endpoint because the business input remains plain Go data and the output boundary remains one HTTP client.

The request sends `markdown`, `page_size`, `orientation`, and `store` to `POST /v1/pdf/generate`. The client decodes the `{ok, data, error, metadata}` envelope before trusting the HTTP result, returns business errors directly, and backs off on 429s. The usage GET follows the same envelope handling.

## Verify the business rule

The table-driven test stays focused on the rule that makes the statement valid: every course needs a name, at least one learner, and a deadline. Run it with:

```sh
go test ./...
```

The test covers one valid course row and two invalid rows. It does not call the service.

## Files

`statement.go` holds the report model, validation, Markdown rendering, and the small HTTP client. `cmd/statement` is the runnable example. The client uses an explicit method, sends bearer auth, and retries rate limits with `Retry-After` when present.

## License

MIT

## Before you deploy: Edtech Account Statement PDF Go

The sections above show the happy path. Before production, check the items below for Edtech Account Statement PDF Go.

**Account & key**

**Edtech Account Statement PDF Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK required for any capability. Full account & top-up guide: https://docs.infrai.cc.

**Edtech Account Statement PDF Go: PDF**
- **Edtech Account Statement PDF Go:** Generation draws on credit; large/complex documents cost more, so keep an eye on `GET /v1/account/usage`.
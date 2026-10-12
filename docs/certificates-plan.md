# Certificates on course-cycle close approval

## Context
When an admin approves a course-cycle close request, the course owner must receive the participants' certificates. The number of participants has no upper bound. Rendering, uploading and emailing many PDFs inside the HTTP request is slow and fragile. It also can't be tied to the approval transaction: today, a failed email rolls the approval back (`notify` callback in `postgres_course_cycle_close_requests_repository.go:120`).

Agreed design:
- A **hand-rolled Postgres job queue**, enqueued in the approval transaction.
- An **in-process worker** in the same binary.
- **Gotenberg** as a new compose service for HTML→PDF rendering.
- Each certificate gets a random **verification code**. Its QR code points to a **public backend-rendered HTML page** (`/certificados/{code}`) that verifies the certificate and offers the PDF download.
- The owner receives **one unique link per participant** (option C) plus a ZIP of all PDFs.

The participants file is **our own XLSX template**, and it lists **only the participants who approved the course**, not the whole enrollment. Columns, in this order: `Nombres`, `Apellidos`, `Cédula`, `Correo` (optional). Course data (name, duration, cycle dates, faculty) already exists. Signatures and instructor name are out of scope for now: the template gets placeholders.

## Data model (new migrations via `make migrate-create`)
1. `000027_create_jobs_table`: `deu.jobs(id bigserial, kind text, payload jsonb, status text CHECK IN ('pending','running','done','failed'), attempts int, max_attempts int default 5, run_at timestamptz, locked_until timestamptz, last_error text, created_at, updated_at)`, plus an index on `(status, run_at)`.
2. `000028_create_certificates_table`: `deu.certificates(id, close_request_id FK, course_cycle_id FK, first_name, last_name, document, email NULL, verification_code text UNIQUE, storage_key NULL, issued_at NULL, revoked_at NULL, created_at, updated_at)`, plus `UNIQUE(close_request_id, document)`.
3. `000029_add_certificates_token_to_close_requests`: `certificates_token text UNIQUE NULL`, used for the ZIP download link.

For each table, follow the CLAUDE.md "database is source of truth" rule: models, queries, scan/mappers, entities.

## Flow

### 0. Template download
- `GET /course-cycle-close-requests/participants-template` (authenticated) returns `participantes_aprobados.xlsx`.
- Generated in code with excelize (no binary committed): header row styled bold and frozen, column widths set, `Cédula` column formatted as text so Excel doesn't strip the `V-` prefix or turn it into a number, and a short note above or in a second sheet ("Incluya solo participantes aprobados").
- The generator and the parser share the header constants, so they can't drift.
- Frontend links to it next to the participants file input.

### 1. Submit (`course_cycle_close_requests.SubmitCloseRequest`, `service.go:62`)
- Parse the participants file **before** creating the request.
  - New `participants.go` in the module, using `github.com/xuri/excelize/v2` (new dependency). XLSX only; no CSV.
  - Since we own the format, parsing is strict: first sheet, row 1 must be exactly `Nombres | Apellidos | Cédula | Correo` (compared trimmed, case- and accent-insensitively), data from row 2. Fully empty rows are skipped.
  - Cédula is normalized (strip dots and spaces, uppercase the `V-`/`E-` prefix; a bare number defaults to `V-`).
  - `Correo` may be blank. If present, it's validated as an email. It's stored for a later "send certificates to participants" feature; it isn't used in this iteration.
  - Return 400 errors (`errors.go`) for: non-`.xlsx` file, wrong or missing headers, no data rows, blank name, last name or cédula, invalid email, duplicate cédula, or more than 2000 rows. Row-level errors include the row number (e.g. `fila 7: cédula vacía`) so the owner can fix the file.
- The file body is an `io.Reader` from `utils.GetFileFrom`. Buffer it (`io.ReadAll` → `bytes.Reader`) so it can be both parsed and uploaded.
- In the create step, insert one `certificates` row per participant. Each gets a `verification_code` from `crypto/rand` (16 bytes, base32 without padding). `issued_at` stays NULL, so the certificate isn't valid yet.

### 2. Approve (`ApproveCloseRequest`, `service.go:107` + repository `:120`)
- In the same transaction:
  - Set `certificates_token` (random).
  - `INSERT INTO deu.jobs(kind='course_cycle_certificates', payload={"close_request_id":…})`.
- The existing approval email/`notify` stays unchanged.

### 3. Worker (new package `internal/jobs`)
- `Worker{repo JobsRepository, handlers map[string]Handler, pollInterval 3s, lease 30m, concurrency 1}` and `Run(ctx)`. The lease is 30m, not 10m: 2000 certificates at ~1s each with 4 in parallel is close to 10 minutes.
- **Claim:** `UPDATE … SET status='running', attempts=attempts+1, locked_until=now()+lease WHERE id=(SELECT id … WHERE attempts<max_attempts AND ((status='pending' AND run_at<=now()) OR (status='running' AND locked_until<now())) ORDER BY run_at LIMIT 1 FOR UPDATE SKIP LOCKED) RETURNING …`
- **Shutdown:** a job interrupted by shutdown stays `running`; it's claimed again once its lease expires.
- **Success:** `status='done'`.
- **Error:** if `attempts>=max_attempts`, set `status='failed'`. Otherwise set `status='pending'` and `run_at=now()+2^attempts*30s`. Store `last_error` either way.
- `JobsRepository` is implemented by `PostgresRepository` (`postgres_jobs_repository.go`, `queries/jobs_queries.go`).

### 4. Certificate job (new module `internal/certificates`)
- `Service.GenerateForCloseRequest(ctx, payload)` is registered as the handler for `course_cycle_certificates`. It must be idempotent, because retries are possible:
  1. Load the close request, cycle and course, plus the certificates. The owner email comes from the submitter, as in `closeRequestRecipient`. Approval closes the cycle (`is_active=false`), and an issued certificate must keep verifying if its course is later deactivated. So the cycle and course are loaded with new `GetCoursePeriodByIDIncludingInactive` / `GetCourseIncludingInactive` lookups instead of the active-only ones.
  2. For each certificate with `storage_key IS NULL`, using `errgroup.SetLimit(4)`:
     - Render the HTML template (embedded `templates/certificate.html`, with the QR code as a base64 PNG from `github.com/skip2/go-qrcode` pointing to `{PublicBaseURL}/certificados/{code}`).
     - Send it to Gotenberg.
     - Upload the PDF to `certificates/{cycleID}/{verification_code}.pdf` with a new `B2Client.PutObject`. `UploadFile` caps the whole upload at 30s, which is too short for a large ZIP.
     - Set `storage_key` and `issued_at`.
  3. Build the ZIP (`archive/zip`, fetching PDFs through `GetObject`, written to a temp file so memory stays bounded) and upload it to `certificates/{cycleID}/{certificates_token}/certificados.zip` (the same path on every retry, so it's overwritten).
  4. Email the owner using a new template `TemplateCourseCycleCertificatesReady`:
     - An HTML table of name → personal link.
     - A ZIP link `{PublicBaseURL}/certificados/lotes/{token}/zip`.
     - A `participantes.xlsx` attachment (nombres, apellidos, cédula, link), built with excelize like the template. This needs a new `SendTemplateWithAttachments` in `email/smtp.go` (go-mail `AttachReader`).
  - Email delivery is at-least-once: a crash after sending but before marking the job done means a resend. That's acceptable.
- **Template:** the certificate is printed on the design of the faculty that approved the course request: `courses.faculty`, which a redirect updates. It is not chosen by `courses.origin_faculty`, which receives the close requests. The designs live in `certificates/templates/assets/` (`layouts` in `templates.go`; DEU is the fallback). The original PDFs are kept as design references in `docs/certificados/`.
- Dependencies are interfaces declared in the service (`Repository`, `StorageClient`, `MailClient`, `Renderer`). Add them to `.mockery.yaml`, then run `make generate-mocks`.
- New `internal/gotenberg` client implementing `Renderer.Render(ctx, html []byte) ([]byte, error)`. It POSTs multipart `index.html` to `{GOTENBERG_URL}/forms/chromium/convert/html`, with a timeout.

### 5. Public endpoints (certificates handler, depends only on the service)
- `GET /certificados/:code`: HTML verification page.
  - Valid certificate: name, masked cédula (`V-12.3**.**9`), course, duration, dates, faculty, issue date, and a download button.
  - Not issued or revoked: a "no válido" page, not a 404 leak distinction.
- `GET /certificados/:code/pdf`: streams from storage with `Content-Disposition: attachment`.
- `GET /certificados/lotes/:token/zip`: streams the ZIP.
- `POST /admin/course-cycle-close-requests/:id/certificates/retry`: resets a `failed` job to `pending`, `attempts=0`.
- Traefik (prod): add a router for `Host(extension.ucv.ve) && PathPrefix(/certificados)` → backend, mirroring the `backend-files` labels in `docker-compose.prod.yml`.

### 6. Wiring & infrastructure
- `config.go`:
  - `GOTENBERG_URL` (default `http://gotenberg:3000`).
  - `PUBLIC_BASE_URL` (default `http://localhost:8080`). This is where QR codes and links point, which is not necessarily `APP_BASE_URL`. Add a comment per the no-renaming rule.
  - Add both to `.env.example`.
- `docker-compose.yml` / `docker-compose.prod.yml`: add a `gotenberg: image gotenberg/gotenberg:8` service on the `web` network with no published port. The backend gets `depends_on`.
- `main.go`:
  - Build the gotenberg client, the certificates service/handler and the jobs worker.
  - Start the worker in a goroutine with a `signal.NotifyContext` context.
  - Replace `e.Logger.Fatal(e.Start…)` with start plus `e.Shutdown` on signal, so the worker and server stop cleanly.

## Known issues (out of scope)
- ~~`GET /files/*` serves any bucket key~~: already fixed on this branch. `files.GetPublicFile` only serves files flagged public.
- `postgresql/migrations` has duplicate numbers `000023_*` and `000024_*` (pre-existing), so golang-migrate refuses to load the directory. See the `fix/migrations-collisions` branch.
- `make generate-mocks` fails: the `vektra/mockery:v3.7.0` image ships Go 1.26.1, but `go.mod` requires 1.26.6. Workaround: `cd backend && go run github.com/vektra/mockery/v3@v3.7.0 --config .mockery.yaml`.

## Verification
- Unit tests:
  - Participants parser: XLSX fixtures built in-test with excelize, each 400 case (with row numbers), cédula normalization, blank vs. invalid email.
  - Template generator: parsing the generated template with one row filled in round-trips.
  - Jobs worker: claim/complete/fail/backoff with a mocked repo.
  - Certificates service: renders only unissued certificates, idempotent re-run, email content and attachment, with mocked renderer/storage/mail.
  - Handler: valid, revoked and unknown codes.
- `make generate-mocks`, then `make quality`.
- End-to-end (`make start-backend`, `make migrate-up`):
  1. Download the template, fill it with fictional names, and submit a close request with it.
  2. Approve it as admin.
  3. Watch the logs for the job going to `done`, and `SELECT status, attempts, last_error FROM deu.jobs`.
  4. Check that the email arrives with the XLSX attachment.
  5. Open a personal link, scan the QR code from the PDF, and download the ZIP.
- Failure path: stop the gotenberg container, approve a request, and confirm the retries and backoff in `deu.jobs`. Then restart gotenberg and confirm the job completes.

# Webhooks

## Overview

External providers push lifecycle events to the API. Signatures are verified in middleware before handlers run.

| Provider | Path | Middleware |
|----------|------|------------|
| Clerk | `POST /webhooks/clerk` | Svix (`UserWebhookMiddleware`) |
| MinIO | `POST /webhooks/minio/object-created` | Bearer/token (`StorageWebhookMiddleware`) |

No Clerk JWT on webhook routes.

## Clerk (user lifecycle)

Handler: `UserWebhookHandler.Execute`

Verified payload is set on `c.Locals("payload", dto.ClerkEvent)` before the handler runs.

| Event | Action |
|-------|--------|
| `user.created` | Create user if not exists (`201`) |
| `user.updated` | Update profile (`204`) |
| `user.deleted` | Delete by Clerk ID (`204`) |
| Unknown | `200` (ignored) |

### Idempotence

`user.created` skips creation if the Clerk ID already exists.

### Errors

| Case | HTTP |
|------|------|
| Invalid signature / payload (middleware) | `401` + `{ "message": "…" }` |
| Invalid JSON body | `400` |
| Validation failed | `400` + field `errors` |
| User not found on update | `404` |
| Handler failure | `500` (generic `message`, no internal leak) |

## MinIO (object created)

Handler: `StorageWebhookHandler.ObjectCreated`

Auth: `Authorization: Bearer {MINIO_WEBHOOK_SECRET}` or raw secret. Verified payload is set on `c.Locals("payload", dto.ObjectCreatedEvent)`.

Compose wires `notify_webhook` on bucket `media` for `put` events with prefix `videos/` and suffix `original.mp4`.

| Case | Action |
|------|--------|
| `s3:ObjectCreated*` on media original key | `ConfirmUpload` (`pending` → `uploaded`) |
| Thumbnail keys / wrong bucket / invalid key | Ignored (`200`) |
| Not found / invalid transition / too large | Soft-skipped (`200`) |

### Errors

| Case | HTTP |
|------|------|
| Missing/invalid secret (middleware) | `401` |
| Invalid JSON body (middleware) | `400` |
| Validation failed | `400` + field `errors` |
| Unexpected confirm failure | `500` (generic `message`) |

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `user_webhook_handler.go`, `storage_webhook_handler.go` |
| Middleware | `user_webhook_middleware.go`, `storage_webhook_middleware.go` |
| Commands | `internal/application/command/user/`, `internal/application/command/mediafile/` |

## Tests

- `internal/interfaces/http/handler/test/user_webhook/`
- `internal/interfaces/http/handler/test/storage_webhook/`

## Security

- Never trust raw webhook bodies — middleware validates signatures/tokens first.
- Handlers validate unmarshalled DTOs via `validation.Struct`.

# L6 — Cloud-Sync

## Zweck
Workspace mit Cloud-Storage (Dropbox/iCloud/S3) sync.

## MVP-Status
- Out-of-MVP (R10)
- Nicht im MVP-Scope

## Architektur-Stub
```go
// internal/obsidian/l6_cloudsync.go (R10) — PLACEHOLDER
package obsidian

type SyncBackend interface {
	Pull(path string) error
	Push(path string) error
}
```

## Nächste Schritte (post-MVP)
- rclone-Integration
- S3-Backend
- Sync-Status-Indicator

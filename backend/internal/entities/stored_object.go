package entities

import "time"

// StoredObject describes an object kept in the file storage (Backblaze B2).
type StoredObject struct {
	Key          string
	Size         int64
	LastModified time.Time
}

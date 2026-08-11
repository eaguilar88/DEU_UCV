package entities

type ContactType string

const (
	ContactTypeEmail ContactType = "email"
	ContactTypePhone ContactType = "phone"
)

type Contact struct {
	ID        string
	Type      ContactType
	Value     string
	OwnerID   string
	OwnerType OwnerType
}

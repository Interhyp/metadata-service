package types

const (
	OwnerPayload NotificationPayloadType = iota
	ServicePayload
	RepositoryPayload
)

func (p NotificationPayloadType) String() string {
	switch p {
	case OwnerPayload:
		return "Owner"
	case ServicePayload:
		return "Service"
	case RepositoryPayload:
		return "Repository"
	default:
		return ""
	}
}

const (
	CreatedEvent NotificationEventType = iota
	ModifiedEvent
	DeletedEvent
)

func (p NotificationEventType) String() string {
	switch p {
	case CreatedEvent:
		return "CREATED"
	case ModifiedEvent:
		return "MODIFIED"
	case DeletedEvent:
		return "DELETED"
	default:
		return ""
	}
}

type NotificationPayloadType uint32

type NotificationEventType uint32

package event

// domain event for the AcitiveClient is activated.
type ActiveClientActivated struct {
	EventEmbd
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
}

func (ActiveClientActivated) Type() Type { return TypeActiveClientActivated }

// domain event for the AcitiveClient is inactivated.
type ActiveClientInactivated struct {
	EventEmbd
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
}

func (ActiveClientInactivated) Type() Type { return TypeActiveClientInactivated }

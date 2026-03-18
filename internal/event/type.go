package event

type UserRegisteredEvent struct {
	EventType string `json:"event_type"`
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

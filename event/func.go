package event

func SendEvent(eventName string, data ...any) {
	svc := NewEventService()
	svc.SendEvent(eventName, data...)
}

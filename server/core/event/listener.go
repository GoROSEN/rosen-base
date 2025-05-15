package event

type EventListener interface {
	HandleEvent(event string, data interface{})
}

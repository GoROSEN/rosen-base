package event

// EventPublisher 事件全局发布者
type EventPublisher struct {
	Listeners map[EventListener]bool
}

var g_eventPublisher EventPublisher = EventPublisher{Listeners: make(map[EventListener]bool)}

func GetPublisher() *EventPublisher {
	return &g_eventPublisher
}

func (p *EventPublisher) Publish(event string, data interface{}) {

	for l, b := range p.Listeners {
		if b {
			l.HandleEvent(event, data)
		}
	}
}

func (p *EventPublisher) PublishAsync(event string, data interface{}) {

	for l, b := range p.Listeners {
		if b {
			go l.HandleEvent(event, data)
		}
	}
}

func (p *EventPublisher) AddListener(listner EventListener) {
	p.Listeners[listner] = true
}

func (p *EventPublisher) RemoveListener(listner EventListener) {
	p.Listeners[listner] = false
	delete(p.Listeners, listner)
}

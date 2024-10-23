package event

import (
	"sync"
	"time"
)

type Event struct {
	listeners []chan interface{}
	mutex     sync.RWMutex
}

func NewEvent() *Event {
	return &Event{
		listeners: make([]chan interface{}, 0),
	}
}

func (e *Event) Subscribe() chan interface{} {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	ch := make(chan interface{})
	e.listeners = append(e.listeners, ch)
	return ch
}

func (e *Event) Unsubscribe(ch chan interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for i, listener := range e.listeners {
		if listener == ch {
			e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
			close(ch)
			break
		}
	}
}

// Publish sends data to all listeners
func (e *Event) Publish(data interface{}) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	for _, listener := range e.listeners {
		listener <- data
	}
}

// event that is actuated ever x seconds specified when creating the event
type TimedEvent struct {
	event  *Event
	ticker *time.Ticker
}

func NewTimedEvent(duration time.Duration) *TimedEvent {
	ticker := time.NewTicker(duration)
	event := NewEvent()

	go func() {
		for range ticker.C {
			event.Publish(nil)
		}
	}()

	return &TimedEvent{
		event:  event,
		ticker: ticker,
	}
}

func (t *TimedEvent) Subscribe() chan interface{} {
	return t.event.Subscribe()
}

func (t *TimedEvent) Unsubscribe(ch chan interface{}) {
	t.event.Unsubscribe(ch)
}

func (t *TimedEvent) Stop() {
	t.ticker.Stop()
}

// Publish sends data to all listeners of the timed event
func (t *TimedEvent) Publish(data interface{}) {
	t.event.Publish(data)
}

/*
usage:
te := event.NewTimedEvent(5 * time.Second)

go func() {
	for {
		<-te.Subscribe() // Wait for the event to trigger
		// Call your function here
		// Replace `yourFunction` with the actual function you want to run
		yourFunction()
	}
}()

*/

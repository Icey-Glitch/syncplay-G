package event

import (
	"fmt"
	"reflect"
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

type ManagedEvent struct {
	Event
	Ticker   *Ticker
	Function reflect.Value
	Params   []reflect.Value
	stopChan chan struct{}
	owner    *EventManager
}

type Ticker struct {
	interval   int // seconds
	Repeat     bool
	StopTicker bool
}

func (em *EventManager) NewManagedEvent(interval int, fn interface{}, repeat bool, params []interface{}, ticker ...*Ticker) *ManagedEvent {
	fnValue := reflect.ValueOf(fn)
	paramsValue := make([]reflect.Value, len(params))
	for i, param := range params {
		paramsValue[i] = reflect.ValueOf(param)
	}

	ne := &ManagedEvent{
		Event:    Event{},
		Ticker:   &Ticker{interval: interval, Repeat: repeat},
		Function: fnValue,
		Params:   paramsValue,
		stopChan: make(chan struct{}),
		owner:    em,
	}

	if len(ticker) > 0 {
		ne.Ticker = ticker[0]
	}

	em.AddEvent(ne)
	return ne
}

func (e *ManagedEvent) Start() {
	go func() {
		ticker := time.NewTicker(time.Duration(e.Ticker.interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Check if any of the parameters are nil
				for _, param := range e.Params {
					if param.Kind() == reflect.Ptr || param.Kind() == reflect.Interface || param.Kind() == reflect.Map || param.Kind() == reflect.Slice || param.Kind() == reflect.Chan {
						if param.IsNil() {
							fmt.Println("Error: One of the parameters is nil")
							e.Stop()
							return
						}
					}
				}
				// check return value of function
				ret := e.Function.Call(e.Params)
				if !e.Ticker.Repeat || ret[0].Bool() {
					e.Stop()
					return
				}
			case <-e.stopChan:
				return
			}
		}
	}()
}

func (e *ManagedEvent) Stop() {
	select {
	case <-e.stopChan:
		// Channel already closed
	default:
		close(e.stopChan)
		if e.owner != nil {
			e.owner.RemoveEvent(e)
		}
	}
}

// NewTicker creates a new Ticker
func NewTicker(interval int, repeat bool) *Ticker {
	return &Ticker{
		interval:   interval,
		Repeat:     repeat,
		StopTicker: false,
	}
}

// EventManager manages events
type EventManager struct {
	events map[*ManagedEvent]bool
	mutex  sync.RWMutex
}

// NewEventManager creates a new EventManager
func NewEventManager() *EventManager {
	return &EventManager{
		events: make(map[*ManagedEvent]bool),
	}
}

// AddEvent adds an event to the EventManager
func (em *EventManager) AddEvent(event *ManagedEvent) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.events[event] = true
}

// RemoveEvent removes an event from the EventManager
func (em *EventManager) RemoveEvent(event *ManagedEvent) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	delete(em.events, event)
}

// Add logging to StopAll to help debug the issue
func (em *EventManager) StopAll() {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	for event := range em.events {
		fmt.Println("Stopping event:", event)
		select {
		case <-event.stopChan:
			// Channel already closed
		default:
			close(event.stopChan)
			if event.owner != nil {
				delete(em.events, event)
			}
		}
	}
}

// Add this method to the EventManager struct
func (em *EventManager) GetEvents() map[*ManagedEvent]bool {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	return em.events
}

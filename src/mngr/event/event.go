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

/*
acctuated event system:

takes a time.Duration and a function to call after the duration has passed, repeatedly if desired.
returns an ManagedEvent that can be stopped with the Stop(), this will then delete the managed event from the event manager.
start the event with Start() and it will be added to the event manager.

func() should accept any number of arguments and return a bool.

if the function returns true, the event will be stopped.
let the function return false to keep the event running, if the event is set to repeat.
if the function errors, the event will be stopped.

ability to create a timer that will call multiple functions in sequence




usage:
func main() {
	event := NewManagedEvent(time.Second * 5, run(), true)
	event.Start()
}
func run() bool {
	fmt.Println("5 seconds have passed")
	return true
}
*/

type ManagedEvent struct {
	Event
	Ticker    *Ticker
	Function  reflect.Value
	Params    []reflect.Value
	StopEvent bool

	owner *EventManager
}

type Ticker struct {
	interval   int // seconds
	Repeat     bool
	StopTicker bool

	linkedEvents []*ManagedEvent
}

func (em *EventManager) NewManagedEvent(interval int, fn interface{}, repeat bool, params []interface{}, ticker ...*Ticker) *ManagedEvent {
	fnValue := reflect.ValueOf(fn)
	paramsValue := make([]reflect.Value, len(params))
	for i, param := range params {
		paramsValue[i] = reflect.ValueOf(param)
	}

	// if ticker is provided, use it
	if len(ticker) > 0 {
		ne := &ManagedEvent{
			Event:     Event{},
			Ticker:    ticker[0],
			Function:  fnValue,
			Params:    paramsValue,
			StopEvent: false,
			owner:     em,
		}
		em.AddEvent(ne)
		return ne
	}

	ne := &ManagedEvent{
		Event:     Event{},
		Ticker:    &Ticker{interval: interval, Repeat: repeat},
		Function:  fnValue,
		Params:    paramsValue,
		StopEvent: false,
		owner:     em,
	}
	em.AddEvent(ne)

	return ne
}

func (e *ManagedEvent) Start() {
	go func() {
		for {
			select {
			case <-time.After(time.Duration(e.Ticker.interval) * time.Second):
				if e.StopEvent {
					return
				}
				// Check if any of the parameters are nil
				for _, param := range e.Params {
					if param.Kind() == reflect.Ptr || param.Kind() == reflect.Interface || param.Kind() == reflect.Map || param.Kind() == reflect.Slice || param.Kind() == reflect.Chan {
						if param.IsNil() {
							fmt.Println("Error: One of the parameters is nil")
							return
						}
					}
				}
				// check return value of function
				ret := e.Function.Call(e.Params)
				if !e.Ticker.Repeat {
					e.Stop()
					return
				}
				if ret[0].Bool() == true {
					e.Stop()
					return
				}

			}
		}
	}()
}

func (e *ManagedEvent) Stop() {
	e.StopEvent = true
	if e.owner != nil {
		e.owner.RemoveEvent(e)
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

func (t *Ticker) Stop() {
	t.StopTicker = true
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

// StopAll stops all events in the EventManager
func (em *EventManager) StopAll() {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	for event := range em.events {
		event.Stop()
	}
}

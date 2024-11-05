package event

import (
	"reflect"
	"testing"
	"time"
)

func TestNewManagedEvent(t *testing.T) {
	em := NewEventManager()

	// Test case 1: Non-repeating event
	fn := func() bool {
		return true
	}
	var params []interface{}
	event := em.NewManagedEvent(1, fn, false, params)

	if event == nil {
		t.Fatal("Expected non-nil event")
	}
	if event.Ticker == nil {
		t.Fatal("Expected non-nil ticker")
	}
	if event.Ticker.Repeat {
		t.Fatal("Expected non-repeating ticker")
	}
	if event.Function.Kind() != reflect.Func {
		t.Fatal("Expected function kind to be Func")
	}

	// Test case 2: Repeating event
	event = em.NewManagedEvent(1, fn, true, params)

	if event == nil {
		t.Fatal("Expected non-nil event")
	}
	if event.Ticker == nil {
		t.Fatal("Expected non-nil ticker")
	}
	if !event.Ticker.Repeat {
		t.Fatal("Expected repeating ticker")
	}

	// Test case 3: Event with parameters
	fnWithParams := func(msg string) bool {
		return msg == "stop"
	}
	params = []interface{}{"stop"}
	event = em.NewManagedEvent(1, fnWithParams, false, params)

	if event == nil {
		t.Fatal("Expected non-nil event")
	}
	if event.Ticker == nil {
		t.Fatal("Expected non-nil ticker")
	}
	if event.Function.Kind() != reflect.Func {
		t.Fatal("Expected function kind to be Func")
	}
	if len(event.Params) != 1 {
		t.Fatal("Expected one parameter")
	}
	if event.Params[0].String() != "stop" {
		t.Fatalf("Expected parameter to be 'stop', got %v", event.Params[0])
	}

	// Test case 4: Event with provided ticker
	ticker := NewTicker(2, true)
	event = em.NewManagedEvent(1, fn, true, params, ticker)

	if event == nil {
		t.Fatal("Expected non-nil event")
	}
	if event.Ticker != ticker {
		t.Fatal("Expected provided ticker to be used")
	}
	if !event.Ticker.Repeat {
		t.Fatal("Expected repeating ticker")
	}
}

func TestManagedEventStartStop(t *testing.T) {
	em := NewEventManager()

	fn := func() bool {
		return true
	}
	var params []interface{}
	event := em.NewManagedEvent(1, fn, false, params)

	event.Start()
	time.Sleep(2 * time.Second)
	event.Stop()

	if !event.StopEvent {
		t.Fatal("Expected event to be stopped")
	}

	// Test case 2: Repeating event
	event = em.NewManagedEvent(1, fn, true, params)

	event.Start()
	time.Sleep(2 * time.Second)
	event.Stop()

	if !event.StopEvent {
		t.Fatal("Expected event to be stopped")
	}

	// Test case 3: Event with provided ticker
	ticker := NewTicker(2, true)
	event = em.NewManagedEvent(1, fn, true, params, ticker)

	event.Start()
	time.Sleep(2 * time.Second)
	event.Stop()

	if !event.StopEvent {
		t.Fatal("Expected event to be stopped")
	}

	ticker.Stop()
	if !ticker.StopTicker {
		t.Fatal("Expected ticker to be stopped")
	}

}

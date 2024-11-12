package event_test

import (
	"testing"
	"time"

	"github.com/Icey-Glitch/Syncplay-G/mngr/event"
)

func TestEvent_SubscribeAndPublish(t *testing.T) {
	e := event.NewEvent()
	ch := e.Subscribe()

	go e.Publish("test message")

	select {
	case msg := <-ch:
		if msg != "test message" {
			t.Errorf("Expected 'test message', got %v", msg)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestEvent_Unsubscribe(t *testing.T) {
	e := event.NewEvent()
	ch := e.Subscribe()
	e.Unsubscribe(ch)

	go e.Publish("test message")

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Received message after unsubscribe")
		}
	case <-time.After(time.Second):
		// Expected behavior
	}
}

func TestManagedEvent_StartAndStop(t *testing.T) {
	em := event.NewEventManager()
	me := em.NewManagedEvent(1, func() bool { return true }, false, nil)

	me.Start()
	time.Sleep(2 * time.Second)
	me.Stop()

	if _, ok := em.GetEvents()[me]; ok {
		t.Error("ManagedEvent was not removed from EventManager after stop")
	}
}

func TestEventManager_StopAll(t *testing.T) {
	em := event.NewEventManager()
	me1 := em.NewManagedEvent(1, func() bool { return false }, true, nil)
	me2 := em.NewManagedEvent(1, func() bool { return false }, true, nil)

	me1.Start()
	me2.Start()
	time.Sleep(2 * time.Second)

	em.StopAll()

	if len(em.GetEvents()) != 0 {
		t.Error("Not all events were stopped")
	}
}

func TestManagedEvent_Repeat(t *testing.T) {
	em := event.NewEventManager()
	counter := 0
	me := em.NewManagedEvent(1, func() bool {
		counter++
		return false
	}, true, nil)

	me.Start()
	time.Sleep(3 * time.Second)
	me.Stop()

	if counter < 2 {
		t.Errorf("Expected counter to be at least 2, got %d", counter)
	}
}

func TestManagedEvent_Params(t *testing.T) {
	em := event.NewEventManager()
	me := em.NewManagedEvent(1, func(a int, b string) bool {
		if a != 42 || b != "test" {
			t.Errorf("Expected params (42, 'test'), got (%d, '%s')", a, b)
		}
		return true
	}, false, []interface{}{42, "test"})

	me.Start()
	time.Sleep(2 * time.Second)
	me.Stop()
}

// NewTickerTest is a test function for the Ticker struct
func TestTicker_NewTicker(t *testing.T) {
	ticker := event.NewTicker(1, true)
	if ticker.Interval != 1 || !ticker.Repeat {
		t.Errorf("Expected interval 1 and Repeat true, got interval %d and Repeat %t", ticker.Interval, ticker.Repeat)
	}
}

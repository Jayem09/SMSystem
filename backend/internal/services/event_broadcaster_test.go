package services

import (
	"testing"
	"time"
)

func TestEventBroadcasterBranchDeliveryAndUnregister(t *testing.T) {
	broadcaster := newEventBroadcaster()

	branchClient := broadcaster.Register(7)
	globalClient := broadcaster.Register(0)
	otherBranchClient := broadcaster.Register(9)

	if !waitForClientCount(t, broadcaster, 3) {
		count := broadcaster.ClientCount()
		t.Fatalf("expected 3 clients after registration, got %d", count)
	}

	broadcaster.BroadcastToBranch(EventOrderCreated, 7, map[string]any{"order_id": 123})

	select {
	case event := <-branchClient.Channel:
		if event.Type != EventOrderCreated {
			t.Fatalf("expected event type %q, got %q", EventOrderCreated, event.Type)
		}
		if event.BranchID != 7 {
			t.Fatalf("expected branch id 7, got %d", event.BranchID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected branch client to receive broadcast")
	}

	select {
	case <-globalClient.Channel:
	case <-time.After(2 * time.Second):
		t.Fatal("expected global client to receive branch broadcast")
	}

	select {
	case event := <-otherBranchClient.Channel:
		t.Fatalf("did not expect other branch client to receive event: %+v", event)
	case <-time.After(200 * time.Millisecond):
	}

	broadcaster.Unregister(branchClient)

	if !waitForClientCount(t, broadcaster, 2) {
		count := broadcaster.ClientCount()
		t.Fatalf("expected 2 clients after unregister, got %d", count)
	}

	select {
	case _, ok := <-branchClient.Channel:
		if ok {
			t.Fatal("expected unregistered client channel to be closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected unregistered client channel to close")
	}
}

func waitForClientCount(t *testing.T, broadcaster *EventBroadcaster, expected int) bool {
	t.Helper()

	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return false
		case <-ticker.C:
			if broadcaster.ClientCount() == expected {
				return true
			}
		}
	}
}

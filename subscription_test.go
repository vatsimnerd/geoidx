package geoidx

import (
	"testing"
)

func expectNoEvents(t *testing.T, ch <-chan Event) bool {
	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return false
	default:
		return true
	}
}

func expectEvent(t *testing.T, ch <-chan Event, et EventType, v interface{}) bool {
	select {
	case e := <-ch:
		if e.Type != et {
			t.Errorf("unexpected event type, expected %v got %v", et, e.Type)
			return false
		}
		if e.Obj.Value() != v {
			t.Errorf("unexpected object, expected '%s', got %v", v, e.Obj.Value())
			return false
		}
		return true
	default:
		t.Error("expected event on channel")
		return false
	}
}

func TestSubscription(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	sub.SetBounds(MakeRect(0, 0, 2, 2))
	expectNoEvents(t, ch)

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	expectEvent(t, ch, EventTypeSet, "test")
	expectNoEvents(t, ch)

	i.Delete(obj)
	expectEvent(t, ch, EventTypeDelete, "test")
	expectNoEvents(t, ch)

}

func TestSubscriptionMoveBounds(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	sub.SetBounds(MakeRect(0, 0, 2, 2))
	expectEvent(t, ch, EventTypeSet, "test")
	expectNoEvents(t, ch)

	sub.SetBounds(MakeRect(3, 3, 5, 5))
	expectEvent(t, ch, EventTypeDelete, "test")
	expectNoEvents(t, ch)
}

// TestSubFilters1 tests filters aren't passing new objects
func TestSubFilters1(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	// should not pass any object
	sub.SetFilters(
		func(obj *Object) bool {
			return false
		},
	)
	sub.SetBounds(MakeRect(0, 0, 2, 2))
	expectNoEvents(t, ch)

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)
	expectNoEvents(t, ch)

	i.Delete(obj)
	expectNoEvents(t, ch)
}

// TestSubFilters2 tests filtered out objects are deleted after applying filter
func TestSubFilters2(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	sub.SetBounds(MakeRect(0, 0, 2, 2))
	expectNoEvents(t, ch)

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)
	expectEvent(t, ch, EventTypeSet, "test")

	// should not pass any object
	sub.SetFilters(
		func(obj *Object) bool {
			return false
		},
	)

	expectEvent(t, ch, EventTypeDelete, "test")
	sub.SetFilters()

	expectEvent(t, ch, EventTypeSet, "test")

	i.Delete(obj)
	expectEvent(t, ch, EventTypeDelete, "test")
	expectNoEvents(t, ch)
}

func TestMultipleUpsertSameObject(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	sub.SetBounds(MakeRect(0, 0, 2, 2))
	expectNoEvents(t, ch)

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	expectEvent(t, ch, EventTypeSet, "test")
	expectNoEvents(t, ch)

	i.Upsert(obj)
	expectEvent(t, ch, EventTypeSet, "test")
	expectNoEvents(t, ch)
}

func TestTrackID(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	sub.SetBounds(MakeRect(0, 0, 2, 2))
	expectNoEvents(t, ch)

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)
	expectEvent(t, ch, EventTypeSet, "test")
	expectNoEvents(t, ch)

	sub.TrackID("1")
	expectNoEvents(t, ch) // we already have the object in sight

	obj.bounds = MakeRect(5, 5, 7, 7) // move outside the bounds
	i.Upsert(obj)
	expectEvent(t, ch, EventTypeSet, "test")
	expectNoEvents(t, ch)

	sub.UntrackID("1")
	expectEvent(t, ch, EventTypeDelete, "test")
	// now the object is out of bounds and not tracked
	expectNoEvents(t, ch)
}

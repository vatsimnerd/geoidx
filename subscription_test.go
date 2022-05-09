package geoidx

import (
	"testing"

	"github.com/dhconnelly/rtreego"
)

func TestSubscription(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	sub.SetBounds(MakeRect(0, 0, 2, 2))

	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	select {
	case e := <-ch:
		if e.Type != EventTypeSet {
			t.Errorf("unexpected event type, expected Set got %v", e.Type)
		}
		if obj.Value() != "test" {
			t.Errorf("unexpected object, expected 'test', got %v", obj.Value())
		}
		return
	default:
		t.Error("expected event on channel")
	}

	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}

	i.Delete(obj)

	select {
	case e := <-ch:
		if e.Type != EventTypeDelete {
			t.Errorf("unexpected event type, expected Delete got %v", e.Type)
		}
		if obj.Value() != "test" {
			t.Errorf("unexpected object, expected 'test', got %v", obj.Value())
		}
		return
	default:
		t.Error("expected event on channel")
	}
}

func TestSubscriptionMoveBounds(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	sub.SetBounds(MakeRect(0, 0, 2, 2))

	select {
	case e := <-ch:
		if e.Type != EventTypeSet {
			t.Errorf("unexpected event type, expected Set got %v", e.Type)
		}
		if obj.Value() != "test" {
			t.Errorf("unexpected object, expected 'test', got %v", obj.Value())
		}
		return
	default:
		t.Error("expected event on channel")
	}

	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}

	sub.SetBounds(MakeRect(3, 3, 5, 5))

	select {
	case e := <-ch:
		if e.Type != EventTypeDelete {
			t.Errorf("unexpected event type, expected Delete got %v", e.Type)
		}
		if obj.Value() != "test" {
			t.Errorf("unexpected object, expected 'test', got %v", obj.Value())
		}
		return
	default:
		t.Error("expected event on channel")
	}

}

// TestSubFilters1 tests filters aren't passing new objects
func TestSubFilters1(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	// should not pass any object
	sub.SetFilters(
		func(results []rtreego.Spatial, object rtreego.Spatial) (refuse bool, abort bool) {
			refuse = true
			return
		},
	)
	sub.SetBounds(MakeRect(0, 0, 2, 2))

	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")

	i.Upsert(obj)
	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}

	i.Delete(obj)
	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}

}

// TestSubFilters2 tests filtered out objects are deleted after applying filter
func TestSubFilters2(t *testing.T) {
	i := NewIndex()
	sub := i.Subscribe(1024)
	ch := sub.Events()

	sub.SetBounds(MakeRect(0, 0, 2, 2))

	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}

	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	select {
	case e := <-ch:
		if e.Type != EventTypeSet {
			t.Errorf("unexpected event type, expected Set got %v", e.Type)
		}
		if obj.Value() != "test" {
			t.Errorf("unexpected object, expected 'test', got %v", obj.Value())
		}
		return
	default:
		t.Error("expected event on channel")
	}

	// should not pass any object
	sub.SetFilters(
		func(results []rtreego.Spatial, object rtreego.Spatial) (refuse bool, abort bool) {
			refuse = true
			return
		},
	)

	select {
	case e := <-ch:
		if e.Type != EventTypeDelete {
			t.Errorf("unexpected event type, expected Delete got %v", e.Type)
		}
		if obj.Value() != "test" {
			t.Errorf("unexpected object, expected 'test', got %v", obj.Value())
		}
		return
	default:
		t.Error("expected event on channel")
	}
	// clear filter
	sub.SetFilters()

	select {
	case e := <-ch:
		if e.Type != EventTypeSet {
			t.Errorf("unexpected event type, expected Set got %v", e.Type)
		}
		if obj.Value() != "test" {
			t.Errorf("unexpected object, expected 'test', got %v", obj.Value())
		}
		return
	default:
		t.Error("expected event on channel")
	}

	i.Delete(obj)
	select {
	case e := <-ch:
		t.Errorf("unexpected event %v on channel", e)
		return
	default:
	}
}

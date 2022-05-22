package geoidx

import "testing"

func TestUpsert(t *testing.T) {
	i := NewIndex()
	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	objs := i.SearchByRect(MakeRect(0, 0, 2, 2))
	if len(objs) != 1 {
		t.Errorf("search is expected to return 1 object, got %d", len(objs))
	}

	obj = objs[0]
	if obj.Value() != "test" {
		t.Errorf("search is expected to return 'test', got %v", obj.Value())
	}

	objs = i.SearchByRect(MakeRect(0.5, 0.5, 0.5, 0.5))
	if len(objs) != 0 {
		t.Errorf("search is expected to return 0 objects, got %d", len(objs))
	}
}

func TestSearchByObjectID(t *testing.T) {
	i := NewIndex()
	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)
	obj2 := NewObject("2", MakeRect(0, 0, 2, 2), "searchbox")
	i.Upsert(obj2)

	objs := i.SearchByObjectID(obj2.id)
	if len(objs) != 1 {
		t.Errorf("search is expected to return 1 object, got %d", len(objs))
	}

	obj = objs[0]
	if obj.Value() != "test" {
		t.Errorf("search is expected to return 'test', got %v", obj.Value())
	}
}

func TestDelete(t *testing.T) {
	i := NewIndex()
	obj := NewObject("1", MakeRect(1, 1, 2, 2), "test")
	i.Upsert(obj)

	objs := i.SearchByRect(MakeRect(0, 0, 2, 2))
	if len(objs) != 1 {
		t.Errorf("search is expected to return 1 object, got %d", len(objs))
	}

	i.Delete(obj)
	objs = i.SearchByRect(MakeRect(0, 0, 2, 2))
	if len(objs) != 0 {
		t.Errorf("search is expected to return 0 object, got %d", len(objs))
	}
}

func TestPartialIntersect(t *testing.T) {
	i := NewIndex()
	obj := NewObject("1", MakeRect(-1, -1, 1, 1), "test")
	i.Upsert(obj)

	objs := i.SearchByRect(MakeRect(0, 0, 2, 2))
	if len(objs) != 1 {
		t.Errorf("search is expected to return 1 object, got %d", len(objs))
	}

	i.Delete(obj)
	objs = i.SearchByRect(MakeRect(0, 0, 2, 2))
	if len(objs) != 0 {
		t.Errorf("search is expected to return 0 object, got %d", len(objs))
	}
}

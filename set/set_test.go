package set

import (
	"reflect"
	"sort"
	"testing"
)

func TestNew(t *testing.T) {
	s := New[int]()
	if s.Size() != 0 {
		t.Error("new set should have size of 0")
	}
}

func TestFromList(t *testing.T) {
	s := FromList([]int{1, 3, 5, 7, 1, 3, 5})
	if s.Size() != 4 {
		t.Errorf("invalid set size, expected 4, got %d", s.Size())
	}

	l := s.List()
	sort.Slice(l, func(i, j int) bool {
		return l[i] < l[j]
	})

	if !reflect.DeepEqual(l, []int{1, 3, 5, 7}) {
		t.Errorf("invalid list from set, expected [1,3,5,7], got %v", l)
	}
}

func TestAddRemove(t *testing.T) {
	s := New[int]()
	s.Add(1)
	if s.Size() != 1 {
		t.Errorf("invalid set size, expected 1, got %d", s.Size())
	}
	s.Add(1)
	if s.Size() != 1 {
		t.Errorf("invalid set size, expected 1, got %d", s.Size())
	}

	if !s.Has(1) {
		t.Error("set is expected to have 1")
	}

	s.Delete(1)
	if s.Size() != 0 {
		t.Errorf("invalid set size, expected 0, got %d", s.Size())
	}

	if s.Has(1) {
		t.Error("set is expected to not have 1")
	}
}

func TestUnion(t *testing.T) {
	s := FromList([]int{1, 3, 5, 7})
	u := FromList([]int{5, 7, 9, 12})
	r := s.Union(u)

	if s.Size() != 4 {
		t.Error("s should not have changed")
	}

	if u.Size() != 4 {
		t.Error("u should not have changed")
	}

	if r.Size() != 6 {
		t.Error("r should have size of 6")
	}
}

func TestSubtract(t *testing.T) {
	s := FromList([]int{1, 3, 5, 7})
	u := FromList([]int{5, 7, 9, 12})
	r := s.Subtract(u)

	if s.Size() != 4 {
		t.Error("s should not have changed")
	}

	if u.Size() != 4 {
		t.Error("u should not have changed")
	}

	if r.Size() != 2 {
		t.Error("r should have size of 2")
	}

	l := r.List()
	sort.Slice(l, func(i, j int) bool {
		return l[i] < l[j]
	})

	if !reflect.DeepEqual(l, []int{1, 3}) {
		t.Errorf("r should be [1,3], got %v", l)
	}

}

func TestIntersection(t *testing.T) {
	s := FromList([]int{1, 3, 5, 7})
	u := FromList([]int{5, 7, 9, 12})
	r := s.Intersection(u)

	if s.Size() != 4 {
		t.Error("s should not have changed")
	}

	if u.Size() != 4 {
		t.Error("u should not have changed")
	}

	if r.Size() != 2 {
		t.Error("r should have size of 2")
	}

	l := r.List()
	sort.Slice(l, func(i, j int) bool {
		return l[i] < l[j]
	})

	if !reflect.DeepEqual(l, []int{5, 7}) {
		t.Errorf("r should be [5,7], got %v", l)
	}
}

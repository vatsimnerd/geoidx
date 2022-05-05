package set

type Set[K comparable] struct {
	m map[K]bool
}

func New[K comparable]() *Set[K] {
	return &Set[K]{m: make(map[K]bool)}
}

func FromList[K comparable](list []K) *Set[K] {
	m := make(map[K]bool)
	for _, value := range list {
		m[value] = true
	}
	return &Set[K]{m: m}
}

func (s *Set[K]) Add(value K) {
	s.m[value] = true
}

func (s *Set[K]) Delete(value K) {
	delete(s.m, value)
}

func (s *Set[K]) Has(value K) bool {
	_, found := s.m[value]
	return found
}

func (s *Set[K]) List() []K {
	list := make([]K, len(s.m))
	i := 0
	for value := range s.m {
		list[i] = value
		i++
	}
	return list
}

func (s *Set[K]) Iter(callback func(value K)) {
	for value := range s.m {
		callback(value)
	}
}

func (s *Set[K]) IterIndex(callback func(value K, i int)) {
	i := 0
	for value := range s.m {
		callback(value, i)
		i++
	}
}

func (s *Set[K]) Size() int {
	return len(s.m)
}

func (s *Set[K]) Union(other *Set[K]) *Set[K] {
	res := New[K]()
	s.Iter(func(value K) { res.Add(value) })
	other.Iter(func(value K) { res.Add(value) })
	return res
}

func (s *Set[K]) Intersection(other *Set[K]) *Set[K] {
	if s.Size() > other.Size() {
		s, other = other, s
	}

	res := New[K]()
	s.Iter(func(value K) {
		if other.Has(value) {
			res.Add(value)
		}
	})

	return res
}

func (s *Set[K]) Copy() *Set[K] {
	res := New[K]()
	for k := range s.m {
		res.m[k] = true
	}
	return res
}

func (s *Set[K]) Subtract(other *Set[K]) *Set[K] {
	res := New[K]()
	for k := range s.m {
		if !other.Has(k) {
			res.m[k] = true
		}
	}
	return res

}

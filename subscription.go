package geoidx

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/util/set"
)

type EventType int

const (
	EventTypeSet EventType = iota
	EventTypeDelete
)

type Event struct {
	Type EventType `json:"type"`
	Obj  *Object   `json:"obj"`
}

type Subscription struct {
	id        string
	idx       *Index
	subBoxes  []*Object
	eventChan chan Event
	filters   []Filter
}

func newSubscription(idx *Index, chSize int) *Subscription {
	id := uuid.New().String()

	sub := &Subscription{
		id:        id,
		idx:       idx,
		subBoxes:  nil,
		eventChan: make(chan Event, chSize),
		filters:   make([]Filter, 0),
	}

	return sub
}

func (s *Subscription) ID() string {
	return s.id
}

func (s *Subscription) SetFilters(filters ...Filter) {
	toRemove := s.findTrackedObjectsIDs()
	s.filters = filters
	toAdd := s.findTrackedObjectsIDs()
	s.notifySetDelete(toAdd, toRemove)
}

func (s *Subscription) TrackID(id string) {
	toRemove := s.findTrackedObjectsIDs()
	s.idx.trackID(s, id)
	toAdd := s.findTrackedObjectsIDs()
	s.notifySetDelete(toAdd, toRemove)
}

func (s *Subscription) UntrackID(id string) {
	toRemove := s.findTrackedObjectsIDs()
	s.idx.untrackID(s, id)
	toAdd := s.findTrackedObjectsIDs()
	s.notifySetDelete(toAdd, toRemove)
}

func (s *Subscription) findTrackedObjectsIDs() *set.Set[string] {
	filters := []Filter{fltNonSubBoxes}
	filters = append(filters, s.filters...)

	ids := set.New[string]()
	for _, box := range s.subBoxes {
		for _, obj := range s.idx.SearchByObject(box, filters...) {
			ids.Add(obj.id)
		}
	}

	if trackedIDs, found := s.idx.sub2ids[s.id]; found {
		ids = ids.Union(trackedIDs)
	}
	return ids
}

func (s *Subscription) SetBounds(bounds Rect) {
	l := log.WithFields(logrus.Fields{
		"func":   "SetBounds",
		"sub_id": s.id,
	})
	// Gather old objects (to remove)

	toRemove := s.findTrackedObjectsIDs()
	l.Debugf("collected %d objects to remove", toRemove.Size())

	// Remove old boxes
	s.removeBoxes()
	s.setupBoxes(bounds)
	// Gather new objects (to add)
	toAdd := s.findTrackedObjectsIDs()
	l.Debugf("collected %d objects to add", toAdd.Size())

	s.notifySetDelete(toAdd, toRemove)
}

func (s *Subscription) notifySetDelete(toAdd *set.Set[string], toRemove *set.Set[string]) {
	l := log.WithFields(logrus.Fields{
		"func":   "notifySetDelete",
		"sub_id": s.id,
	})

	toAdd, toRemove = reduceSets(toAdd, toRemove)
	l.Debugf("calculated diff, %d objects to remove, %d objects to add",
		toRemove.Size(),
		toAdd.Size())

	l.Trace("emitting set")
	// - Notify set/delete
	toAdd.Iter(func(id string) {
		obj := s.idx.GetObjectByID(id)
		s.emitSet(obj)
	})

	l.Trace("emitting delete")
	toRemove.Iter(func(id string) {
		obj := s.idx.GetObjectByID(id)
		s.emitDelete(obj)
	})
}

func (s *Subscription) filterObject(obj *Object) *Object {
	for _, flt := range s.filters {
		if !flt(obj) {
			return nil
		}
	}
	return obj
}

func (s *Subscription) emitSet(obj *Object) {
	if obj != nil {
		event := Event{Type: EventTypeSet, Obj: obj}
		s.send(event)
	}
}

func (s *Subscription) emitDelete(obj *Object) {
	if obj != nil {
		event := Event{Type: EventTypeDelete, Obj: obj}
		s.send(event)
	}
}

func (s *Subscription) Events() <-chan Event {
	return s.eventChan
}

func (s *Subscription) send(event Event) {
	s.eventChan <- event
}

func (s *Subscription) removeBoxes() {
	log.WithFields(logrus.Fields{
		"func":   "removeBoxes",
		"sub_id": s.id,
	}).Debug("removing boxes")
	for _, box := range s.subBoxes {
		s.idx.DeleteNoNotify(box)
	}
	s.subBoxes = nil
}

func (s *Subscription) setupBoxes(bounds Rect) {
	log.WithFields(logrus.Fields{
		"func":   "setupBoxes",
		"sub_id": s.id,
		"bounds": bounds,
	}).Debug("setting up boxes")

	rects := split(bounds)
	boxes := make([]*Object, len(rects))
	for i, rect := range rects {
		sboxID := fmt.Sprintf("%s:%d", s.id, i)
		bounds := rect
		boxes[i] = NewObject(sboxID, bounds, s)
		s.idx.UpsertNoNotify(boxes[i])
	}
	s.subBoxes = boxes
}

func (s *Subscription) release() {
	log.WithFields(logrus.Fields{
		"func":   "release",
		"sub_id": s.id,
	}).Debug("release subscription")
	s.removeBoxes()
	close(s.eventChan)
}

func reduceSets[T comparable](add *set.Set[T], remove *set.Set[T]) (newAdd *set.Set[T], newRemove *set.Set[T]) {
	common := add.Intersection(remove)
	// - Subtract old from new -> get final set of objects to add
	newAdd = add.Subtract(common)
	// - Subtract new from old -> get final set of objects to remove
	newRemove = remove.Subtract(common)
	return
}

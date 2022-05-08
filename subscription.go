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
	Type EventType
	Obj  *Object
}

type Subscription struct {
	id        string
	idx       *Index
	subBoxes  []*Object
	eventChan chan Event
}

func newSubscription(idx *Index, chSize int) *Subscription {
	id := uuid.New().String()

	sub := &Subscription{
		id:        id,
		idx:       idx,
		subBoxes:  nil,
		eventChan: make(chan Event, chSize),
	}

	return sub
}

func (s *Subscription) SetBounds(bounds Rect) {
	l := log.WithFields(logrus.Fields{
		"func":   "SetBounds",
		"sub_id": s.id,
	})
	// Gather old objects (to remove)
	toRemove := set.New[string]()
	for _, box := range s.subBoxes {
		for _, obj := range s.idx.SearchByObject(box) {
			toRemove.Add(obj.id)
		}
	}
	l.Debugf("collected %d objects to remove", toRemove.Size())

	// Remove old boxes
	s.removeBoxes()
	s.setupBoxes(bounds)
	// Gather new objects (to add)
	toAdd := set.New[string]()
	for _, box := range s.subBoxes {
		for _, obj := range s.idx.SearchByObject(box, fltNonSubBoxes) {
			toAdd.Add(obj.id)
		}
	}
	l.Debugf("collected %d objects to add", toAdd.Size())

	common := toAdd.Intersection(toRemove)
	// - Subtract old from new -> get final set of objects to add
	toAdd = toAdd.Subtract(common)
	// - Subtract new from old -> get final set of objects to remove
	toRemove = toRemove.Subtract(common)
	l.Debugf("calculated diff, %d objects to remove, %d objects to add",
		toRemove.Size(),
		toAdd.Size())

	l.Debug("emitting set")
	// - Notify set/delete
	toAdd.Iter(func(id string) {
		obj := s.idx.GetObjectByID(id)
		s.setObject(obj)
	})

	l.Debug("emitting delete")
	toRemove.Iter(func(id string) {
		obj := s.idx.GetObjectByID(id)
		s.deleteObject(obj)
	})
}

func (s *Subscription) setObject(obj *Object) {
	if obj != nil {
		event := Event{Type: EventTypeSet, Obj: obj}
		s.send(event)
	}
}

func (s *Subscription) deleteObject(obj *Object) {
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

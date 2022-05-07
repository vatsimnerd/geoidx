package geoidx

import (
	"sync"

	"github.com/dhconnelly/rtreego"
	"github.com/vatsimnerd/util/set"
)

type Index struct {
	tree  *rtreego.Rtree
	idIdx map[string]*Object
	subs  map[string]*Subscription
	lock  sync.RWMutex
}

func NewIndex() *Index {
	return &Index{
		tree:  rtreego.NewTree(2, 25, 50),
		idIdx: make(map[string]*Object),
		subs:  make(map[string]*Subscription),
	}
}

func (i *Index) UpsertNoNotify(obj *Object) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if ex, found := i.idIdx[obj.id]; found {
		i.Delete(ex)
	}

	i.tree.Insert(obj)
	i.idIdx[obj.id] = obj
}

func (i *Index) Upsert(obj *Object) {
	i.UpsertNoNotify(obj)
	// find all sub boxes
	boxes := i.searchByRectUnsafe(obj.bounds, fltSubBoxes)
	// reduce them to a set of sub ids
	subIDs := set.New[string]()
	for _, box := range boxes {
		if sub, ok := box.val.(*Subscription); ok {
			subIDs.Add(sub.id)
		}
	}

	subIDs.Iter(func(id string) {
		if sub, found := i.subs[id]; found {
			sub.setObject(obj)
		}
	})
}

func (i *Index) DeleteNoNotify(obj *Object) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.tree.Delete(obj)
	delete(i.idIdx, obj.id)
}

func (i *Index) Delete(obj *Object) {
	i.DeleteNoNotify(obj)
	// find all sub boxes
	boxes := i.searchByRectUnsafe(obj.bounds, fltSubBoxes)
	// reduce them to a set of sub ids
	subIDs := set.New[string]()
	for _, box := range boxes {
		if sub, ok := box.val.(*Subscription); ok {
			subIDs.Add(sub.id)
		}
	}

	subIDs.Iter(func(id string) {
		if sub, found := i.subs[id]; found {
			sub.deleteObject(obj)
		}
	})
}

func (i *Index) searchByRectUnsafe(rect Rect, filters ...rtreego.Filter) []*Object {
	objects := make([]*Object, 0)
	for _, spatial := range i.tree.SearchIntersect(rect.ToRTreeRect(), filters...) {
		obj, ok := spatial.(*Object)
		if ok {
			objects = append(objects, obj)
		}
	}
	return objects
}

func (i *Index) SearchByRect(rect Rect, filters ...rtreego.Filter) []*Object {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.searchByRectUnsafe(rect, filters...)
}

func (i *Index) SearchByObject(obj *Object, filters ...rtreego.Filter) []*Object {
	if obj == nil {
		return nil
	}

	if filters == nil {
		filters = []rtreego.Filter{}
	}
	filters = append(filters, fltIDNMatch(obj.id))

	return i.SearchByRect(obj.bounds, filters...)
}

func (i *Index) SearchByObjectID(id string, filters ...rtreego.Filter) []*Object {
	i.lock.RLock()
	obj := i.idIdx[id]
	i.lock.RUnlock()
	return i.SearchByObject(obj, filters...)
}

func (i *Index) GetObjectByID(id string) *Object {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.idIdx[id]
}

func (i *Index) Subscribe(chSize int) *Subscription {
	i.lock.Lock()
	defer i.lock.Unlock()
	sub := newSubscription(i, chSize)
	i.subs[sub.id] = sub
	return sub
}

func (i *Index) Unsubscribe(sub *Subscription) {
	sub.release()

	i.lock.Lock()
	defer i.lock.Unlock()
	delete(i.subs, sub.id)
}

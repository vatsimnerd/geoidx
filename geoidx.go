package geoidx

import (
	"reflect"
	"sync"

	"github.com/dhconnelly/rtreego"
	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/util/set"
)

var (
	log = logrus.WithField("module", "geoidx")
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
	l := log.WithFields(logrus.Fields{
		"func": "UpsertNoNotify",
		"obj":  obj,
	})
	l.Debug("perform upsert")

	i.lock.Lock()
	defer i.lock.Unlock()

	if ex, found := i.idIdx[obj.id]; found {
		l.Debug("existing object found")
		i.Delete(ex)
	}

	l.Debug("inserting to tree")
	i.tree.Insert(obj)
	l.Debug("inserting to id index")
	i.idIdx[obj.id] = obj
}

func (i *Index) Upsert(obj *Object) {
	l := log.WithFields(logrus.Fields{
		"func": "Upsert",
		"obj":  obj,
	})
	i.UpsertNoNotify(obj)

	// find all sub boxes
	l.Debug("searching for subscription boxes")
	boxes := i.searchByRectUnsafe(obj.bounds, fltSubBoxes)
	// reduce them to a set of sub ids
	subIDs := set.New[string]()
	for _, box := range boxes {
		if sub, ok := box.val.(*Subscription); ok {
			subIDs.Add(sub.id)
		}
	}

	l.Debugf("found %d subscriptions, notifying", subIDs.Size())
	subIDs.Iter(func(id string) {
		if sub, found := i.subs[id]; found {
			sub.setObject(sub.filterObject(obj))
		}
	})
}

func (i *Index) DeleteNoNotify(obj *Object) {
	l := log.WithFields(logrus.Fields{
		"func": "DeleteNoNotify",
		"obj":  obj,
	})
	i.lock.Lock()
	defer i.lock.Unlock()

	l.Debug("deleting from tree")
	i.tree.Delete(obj)
	l.Debug("deleting from id index")
	delete(i.idIdx, obj.id)
}

func (i *Index) Delete(obj *Object) {
	l := log.WithFields(logrus.Fields{
		"func": "Delete",
		"obj":  obj,
	})
	i.DeleteNoNotify(obj)

	// find all sub boxes
	l.Debug("searching for subscription boxes")
	boxes := i.searchByRectUnsafe(obj.bounds, fltSubBoxes)
	// reduce them to a set of sub ids
	subIDs := set.New[string]()
	for _, box := range boxes {
		if sub, ok := box.val.(*Subscription); ok {
			subIDs.Add(sub.id)
		}
	}

	l.Debugf("found %d subscriptions, notifying", subIDs.Size())
	subIDs.Iter(func(id string) {
		if sub, found := i.subs[id]; found {
			sub.deleteObject(sub.filterObject(obj))
		}
	})
}

func (i *Index) searchByRectUnsafe(rect Rect, filters ...rtreego.Filter) []*Object {
	l := log.WithFields(logrus.Fields{
		"func":          "searchByRectUnsafe",
		"rect":          rect,
		"filters_count": len(filters),
	})

	l.Debug("performing search")
	objects := make([]*Object, 0)
	spatials := i.tree.SearchIntersect(rect.ToRTreeRect(), filters...)
	l.Debugf("found %d objects in tree", len(spatials))
	for _, spatial := range spatials {
		obj, ok := spatial.(*Object)
		if ok {
			objects = append(objects, obj)
		} else {
			l.WithField("type", reflect.TypeOf(obj).String()).Error("unexpected object type")
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

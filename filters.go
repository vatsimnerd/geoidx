package geoidx

import "github.com/dhconnelly/rtreego"

type Filter func(obj *Object) bool
type FilterList []Filter

func (f Filter) toRTreeGoFilter() rtreego.Filter {
	return func(results []rtreego.Spatial, object rtreego.Spatial) (refuse bool, abort bool) {
		obj, ok := object.(*Object)
		refuse = !ok || !f(obj)
		return
	}
}

func fltSubBoxes(obj *Object) bool {
	_, ok := obj.val.(*Subscription)
	return ok
}

func fltNonSubBoxes(obj *Object) bool {
	return !fltSubBoxes(obj)
}

func fltIDNMatch(id string) Filter {
	return func(obj *Object) bool {
		return obj.id != id
	}
}

func (fl FilterList) toRTreeGoFilterList() []rtreego.Filter {
	if fl == nil {
		return []rtreego.Filter{}
	}
	filters := make([]rtreego.Filter, len(fl))
	for i := 0; i < len(fl); i++ {
		filters[i] = fl[i].toRTreeGoFilter()
	}
	return filters
}

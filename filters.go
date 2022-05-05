package geoidx

import "github.com/dhconnelly/rtreego"

func isSubBox(spatial rtreego.Spatial) bool {
	obj, ok := spatial.(*Object)
	if !ok {
		return false
	}
	_, ok = obj.val.(*Subscription)
	return ok
}

func fltSubBoxes(results []rtreego.Spatial, spatial rtreego.Spatial) (refuse, abort bool) {
	refuse = !isSubBox(spatial)
	return
}

func fltNonSubBoxes(results []rtreego.Spatial, spatial rtreego.Spatial) (refuse, abort bool) {
	refuse = isSubBox(spatial)
	return
}

func fltIDNMatch(id string) func(results []rtreego.Spatial, spatial rtreego.Spatial) (refuse, abort bool) {
	return func(results []rtreego.Spatial, spatial rtreego.Spatial) (refuse bool, abort bool) {
		obj, ok := spatial.(*Object)
		if !ok {
			refuse = true
			return
		}
		refuse = obj.id == id
		return
	}
}

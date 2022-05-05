package geoidx

import "github.com/dhconnelly/rtreego"

type Object struct {
	id     string
	bounds Rect
	val    interface{}
}

func (o *Object) Bounds() *rtreego.Rect {
	return o.bounds.ToRTreeRect()
}

func (o *Object) Value() interface{} {
	return o.val
}

func (o *Object) ID() string {
	return o.id
}

func NewObject(id string, bounds Rect, val interface{}) *Object {
	return &Object{
		id:     id,
		bounds: bounds,
		val:    val,
	}
}

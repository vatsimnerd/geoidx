package geoidx

import "github.com/dhconnelly/rtreego"

type Point struct {
	Latitude  float64
	Longitude float64
}

type Rect struct {
	SouthWest Point
	NorthEast Point
}

const (
	eastmostLongitude = 179.9999999
	northmostLatitude = 89.9999999
)

func (r *Rect) ToRTreeRect() *rtreego.Rect {
	w := r.NorthEast.Longitude - r.SouthWest.Longitude
	h := r.NorthEast.Latitude - r.SouthWest.Latitude
	rect, err := rtreego.NewRect(rtreego.Point{r.SouthWest.Longitude, r.SouthWest.Latitude}, []float64{w, h})
	if err != nil {
		rect = rtreego.Point{r.SouthWest.Longitude, r.SouthWest.Latitude}.ToRect(0.001)
	}
	return rect
}

func MakeRect(minX, minY, maxX, maxY float64) Rect {
	return Rect{
		SouthWest: Point{
			Longitude: minX,
			Latitude:  minY,
		},
		NorthEast: Point{
			Longitude: maxX,
			Latitude:  maxY,
		},
	}
}

func split(r Rect) []Rect {
	rects := make([]Rect, 1)
	rects[0] = r

	if r.SouthWest.Longitude > r.NorthEast.Longitude {
		temp := make([]Rect, 0)
		for _, rect := range rects {
			temp = append(temp,
				// western box
				Rect{
					SouthWest: Point{
						Longitude: rect.SouthWest.Longitude,
						Latitude:  rect.SouthWest.Latitude,
					},
					NorthEast: Point{
						Longitude: eastmostLongitude,
						Latitude:  rect.NorthEast.Latitude,
					},
				},
				// eastern box
				Rect{
					SouthWest: Point{
						Longitude: -eastmostLongitude,
						Latitude:  rect.SouthWest.Latitude,
					},
					NorthEast: Point{
						Longitude: rect.NorthEast.Longitude,
						Latitude:  rect.NorthEast.Latitude,
					},
				},
			)
		}
		rects = temp
	}

	if r.SouthWest.Latitude > r.NorthEast.Latitude {
		temp := make([]Rect, 0)
		for _, rect := range rects {
			temp = append(temp,
				// northern box
				Rect{
					SouthWest: Point{
						Longitude: rect.SouthWest.Longitude,
						Latitude:  rect.SouthWest.Latitude,
					},
					NorthEast: Point{
						Longitude: rect.NorthEast.Longitude,
						Latitude:  northmostLatitude,
					},
				},
				// eastern box
				Rect{
					SouthWest: Point{
						Longitude: rect.SouthWest.Longitude,
						Latitude:  -northmostLatitude,
					},
					NorthEast: Point{
						Longitude: rect.NorthEast.Longitude,
						Latitude:  rect.NorthEast.Latitude,
					},
				},
			)
		}
		rects = temp
	}
	return rects
}

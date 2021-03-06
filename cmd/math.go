package cmd

import (
	"math"

	"gonum.org/v1/gonum/spatial/r3"
)

const EPS = 0.000001

type Plane struct {
	V r3.Vec
	D float64
}

func (p Plane) Classify(v r3.Vec) float64 {
	return r3.Dot(p.V, v) + p.D
}

func planeFromPoints(vv [3]r3.Vec) Plane {
	plane := Plane{}
	plane.V = r3.Unit(r3.Cross(r3.Sub(vv[1], vv[0]), r3.Sub(vv[2], vv[0])))
	plane.D = -r3.Dot(plane.V, vv[0])
	return plane
}

func calcIntersection(p0, p1, p2 Plane) *r3.Vec {
	denom := r3.Dot(p0.V, r3.Cross(p1.V, p2.V))
	if -EPS < denom && denom < EPS {
		return nil
	}
	d0 := r3.Cross(p1.V, p2.V).Scale(-p0.D)
	d1 := r3.Cross(p2.V, p0.V).Scale(-p1.D)
	d2 := r3.Cross(p0.V, p1.V).Scale(-p2.D)
	r := r3.Scale(1.0/denom, r3.Add(d0, r3.Add(d1, d2)))
	return &r
}

func calcOrigin(vs []r3.Vec) r3.Vec {
	max := r3.Vec{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64}
	min := r3.Vec{X: +math.MaxFloat64, Y: +math.MaxFloat64, Z: +math.MaxFloat64}
	for _, v := range vs {
		if v.X > max.X {
			max.X = v.X
		}
		if v.Y > max.Y {
			max.Y = v.Y
		}
		if v.Z > max.Z {
			max.Z = v.Z
		}
		if v.X < min.X {
			min.X = v.X
		}
		if v.Y < min.Y {
			min.Y = v.Y
		}
		if v.Z < min.Z {
			min.Z = v.Z
		}
	}
	return r3.Scale(0.5, r3.Add(min, max))
}

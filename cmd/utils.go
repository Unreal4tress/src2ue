package cmd

import (
	"fmt"

	"gonum.org/v1/gonum/spatial/r3"
)

func parse3Vec(text string) [3]r3.Vec {
	var v [3]r3.Vec
	fmt.Sscanf(text, "(%f %f %f) (%f %f %f) (%f %f %f)",
		&v[0].X, &v[0].Y, &v[0].Z, &v[1].X, &v[1].Y, &v[1].Z, &v[2].X, &v[2].Y, &v[2].Z)
	return v
}

func parseAxis(text string) [5]float64 {
	var v [5]float64
	fmt.Sscanf(text, "[%f %f %f %f] %f",
		&v[0], &v[1], &v[2], &v[3], &v[4])
	return v
}

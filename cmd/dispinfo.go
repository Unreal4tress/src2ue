package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Unreal4tress/go-sourceformat/vmf"
	"gonum.org/v1/gonum/spatial/r3"
)

type DispInfo struct {
	Power         int
	Seps          int
	StartPosition r3.Vec
	Normals       [][]r3.Vec
	Distances     [][]float64
}

func parseDispInfo(side *vmf.Node) *DispInfo {
	nodes := side.Nodes("dispinfo")
	if len(nodes) == 0 {
		return nil
	}
	dispinfo := nodes[0]
	r := new(DispInfo)
	r.Power = dispinfo.Int("power")
	r.Seps = []int{2, 3, 5, 9, 17}[r.Power]
	fmt.Sscanf(dispinfo.String("startposition"), "[%f %f %f]",
		&r.StartPosition.X, &r.StartPosition.Y, &r.StartPosition.Z)
	normals := dispinfo.Nodes("normals")[0]
	r.Normals = make([][]r3.Vec, r.Seps)
	for i := 0; i < r.Seps; i++ {
		key := fmt.Sprintf("row%v", i)
		vs := strings.Split(normals.String(key), " ")
		row := make([]r3.Vec, r.Seps)
		for j := 0; j < r.Seps; j++ {
			v := r3.Vec{}
			v.X, _ = strconv.ParseFloat(vs[j+0], 64)
			v.Y, _ = strconv.ParseFloat(vs[j+1], 64)
			v.Z, _ = strconv.ParseFloat(vs[j+2], 64)
			row[j] = v
		}
		r.Normals[i] = row
	}

	distances := dispinfo.Nodes("distances")[0]
	r.Distances = make([][]float64, r.Seps)
	for i := 0; i < r.Seps; i++ {
		key := fmt.Sprintf("row%v", i)
		text := distances.String(key)
		row := make([]float64, r.Seps)
		for j, v := range strings.Split(text, " ") {
			row[j], _ = strconv.ParseFloat(v, 64)
		}
		r.Distances[i] = row
	}
	return r
}

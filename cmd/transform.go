package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Unreal4tress/go-sourceformat/vmf"
	ulc "github.com/Unreal4tress/uelevelclip"
	cprint "github.com/fatih/color"
	"gonum.org/v1/gonum/spatial/r3"
)

func transform(vmf *vmf.Node) *ulc.Block {
	level := &ulc.Block{Class: "Level"}
	maap := &ulc.Block{Class: "Map", Children: []ulc.Node{level}}
	level.Children = make([]ulc.Node, 0, 1000000)

	////////////////////////////////////////
	//// Solid
	////////////////////////////////////////
	for _, solid := range vmf.Nodes("world")[0].Nodes("solid") {
		id := solid.ID()
		name := "Solid_" + strconv.Itoa(id)
		actor := &ulc.Block{Class: "Actor"}
		actor.Option = map[string]string{
			"Class": "/Script/Engine.Brush",
			"Name":  name,
		}
		brush := &ulc.Block{Class: "Brush", Option: map[string]string{"Name": fmt.Sprintf("Model_%v", id)}}
		polylist := &ulc.Block{Class: "PolyList"}

		var dispinfo *DispInfo
		dispinfoSide := -1

		type Face struct {
			Plane          Plane
			VertexIndecies []int
			Material       string
			UAxis          [5]float64
			VAxis          [5]float64
		}
		faces := make([]Face, solid.CountNodes("side"))
		toolMaterial := false
		for i, side := range solid.Nodes("side") {
			plane := parse3Vec(side.String("plane"))
			faces[i].Plane = planeFromPoints(plane)
			material := strings.ToLower(side.String("material"))
			if strings.HasPrefix(material, "tools/") {
				toolMaterial = true
				break
			}
			faces[i].Material = material
			faces[i].VertexIndecies = make([]int, 0, 16)
			faces[i].UAxis = parseAxis(side.String("uaxis"))
			faces[i].VAxis = parseAxis(side.String("vaxis"))
			di := parseDispInfo(side)
			if di != nil {
				dispinfo = di
				dispinfoSide = i
			}
		}
		if toolMaterial {
			continue
		}

		facesLen := len(faces)
		verteces := make([]r3.Vec, 0, 256)

		for i := 0; i < facesLen-2; i++ {
			faceI := faces[i]
			for j := i + 1; j < facesLen-1; j++ {
				faceJ := faces[j]
				for k := j + 1; k < facesLen; k++ {
					faceK := faces[k]
					ok := true
					v := calcIntersection(faceI.Plane, faceJ.Plane, faceK.Plane)
					if v == nil {
						continue
					}
					for _, faceL := range faces {
						if r3.Dot(faceL.Plane.V, *v)+faceL.Plane.D < -EPS {
							ok = false
							break
						}
					}
					if ok {
						verteces = append(verteces, *v)
						index := len(verteces) - 1
						faces[i].VertexIndecies = append(faces[i].VertexIndecies, index)
						faces[j].VertexIndecies = append(faces[j].VertexIndecies, index)
						faces[k].VertexIndecies = append(faces[k].VertexIndecies, index)
					}
				}
			}
		}

		center := r3.Vec{}
		for _, v := range verteces {
			center = r3.Add(center, v)
		}
		center = r3.Scale(float64(len(verteces)/1.0), center)

		for _, face := range faces {
			vertex := func(i int) r3.Vec {
				return verteces[face.VertexIndecies[i]]
			}
			average := r3.Vec{}
			for _, i := range face.VertexIndecies {
				average = r3.Add(average, verteces[i])
			}
			average = r3.Scale(1.0/float64(len(face.VertexIndecies)), average)
			viLen := len(face.VertexIndecies)

			for n := 0; n < viLen-2; n++ {
				a := r3.Unit(r3.Sub(vertex(n), average))
				p := planeFromPoints([3]r3.Vec{vertex(n), average, r3.Add(average, face.Plane.V)})
				smallestAngle := -1.0
				smallest := -1
				for m := n + 1; m < viLen; m++ {
					side := p.Classify(vertex(m))
					if side < EPS {
						continue
					}
					b := r3.Unit(r3.Sub(vertex(m), average))
					angle := r3.Dot(a, b)
					if angle > smallestAngle {
						smallestAngle = angle
						smallest = m
					}
				}
				face.VertexIndecies[n+1], face.VertexIndecies[smallest] = face.VertexIndecies[smallest], face.VertexIndecies[n+1]
			}
		}
		if dispinfo == nil {
			polylist.Children = make([]ulc.Node, len(faces))
		} else {
			polylist.Children = make([]ulc.Node, 0, 1000)
		}
		for i, face := range faces {
			if dispinfo != nil && dispinfoSide != i {
				continue
			}
			vertex := func(i int) r3.Vec {
				return verteces[face.VertexIndecies[i]]
			}
			material, materialOk := assets.Materials[strings.ToLower(face.Material)]
			if !materialOk {
				mat := strings.ToLower(face.Material)
				if _, found := unknownMaterials[mat]; !found {
					unknownMaterials[mat] = struct{}{}
					cprint.Yellow("Warning: Unknown material found: \"%v\"", mat)
				}
			}
			if dispinfo == nil {
				polygon := &ulc.Block{Class: "Polygon"}
				if materialOk {
					polygon.Option = map[string]string{"Texture": material.Asset}
				}
				polygon.Children = make([]ulc.Node, 0, 16)
				for _, vi := range face.VertexIndecies {
					v := verteces[vi]
					line := ulc.Line(fmt.Sprintf("Vertex   %+013.6f,%+013.6f,%+013.6f", v.Y*gScale, v.X*gScale, v.Z*gScale))
					polygon.Children = append(polygon.Children, &line)
				}
				if materialOk {
					uvu := r3.Vec{X: face.UAxis[0], Y: face.UAxis[1], Z: face.UAxis[2]}
					tu := r3.Scale(64/float64(material.W)/face.UAxis[4], uvu)
					uvv := r3.Vec{X: face.VAxis[0], Y: face.VAxis[1], Z: face.VAxis[2]}
					tv := r3.Scale(64/float64(material.H)/face.VAxis[4], uvv)
					lineu := ulc.Line(fmt.Sprintf("TextureU %+013.6f,%+013.6f,%+013.6f", tu.Y/1.0, tu.X/1.0, tu.Z/1.0))
					polygon.Children = append(polygon.Children, &lineu)
					linev := ulc.Line(fmt.Sprintf("TextureV %+013.6f,%+013.6f,%+013.6f", tv.Y/1.0, tv.X/1.0, tv.Z/1.0))
					polygon.Children = append(polygon.Children, &linev)
					oru := r3.Scale(face.UAxis[3]*float64(material.W)/-512.0, tu)
					orv := r3.Scale(face.VAxis[3]*float64(material.H)/-512.0, tv)
					or := r3.Add(oru, orv)
					lineo := ulc.Line(fmt.Sprintf("Origin   %+013.6f,%+013.6f,%+013.6f", or.Y, or.X, or.Z))
					polygon.Children = append(polygon.Children, &lineo)
				}
				polylist.Children[i] = polygon
			} else {
				var blv, brv, tlv, trv r3.Vec //Bottom Left Vertex
				for _, i := range face.VertexIndecies {
					if r3.Norm(r3.Sub(vertex(i), dispinfo.StartPosition)) < EPS {
						blv = vertex((i + 0) % 4)
						tlv = vertex((i + 1) % 4)
						trv = vertex((i + 2) % 4)
						brv = vertex((i + 3) % 4)
						break
					}
				}
				table := make([][]r3.Vec, dispinfo.Seps)
				for y := range table {
					table[y] = make([]r3.Vec, dispinfo.Seps)
					yp := float64(y) / float64(dispinfo.Seps-1)
					lv := r3.Scale(0.5, r3.Add(r3.Scale(1.0-yp, blv), r3.Scale(yp, tlv)))
					rv := r3.Scale(0.5, r3.Add(r3.Scale(1.0-yp, brv), r3.Scale(yp, trv)))
					for x := range table[y] {
						xp := float64(x) / float64(dispinfo.Seps-1)
						v := r3.Scale(0.5, r3.Add(r3.Scale(1.0-xp, rv), r3.Scale(xp, lv)))
						//TODO
						table[y][x] = v
					}
				}
				for y := 0; y < dispinfo.Seps-1; y++ {
					for x := 0; x < dispinfo.Seps-1; x++ {
						blv, brv, tlv, trv := table[y][x], table[y][x+1], table[y+1][x], table[y+1][x+1]

						polygon1 := &ulc.Block{Class: "Polygon", Children: make([]ulc.Node, 0, 3)}
						polygon2 := &ulc.Block{Class: "Polygon", Children: make([]ulc.Node, 0, 3)}
						if (x+y)%2 == 0 {
							polygon1.Children = append(polygon1.Children,
								makeUEVertexLine(blv), makeUEVertexLine(trv), makeUEVertexLine(tlv))
							polygon2.Children = append(polygon2.Children,
								makeUEVertexLine(blv), makeUEVertexLine(brv), makeUEVertexLine(trv))
						} else {
							polygon1.Children = append(polygon1.Children,
								makeUEVertexLine(blv), makeUEVertexLine(brv), makeUEVertexLine(tlv))
							polygon2.Children = append(polygon2.Children,
								makeUEVertexLine(brv), makeUEVertexLine(trv), makeUEVertexLine(tlv))
						}
						polylist.Children = append(polylist.Children, polygon1, polygon2)

						if y == 0 {
							polygon := &ulc.Block{Class: "Polygon", Children: make([]ulc.Node, 0, 3)}
							polygon.Children = append(polygon.Children,
								makeUEVertexLine(brv), makeUEVertexLine(blv), makeUEVertexLine(center),
							)
							polylist.Children = append(polylist.Children, polygon)
						}
						if x == 0 {
							polygon := &ulc.Block{Class: "Polygon", Children: make([]ulc.Node, 0, 3)}
							polygon.Children = append(polygon.Children,
								makeUEVertexLine(blv), makeUEVertexLine(tlv), makeUEVertexLine(center),
							)
							polylist.Children = append(polylist.Children, polygon)
						}
						if y == dispinfo.Seps-2 {
							polygon := &ulc.Block{Class: "Polygon", Children: make([]ulc.Node, 0, 3)}
							polygon.Children = append(polygon.Children,
								makeUEVertexLine(tlv), makeUEVertexLine(trv), makeUEVertexLine(center),
							)
							polylist.Children = append(polylist.Children, polygon)
						}
						if x == dispinfo.Seps-2 {
							polygon := &ulc.Block{Class: "Polygon", Children: make([]ulc.Node, 0, 3)}
							polygon.Children = append(polygon.Children,
								makeUEVertexLine(trv), makeUEVertexLine(brv), makeUEVertexLine(center),
							)
							polylist.Children = append(polylist.Children, polygon)
						}

					}
				}
			}
		}

		brush.Children = []ulc.Node{polylist}
		brushModel := ulc.Line(fmt.Sprintf("Brush=Model'\"Model_%v\"'", id))
		actorLabel := ulc.Line(fmt.Sprintf("ActorLabel=\"Solid_%v\"", id))
		folderPath := ulc.Line("FolderPath=\"Solids\"")
		actor.Children = []ulc.Node{brush, &brushModel, &actorLabel, &folderPath}
		level.Children = append(level.Children, actor)
	}

	return maap
}

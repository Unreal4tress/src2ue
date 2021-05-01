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
		name := "S_MapBrush_" + rule.MapName + "_" + strconv.Itoa(id)
		actor := &ulc.Block{Class: "Actor"}
		actor.Option = map[string]string{
			"Class": "/Script/Engine.StaticMeshActor",
			"Name":  name,
		}
		object := &ulc.Block{Class: "Object", Option: map[string]string{"Name": "StaticMeshComponent0"}}

		//var dispinfo *DispInfo
		//dispinfoSide := -1

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
				//dispinfo = di
				//dispinfoSide = i
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

		center := calcOrigin(verteces)

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

		obj := new(OBJ)
		obj.V = make([]r3.Vec, len(verteces))
		for i, v := range verteces {
			obj.V[i] = r3.Sub(v, center)
		}
		obj.FS = make([]OBJF, 0, 16)
		for _, face := range faces {
			material, materialOk := assets.Materials[strings.ToLower(face.Material)]
			if !materialOk {
				material = struct {
					Asset string `json:"asset"`
					W     int    `json:"w"`
					H     int    `json:"h"`
				}{
					Asset: "MaterialNotFound.MaterialNotFound",
				}
			}
			objfs := OBJF{
				Mtl: strings.Split(material.Asset, ".")[1],
				F:   make([]OBJFV, 0, 16),
			}
			for _, vi := range face.VertexIndecies {
				objfs.F = append(objfs.F, OBJFV{
					V: vi + 1,
				})
			}
			obj.FS = append(obj.FS, objfs)
		}
		objPath := rule.DstPath + rule.MapName + "/Brushes/" + name + ".obj"
		obj.Export(objPath)
		cprint.Blue("Info: Wrote: %v", objPath)
		gameName := rule.GamePath + rule.MapName + "/Brushes/" + name + "." + name
		objectStaticMesh := ulc.Line(fmt.Sprintf("StaticMesh=%v", gameName))
		objectRelativeLocation := ulc.Line(fmt.Sprintf("RelativeLocation=(X=%.6f,Y=%.6f,Z=%.6f)",
			center.Y*gScale, center.X*gScale, center.Z*gScale))
		object.Children = []ulc.Node{&objectStaticMesh, &objectRelativeLocation}
		actorLabel := ulc.Line(fmt.Sprintf("ActorLabel=\"Solid_%v\"", id))
		folderPath := ulc.Line("FolderPath=\"Solids\"")
		actor.Children = []ulc.Node{object, &actorLabel, &folderPath}
		level.Children = append(level.Children, actor)
	}

	return maap
}

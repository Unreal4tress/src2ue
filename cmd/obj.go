package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gonum.org/v1/gonum/spatial/r3"
)

type OBJ struct {
	V []r3.Vec
	//VT []r2.Vec
	//VN []r3.Vec
	FS []OBJF
}

type OBJF struct {
	Mtl string
	F   []OBJFV
}

type OBJFV struct {
	V int
}

func (o *OBJ) Export(path string) error {
	if !strings.HasSuffix(path, ".obj") {
		return errors.New("are")
	}
	mtlPath := path[:len(path)-4] + ".mtl"
	obj, err := os.Create(path)
	if err != nil {
		return err
	}
	defer obj.Close()
	mtl, err := os.Create(mtlPath)
	if err != nil {
		return err
	}
	defer mtl.Close()
	fmt.Fprintf(obj, "mtllib %v\n", filepath.Base(mtlPath))
	for _, v := range o.V {
		fmt.Fprintf(obj, "v %.6f %.6f %.6f\n", v.Y, -v.X, v.Z)
	}
	mats := make(map[string]struct{})
	for _, fs := range o.FS {
		fmt.Fprintf(obj, "usemtl %v\n", fs.Mtl)
		fmt.Fprintf(obj, "f")
		for i := len(fs.F) - 1; i >= 0; i-- {
			f := fs.F[i]
			fmt.Fprintf(obj, " %v", f.V)
		}
		fmt.Fprintln(obj)
		mats[fs.Mtl] = struct{}{}
	}
	for mat := range mats {
		fmt.Fprintf(mtl, "newmtl %v\n", mat)
	}
	return nil
}

package dix

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
)

var annRe = regexp.MustCompile(`^@([A-Za-z0-9_]+):\s*(.+)$`)

func parseFileComments(path string) ([]Annotation, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	dir := filepath.ToSlash(filepath.Dir(path))
	var out []Annotation
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			txt := strings.TrimSpace(c.Text)
			// remove leading // or /* and trailing */
			txt = strings.TrimPrefix(txt, "//")
			txt = strings.TrimPrefix(txt, "/*")
			txt = strings.TrimSuffix(txt, "*/")
			txt = strings.TrimSpace(txt)

			if strings.HasPrefix(txt, "@") {
				if m := annRe.FindStringSubmatch(txt); m != nil {
					pos := fset.Position(c.Pos())

					metadata := &AnnotationMetadata{
						Key:   m[1],
						Value: strings.TrimSpace(m[2]),
						File:  path,
						Line:  pos.Line,
						Path:  dir,
					}

					if metadata.Key == "factory" {
						out = append(out, parseFactoryAnnotation(metadata))
					}

					if metadata.Key == "wire" {
						out = append(out, parseWireAnnotation(metadata))
					}
				}
			}
		}
	}
	return out, nil
}

func parseFactoryAnnotation(ann *AnnotationMetadata) *FactoryAnnotation {
	value := strings.TrimSpace(ann.Value)
	splitted := strings.Split(value, "->")
	funcName := strings.TrimSpace(splitted[0])
	alias := strings.TrimSpace(splitted[1])
	return &FactoryAnnotation{
		Path:     ann.Path,
		Function: funcName,
		Alias:    alias,
	}
}

func parseWireAnnotation(ann *AnnotationMetadata) *WireAnnotation {
	wireRe := regexp.MustCompile(`^([A-Za-z0-9_]+)\(([^)]*)\)$`)
	m := wireRe.FindStringSubmatch(strings.TrimSpace(ann.Value))

	funcName := m[1]
	var depsOut []string
	if strings.TrimSpace(m[2]) != "" {
		depsOut = strings.Split(m[2], ",")
		for i := range depsOut {
			depsOut[i] = strings.TrimSpace(depsOut[i])
		}
	}

	return &WireAnnotation{
		Path:   ann.Path,
		Target: funcName,
		Deps:   depsOut,
	}
}

func scanDir(root string) ([]Annotation, error) {
	var res []Annotation
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		fmt.Println("[Dix] Scanning file " + p)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(p) == ".go" {
			anns, err := parseFileComments(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse error %s: %v\n", p, err)
				return nil
			}

			res = append(res, anns...)
		}
		return nil
	})
	return res, err
}

func ScanProjectAndGenerateDI(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		panic(err)
	}

	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("[Dix] Scanning module name:", f.Module.Mod.Path)

	anns, err := scanDir(root)
	if err != nil {
		return "", errors.New("scan error: " + err.Error())
	}

	diConfig := &DIConfig{
		Container: make(map[string]*Factory),
	}

	for _, a := range anns {
		if a.Type() == "Factory" {
			factory := a.(*FactoryAnnotation)
			modName := filepath.ToSlash(filepath.Join(f.Module.Mod.Path, factory.Path))
			fmt.Println("[Dix] Detect factory " + factory.Function + " -> " + factory.Alias + " in " + modName)
			if _, ok := diConfig.Container[factory.Alias]; ok {
				if diConfig.Container[factory.Alias].Module == factory.Path {
					return "", errors.New("Duplicate Alias " + factory.Alias + " in " + diConfig.Container[factory.Alias].Module)
				} else {
					return "", errors.New("Alias " + factory.Alias + " used by " + diConfig.Container[factory.Alias].Module)
				}
			} else {
				diConfig.Container[factory.Alias] = &Factory{
					Function: factory.Function,
					Deps:     make([]*Dependency, 0),
					Module:   modName,
				}
			}
		}
	}

	for _, a := range anns {
		if a.Type() == "Wire" {
			wire := a.(*WireAnnotation)
			if _, ok := diConfig.Container[wire.Target]; ok {
				fmt.Println("[Dix] Detect dependency [" + strings.Join(wire.Deps, ",") + "] -> " + wire.Target + " in " + diConfig.Container[wire.Target].Module)
				deps := []*Dependency{}
				for _, d := range wire.Deps {
					depName := strings.TrimSpace(d)
					standalone := false

					if strings.HasPrefix(depName, "^") {
						standalone = true
						depName = strings.Split(depName, "^")[1]
					}

					_, hasDep := diConfig.Container[depName]
					if hasDep {
						deps = append(deps, &Dependency{Name: depName, Standalone: standalone})
					} else {
						return "", errors.New("Can't resolve dependency " + depName + " of " + wire.Target + " in " + diConfig.Container[wire.Target].Module)
					}

				}
				diConfig.Container[wire.Target].Deps = deps
			}

		}
	}
	return GenerateCode(f.Module.Mod.Path, diConfig)
}

func Mark(values ...any) {}

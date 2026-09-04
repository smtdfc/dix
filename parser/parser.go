package parser

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

type Parser struct{}

type ScanOptions struct {
	Workspace bool
	NoCache   bool
}

func isDixSingletonNamed(named *types.Named) bool {
	if named == nil {
		return false
	}

	origin := named.Origin()
	if origin == nil || origin.Obj() == nil || origin.Obj().Pkg() == nil {
		return false
	}

	return origin.Obj().Name() == "Singleton" &&
		origin.Obj().Pkg().Path() == "github.com/smtdfc/dix/di"
}

func (p *Parser) ParseProvider(pkg *packages.Package, file *ast.File, fn *ast.FuncDecl) (*Provider, error) {
	c := &Provider{
		Name:        fn.Name.Name,
		File:        pkg.Fset.Position(file.Package).Filename,
		PackagePath: pkg.PkgPath,
		PackageName: pkg.Name,
	}

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			_, isPointerAtAST := field.Type.(*ast.StarExpr)

			tv := pkg.TypesInfo.TypeOf(field.Type)
			if tv == nil {
				continue
			}

			for _, name := range field.Names {
				var depType *types.Type
				isSingleton := false

				var singletonNamed *types.Named
				switch t := tv.(type) {
				case *types.Named:
					if isDixSingletonNamed(t) {
						singletonNamed = t
					}
				case *types.Pointer:
					if n, ok := t.Elem().(*types.Named); ok && isDixSingletonNamed(n) {
						return nil, NewValidationError(
							"singleton dependency must be `di.Singleton[T]`, not `*di.Singleton[T]`",
							fn.Name.Name,
							name.Name,
							c.File,
						)
					}
				}

				if singletonNamed != nil {
					if isPointerAtAST {
						return nil, NewValidationError(
							"singleton dependency must be `di.Singleton[T]`, not pointer form",
							fn.Name.Name,
							name.Name,
							c.File,
						)
					}

					targs := singletonNamed.TypeArgs()
					if targs != nil && targs.Len() > 0 {
						innerType := targs.At(0)
						depType = &innerType
						isSingleton = true
					}
				}

				if depType == nil {
					depType = &tv
				}

				typeName, isPtr := parseTypeDetails(*depType)
				pkgPath := getPackagePath(*depType)

				c.Deps = append(c.Deps, &Dependency{
					Name: name.Name,
					Type: &TypeInfo{
						Name:      typeName,
						Pkg:       pkgPath,
						IsPointer: isPtr,
					},
					IsSingleton: isSingleton,
				})
			}
		}
	}

	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return nil, NewValidationError(
			"provider function must return one value or (value, error)",
			fn.Name.Name,
			"",
			c.File,
		)
	}

	if len(fn.Type.Results.List) > 2 || len(fn.Type.Results.List[0].Names) > 1 {
		return nil, NewValidationError(
			"provider function must return one value or (value, error)",
			fn.Name.Name,
			"",
			c.File,
		)
	}
	if len(fn.Type.Results.List) == 2 {
		errorResult := fn.Type.Results.List[1]
		if len(errorResult.Names) > 1 || !isBuiltinError(pkg.TypesInfo.TypeOf(errorResult.Type)) {
			return nil, NewValidationError(
				"the second provider return value must be error",
				fn.Name.Name,
				"",
				c.File,
			)
		}
		c.ReturnsError = true
	}

	valueResult := fn.Type.Results.List[0]
	rtv := pkg.TypesInfo.TypeOf(valueResult.Type)
	rName, rIsPtr := parseTypeDetails(rtv)
	rPkg := getPackagePath(rtv)
	c.Return = &ReturnValue{Type: &TypeInfo{Name: rName, Pkg: rPkg, IsPointer: rIsPtr}}

	return c, nil
}

func isBuiltinError(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.Identical(t, types.Universe.Lookup("error").Type())
}

func (p *Parser) Parse(dir string) (*Metadata, error) {
	return p.ParseWithOptions(dir, ScanOptions{})
}

func (p *Parser) ParseWithOptions(dir string, options ScanOptions) (*Metadata, error) {
	dirs := []string{dir}
	if options.Workspace {
		var err error
		dirs, err = workspaceModuleDirs(dir)
		if err != nil {
			return nil, err
		}
	}

	metadata := new(Metadata)
	var cachePath, fingerprint string
	if !options.NoCache {
		if key, err := cacheKey(dirs, options); err == nil {
			fingerprint = key
			baseDir := dirs[0]
			if options.Workspace {
				if workFile, findErr := findGoWork(dir); findErr == nil {
					baseDir = filepath.Dir(workFile)
				}
			} else if moduleRoot, rootErr := findModuleRoot(baseDir); rootErr == nil {
				baseDir = moduleRoot
			}
			cachePath = filepath.Join(baseDir, ".dix", "cache", key+".json")
			if cached, ok := loadCache(cachePath, fingerprint); ok {
				fmt.Printf("\033[36m[Cache]\033[0m Hit: %s\n", cachePath)
				return cached, nil
			}
		}
	}
	providers := make(map[string]*Provider)
	for _, moduleDir := range dirs {
		if err := p.parseModule(moduleDir, metadata, providers); err != nil {
			return nil, err
		}
	}

	if cachePath != "" {
		saveCache(cachePath, fingerprint, metadata)
	}
	return metadata, nil
}

func findModuleRoot(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(absDir, "go.mod")); err == nil {
			return absDir, nil
		}
		parent := filepath.Dir(absDir)
		if parent == absDir {
			return "", fmt.Errorf("no go.mod found from %s", dir)
		}
		absDir = parent
	}
}

func (p *Parser) parseModule(dir string, metadata *Metadata, providers map[string]*Provider) error {

	cfg := &packages.Config{
		Dir:        dir,
		BuildFlags: []string{"-tags=dix"},
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return &ParseError{
			Kind:    ParseErrorPackageLoad,
			Message: "failed to load Go packages",
			Cause:   err,
		}
	}

	var parseErr error
	for _, pkg := range pkgs {

		if len(pkg.Errors) > 0 {
			return NewPackageLoadError(pkg.Errors[0])
		}

		for _, file := range pkg.Syntax {

			fileName := pkg.Fset.Position(file.Package).Filename

			fmt.Printf("\033[32m[Scan]\033[0m File: %s ... ", fileName)

			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok || fn.Doc == nil {
					return true
				}

				if containsInjectableAnnotation(fn.Doc.Text()) {
					m, err := p.ParseProvider(pkg, file, fn)
					if err != nil {
						parseErr = err
						return false
					}

					if containsRootAnnotation(fn.Doc.Text()) {
						if metadata.Root != nil {
							parseErr = NewValidationError(
								fmt.Sprintf("multiple @Root providers: %s and %s", metadata.Root.Name, m.Name),
								m.Name, "", m.File,
							)
							return false
						}
						metadata.Root = m
					} else {
						key := m.Return.Type.Signature()
						if existing, exists := providers[key]; exists {
							parseErr = NewValidationError(
								fmt.Sprintf("duplicate provider for %s: %s (%s) and %s (%s)", m.Return.Type.String(), existing.Name, existing.File, m.Name, m.File),
								m.Name, "", m.File,
							)
							return false
						}
						providers[key] = m
						metadata.Providers = append(metadata.Providers, m)
					}

					if containsDisableAnnotation(fn.Doc.Text()) {
						m.IsDisable = true
					}
				}
				return true
			})
			if parseErr != nil {
				// Close the in-progress scan line before printing fatal error output.
				fmt.Println()
				return parseErr
			}

			color.New(color.FgGreen).Printf("OK\n")

		}

	}

	return nil
}

func workspaceModuleDirs(dir string) ([]string, error) {
	workFile, err := findGoWork(dir)
	if err != nil {
		return nil, err
	}

	body, err := os.ReadFile(workFile)
	if err != nil {
		return nil, &ParseError{
			Kind:    ParseErrorWorkspace,
			Message: fmt.Sprintf("cannot read workspace file %s", workFile),
			Cause:   err,
		}
	}
	work, err := modfile.ParseWork(workFile, body, nil)
	if err != nil {
		return nil, &ParseError{
			Kind:    ParseErrorWorkspace,
			Message: fmt.Sprintf("cannot parse workspace file %s", workFile),
			Cause:   err,
		}
	}

	workDir := filepath.Dir(workFile)
	moduleDirs := make([]string, 0, len(work.Use))
	for _, use := range work.Use {
		moduleDir := use.Path
		if !filepath.IsAbs(moduleDir) {
			moduleDir = filepath.Join(workDir, moduleDir)
		}
		moduleDir = filepath.Clean(moduleDir)
		if _, err := os.Stat(filepath.Join(moduleDir, "go.mod")); err != nil {
			return nil, &ParseError{
				Kind:    ParseErrorWorkspace,
				Message: fmt.Sprintf("workspace module %s does not contain go.mod", moduleDir),
				Cause:   err,
			}
		}
		moduleDirs = append(moduleDirs, moduleDir)
	}
	if len(moduleDirs) == 0 {
		return nil, NewParseError(ParseErrorWorkspace, fmt.Sprintf("workspace file %s does not declare any modules", workFile))
	}
	return moduleDirs, nil
}

func findGoWork(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", NewParseError(ParseErrorWorkspace, "cannot resolve scan directory")
	}
	for {
		candidate := filepath.Join(absDir, "go.work")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(absDir)
		if parent == absDir {
			return "", NewParseError(ParseErrorWorkspace, fmt.Sprintf("no go.work found from %s", dir))
		}
		absDir = parent
	}
}
func NewParser() *Parser {
	return &Parser{}
}

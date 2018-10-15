package liana

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"github.com/knq/snaker"
	"github.com/reddec/astools"
	"go/ast"
	"path/filepath"
	"strings"
)

// Parameters for HTTP wrapper generator
type WrapperParams struct {
	File              string   // required, path for source file
	AdditionalImports []string // optional, additional imports for result render (almost never should be used)
	OutPackageName    string   // optional, package name in render (default is same as in file)
	OutPackagePath    string   // optional, package path in render (default is same as in file). If it is  a same as InPackagePath import will be not used
	InPackagePath     string   // optional, source file import path (default is a directory name of file)
}

// Generate golang code that's expose functions defined in a exported interfaces.
// For each suitable interface Gin wrapper over HTTP POST method is generated, where input parameters can be form fields, json or XML
// data. Input parameter names are same as input parameters but in a snake case.
// If function has more then one non-error output it is skipped.
func GenerateInterfacesWrapperHTTP(params WrapperParams) (string, error) {
	f, err := atool.Scan(params.File)
	if err != nil {
		return "", err
	}

	InPackagePath := filepath.Dir(params.File)
	if params.InPackagePath != "" {
		InPackagePath = params.InPackagePath
	}

	OutPackagePath := ""

	OutPackageName := f.Package
	if params.OutPackageName != "" {
		OutPackageName = params.OutPackageName
	}

	if f.Package == OutPackageName {
		OutPackagePath = InPackagePath
	}

	out := jen.NewFilePathName(OutPackagePath, OutPackageName)
	for _, imp := range params.AdditionalImports {
		out.ImportName(imp, "")
	}
	out.ImportName("github.com/gin-gonic/gin", "gin")
	out.ImportName(params.InPackagePath, f.Package)
	out.HeaderComment("DO NOT EDIT! This is automatically generated wrapper")
	for _, ifs := range f.Interfaces {
		if !ast.IsExported(ifs.Name) {
			continue
		}
		typeName := "handler" + ifs.Name
		out.Type().Id(typeName).StructFunc(func(group *jen.Group) {
			group.Id("wrap").Qual(InPackagePath, ifs.Name)
		})
		out.Line()
		var wrappedMethods []*atool.Method
		for _, method := range ifs.Methods {
			if method.HasOutput() && len(method.NonErrorOutputs()) > 1 {
				continue
			}
			wrappedMethods = append(wrappedMethods, method)
			argType := "args" + method.Name + "Handler"
			out.Type().Id(argType).StructFunc(func(group *jen.Group) {
				for _, param := range method.In {
					tagName := snaker.CamelToSnake(param.Name)

					st, err := f.ExtractType(param.Type)
					if err != nil {
						panic(err)
					}
					qualType := jen.Id(param.GolangType())
					if st.File.Import != "" {
						_, name := param.GoPkgType()
						qualType = jen.Qual(st.File.Import, name)
						if param.IsPointer() {
							qualType = jen.Op("*").Add(qualType)
						}
					}
					group.Id(strings.Title(param.Name)).Add(qualType).Tag(map[string]string{
						"json":  tagName,
						"form":  tagName,
						"query": tagName,
						"xml":   tagName,
					})
				}
			})

			if method.Comment != "" {
				out.Comment(strings.TrimSpace(method.Comment))
			}
			out.Func().Parens(jen.Id("h").Op("*").Id(typeName)).Id("handle" + method.Name).Params(jen.Id("gctx").Op("*").Qual("github.com/gin-gonic/gin", "Context")).BlockFunc(func(group *jen.Group) {
				group.Var().Id("params").Id(argType)
				group.IfFunc(func(ifG *jen.Group) {
					ifG.Err().Op(":=").Id("gctx").Dot("Bind").Call(jen.Op("&").Id("params"))
					ifG.Err().Op("!=").Nil()
				}).BlockFunc(func(g *jen.Group) {
					g.Qual("log", "Println").Call(jen.Lit("["+method.Name+"]"), jen.Lit("failed to parse arguments:"), jen.Err())
					g.Id("gctx").Dot("AbortWithError").Call(jen.Qual("net/http", "StatusBadRequest"), jen.Err())
					g.Return()
				})
				call := jen.Id("h").Dot("wrap").Dot(method.Name).CallFunc(func(args *jen.Group) {
					for _, inParam := range method.In {
						args.Id("params").Dot(strings.Title(inParam.Name))
					}
				})
				if method.HasOutput() {
					group.ListFunc(func(result *jen.Group) {
						for _, outParam := range method.Out {
							result.Id(outParam.Name)
						}
					}).Op(":=").Add(call)
					for _, errOut := range method.ErrorOutputs() {
						group.If(jen.Id(errOut.Name).Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
							g.Qual("log", "Println").Call(jen.Lit("["+method.Name+"]"), jen.Lit("invoke returned error:"), jen.Id(errOut.Name))
							g.Id("gctx").Dot("AbortWithError").Call(jen.Qual("net/http", "StatusInternalServerError"), jen.Id(errOut.Name))
							g.Return()
						})
					}
					if len(method.NonErrorOutputs()) == 0 {
						group.Id("gctx").Dot("AbortWithStatus").Call(jen.Qual("net/http", "StatusNoContent"))
					} else {
						for _, result := range method.NonErrorOutputs() {
							group.Id("gctx").Dot("IndentedJSON").Call(jen.Qual("net/http", "StatusOK"), jen.Id(result.Name))
						}
					}
				} else {
					group.Add(call)
					group.Id("gctx").Dot("AbortWithStatus").Call(jen.Qual("net/http", "StatusNoContent"))
				}
			})

		}

		out.Line()
		// Gin wrapper
		comment := "Wrapper of " + f.Package + "." + ifs.Name + " that expose functions over simple JSON HTTP interface.\n Those methods are wrapped: "
		var parts []string
		for _, method := range wrappedMethods {
			parts = append(parts, method.Name+" (POST /"+toKebab(method.Name)+")")
		}
		comment += strings.Join(parts, ",\n ")
		out.Comment(comment)
		out.Func().Id("Wrap"+ifs.Name).Params(jen.Id("wrapper").Qual(InPackagePath, ifs.Name)).Qual("net/http", "Handler").BlockFunc(func(group *jen.Group) {
			group.Id("router").Op(":=").Qual("github.com/gin-gonic/gin", "Default").Call()
			group.Id("GinWrap"+ifs.Name).Call(jen.Id("wrapper"), jen.Id("router"))
			group.Return(jen.Id("router"))
		})
		out.Line()
		out.Comment("Same as Wrap but allows to use your own Gin instance")
		out.Func().Id("GinWrap"+ifs.Name).Params(jen.Id("wrapper").Qual(InPackagePath, ifs.Name), jen.Id("router").Qual("github.com/gin-gonic/gin", "IRoutes")).BlockFunc(func(group *jen.Group) {
			group.Id("handler").Op(":=").Id(typeName).Values(jen.Id("wrapper"))
			for _, method := range wrappedMethods {
				group.Id("router").Dot("POST").Call(jen.Lit("/"+toKebab(method.Name)), jen.Id("handler").Dot("handle"+method.Name))
			}
		})
	}

	buffer := &bytes.Buffer{}
	err = out.Render(buffer)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func toKebab(v string) string {
	return strings.Replace(snaker.CamelToSnake(v), "_", "-", -1)
}

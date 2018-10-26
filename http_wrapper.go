package liana

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"github.com/knq/snaker"
	"github.com/reddec/astools"
	"go/ast"
	"gopkg.in/yaml.v2"
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
	DisableSwagger    bool     // optional, if specified skip generates swagger files for all interfaces found in the file
	FilterInterfaces  []string //optional, if specified generates only for specified interfaces
	Lock              bool     //optional, if specified expects global lockable object (global sync)
}

// Result of generator
type GenerateResult struct {
	Wrapper  string            // generate go code
	Swaggers map[string]string // if not disabled, generated swagger against interfaces names
}

// Generate golang code that's expose functions defined in a exported interfaces.
// For each suitable interface Gin wrapper over HTTP POST method is generated, where input parameters can be form fields, json or XML
// data. Input parameter names are same as input parameters but in a snake case.
// If function has more then one non-error output it is skipped.
func GenerateInterfacesWrapperHTTP(params WrapperParams) (GenerateResult, error) {

	f, err := atool.Scan(params.File)
	if err != nil {
		return GenerateResult{}, err
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

	var result GenerateResult
	result.Swaggers = make(map[string]string)

	out := jen.NewFilePathName(OutPackagePath, OutPackageName)
	for _, imp := range params.AdditionalImports {
		out.ImportName(imp, "_")
	}
	out.ImportName("github.com/gin-gonic/gin", "gin")
	out.ImportAlias(params.InPackagePath, f.Package)
	out.HeaderComment("DO NOT EDIT! This is automatically generated wrapper")
	for _, ifs := range f.Interfaces {
		if !ast.IsExported(ifs.Name) {
			continue
		}
		if len(params.FilterInterfaces) > 0 {
			var ok bool
			for _, ifn := range params.FilterInterfaces {
				if ifn == ifs.Name {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		typeName := "handler" + ifs.Name
		out.Type().Id(typeName).StructFunc(func(group *jen.Group) {
			group.Id("wrap").Qual(InPackagePath, ifs.Name)
			if params.Lock {
				group.Id("lock").Qual("sync", "Locker")
			}
		})
		out.Line()
		clientTypeName := "client" + ifs.Name
		out.Type().Id(clientTypeName).StructFunc(func(group *jen.Group) {
			group.Id("baseURL").String().Comment("Base url for requests")
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
					findRender(param, f).OnStructField(out, group, param, f, params)
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
				for _, param := range method.In {
					findRender(param, f).OnParseField(out, group, param, f)
				}

				call := jen.Id("h").Dot("wrap").Dot(method.Name).CallFunc(func(args *jen.Group) {
					for _, inParam := range method.In {
						args.Id("params").Dot(strings.Title(inParam.Name))
					}
				})
				if params.Lock {
					group.Id("h").Dot("lock").Dot("Lock").Call()
				}
				if method.HasOutput() {
					group.ListFunc(func(result *jen.Group) {
						for _, outParam := range method.Out {
							result.Id(outParam.Name)
						}
					}).Op(":=").Add(call)
					if params.Lock {
						group.Id("h").Dot("lock").Dot("Unlock").Call()
					}
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
					if params.Lock {
						group.Id("h").Dot("lock").Dot("Unlock").Call()
					}
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
		var subParams jen.Code
		var subCall jen.Code
		if params.Lock {
			subParams = jen.Id("lock").Qual("sync", "Locker")
			subCall = jen.Id("lock")
		}
		out.Func().Id("Wrap"+ifs.Name).Params(jen.Id("wrapper").Qual(InPackagePath, ifs.Name), subParams).Qual("net/http", "Handler").BlockFunc(func(group *jen.Group) {
			group.Id("router").Op(":=").Qual("github.com/gin-gonic/gin", "Default").Call()
			group.Id("GinWrap"+ifs.Name).Call(jen.Id("wrapper"), jen.Id("router"), subCall)
			group.Return(jen.Id("router"))
		})
		out.Line()
		out.Comment("Same as Wrap but allows to use your own Gin instance")
		out.Func().Id("GinWrap"+ifs.Name).Params(jen.Id("wrapper").Qual(InPackagePath, ifs.Name), jen.Id("router").Qual("github.com/gin-gonic/gin", "IRoutes"), subParams).BlockFunc(func(group *jen.Group) {
			group.Id("handler").Op(":=").Id(typeName).Values(jen.Id("wrapper"), subCall)
			for _, method := range wrappedMethods {
				group.Id("router").Dot("POST").Call(jen.Lit("/"+toKebab(method.Name)), jen.Id("handler").Dot("handle"+method.Name))
			}
		})

		if !params.DisableSwagger {
			sw := generateSwaggerDefinition(f, ifs, wrappedMethods)
			v, err := yaml.Marshal(sw)
			if err != nil {
				return GenerateResult{}, err
			}
			result.Swaggers[ifs.Name] = string(v)
		}
	}

	buffer := &bytes.Buffer{}
	err = out.Render(buffer)
	if err != nil {
		return GenerateResult{}, err
	}
	result.Wrapper = buffer.String()
	return result, nil
}

func toKebab(v string) string {
	return strings.Replace(snaker.CamelToSnake(v), "_", "-", -1)
}

var renders = []renderHandler{
	&defaultRender{},
}

func buildSignature(m *atool.Method, file *atool.File) jen.Code {
	var params []jen.Code

	for _, in := range m.In {
		params = append(params, jen.Id(in.Name).Add(getType(in, file)))
	}

	var out []jen.Code
	for _, mout := range m.Out {
		out = append(out, getType(mout, file))
	}
	return jen.Params(params...).Parens(jen.List(out...))
}

func getType(param *atool.Arg, file *atool.File) jen.Code {
	st, err := file.ExtractType(param.Type)
	qualType := jen.Id(param.GolangType())
	if err == nil && st.File.Import != "" {
		_, name := param.GoPkgType()
		// as-is
		qualType = jen.Qual(st.File.Import, name)
		if param.IsPointer() {
			qualType = jen.Op("*").Add(qualType)
		}
	}
	return qualType
}

func findRender(field *atool.Arg, file *atool.File) renderHandler {
	for _, r := range renders {
		if r.IsMatch(field, file) {
			return r
		}
	}
	panic("render not found for field " + field.Name) // should be never ever happen due to default render
}

type renderHandler interface {
	IsMatch(field *atool.Arg, file *atool.File) bool
	OnStructField(out *jen.File, structDefinition *jen.Group, field *atool.Arg, file *atool.File, params WrapperParams)
	OnParseField(out *jen.File, methodDefinition *jen.Group, field *atool.Arg, file *atool.File)
}

type defaultRender struct{}

func (dr *defaultRender) IsMatch(field *atool.Arg, file *atool.File) bool { return true }
func (dr *defaultRender) OnStructField(out *jen.File, structDefinition *jen.Group, param *atool.Arg, f *atool.File, params WrapperParams) {
	tagName := snaker.CamelToSnake(param.Name)
	qualType := jen.Id(param.GolangType())

	st, err := f.ExtractType(param.Type)
	if err == nil && st.File.Import != "" {
		_, name := param.GoPkgType()
		// as-is
		qualType = jen.Qual(st.File.Import, name)
		if param.IsPointer() {
			qualType = jen.Op("*").Add(qualType)
		}
	} else if err == nil && st.File.Import == "" && params.InPackagePath != "" {
		_, name := param.GoPkgType()
		qualType = jen.Qual(params.InPackagePath, name)
	}
	//TODO: think what to do if type can't be extracted (like type-alias)
	structDefinition.Id(strings.Title(param.Name)).Add(qualType).Tag(map[string]string{
		"json":  tagName,
		"form":  tagName,
		"query": tagName,
		"xml":   tagName,
	})
}

// nothing to parse - just use field from structure
func (dr *defaultRender) OnParseField(out *jen.File, methodDefinition *jen.Group, field *atool.Arg, file *atool.File) {
}

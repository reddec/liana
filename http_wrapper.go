package liana

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"github.com/knq/snaker"
	"github.com/reddec/astools"
	"github.com/reddec/liana/types"
	"go/ast"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"strings"
)

// Parameters for HTTP wrapper generator
type WrapperParams struct {
	File                string            // required, path for source file
	AdditionalImports   []string          // optional, additional imports for result render (almost never should be used)
	OutPackageName      string            // optional, package name in render (default is same as in file)
	OutPackagePath      string            // optional, package path in render (default is same as in file). If it is  a same as InPackagePath import will be not used
	InPackagePath       string            // optional, source file import path (default is a directory name of file)
	DisableSwagger      bool              // optional, if specified skip generates swagger files for all interfaces found in the file
	FilterInterfaces    []string          // optional, if specified generates only for specified interfaces
	Lock                bool              // optional, if specified expects global lockable object (global sync)
	GetOnEmptyParams    bool              // optional, if specified methods without args will be also available over GET method
	GetOnSimpleParams   bool              // optional, if specified methods that contains only simple (built-in) params will be available over GET method with query params
	UseShortNames       bool              // optional, generate swagger types names shortly without hashed package
	BasePath            string            // optional, generate swagger base path (default is '/')
	UrlName             bool              // optional, split method name to parts of url
	InterfaceAsTag      bool              // optional, add tag to swagger as interface name
	PrefixTag           map[string]string // optional, add a tag to swagger if method has prefix
	EmbeddedSwaggerURL  string            // optional, when specified the swagger will be generated, merged and included by specified url
	BypassContext       bool              // optional, when specified pass do not parse context.Context
	AuthPrefixes        []string          // optional, call auth provider for method with specified prefixes. useful only with BypassContext
	AuthType            []Auth            // optional, required for auth prefixes - type of auth
	NormalizeFieldsName bool              // optional, make first letter in fields in model to lower case
	CustomErrCode       int               // optional, custom http code for error
	PreProcessor        bool              // optional, if defined additional function will be invoked right before handler
}

// Result of generator
type GenerateResult struct {
	Wrapper  string                    // generate go code
	Swaggers map[string]*types.Swagger // if not disabled, generated swagger against interfaces names
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
	result.Swaggers = make(map[string]*types.Swagger)

	out := jen.NewFilePathName(OutPackagePath, OutPackageName)
	for _, imp := range params.AdditionalImports {
		out.ImportName(imp, "_")
	}
	out.ImportName("github.com/gin-gonic/gin", "gin")
	out.ImportAlias(params.InPackagePath, f.Package)
	out.HeaderComment("Code generated by liana. DO NOT EDIT.")

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
		if params.PreProcessor {
			out.Type().Id("Processor"+ifs.Name).Func().Params(jen.Op("*").Qual("github.com/gin-gonic/gin", "Context"), jen.Qual("context", "Context")).Bool()
		}
		var numAuthMethods int
		typeName := "handler" + ifs.Name
		var ifStructDef *jen.Group
		out.Type().Id(typeName).StructFunc(func(group *jen.Group) {
			group.Id("wrap").Qual(InPackagePath, ifs.Name)
			if params.Lock {
				group.Id("lock").Qual("sync", "Locker")
			}
			ifStructDef = group
		})
		out.Line()
		var wrappedMethods []*atool.Method
		for _, method := range ifs.Methods {
			if method.HasOutput() && len(method.NonErrorOutputs()) > 1 {
				continue
			}
			var authRequired bool
			for _, prefix := range params.AuthPrefixes {
				if strings.HasPrefix(method.Name, prefix) {
					numAuthMethods++
					authRequired = true
					break
				}
			}
			wrappedMethods = append(wrappedMethods, method)
			argType := "args" + method.Name + "Handler"
			var numParsableInArgs int
			out.Type().Id(argType).StructFunc(func(group *jen.Group) {
				for _, param := range method.In {
					if param.GolangType() == "context.Context" && params.BypassContext {
						continue
					}
					if param.GolangType() == "*gin.Context" {
						continue
					}
					findRender(param, f).OnStructField(out, group, param, f, params)
					numParsableInArgs++
				}
				if authRequired {
					for _, auth := range params.AuthType {
						numParsableInArgs += auth.AddRequestField(group)
					}
				}
			})

			if method.Comment != "" {
				out.Comment(strings.TrimSpace(method.Comment))
			}
			out.Func().Parens(jen.Id("h").Op("*").Id(typeName)).Id("handle" + method.Name).Params(jen.Id("gctx").Op("*").Qual("github.com/gin-gonic/gin", "Context")).BlockFunc(func(group *jen.Group) {
				var needsParse = numParsableInArgs > 0
				var needsBody = numParsableInArgs > 0
				if !needsParse && authRequired {
					for _, auth := range params.AuthType {
						needsParse = needsParse || auth.NeedsParse()
						needsBody = needsBody || auth.NeedsBody()
					}
				}
				if needsBody {
					group.List(jen.Id("body"), jen.Err()).Op(":=").Id("gctx").Dot("GetRawData").Call()
					group.If(jen.Err().Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
						g.Qual("log", "Println").Call(jen.Lit("["+method.Name+"]"), jen.Lit("failed read body:"), jen.Err())
						g.Id("gctx").Dot("AbortWithError").Call(jen.Qual("net/http", "StatusBadRequest"), jen.Err())
						g.Return()
					})
				}
				if needsParse {
					group.Var().Id("params").Id(argType)
					group.IfFunc(func(ifG *jen.Group) {
						ifG.Err().Op(":=").Qual("encoding/json", "Unmarshal").Call(jen.Id("body"), jen.Op("&").Id("params"))
						ifG.Err().Op("!=").Nil()
					}).BlockFunc(func(g *jen.Group) {
						g.Qual("log", "Println").Call(jen.Lit("["+method.Name+"]"), jen.Lit("failed to parse arguments:"), jen.Err())
						g.Id("gctx").Dot("AbortWithError").Call(jen.Qual("net/http", "StatusBadRequest"), jen.Err())
						g.Return()
					})
				}
				for _, param := range method.In {
					if param.GolangType() == "context.Context" && params.BypassContext {
						continue
					}
					findRender(param, f).OnParseField(out, group, param, f)
				}
				group.Id("ctx").Op(":=").Qual("context", "Background").Call()
				if authRequired {
					for _, auth := range params.AuthType {
						auth.Parse(group)
					}
					var authGroup = group.Empty()

					for i, auth := range params.AuthType {
						if i != 0 {
							authGroup = authGroup.Else()
						}
						authGroup = authGroup.IfFunc(func(ifSt *jen.Group) {
							ifSt.List(jen.Id("nctx"), jen.Id("ok")).Op(":=").Add(auth.ValidateRequest(jen.Id("h"), jen.Id("params"), jen.Id("gctx")))
							ifSt.Id("ok")
						}).BlockFunc(func(ifOk *jen.Group) {
							ifOk.Id("ctx").Op("=").Id("nctx")
						})
					}
					authGroup.Else().BlockFunc(func(ifNotAuth *jen.Group) {
						ifNotAuth.Qual("log", "Println").Call(jen.Lit("["+method.Name+"]"), jen.Lit("unauthorized request from"), jen.Id("gctx").Dot("Request").Dot("RemoteAddr"))
						ifNotAuth.Id("gctx").Dot("AbortWithStatus").Call(jen.Qual("net/http", "StatusUnauthorized"))
						ifNotAuth.Return()
					})

				}
				call := jen.Id("h").Dot("wrap").Dot(method.Name).CallFunc(func(args *jen.Group) {
					for _, inParam := range method.In {
						if inParam.GolangType() == "context.Context" && params.BypassContext {
							args.Id("ctx")
						} else if inParam.GolangType() == "*gin.Context" {
							args.Id("gctx")
						} else {
							var arg = args.Empty()
							if inParam.IsPointer() {
								arg = arg.Op("&")
							}
							arg.Id("params").Dot(strings.Title(inParam.Name))
						}
					}
				})
				if params.PreProcessor {
					group.If(jen.Op("!").Id("h").Dot("preProcessor").Call(jen.Id("gctx"), jen.Id("ctx"))).Block(jen.Return())
				}
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
							g.Id("gctx").Dot("AbortWithStatusJSON").Call(jen.Lit(params.CustomErrCode), jen.Id(errOut.Name).Dot("Error").Call())
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
		if numAuthMethods > 0 {
			for _, auth := range params.AuthType {
				out.Add(auth.GenerateInterface("Auth" + auth.Name()))
				ifStructDef.Id("auth" + auth.Name()).Id("Auth" + auth.Name())
			}
		}
		if params.PreProcessor {
			ifStructDef.Id("preProcessor").Id("Processor" + ifs.Name)
		}

		out.Line()
		// Gin wrapper
		comment := "Wrapper of " + f.Package + "." + ifs.Name + " that expose functions over simple JSON HTTP interface.\n Those methods are wrapped: "
		var parts []string
		for _, method := range wrappedMethods {
			path := toKebab(method.Name)
			if params.UrlName {
				path = strings.Replace(path, "-", "/", -1)
			}
			parts = append(parts, method.Name+" (POST /"+path+")")
		}
		comment += strings.Join(parts, ",\n ")
		out.Comment(comment)
		var subCall *jen.Group
		jen.ListFunc(func(sc *jen.Group) {
			subCall = sc
		})
		var subParams *jen.Group
		jen.ListFunc(func(sp *jen.Group) {
			subParams = sp
		})

		if params.Lock {
			subParams.Id("lock").Qual("sync", "Locker")
			subCall.Id("lock")
		}
		if numAuthMethods > 0 {
			for _, auth := range params.AuthType {
				subParams.Id("auth" + auth.Name()).Id("Auth" + auth.Name())
				subCall.Id("auth" + auth.Name())
			}
		}
		if params.PreProcessor {
			subParams.Id("preProcessor").Id("Processor" + ifs.Name)
			subCall.Id("preProcessor")
		}

		if !params.DisableSwagger {
			usn := swaggerGen{
				UseShortNames:  params.UseShortNames,
				BasePath:       params.BasePath,
				GetOnEmpty:     params.GetOnEmptyParams,
				NameURL:        params.UrlName,
				InterfaceAsTag: params.InterfaceAsTag,
				PrefixTag:      params.PrefixTag,
				EmbeddedURL:    params.EmbeddedSwaggerURL,
				BypassContext:  params.BypassContext,
				WrapperParams:  params,
			}
			sw := usn.generateSwaggerDefinition(f, ifs, wrappedMethods)
			result.Swaggers[ifs.Name] = &sw
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
				var getGenerated bool
				path := toKebab(method.Name)
				if params.UrlName {
					path = strings.Replace(path, "-", "/", -1)
				}
				var numParsableParam int
				var onlySimple = true
				for _, p := range method.In {
					if p.GolangType() == "context.Context" && params.BypassContext {
						continue
					}
					numParsableParam++
					onlySimple = p.IsSimple() && onlySimple
				}
				group.Id("router").Dot("POST").Call(jen.Lit("/"+path), jen.Id("handler").Dot("handle"+method.Name))
				if params.GetOnEmptyParams && numParsableParam == 0 {
					group.Id("router").Dot("GET").Call(jen.Lit("/"+path), jen.Id("handler").Dot("handle"+method.Name))
					getGenerated = true
				}
				if !getGenerated && params.GetOnSimpleParams && onlySimple {
					getGenerated = true
					group.Id("router").Dot("GET").Call(jen.Lit("/"+path), jen.Id("handler").Dot("handle"+method.Name))
				}
			}
			if params.EmbeddedSwaggerURL != "" {
				group.Id("router").Dot("GET").Call(jen.Lit(params.EmbeddedSwaggerURL), jen.Func().Params(jen.Id("gctx").Op("*").Qual("github.com/gin-gonic/gin", "Context")).BlockFunc(func(handler *jen.Group) {
					handler.Id("gctx").Dot("Data").Call(jen.Qual("net/http", "StatusOK"), jen.Lit("application/yaml"), jen.Index().Byte().Parens(jen.Id("Swagger"+ifs.Name)))
				}))
			}
		})

		if params.EmbeddedSwaggerURL != "" {
			var root *types.Swagger
			for _, sw := range result.Swaggers {
				if root == nil {
					root = sw
				} else {
					MergeSwagger(root, sw)
				}
			}
			data, err := yaml.Marshal(root)
			if err != nil {
				return result, err
			}
			out.Const().Id("Swagger" + ifs.Name).Op("=").Lit(string(data))
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
	qualType := jen.Id(strings.Replace(param.GolangType(), "*", "", -1))

	st, err := f.ExtractType(param.Type)
	if err == nil && st.File.Import != "" {
		_, name := param.GoPkgType()
		name = strings.Replace(name, "*", "", -1)
		// as-is
		qualType = jen.Qual(st.File.Import, name)
	} else if err == nil && st.File.Import == "" && params.InPackagePath != "" {
		_, name := param.GoPkgType()
		name = strings.Replace(name, "*", "", -1)
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

func MergeSwagger(target *types.Swagger, source *types.Swagger) {
	for url, p := range source.Paths {
		target.Paths[url] = p
	}
}

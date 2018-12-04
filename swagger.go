package liana

import (
	"github.com/knq/snaker"
	"github.com/reddec/astools"
	"github.com/reddec/liana/typemap"
	"github.com/reddec/liana/types"
	"go/ast"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type swaggerGen struct {
	UseShortNames bool
	BasePath      string
	GetOnEmpty    bool
}

func (usn *swaggerGen) generateSwaggerDefinition(file *atool.File, iface *atool.Interface, exportedMethods []*atool.Method) types.Swagger {
	var sw types.Swagger
	sw.Swagger = "2.0"
	if usn.BasePath != "" {
		sw.BasePath = usn.BasePath
	} else {
		sw.BasePath = "/"
	}
	sw.Info.Title = iface.Name
	sw.Info.Description = strings.TrimSpace(iface.Comment)
	sw.Info.Version = "1.0"

	sw.Paths = make(map[string]types.Path)
	sw.Definitions = make(map[string]*types.Definition)
	for _, method := range exportedMethods {

		var pt types.Path

		var act types.Action
		act.OperationID = method.Name
		act.Summary = strings.TrimSpace(method.Comment)
		if method.HasInput() {
			act.Consumes = append(act.Consumes, "application/json")

			act.Parameters = append(act.Parameters, types.Param{
				In:          "body",
				Name:        "request",
				Description: "Request params",
				Required:    true,
				Schema: &types.Definition{
					Ref: "#/definitions/" + method.Name + "Params",
				},
			})

			sw.Definitions[method.Name+"Params"] = usn.generateParamsDefinition(file, method, &sw)
		}
		if len(method.NonErrorOutputs()) > 0 {
			act.Produces = append(act.Produces, "application/json")
		}
		act.Responses = make(map[int]types.Response)

		if method.HasInput() {
			act.Responses[http.StatusBadRequest] = types.Response{
				Description: "Request data contains invalid symbols",
				Schema: &types.Definition{
					Type: "string",
				},
			}
		}
		if len(method.ErrorOutputs()) > 0 {
			act.Responses[http.StatusInternalServerError] = types.Response{
				Description: "Failed to process request by the handler",
				Schema: &types.Definition{
					Type: "string",
				},
			}
		}
		if len(method.NonErrorOutputs()) == 0 {
			act.Responses[http.StatusNoContent] = types.Response{
				Description: "Success",
			}
		} else {
			act.Responses[http.StatusOK] = types.Response{
				Description: "Success",
				Schema:      usn.generateTypeSchema(file, method.NonErrorOutputs()[0], &sw),
			}
		}
		if len(method.In) == 0 && usn.GetOnEmpty {
			pt.Get = &act
		} else {
			pt.Post = &act
		}
		sw.Paths["/"+toKebab(method.Name)] = pt
	}
	return sw
}

func (usn *swaggerGen) generateParamsDefinition(file *atool.File, tp *atool.Method, sw *types.Swagger) *types.Definition {
	var def types.Definition
	def.Type = "object"
	def.Properties = make(map[string]*types.Definition)
	for _, param := range tp.In {
		def.Properties[param.Name] = usn.generateTypeSchema(file, param, sw)
	}
	return &def
}

func (usn *swaggerGen) generateTypeSchema(file *atool.File, param *atool.Arg, sw *types.Swagger) *types.Definition {
	var sh types.Definition
	pkg, tpName := param.GoPkgType()
	if param.IsString() {
		sh.Type = "string"
	} else if param.IsBoolean() {
		sh.Type = "boolean"
	} else if param.IsInteger() {
		sh.Type = "integer"
	} else if param.IsFloat() {
		sh.Type = "number"
	} else if param.IsMap() {
		sh.Type = "object"
	} else if def := typemap.TypeMap("", tpName); pkg == "" && def != nil {
		typeName := def.Alias
		if typeName == "" {
			typeName = usn.hashTypeName("", tpName)
		}
		sh.Ref = "#/definitions/" + typeName
		sw.Definitions[typeName] = &def.Definition
	} else if param.IsArray() {
		sh.Type = "array"
		sh.Items = usn.generateTypeSchema(file, param.ArrayItem(), sw)
	} else if tp, err := file.ExtractType(param.Type); err == nil {
		def := typemap.TypeMap(tp.File.Import, tp.Name)

		var typeName string
		if def != nil {
			typeName = def.Alias
		}
		if typeName == "" {
			typeName = usn.hashTypeName(tp.File.Import, tp.Name)
		}

		sh.Ref = "#/definitions/" + typeName
		if _, exists := sw.Definitions[typeName]; !exists {
			if def != nil {
				sw.Definitions[typeName] = &def.Definition
			} else {
				var x *types.Definition
				sw.Definitions[typeName] = x // forward declaration to prevent infinite cycle
				sw.Definitions[typeName] = usn.generateStructDefinition(tp, sw)
			}
		}
	} else {
		sh.Type = "object"
	}
	if sh.Ref == "" {
		sh.Description = strings.TrimSpace(param.Comment)
	}
	return &sh
}

func (usn *swaggerGen) generateStructDefinition(st *atool.Struct, sw *types.Swagger) *types.Definition {
	var def types.Definition
	def.Type = "object"
	def.Properties = make(map[string]*types.Definition)
	def.Description = strings.TrimSpace(st.Comment)
	for _, f := range st.Fields {
		if ast.IsExported(f.Name) {
			rawTags, _ := strconv.Unquote(f.AsField().Tag.Value)
			tags := reflect.StructTag(rawTags)
			if tags.Get("json") != "-" { // skip json excluded tags
				def.Properties[f.Name] = usn.generateTypeSchema(st.File, f, sw)
			}
		}
	}
	return &def
}

func (usn *swaggerGen) hashTypeName(pkg, name string) string {
	if usn.UseShortNames {
		return name
	}
	typeName := strings.Replace(strings.Replace(strings.Replace(pkg+"_"+name, "/", "_", -1), "-", "_", -1), ".", "_", -1)
	typeName = snaker.SnakeToCamel(typeName)
	return strings.Replace(typeName, "*", "", -1)
}

package liana

import (
	"github.com/knq/snaker"
	"github.com/reddec/astools"
	"go/ast"
	"net/http"
	"strings"
)

type swagger struct {
	Swagger     string                `yaml:"swagger"`
	Info        info                  `yaml:"info"`
	Host        string                `yaml:"host,omitempty"`
	BasePath    string                `yaml:"basePath,omitempty"`
	Paths       map[string]path       `yaml:"paths,omitempty"`
	Definitions map[string]Definition `yaml:"definitions,omitempty"`
}

type info struct {
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version"`
	Title       string `yaml:"title"`
}

type path struct {
	Post action `yaml:"post"`
}

type action struct {
	Summary     string           `yaml:"summary,omitempty"`
	OperationID string           `yaml:"operationId"`
	Consumes    []string         `yaml:"consumes,omitempty"`
	Produces    []string         `yaml:"produces,omitempty"`
	Parameters  []param          `yaml:"parameters,omitempty"`
	Responses   map[int]response `yaml:"responses"`
}

type param struct {
	In          string `yaml:"in,omitempty"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Schema      Schema `yaml:"schema"`
}

type response struct {
	Description string `yaml:"description,omitempty"`
	Schema      Schema `yaml:"schema,omitempty"`
}

type Schema struct {
	Type        string `yaml:"type,omitempty"`
	Ref         string `yaml:"$ref,omitempty"`
	Format      string `yaml:"format,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Definition struct {
	Type        string            `yaml:"type"`
	Format      string            `yaml:"format,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Example     string            `yaml:"example,omitempty"`
	Properties  map[string]Schema `yaml:"properties,omitempty"`
}

func generateSwaggerDefinition(file *atool.File, iface *atool.Interface, exportedMethods []*atool.Method) swagger {
	var sw swagger
	sw.Swagger = "2.0"
	sw.BasePath = "/"
	sw.Info.Title = iface.Name
	sw.Info.Description = strings.TrimSpace(iface.Comment)
	sw.Info.Version = "1.0"

	sw.Paths = make(map[string]path)
	sw.Definitions = make(map[string]Definition)
	for _, method := range exportedMethods {

		var pt path

		var act action
		act.OperationID = method.Name
		act.Summary = strings.TrimSpace(method.Comment)
		if method.HasInput() {
			act.Consumes = append(act.Consumes, "application/json")

			act.Parameters = append(act.Parameters, param{
				In:          "body",
				Name:        "request",
				Description: "Request params",
				Required:    true,
				Schema: Schema{
					Ref: "#/definitions/" + method.Name + "Params",
				},
			})

			sw.Definitions[method.Name+"Params"] = generateParamsDefinition(file, method, &sw)
		}
		if len(method.NonErrorOutputs()) > 0 {
			act.Produces = append(act.Produces, "application/json")
		}
		act.Responses = make(map[int]response)

		if method.HasInput() {
			act.Responses[http.StatusBadRequest] = response{
				Description: "Request data contains invalid symbols",
				Schema: Schema{
					Type: "string",
				},
			}
		}
		if len(method.ErrorOutputs()) > 0 {
			act.Responses[http.StatusInternalServerError] = response{
				Description: "Failed to process request by the handler",
				Schema: Schema{
					Type: "string",
				},
			}
		}
		if len(method.NonErrorOutputs()) == 0 {
			act.Responses[http.StatusNoContent] = response{
				Description: "Success",
			}
		} else {
			act.Responses[http.StatusOK] = response{
				Description: "Success",
				Schema:      generateTypeSchema(file, method.NonErrorOutputs()[0], &sw),
			}
		}

		pt.Post = act
		sw.Paths["/"+toKebab(method.Name)] = pt
	}
	return sw
}

func generateParamsDefinition(file *atool.File, tp *atool.Method, sw *swagger) Definition {
	var def Definition
	def.Type = "object"
	def.Properties = make(map[string]Schema)
	for _, param := range tp.In {
		def.Properties[param.Name] = generateTypeSchema(file, param, sw)
	}
	return def
}

func generateTypeSchema(file *atool.File, param *atool.Arg, sw *swagger) Schema {
	var sh Schema
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
	} else if def, ok := queryHook(file, param); ok {
		sh.Ref = "#/definitions/" + def.Alias
		sw.Definitions[def.Alias] = def.Definition
	} else if param.IsArray() {
		sh.Type = "array"
	} else if tp, err := file.ExtractType(param.Type); err == nil {
		typeName := strings.Replace(strings.Replace(strings.Replace(tp.File.Import+"_"+tp.Name, "/", "_", -1), "-", "_", -1), ".", "_", -1)
		typeName = snaker.SnakeToCamel(typeName)
		sh.Ref = "#/definitions/" + typeName
		sw.Definitions[typeName] = generateStructDefinition(tp, sw)
	} else {
		sh.Type = "object"
	}
	if sh.Ref == "" {
		sh.Description = strings.TrimSpace(param.Comment)
	}
	return sh
}

func generateStructDefinition(st *atool.Struct, sw *swagger) Definition {
	var def Definition
	def.Type = "object"
	def.Properties = make(map[string]Schema)
	def.Description = strings.TrimSpace(st.Comment)
	for _, f := range st.Fields {
		if ast.IsExported(f.Name) {
			def.Properties[f.Name] = generateTypeSchema(st.File, f, sw)

		}
	}
	return def
}

type Alias struct {
	Definition
	Alias string
}

var globalTypeHooks = make(map[string]Alias)

// Add swagger definition for custom type
func AddSwaggerType(importName, typeName string, definition Alias) {
	fqdn := importName + "." + typeName
	globalTypeHooks[fqdn] = definition
}

func queryHook(file *atool.File, arg *atool.Arg) (Alias, bool) {
	pkg, name := arg.GoPkgType()
	fqdn := "." + arg.GolangType()
	if pkg != "" {
		tp, err := file.ExtractType(arg.Type)
		if err != nil {
			return Alias{}, false
		}
		fqdn = tp.File.Import + "." + name
	}
	def, ok := globalTypeHooks[fqdn]
	return def, ok
}

func init() {
	AddSwaggerType("time", "Duration", Alias{
		Definition: Definition{
			Type:        "string",
			Description: "duration time with suffixes (ms, s, m, h)",
			Example:     "3s",
		},
		Alias: "Duration",
	})

	AddSwaggerType("time", "Time", Alias{
		Definition: Definition{
			Type:        "string",
			Description: "RFC3339 time with optional nanoseconds and timezone",
			Example:     "2018-10-15T21:59:13.915939243+08:00",
		},
		Alias: "RFC3339",
	})

	AddSwaggerType("github.com/shopspring/decimal", "Decimal", Alias{
		Definition: Definition{
			Type:        "string",
			Description: "decimal number with up to 254 symbols after floating point",
			Example:     "123.456",
		},
		Alias: "Decimal",
	})
	AddSwaggerType("", "[]byte", Alias{
		Definition: Definition{
			Type:        "string",
			Format:      "base64",
			Description: "Base64 encoded byte array",
			Example:     "U3dhZ2dlciByb2Nrcw==",
		},
		Alias: "Base64",
	})

	AddSwaggerType("database/sql", "NullInt64", Alias{
		Definition: Definition{
			Type:        "integer",
			Description: "optional integer",
		},
		Alias: "NullInt64",
	})
	AddSwaggerType("database/sql", "NullString", Alias{
		Definition: Definition{
			Type:        "string",
			Description: "optional string",
		},
		Alias: "NullString",
	})

	AddSwaggerType("database/sql", "NullFloat64", Alias{
		Definition: Definition{
			Type:        "number",
			Description: "optional number",
		},
		Alias: "NullFloat64",
	})

	AddSwaggerType("database/sql", "NullBool", Alias{
		Definition: Definition{
			Type:        "boolean",
			Description: "optional boolean",
		},
		Alias: "NullBool",
	})

	AddSwaggerType("github.com/lib/pq", "NullTime", Alias{
		Definition: Definition{
			Type:        "string",
			Description: "RFC3339 time with optional nanoseconds and timezone",
			Example:     "2018-10-15T21:59:13.915939243+08:00",
		},
		Alias: "NullTime",
	})
}

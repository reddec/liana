package main

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"github.com/dave/jennifer/jen"
	"github.com/fatih/camelcase"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/reddec/liana/abu"
	"github.com/reddec/liana/abu/utils"
	"github.com/reddec/symbols"
	"gopkg.in/yaml.v2"

	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

var config struct {
	List  List  `command:"list"  description:"generate page for tables"`
	Page  Page  `command:"page"  description:"generate page for single item"`
	Form  Form  `command:"form"  description:"generate form for single item"`
	CSS   CSS   `command:"css"   description:"make css static handler"`
	Batch Batch `command:"batch" description:"execute batch commands"`
}

func main() {
	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}
}

type Common struct {
	Title           string   `long:"title" env:"TITLE" description:"Title of page" default:"" yaml:",omitempty"`
	Description     string   `long:"description" env:"Description" description:"Description on page" default:"" yaml:",omitempty"`
	Handler         string   `long:"handler" env:"HANDLER" description:"Handler custom name"  yaml:",omitempty"`
	Type            string   `long:"type" env:"TYPE" description:"Type name of item (should be imported in a current package)" required:"yes" yaml:",omitempty"`
	Fields          []string `long:"field" short:"f" env:"FIELD" env-delim:"," description:"Fields to include. If set only this fields will be used otherwise - everything. Conflicts with EXCLUDE parameter" yaml:",omitempty"`
	Exclude         []string `long:"exclude" short:"e" env:"EXCLUDE" env-delim:"," description:"Exclude fields. If set then all fields will be used except specified, otherwise - everything. Conflicts with FIELDS parameter" yaml:",omitempty"`
	SymbolScanLimit int      `long:"symbol-scan-limit" env:"SYMBOL_SCAN_LIMIT" description:"Limit to scan for an imports" default:"-1" yaml:",omitempty"`
	Sample          bool     `long:"sample" env:"SAMPLE" description:"Dump config as sample for batch" yaml:",omitempty"`
	// ui features
	BootstrapURL   string `long:"bootstrap-url" short:"B" env:"BOOTSTRAP_URL" description:"Bootstrap link for CSS" default:"https://bootswatch.com/4/superhero/bootstrap.min.css" yaml:",omitempty"`
	TemplatePath   string `long:"template" env:"TEMPLATE" description:"Custom template path. If not set - used default" yaml:",omitempty"`
	ExportTemplate bool   `long:"export-template" env:"EXPORT_TEMPLATE" description:"Export template"  yaml:",omitempty"`

	Menu       map[string]string `long:"menu" short:"m" env:"MENU" env-delim:"," description:"Top menu map (name is title, value is link)" yaml:",omitempty"`
	Active     string            `long:"active" short:"a" env:"ACTIVE" description:"Active title" yaml:",omitempty"`
	Package    string            `long:"package" env:"PACKAGE" description:"Package name (default is current)" yaml:",omitempty"`
	Positional struct {
		RootDir string `positional-arg-name:"directory" default:"." description:"GoLang files locations" yaml:",omitempty"`
	} `positional-args:"yes" yaml:",omitempty"`

	out *jen.File
}

type commonParams struct {
	Types     []*symbols.Symbol
	Fields    []string
	Titles    []string
	RawFields []*symbols.Field
	Templ     *template.Template
	Sym       *symbols.Symbol
	Project   *symbols.Project
	file      *jen.File
}

var parsedCache = make(map[string]*symbols.Project)

func (c *Common) prepare(args []string, defaultTemplate string, shims []*Shim) (*commonParams, error) {
	if len(c.Fields) > 0 && len(c.Exclude) > 0 {
		return nil, errors.New("fields and exclude parameter are conflicted")
	}
	var blackList = make(map[string]bool)
	var whiteList = make(map[string]bool)
	for _, f := range c.Fields {
		whiteList[f] = true
	}
	for _, f := range c.Exclude {
		blackList[f] = true
	}

	proj, ok := parsedCache[c.Positional.RootDir]
	if !ok {
		prj, err := symbols.ProjectByDir(c.Positional.RootDir, c.SymbolScanLimit)
		if err != nil {
			return nil, err
		}
		proj = prj
		parsedCache[c.Positional.RootDir] = proj
	}

	sym, err := proj.FindLocalSymbol(c.Type)
	if err != nil {
		return nil, err
	}

	fields, err := sym.Fields(proj)
	if err != nil {
		return nil, err
	}

	var (
		fieldsRender []string
		titleRender  []string
		fieldTypes   []*symbols.Symbol
		rawFields    []*symbols.Field
	)
	for _, f := range fields {
		if len(whiteList) > 0 {
			// only selected
			if !whiteList[f.Name] {
				continue
			}
		} else if len(blackList) > 0 {
			// all except blocked
			if blackList[f.Name] {
				continue
			}
		}
		fieldsRender = append(fieldsRender, f.Name)
		titleRender = append(titleRender, strings.Join(camelcase.Split(f.Name), " "))
		fieldTypes = append(fieldTypes, f.Type)
		rawFields = append(rawFields, f)
	}
	if len(fieldsRender) == 0 {
		return nil, errors.New("no fields to render")
	}
	var out *jen.File
	if c.Package == "" {
		out = jen.NewFilePathName(proj.Package.Import, proj.Package.Package)
	} else {
		out = jen.NewFile(c.Package)
	}
	funcs := sprig.TxtFuncMap()
	funcs["gtpl"] = func(txt string) string { return "{{" + txt + "}}" }
	funcs["unref"] = func(rf interface{}) interface{} {
		if rf == nil {
			return nil
		}
		v := reflect.ValueOf(rf)
		if v.IsNil() {
			return nil
		}
		if v.Type().Kind() == reflect.Ptr {
			return v.Elem().Interface()
		}
		return rf
	}
	funcs["isType"] = func(sm *symbols.Symbol, importPackage, name string) bool {
		if sm.BuiltIn {
			return false
		}
		if sm.Name != name {
			return false
		}
		return sm.Import.Import == importPackage
	}
	funcs["isByteArray"] = func(fieldIndex int) bool {
		return utils.IsByteArray(rawFields[fieldIndex].Raw.Type)
	}
	funcs["shim"] = func(fieldIndex int) string {
		sm := fieldTypes[fieldIndex]
		if sm.BuiltIn {
			return ""
		}
		for _, sh := range shims {
			if sm.Name == sh.Type && sm.Import.Package == sh.Package {
				return "{{with ." + fieldsRender[fieldIndex] + "}}" + sh.Render + "{{end}}"
			}
		}
		return ""
	}

	var templ = template.New("").Funcs(funcs)
	if c.TemplatePath != "" {
		data, err := ioutil.ReadFile(c.TemplatePath)
		if err != nil {
			return nil, err
		}
		templ, err = templ.Parse(string(data))
		if err != nil {
			return nil, err
		}
	} else {
		t, err := templ.Parse(defaultTemplate)
		if err != nil {
			return nil, err
		}
		templ = t
	}
	c.out = out
	return &commonParams{
		Fields:    fieldsRender,
		Titles:    titleRender,
		Types:     fieldTypes,
		RawFields: rawFields,
		Project:   proj,
		Sym:       sym,
		Templ:     templ,
		file:      out,
	}, nil
}

type List struct {
	MaxLimit     int    `long:"max-limit" env:"MAX_LIMIT" description:"Maximum value of limit" default:"50" yaml:",omitempty"`
	DefaultLimit int    `long:"default-limit" env:"DEFAULT_LIMIT" description:"Default limit" default:"20" yaml:",omitempty"`
	Query        string `long:"query" env:"QUERY" description:"Query placeholder. If not specified - not supported" yaml:",omitempty"`
	ItemLink     string `long:"item-link" env:"ITEM_LINK" description:"Link for item. Supports GoTemplate as root of provied item" yaml:"itemLink,omitempty"`
	Common       `yaml:",inline"`
}

func (l *List) Execute(args []string) error {
	if l.Sample {
		dec := yaml.NewEncoder(os.Stdout)
		defer dec.Close()
		return dec.Encode([]BatchItem{{List: l}})
	}
	code, err := l.execute(args, nil)
	if err != nil {
		return err
	}
	l.out.Add(code)
	return l.out.Render(os.Stdout)
}

func (l *List) execute(args []string, shims []*Shim) (jen.Code, error) {
	params, err := l.prepare(args, string(abu.MustAsset("templates/table.gotemplate")), shims)
	if err != nil {
		return nil, err
	}

	renderParams := renderListParams{
		commonParams: *params,
		Params:       *l,
	}
	preRender := &bytes.Buffer{}
	// render template
	err = params.Templ.Execute(preRender, renderParams)
	if err != nil {
		return nil, err
	}

	if l.ExportTemplate {
		_, err = os.Stdout.Write(preRender.Bytes())
		return nil, err
	}
	name := "HandlerList" + params.Sym.Name
	if l.Handler != "" {
		name = l.Handler
	}
	return createListHandler(name, params.Sym, preRender.String(), l.MaxLimit, l.DefaultLimit, l.Query != "")
}

type renderListParams struct {
	Params List
	commonParams
}

func createListHandler(name string, sym *symbols.Symbol, preRender string, maxLimit, defaultLimit int, query bool) (jen.Code, error) {
	handlerFuncType := jen.Func().Params(jen.Id("rw").Qual("net/http", "ResponseWriter"), jen.Id("rq").Op("*").Qual("net/http", "Request"))
	inType := jen.Func().ParamsFunc(func(params *jen.Group) {
		params.Id("offset").Int64()
		params.Id("limit").Int64()
		if query {
			params.Id("query").String()
		}
	}).Params(jen.Index().Op("*").Qual(sym.Import.Import, sym.Name), jen.Error())
	return jen.Func().Id(name).Params(jen.Id("provider").Add(inType)).Params(handlerFuncType).BlockFunc(func(group *jen.Group) {
		group.Const().Id("templateData").Op("=").Lit(preRender)
		group.Const().Id("maxLimit").Op("=").Lit(maxLimit)
		group.Const().Id("defaultLimit").Op("=").Lit(defaultLimit)
		group.Id("funcMap").Op(":=").Map(jen.String()).Interface().Values()
		group.Id("funcMap").Index(jen.Lit("b64")).Op("=").Func().Params(jen.Id("data").Index().Byte()).String().BlockFunc(func(converter *jen.Group) {
			converter.Return(jen.Qual("encoding/base64", "StdEncoding").Dot("EncodeToString").Call(jen.Id("data")))
		})
		group.List(jen.Id("tpl"), jen.Err()).Op(":=").Qual("html/template", "New").Call(jen.Lit("")).Dot("Funcs").Call(jen.Id("funcMap")).Dot("Parse").Call(jen.Id("templateData"))
		group.If(jen.Err().Op("!=").Nil()).Block(jen.Panic(jen.Err()))
		group.Type().Id("params").StructFunc(func(strct *jen.Group) {
			strct.Id("Limit").Int64()
			strct.Id("Offset").Int64()
			strct.Id("Next").Int64()
			strct.Id("Prev").Int64()
			strct.Id("Num").Int64()
			strct.Id("Data").Index().Op("*").Qual(sym.Import.Import, sym.Name)
			if query {
				strct.Id("Query").String()
			}
		})
		group.Return(jen.Add(handlerFuncType).BlockFunc(func(handler *jen.Group) {
			handler.Defer().Id("rq").Dot("Body").Dot("Close").Call()
			if query {
				handler.Id("query").Op(":=").Id("rq").Dot("FormValue").Call(jen.Lit("query"))
			}

			handler.Id("offTxt").Op(":=").Id("rq").Dot("FormValue").Call(jen.Lit("offset"))
			handler.List(jen.Id("offset"), jen.Err()).Op(":=").Qual("strconv", "ParseInt").Call(jen.Id("offTxt"), jen.Lit(10), jen.Lit(64))
			handler.If(jen.Err().Op("!=").Nil()).Block(jen.Id("offset").Op("=").Lit(0))

			handler.Id("limitTxt").Op(":=").Id("rq").Dot("FormValue").Call(jen.Lit("limit"))
			handler.List(jen.Id("limit"), jen.Err()).Op(":=").Qual("strconv", "ParseInt").Call(jen.Id("limitTxt"), jen.Lit(10), jen.Lit(64))
			handler.If(jen.Err().Op("!=").Nil()).Block(jen.Id("limit").Op("=").Id("defaultLimit"))

			handler.If(jen.Id("offset").Op("<").Lit(0)).Block(jen.Id("offset").Op("=").Lit(0))
			handler.If(jen.Id("limit").Op("<").Lit(0)).Block(jen.Id("limit").Op("=").Id("defaultLimit"))
			handler.If(jen.Id("limit").Op(">").Id("maxLimit")).Block(jen.Id("limit").Op("=").Id("maxLimit"))

			handler.Id("next").Op(":=").Id("offset").Op("+").Id("limit")
			handler.Id("prev").Op(":=").Id("offset").Op("-").Id("limit")
			handler.If(jen.Id("prev").Op("<").Lit(0)).Block(jen.Id("prev").Op("=").Lit(0))

			handler.List(jen.Id("data"), jen.Err()).Op(":=").Id("provider").CallFunc(func(call *jen.Group) {
				call.Id("offset")
				call.Id("limit")
				if query {
					call.Id("query")
				}
			})
			handler.If(jen.Err().Op("!=").Nil()).BlockFunc(func(errGroup *jen.Group) {
				errGroup.Qual("log", "Println").Call(jen.Lit("["+sym.Name+"-list]"), jen.Err())
				errGroup.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Err().Dot("Error").Call(), jen.Qual("net/http", "StatusBadGateway"))
				errGroup.Return()
			})
			handler.Id("num").Op(":=").Int64().Call(jen.Len(jen.Id("data")))

			handler.Id("rw").Dot("Header").Call().Dot("Set").Call(jen.Lit("Content-Type"), jen.Lit("text/html"))
			handler.Id("rw").Dot("WriteHeader").Call(jen.Qual("net/http", "StatusOK"))

			handler.Id("tpl").Dot("Execute").CallFunc(func(call *jen.Group) {
				call.Id("rw")
				call.Op("&").Id("params").ValuesFunc(func(values *jen.Group) {
					values.Id("limit")
					values.Id("offset")
					values.Id("next")
					values.Id("prev")
					values.Id("num")
					values.Id("data")
					if query {
						values.Id("query")
					}
				})
			})
		}))
	}), nil
}

type Page struct {
	Common   `yaml:",inline"`
	KeyType  string            `long:"key-type" env:"KEY_TYPE" description:"Key type" choice:"string" choice:"int64" default:"string" yaml:"keyType"`
	Pattern  string            `long:"pattern" env:"PATTERN" description:"Regexp pattern to extract key from URL" default:"([^/]+)$"`
	ItemLink map[string]string `long:"item-link" env:"ITEM_LINK" description:"Link for item (field name => link template). Supports GoTemplate as root of provided item" yaml:"itemLink,omitempty"`
}

type renderPageParams struct {
	commonParams
	Params Page
}

func (l *Page) Execute(args []string) error {
	if l.Sample {
		dec := yaml.NewEncoder(os.Stdout)
		defer dec.Close()
		return dec.Encode([]BatchItem{{Page: l}})
	}
	code, err := l.execute(args, nil)
	if err != nil {
		return err
	}
	l.out.Add(code)
	return l.out.Render(os.Stdout)
}

func (l *Page) execute(args []string, shims []*Shim) (jen.Code, error) {
	params, err := l.prepare(args, string(abu.MustAsset("templates/page.gotemplate")), shims)
	if err != nil {
		return nil, err
	}
	renderParams := renderPageParams{
		commonParams: *params,
		Params:       *l,
	}
	preRender := &bytes.Buffer{}
	// render template
	err = params.Templ.Execute(preRender, renderParams)
	if err != nil {
		return nil, err
	}

	if l.ExportTemplate {
		_, err = os.Stdout.Write(preRender.Bytes())
		return nil, err
	}
	name := "HandlerPage" + params.Sym.Name
	if l.Handler != "" {
		name = l.Handler
	}
	return createPageHandler(name, params.Sym, preRender.String(), l.KeyType, l.Pattern)
}

func createPageHandler(name string, sym *symbols.Symbol, preRender string, keyType string, pattern string) (jen.Code, error) {
	parser, err := keyParser(keyType, sym.Name+"-page")
	if err != nil {
		return nil, err
	}
	handlerFuncType := jen.Func().Params(jen.Id("rw").Qual("net/http", "ResponseWriter"), jen.Id("rq").Op("*").Qual("net/http", "Request"))
	inType := jen.Func().Params(jen.Id("key").Id(keyType)).Params(jen.Op("*").Qual(sym.Import.Import, sym.Name), jen.Error())
	return jen.Func().Id(name).Params(jen.Id("provider").Add(inType)).Params(handlerFuncType).BlockFunc(func(group *jen.Group) {
		group.Const().Id("templateData").Op("=").Lit(preRender)
		group.Var().Id("pattern").Op("=").Qual("regexp", "MustCompile").Call(jen.Lit(pattern))
		group.Id("funcMap").Op(":=").Map(jen.String()).Interface().Values()
		group.Id("funcMap").Index(jen.Lit("b64")).Op("=").Func().Params(jen.Id("data").Index().Byte()).String().BlockFunc(func(converter *jen.Group) {
			converter.Return(jen.Qual("encoding/base64", "StdEncoding").Dot("EncodeToString").Call(jen.Id("data")))
		})
		group.List(jen.Id("tpl"), jen.Err()).Op(":=").Qual("html/template", "New").Call(jen.Lit("")).Dot("Funcs").Call(jen.Id("funcMap")).Dot("Parse").Call(jen.Id("templateData"))
		group.If(jen.Err().Op("!=").Nil()).Block(jen.Panic(jen.Err()))
		group.Type().Id("params").StructFunc(func(strct *jen.Group) {
			strct.Id("Key").Id(keyType)
			strct.Id("Data").Op("*").Qual(sym.Import.Import, sym.Name)
		})
		group.Return(jen.Add(handlerFuncType).BlockFunc(func(handler *jen.Group) {
			handler.Defer().Id("rq").Dot("Body").Dot("Close").Call()
			handler.Id("matches").Op(":=").Id("pattern").Dot("FindStringSubmatch").Call(jen.Id("rq").Dot("URL").Dot("Path"))
			handler.If(jen.Len(jen.Id("matches")).Op("<").Lit(2)).BlockFunc(func(errGroup *jen.Group) {
				errGroup.Qual("log", "Println").Call(jen.Lit("[" + sym.Name + "-page] no params"))
				errGroup.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Lit(" no params"), jen.Qual("net/http", "StatusNotFound"))
				errGroup.Return()
			})

			handler.Id("param").Op(":=").Id("matches").Index(jen.Lit(1))
			handler.If(jen.Id("param").Op("==").Lit("")).BlockFunc(func(errGroup *jen.Group) {
				errGroup.Qual("log", "Println").Call(jen.Lit("[" + sym.Name + "-page] empty param"))
				errGroup.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Lit("empty param"), jen.Qual("net/http", "StatusNotFound"))
				errGroup.Return()
			})

			handler.Add(parser)

			handler.List(jen.Id("data"), jen.Err()).Op(":=").Id("provider").Call(jen.Id("key"))
			handler.If(jen.Err().Op("!=").Nil()).BlockFunc(func(errGroup *jen.Group) {
				errGroup.Qual("log", "Println").Call(jen.Lit("["+sym.Name+"-page]"), jen.Err())
				errGroup.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Err().Dot("Error").Call(), jen.Qual("net/http", "StatusBadGateway"))
				errGroup.Return()
			}).Else().If(jen.Id("data").Op("==").Nil()).BlockFunc(func(errGroup *jen.Group) {
				errGroup.Qual("log", "Println").Call(jen.Lit("[" + sym.Name + "-page] no data"))
				errGroup.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Lit(" no data"), jen.Qual("net/http", "StatusNotFound"))
				errGroup.Return()
			})

			handler.Id("rw").Dot("Header").Call().Dot("Set").Call(jen.Lit("Content-Type"), jen.Lit("text/html"))
			handler.Id("rw").Dot("WriteHeader").Call(jen.Qual("net/http", "StatusOK"))
			handler.Id("tpl").Dot("Execute").Call(jen.Id("rw"), jen.Op("&").Id("params").Values(jen.Id("key"), jen.Id("data")))
		}))
	}), nil
}

func keyParser(keyType string, errCaption string) (jen.Code, error) {
	var parser jen.Code
	switch keyType {
	case "string":
		parser = jen.Id("key").Op(":=").Id("param")
	case "int64":
		parser = jen.List(jen.Id("key"), jen.Err()).Op(":=").Qual("strconv", "ParseInt").Call(jen.Id("param"), jen.Lit(10), jen.Lit(64)).Line().If(jen.Err().Op("!=").Nil()).BlockFunc(func(group *jen.Group) {
			group.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Err())
			group.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Err().Dot("Error").Call(), jen.Qual("net/http", "StatusBadRequest"))
			group.Return()
		})
	default:
		return nil, errors.New("unknown key type " + keyType)
	}
	return parser, nil
}

type Form struct {
	Common         `yaml:",inline"`
	Redirect       string `long:"redirect" env:"REDIRECT" description:"To what page redirect after success"`
	SuccessMessage string `long:"success-message" env:"SUCCESS_MESSAGE" description:"Success message" default:"Success"`
}
type renderFormParams struct {
	commonParams
	Params Form
}

func (l *Form) Execute(args []string) error {
	if l.Sample {
		dec := yaml.NewEncoder(os.Stdout)
		defer dec.Close()
		return dec.Encode([]BatchItem{{Form: l}})
	}
	code, err := l.execute(args, nil)
	if err != nil {
		return err
	}
	l.out.Add(code)
	return l.out.Render(os.Stdout)
}

func (f *Form) execute(args []string, shims []*Shim) (jen.Code, error) {
	params, err := f.prepare(args, string(abu.MustAsset("templates/form.gotemplate")), shims)
	if err != nil {
		return nil, err
	}
	// check that only simple type
	for idx, fieldType := range params.Types {
		name := params.Fields[idx]
		if fieldType.Is("time", "Time") || fieldType.Is("time", "Duration") {
			continue
		}
		if !fieldType.BuiltIn {
			return nil, errors.Errorf("field %v is not built-in type. Custom types except time.Time and time.Duration are not yet supported", name)
		}
		if fieldType.IsStructDefinition() {
			return nil, errors.Errorf("field %v: struct as field type not yet supported", name)
		}
		if fieldType.IsMap() {
			return nil, errors.Errorf("field %v: map as field type not yet supported", name)
		}
		if fieldType.Name == "error" {
			return nil, errors.Errorf("field %v: error as field type not yet supported", name)
		}
		rf := params.RawFields[idx]
		if symbols.IsArray(rf.Raw.Type) && !utils.IsByteArray(rf.Raw.Type) {
			return nil, errors.Errorf("field %v: array as field type not yet supported", name)
		}
	}
	renderParams := renderFormParams{
		commonParams: *params,
		Params:       *f,
	}
	preRender := &bytes.Buffer{}
	// render template
	err = params.Templ.Execute(preRender, renderParams)
	if err != nil {
		return nil, err
	}

	if f.ExportTemplate {
		_, err = os.Stdout.Write(preRender.Bytes())
		return nil, err
	}
	name := "HandlerForm" + params.Sym.Name
	if f.Handler != "" {
		name = f.Handler
	}
	return createFormHandler(name, preRender.String(), params.Sym, params.Fields, params.Types, params.RawFields, f.Redirect, f.SuccessMessage), nil
}

func createFormHandler(name string, preRender string, sym *symbols.Symbol, fieldNames []string, fieldTypes []*symbols.Symbol, fields []*symbols.Field, redirect string, successMsg string) jen.Code {
	handlerFuncType := jen.Func().Params(jen.Id("rw").Qual("net/http", "ResponseWriter"), jen.Id("rq").Op("*").Qual("net/http", "Request"))
	providerType := jen.Func().ParamsFunc(func(fn *jen.Group) {
		fn.Id("item").Op("*").Qual(sym.Import.Import, sym.Name)
	}).Error()
	return jen.Func().Id(name).Params(jen.Id("provider").Add(providerType)).Params(handlerFuncType).BlockFunc(func(group *jen.Group) {
		group.Const().Id("templateData").Op("=").Lit(preRender)
		group.List(jen.Id("tpl"), jen.Err()).Op(":=").Qual("html/template", "New").Call(jen.Lit("")).Dot("Parse").Call(jen.Id("templateData"))
		group.If(jen.Err().Op("!=").Nil()).Block(jen.Panic(jen.Err()))
		if redirect != "" {
			group.List(jen.Id("redirectTpl"), jen.Err()).Op(":=").Qual("text/template", "New").Call(jen.Lit("")).Dot("Parse").Call(jen.Lit(redirect))
		}
		group.Type().Id("params").StructFunc(func(strct *jen.Group) {
			strct.Id("Error").Error()
			strct.Id("Success").String()
			strct.Id("Data").Op("*").Qual(sym.Import.Import, sym.Name)
		})
		group.Return(jen.Add(handlerFuncType).BlockFunc(func(handler *jen.Group) {
			handler.Defer().Id("rq").Dot("Body").Dot("Close").Call()
			handler.Var().Id("renderParams").Id("params")
			handler.Var().Id("item").Qual(sym.Import.Import, sym.Name)
			handler.Var().Id("status").Op("=").Qual("net/http", "StatusOK")
			handler.Var().Id("err").Error()
			// parser

			handler.If(jen.Id("rq").Dot("Method").Op("==").Qual("net/http", "MethodPost")).BlockFunc(func(parser *jen.Group) {
				parser.Add(utils.FormParser(fieldNames, fieldTypes, fields, sym.Name+"-form"))
				parser.Err().Op("=").Id("provider").Call(jen.Op("&").Id("item"))
				parser.If(jen.Err().Op("!=").Nil()).BlockFunc(func(failed *jen.Group) {
					failed.Id("renderParams").Dot("Error").Op("=").Err()
				}).Else().BlockFunc(func(success *jen.Group) {
					if redirect != "" {
						success.Var().Id("buffer").Op("=").Op("&").Qual("bytes", "Buffer").Values()
						success.Id("redirectTpl").Dot("Execute").Call(jen.Id("buffer"), jen.Id("item"))
						success.Qual("net/http", "Redirect").Call(jen.Id("rw"), jen.Id("rq"), jen.Id("buffer").Dot("String").Call(), jen.Qual("net/http", "StatusSeeOther"))
						success.Return()
					} else {
						success.Id("renderParams").Dot("Success").Op("=").Lit(successMsg)
					}
				})
			})
			handler.Id("renderParams").Dot("Data").Op("=").Op("&").Id("item")
			handler.Id("rw").Dot("Header").Call().Dot("Set").Call(jen.Lit("Content-Type"), jen.Lit("text/html"))
			handler.Id("rw").Dot("WriteHeader").Call(jen.Id("status"))
			handler.Id("tpl").Dot("Execute").Call(jen.Id("rw"), jen.Op("&").Id("renderParams"))
		}))
	})
}

type CSS struct {
	Package string `long:"package" env:"PACKAGE" description:"Package name" required:"yes"`
}

func (c *CSS) Execute([]string) error {
	var out *jen.File
	out = jen.NewFile(c.Package)
	out.Add(createCssHandler(string(abu.MustAsset("static/bootstrap.min.css"))))
	return out.Render(os.Stdout)
}

func createCssHandler(data string) jen.Code {
	handlerFuncType := jen.Func().Params(jen.Id("rw").Qual("net/http", "ResponseWriter"), jen.Id("rq").Op("*").Qual("net/http", "Request"))
	return jen.Func().Id("StaticCSS").Params().Params(handlerFuncType).BlockFunc(func(group *jen.Group) {
		group.Const().Id("css").Op("=").Lit(data)
		group.Return(jen.Add(handlerFuncType).BlockFunc(func(handler *jen.Group) {
			handler.Defer().Id("rq").Dot("Body").Dot("Close").Call()
			handler.Id("rw").Dot("Header").Call().Dot("Set").Call(jen.Lit("Content-Type"), jen.Lit("text/css"))
			handler.Id("rw").Dot("Header").Call().Dot("Set").Call(jen.Lit("Content-Length"), jen.Lit(strconv.Itoa(len(data))))
			handler.Id("rw").Dot("WriteHeader").Call(jen.Qual("net/http", "StatusOK"))
			handler.Id("rw").Dot("Write").Call(jen.Index().Byte().Call(jen.Id("css")))
		}))
	})
}

type Shim struct {
	Package string
	Type    string
	Render  string
}

type BatchItem struct {
	List *List `yaml:",omitempty"`
	Page *Page `yaml:",omitempty"`
	Form *Form `yaml:",omitempty"`
	Shim *Shim `yaml:",omitempty"`
}

type Batch struct {
	Package    string `long:"package" env:"PACKAGE" description:"Package name (default is current)" yaml:",omitempty"`
	Positional struct {
		RootDir string `positional-arg-name:"directory" default:"." description:"GoLang files locations" yaml:",omitempty"`
	} `positional-args:"yes" yaml:",omitempty"`
}

func (b *Batch) Execute(args []string) error {
	proj, err := symbols.ProjectByDir(b.Positional.RootDir, 1)
	if err != nil {
		return err
	}

	var items []BatchItem
	err = yaml.NewDecoder(os.Stdin).Decode(&items)
	if err != nil {
		return err
	}

	var out *jen.File
	if b.Package == "" {
		out = jen.NewFilePathName(proj.Package.Import, proj.Package.Package)
	} else {
		out = jen.NewFile(b.Package)
	}

	sample := "Sample usage:\n\n"
	var shims []*Shim
	for _, item := range items {
		if item.Shim != nil {
			shims = append(shims, item.Shim)
		}
	}
	for _, item := range items {
		var err error
		var code jen.Code
		if item.List != nil {
			code, err = item.List.execute(args, shims)
			sample += `http.HandleFunc("/` + strings.ToLower(item.List.Type) + "s" + `", nil) // TODO:` + "\n"
		} else if item.Page != nil {
			code, err = item.Page.execute(args, shims)
			sample += `http.HandleFunc("/` + strings.ToLower(item.Page.Type) + `/", nil) // TODO:` + "\n"
		} else if item.Form != nil {
			sample += `http.HandleFunc("/` + strings.ToLower(item.Form.Type) + "/actionname" + `", nil) // TODO:` + "\n"
			code, err = item.Form.execute(args, shims)
		}
		if err != nil {
			return err
		} else if code != nil {
			out.Add(code)
		}
	}
	err = out.Render(os.Stdout)
	if err == nil {
		os.Stderr.WriteString(sample)
	}
	return err
}

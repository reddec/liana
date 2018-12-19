package main

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"github.com/dave/jennifer/jen"
	"github.com/fatih/camelcase"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/reddec/liana/abu"
	"github.com/reddec/symbols"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

var config struct {
	List List `command:"list" description:"generate page for tables"`
	Page Page `command:"page" description:"generate page for single item"`
}

func main() {
	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}
}

type common struct {
	Title           string   `long:"title" env:"TITLE" description:"Title of page" default:""`
	Type            string   `long:"type" env:"TYPE" description:"Type name of item (should be imported in a current package)" required:"yes"`
	Fields          []string `long:"field" short:"f" env:"FIELD" env-delim:"," description:"Fields to include. If set only this fields will be used otherwise - everything. Conflicts with EXCLUDE parameter"`
	Exclude         []string `long:"exclude" short:"e" env:"EXCLUDE" env-delim:"," description:"Exclude fields. If set then all fields will be used except specified, otherwise - everything. Conflicts with FIELDS parameter"`
	SymbolScanLimit int      `long:"symbol-scan-limit" env:"SYMBOL_SCAN_LIMIT" description:"Limit to scan for an imports" default:"-1"`
	Package         string   `long:"package" env:"PACKAGE" description:"Package name (default is current)"`
	// ui features
	BootstrapURL string            `long:"bootstrap-url" env:"BOOTSTRAP_URL" description:"Bootstrap link for CSS" default:"https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css"`
	TemplatePath string            `long:"template" env:"TEMPLATE" description:"Custom template path. If not set - used default"`
	ItemLink     string            `long:"item-link" env:"ITEM_LINK" description:"Link for item. Supports GoTemplate as root of provied item"`
	Menu         map[string]string `long:"menu" short:"m" env:"MENU" env-delim:"," description:"Top menu map (name is title, value is link)"`
	Active       string            `long:"active" short:"a" env:"ACTIVE" description:"Active title"`
	Positional   struct {
		RootDir string `positional-arg-name:"directory" default:"." description:"GoLang files locations"`
	} `positional-args:"yes"`
}

type commonParams struct {
	Fields  []string
	Titles  []string
	Templ   *template.Template
	Sym     *symbols.Symbol
	Project *symbols.Project
	file    *jen.File
}

func (c *common) prepare(args []string, defaultTemplate string) (*commonParams, error) {
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

	funcs := sprig.TxtFuncMap()
	funcs["gtpl"] = func(txt string) string { return "{{" + txt + "}}" }
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
		t, err := templ.Parse(string(abu.MustAsset("templates/table.gotemplate")))
		if err != nil {
			return nil, err
		}
		templ = t
	}

	proj, err := symbols.ProjectByDir(c.Positional.RootDir, c.SymbolScanLimit)
	if err != nil {
		return nil, err
	}

	sym, err := proj.FindLocalSymbol(c.Type)
	if err != nil {
		return nil, err
	}

	fields, err := sym.FieldsNames()
	if err != nil {
		return nil, err
	}

	var (
		fieldsRender []string
		titleRender  []string
	)
	for _, f := range fields {
		if len(whiteList) > 0 {
			// only selected
			if !whiteList[f] {
				continue
			}
		} else if len(blackList) > 0 {
			// all except blocked
			if blackList[f] {
				continue
			}
		}
		fieldsRender = append(fieldsRender, f)
		titleRender = append(titleRender, strings.Join(camelcase.Split(f), " "))
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
	return &commonParams{
		Fields:  fields,
		Titles:  titleRender,
		Project: proj,
		Sym:     sym,
		Templ:   templ,
		file:    out,
	}, nil
}

type List struct {
	MaxLimit     int `long:"max-limit" env:"MAX_LIMIT" description:"Maximum value of limit" default:"50"`
	DefaultLimit int `long:"default-limit" env:"DEFAULT_LIMIT" description:"Default limit" default:"20"`
	common
}

func (l *List) Execute(args []string) error {
	params, err := l.prepare(args, string(abu.MustAsset("templates/table.gotemplate")))
	if err != nil {
		return err
	}
	renderParams := renderListParams{
		commonParams: *params,
		Params:       *l,
	}
	preRender := &bytes.Buffer{}
	// render template
	err = params.Templ.Execute(preRender, renderParams)
	if err != nil {
		return err
	}

	code, err := createListHandler(params.Sym, preRender.String(), l.MaxLimit, l.DefaultLimit)
	if err != nil {
		return err
	}
	out := params.file
	out.Add(code)
	return out.Render(os.Stdout)
}

type renderListParams struct {
	Params List
	commonParams
}

func createListHandler(sym *symbols.Symbol, preRender string, maxLimit, defaultLimit int) (jen.Code, error) {
	handlerFuncType := jen.Func().Params(jen.Id("rw").Qual("net/http", "ResponseWriter"), jen.Id("rq").Op("*").Qual("net/http", "Request"))
	inType := jen.Func().Params(jen.Id("offset").Int64(), jen.Id("limit").Int64()).Params(jen.Index().Op("*").Qual(sym.Import.Import, sym.Name), jen.Error())
	return jen.Func().Id("HandlerList" + sym.Name).Params(jen.Id("provider").Add(inType)).Params(handlerFuncType).BlockFunc(func(group *jen.Group) {
		group.Const().Id("templateData").Op("=").Lit(preRender)
		group.Const().Id("maxLimit").Op("=").Lit(maxLimit)
		group.Const().Id("defaultLimit").Op("=").Lit(defaultLimit)
		group.List(jen.Id("tpl"), jen.Err()).Op(":=").Qual("html/template", "New").Call(jen.Lit("")).Dot("Parse").Call(jen.Id("templateData"))
		group.If(jen.Err().Op("!=").Nil()).Block(jen.Panic(jen.Err()))
		group.Type().Id("params").StructFunc(func(strct *jen.Group) {
			strct.Id("Limit").Int64()
			strct.Id("Offset").Int64()
			strct.Id("Next").Int64()
			strct.Id("Prev").Int64()
			strct.Id("Num").Int64()
			strct.Id("Data").Index().Op("*").Qual(sym.Import.Import, sym.Name)
		})
		group.Return(jen.Add(handlerFuncType).BlockFunc(func(handler *jen.Group) {
			handler.Defer().Id("rq").Dot("Body").Dot("Close").Call()
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

			handler.List(jen.Id("data"), jen.Err()).Op(":=").Id("provider").Call(jen.Id("offset"), jen.Id("limit"))
			handler.If(jen.Err().Op("!=").Nil()).BlockFunc(func(errGroup *jen.Group) {
				errGroup.Qual("log", "Println").Call(jen.Lit("["+sym.Name+"-list]"), jen.Err())
				errGroup.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Err().Dot("Error").Call(), jen.Qual("net/http", "StatusBadGateway"))
				errGroup.Return()
			})
			handler.Id("num").Op(":=").Int64().Call(jen.Len(jen.Id("data")))

			handler.Id("rw").Dot("Header").Call().Dot("Set").Call(jen.Lit("Content-Type"), jen.Lit("text/html"))
			handler.Id("rw").Dot("WriteHeader").Call(jen.Qual("net/http", "StatusOK"))
			handler.Id("tpl").Dot("Execute").Call(jen.Id("rw"), jen.Op("&").Id("params").Values(jen.Id("limit"), jen.Id("offset"), jen.Id("next"), jen.Id("prev"), jen.Id("num"), jen.Id("data")))
		}))
	}), nil
}

type Page struct {
	common
	KeyType string `long:"key-type" env:"KEY_TYPE" description:"Key type" choice:"string" choice:"int64" default:"string"`
	Pattern string `long:"pattern" env:"PATTERN" description:"Regexp pattern to extract key from URL" default:"([^/]+)$"`
}

type renderPageParams struct {
	commonParams
	Params Page
}

func (l *Page) Execute(args []string) error {
	params, err := l.prepare(args, string(abu.MustAsset("templates/page.gotemplate")))
	if err != nil {
		return err
	}
	renderParams := renderPageParams{
		commonParams: *params,
		Params:       *l,
	}
	preRender := &bytes.Buffer{}
	// render template
	err = params.Templ.Execute(preRender, renderParams)
	if err != nil {
		return err
	}

	code, err := createPageHandler(params.Sym, preRender.String(), l.KeyType, l.Pattern)
	if err != nil {
		return err
	}
	out := params.file
	out.Add(code)
	return out.Render(os.Stdout)
}

func createPageHandler(sym *symbols.Symbol, preRender string, keyType string, pattern string) (jen.Code, error) {
	var parser jen.Code
	switch keyType {
	case "string":
		parser = jen.Id("key").Op(":=").Id("param")
	case "int64":
		parser = jen.List(jen.Id("key"), jen.Err()).Qual("strconv", "ParseInt").Call(jen.Id("param"), jen.Lit(10), jen.Lit(64)).Line().If(jen.Err().Op("!=").Nil()).BlockFunc(func(group *jen.Group) {
			group.Qual("log", "Println").Call(jen.Lit("["+sym.Name+"-page]"), jen.Err())
			group.Qual("net/http", "Error").Call(jen.Id("rw"), jen.Err().Dot("Error").Call(), jen.Qual("net/http", "StatusBadRequest"))
			group.Return()
		})
	default:
		return nil, errors.New("unknown key type")
	}
	handlerFuncType := jen.Func().Params(jen.Id("rw").Qual("net/http", "ResponseWriter"), jen.Id("rq").Op("*").Qual("net/http", "Request"))
	inType := jen.Func().Params(jen.Id("key").Id(keyType)).Params(jen.Op("*").Qual(sym.Import.Import, sym.Name), jen.Error())
	return jen.Func().Id("HandlerPage" + sym.Name).Params(jen.Id("provider").Add(inType)).Params(handlerFuncType).BlockFunc(func(group *jen.Group) {
		group.Const().Id("templateData").Op("=").Lit(preRender)
		group.Var().Id("pattern").Op("=").Qual("regexp", "MustCompile").Call(jen.Lit(pattern))
		group.List(jen.Id("tpl"), jen.Err()).Op(":=").Qual("html/template", "New").Call(jen.Lit("")).Dot("Parse").Call(jen.Id("templateData"))
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

package utils

import (
	"github.com/dave/jennifer/jen"
	"github.com/reddec/symbols"
)

func FormParser(fieldNames []string, fieldTypes []*symbols.Symbol, errCaption string) jen.Code {
	return jen.CustomFunc(jen.Options{Multi: true, Close: "\n"}, func(parser *jen.Group) {

		parser.Var().Id("errorsText").Index().String()
		for index, name := range fieldNames {
			fType := fieldTypes[index]
			switch fType.Name {
			case "int":
				parser.List(jen.Id(name), jen.Err()).Op(":=").Qual("strconv", "Atoi").Call(jen.Id("rq").Dot("FormValue").Call(jen.Lit(name)))
				parser.If(jen.Err().Op("!=").Nil()).BlockFunc(func(notParsed *jen.Group) {
					notParsed.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Lit(name), jen.Err())
					notParsed.Id("errorsText").Op("=").Append(jen.Id("errorsText"), jen.Lit(name+": ").Op("+").Err().Dot("Error").Call())
				})
			case "int8":
				parser.Add(parseFormIntField(errCaption, name, 8))
			case "int16":
				parser.Add(parseFormIntField(errCaption, name, 16))
			case "int32":
				parser.Add(parseFormIntField(errCaption, name, 32))
			case "int64":
				parser.Add(parseFormIntField(errCaption, name, 64))
			case "uint":
				parser.List(jen.Id("_"+name+"_u64"), jen.Err()).Op(":=").Qual("strconv", "ParseUint").Call(jen.Id("rq").Dot("FormValue").Call(jen.Lit(name)), jen.Lit(10), jen.Lit(64)).Line().
					If(jen.Err().Op("!=").Nil()).BlockFunc(func(notParsed *jen.Group) {
					notParsed.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Lit(name), jen.Err())
					notParsed.Id("errorsText").Op("=").Append(jen.Id("errorsText"), jen.Lit(name+": ").Op("+").Err().Dot("Error").Call())
				})
				parser.Id(name).Op(":=").Uint().Call(jen.Id("_" + name + "_u64"))
			case "uint8":
				parser.Add(parseFormUIntField(errCaption, name, 8))
			case "uint16":
				parser.Add(parseFormUIntField(errCaption, name, 16))
			case "uint32":
				parser.Add(parseFormUIntField(errCaption, name, 32))
			case "uint64":
				parser.Add(parseFormUIntField(errCaption, name, 64))
			case "float32":
				parser.Add(parseFormFloatField(errCaption, name, 32))
			case "float64":
				parser.Add(parseFormFloatField(errCaption, name, 64))
			case "bool":
				parser.Id(name).Op(":=").Id("rq").Dot("FormValue").Call(jen.Lit(name)).Op("==").Lit("on")
			case "Time":
				parser.List(jen.Id(name), jen.Err()).Op(":=").Qual("time", "Parse").Call(jen.Lit("2006-01-02T15:04"), jen.Id("rq").Dot("FormValue").Call(jen.Lit(name)))
				parser.If(jen.Err().Op("!=").Nil()).BlockFunc(func(notParsed *jen.Group) {
					notParsed.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Lit(name), jen.Err())
					notParsed.Id("errorsText").Op("=").Append(jen.Id("errorsText"), jen.Lit(name+": ").Op("+").Err().Dot("Error").Call())
				})
			case "Duration":
				parser.List(jen.Id(name), jen.Err()).Op(":=").Qual("time", "ParseDuration").Call(jen.Id("rq").Dot("FormValue").Call(jen.Lit(name)))
				parser.If(jen.Err().Op("!=").Nil()).BlockFunc(func(notParsed *jen.Group) {
					notParsed.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Lit(name), jen.Err())
					notParsed.Id("errorsText").Op("=").Append(jen.Id("errorsText"), jen.Lit(name+": ").Op("+").Err().Dot("Error").Call())
				})
			case "string":
				parser.Id(name).Op(":=").Id("rq").Dot("FormValue").Call(jen.Lit(name))
			default:
				panic("unknown type:" + fType.Name)
			}
		}
		for index, name := range fieldNames {
			fType := fieldTypes[index]
			if fType.IsPointer() {
				parser.Id("item").Dot(name).Op("=").Op("&").Id(name)
			} else {
				parser.Id("item").Dot(name).Op("=").Id(name)
			}
		}
		parser.If(jen.Len(jen.Id("errorsText")).Op(">").Lit(0)).BlockFunc(func(notParser *jen.Group) {
			notParser.Id("status").Op("=").Qual("net/http", "StatusBadRequest")
			notParser.Err().Op("=").Qual("errors", "New").Call(jen.Qual("strings", "Join").Call(jen.Id("errorsText"), jen.Lit("\n")))
		})
	})
}

func parseFormIntField(errCaption, name string, bits int) jen.Code {
	return jen.List(jen.Id(name), jen.Err()).Op(":=").Qual("strconv", "ParseInt").Call(jen.Id("rq").Dot("FormValue").Call(jen.Lit(name)), jen.Lit(10), jen.Lit(bits)).Line().
		If(jen.Err().Op("!=").Nil()).BlockFunc(func(notParsed *jen.Group) {
		notParsed.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Lit(name), jen.Err())
		notParsed.Id("errorsText").Op("=").Append(jen.Id("errorsText"), jen.Lit(name+": ").Op("+").Err().Dot("Error").Call())
	})
}

func parseFormFloatField(errCaption, name string, bits int) jen.Code {
	return jen.List(jen.Id(name), jen.Err()).Op(":=").Qual("strconv", "ParseFloat").Call(jen.Id("rq").Dot("FormValue").Call(jen.Lit(name)), jen.Lit(bits)).Line().
		If(jen.Err().Op("!=").Nil()).BlockFunc(func(notParsed *jen.Group) {
		notParsed.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Lit(name), jen.Err())
		notParsed.Id("errorsText").Op("=").Append(jen.Id("errorsText"), jen.Lit(name+": ").Op("+").Err().Dot("Error").Call())
	})
}

func parseFormUIntField(errCaption, name string, bits int) jen.Code {
	return jen.List(jen.Id(name), jen.Err()).Op(":=").Qual("strconv", "ParseUint").Call(jen.Id("rq").Dot("FormValue").Call(jen.Lit(name)), jen.Lit(10), jen.Lit(bits)).Line().
		If(jen.Err().Op("!=").Nil()).BlockFunc(func(notParsed *jen.Group) {
		notParsed.Qual("log", "Println").Call(jen.Lit("["+errCaption+"]"), jen.Err())
		notParsed.Id("errorsText").Op("=").Append(jen.Id("errorsText"), jen.Lit(name+": ").Op("+").Err().Dot("Error").Call())
	})
}

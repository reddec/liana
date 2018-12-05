package liana

import (
	"github.com/dave/jennifer/jen"
	"github.com/reddec/liana/types"
)

type AuthType int

const (
	JWT   AuthType = 0
	Token AuthType = 1
)

func (a AuthType) GenerateInterface(name string) jen.Code {
	switch a {
	case Token, JWT:
		return jen.Type().Id(name).InterfaceFunc(func(auth *jen.Group) {
			auth.Id("Validate").Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("token").String()).Parens(jen.List(jen.Qual("context", "Context"), jen.Bool()))
		})
	default:
		panic("unknown auth type")
	}
}

func (a AuthType) AddRequestField(group *jen.Group) {
	switch a {
	case JWT:
	case Token:
		group.Id("Token").String().Tag(map[string]string{"json": "token", "xml": "token", "form": "token", "query": "token"})
	default:
		panic("unknown auth type")
	}
}

func (a AuthType) ValidateRequest(self *jen.Statement, req *jen.Statement, gctx *jen.Statement) jen.Code {
	switch a {
	case JWT:
		return jen.Add(self).Dot("auth").Dot("Validate").Call(jen.Id("ctx"), jen.Add(gctx).Dot("GetHeader").Call(jen.Lit("Authorization")))
	case Token:
		return jen.Add(self).Dot("auth").Dot("Validate").Call(jen.Id("ctx"), jen.Add(req).Dot("Token"))
	default:
		panic("unknown auth type")
	}
}

func (a AuthType) SwaggerSecuirty(sw *types.Swagger) {
	switch a {
	case JWT:
		if sw.SecurityDefinitions == nil {
			sw.SecurityDefinitions = make(map[string]types.Auth)
		}
		sw.SecurityDefinitions["JWT"] = types.Auth{
			Type: "apiKey",
			In:   "header",
			Name: "Authorization",
			Description: "JWT based authorization. " +
				"Token should be presented as-is in the Authorization header",
		}

	case Token:
		return
	default:
		panic("unknown auth type")
	}
}

func (a AuthType) SwaggerSecTag() string {
	switch a {
	case JWT:
		return "JWT"

	case Token:
		return "Token"
	default:
		panic("unknown auth type")
	}
}

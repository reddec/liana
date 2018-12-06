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

type Auth interface {
	Name() string
	GenerateInterface(name string) jen.Code
	AddRequestField(group *jen.Group) int
	Parse(group *jen.Group)
	ValidateRequest(self *jen.Statement, req *jen.Statement, gctx *jen.Statement) jen.Code
	SwaggerSecurity(sw *types.Swagger)
	SwaggerSecTag() string
	NeedsParse() bool
	NeedsBody() bool
}

func (a AuthType) Name() string {
	switch a {
	case Token:
		return "Token"
	case JWT:
		return "JWT"
	default:
		panic("unknown auth type")
	}
}

func (a AuthType) NeedsParse() bool {
	switch a {
	case Token:
		return true
	default:
		return false
	}
}

func (a AuthType) NeedsBody() bool { return false }

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

func (a AuthType) AddRequestField(group *jen.Group) int {
	switch a {
	case JWT:
		return 0
	case Token:
		group.Id("Token").String().Tag(map[string]string{"json": "token", "xml": "token", "form": "token", "query": "token"})
		return 1
	default:
		panic("unknown auth type")
	}
}

func (a AuthType) Parse(group *jen.Group) {}

func (a AuthType) ValidateRequest(self *jen.Statement, req *jen.Statement, gctx *jen.Statement) jen.Code {
	switch a {
	case JWT:
		return jen.Add(self).Dot("auth"+a.Name()).Dot("Validate").Call(jen.Id("ctx"), jen.Add(gctx).Dot("GetHeader").Call(jen.Lit("Authorization")))
	case Token:
		return jen.Add(self).Dot("auth"+a.Name()).Dot("Validate").Call(jen.Id("ctx"), jen.Add(req).Dot("Token"))
	default:
		panic("unknown auth type")
	}
}

func (a AuthType) SwaggerSecurity(sw *types.Swagger) {
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

type AuthApiSignature int

func (a AuthApiSignature) Name() string {
	return "SignedToken"
}
func (a AuthApiSignature) NeedsBody() bool  { return true }
func (a AuthApiSignature) NeedsParse() bool { return false }

func (AuthApiSignature) GenerateInterface(name string) jen.Code {
	return jen.Type().Id(name).InterfaceFunc(func(auth *jen.Group) {
		auth.Id("Validate").ParamsFunc(func(g *jen.Group) {
			g.Id("ctx").Qual("context", "Context")
			g.Id("apiToken").String()
			g.Id("signature").String()
			g.Id("body").Index().Byte()
		}).Parens(jen.List(jen.Qual("context", "Context"), jen.Bool()))
	})
}

func (AuthApiSignature) AddRequestField(group *jen.Group) int { return 0 }

func (AuthApiSignature) Parse(group *jen.Group) {
	group.Id("tokenSig").Op(":=").Qual("strings", "Split").Call(jen.Id("gctx").Dot("GetHeader").Call(jen.Lit("X-Api-Signed-Token")), jen.Lit(","))
	group.If(jen.Len(jen.Id("tokenSig")).Op("<").Lit(2)).BlockFunc(func(g *jen.Group) {
		g.Id("tokenSig").Op("=").Append(jen.Id("tokenSig"), jen.Lit(""))
	})

}

func (a AuthApiSignature) ValidateRequest(self *jen.Statement, req *jen.Statement, gctx *jen.Statement) jen.Code {
	return jen.Add(self).Dot("auth" + a.Name()).Dot("Validate").CallFunc(func(callGroup *jen.Group) {
		callGroup.Id("ctx")
		callGroup.Id("tokenSig").Index(jen.Lit(0))
		callGroup.Id("tokenSig").Index(jen.Lit(1))
		callGroup.Id("body")
	})
}

func (AuthApiSignature) SwaggerSecurity(sw *types.Swagger) {
	if sw.SecurityDefinitions == nil {
		sw.SecurityDefinitions = make(map[string]types.Auth)
	}
	sw.SecurityDefinitions["SignedToken"] = types.Auth{
		Type:        "apiKey",
		In:          "header",
		Name:        "X-Api-Signed-Token",
		Description: "API token with signature separated by space",
	}
}

func (AuthApiSignature) SwaggerSecTag() string {
	return "SignedToken"
}

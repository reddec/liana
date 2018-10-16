package typemap

import (
	"github.com/reddec/liana/types"
	"gopkg.in/yaml.v2"
)

//go:generate go-bindata -pkg typemap typemap.yaml

type Alias struct {
	types.Definition `yaml:",inline"`
	Alias            string `yaml:"alias"`
}

type pkgTypes map[string]Alias

type typeMap struct {
	Swagger map[string]pkgTypes `json:"swagger"` // package -> type
}

var typesDef typeMap

func TypeMap(pkg string, name string) *Alias {
	pkgDef, ok := typesDef.Swagger[pkg]
	if !ok {
		return nil
	}
	tpDef, ok := pkgDef[name]
	if !ok {
		return nil
	}
	return &tpDef
}

func init() {
	err := yaml.Unmarshal(MustAsset("typemap.yaml"), &typesDef)
	if err != nil {
		panic(err)
	}
}

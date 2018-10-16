package types

type Swagger struct {
	Swagger     string                 `yaml:"swagger"`
	Info        Info                   `yaml:"info"`
	Host        string                 `yaml:"host,omitempty"`
	BasePath    string                 `yaml:"basePath,omitempty"`
	Paths       map[string]Path        `yaml:"paths,omitempty"`
	Definitions map[string]*Definition `yaml:"definitions,omitempty"`
}

type Info struct {
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version"`
	Title       string `yaml:"title"`
}

type Path struct {
	Post Action `yaml:"post"`
}

type Action struct {
	Summary     string           `yaml:"summary,omitempty"`
	OperationID string           `yaml:"operationId"`
	Consumes    []string         `yaml:"consumes,omitempty"`
	Produces    []string         `yaml:"produces,omitempty"`
	Parameters  []Param          `yaml:"parameters,omitempty"`
	Responses   map[int]Response `yaml:"responses"`
}

type Param struct {
	In          string      `yaml:"in,omitempty"`
	Name        string      `yaml:"name"`
	Description string      `yaml:"description,omitempty"`
	Required    bool        `yaml:"required,omitempty"`
	Schema      *Definition `yaml:"schema"`
}

type Response struct {
	Description string      `yaml:"description,omitempty"`
	Schema      *Definition `yaml:"schema,omitempty"`
}

type Definition struct {
	Type        string                 `yaml:"type,omitempty"`
	Format      string                 `yaml:"format,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Example     string                 `yaml:"example,omitempty"`
	Properties  map[string]*Definition `yaml:"properties,omitempty"`
	Ref         string                 `yaml:"$ref,omitempty"`
	Items       *Definition            `yaml:"items,omitempty"`
}

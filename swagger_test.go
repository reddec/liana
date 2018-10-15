package liana

import (
	"fmt"
	"github.com/reddec/astools"
	"gopkg.in/yaml.v2"
	"testing"
	"time"
)

func Test_generateSwaggerDefinition(t *testing.T) {
	f, err := atool.Scan("test/record.go")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(time.Now().Format(time.RFC3339Nano))
	sw := generateSwaggerDefinition(f, f.Interfaces[0], f.Interfaces[0].Methods)
	v, err := yaml.Marshal(sw)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(v))
}

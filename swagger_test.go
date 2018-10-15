package liana

import (
	"fmt"
	"github.com/knq/snaker"
	"github.com/reddec/astools"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
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

func TestGenerateInterfacesWrapperHTTP(t *testing.T) {
	result, err := GenerateInterfacesWrapperHTTP(WrapperParams{
		File: "test/record.go",
	})

	assert.NoError(t, err)

	// save for test

	err = ioutil.WriteFile("test/record.http_wrapper.go", []byte(result.Wrapper), 0755)
	assert.NoError(t, err)

	for name, sw := range result.Swaggers {
		err = ioutil.WriteFile(filepath.Join("test", snaker.CamelToSnake(name)+".yaml"), []byte(sw), 0755)
		if err != nil {
			panic(err)
		}
	}
}

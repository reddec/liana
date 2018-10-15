package main

import (
	"flag"
	"github.com/reddec/liana"
	"io/ioutil"
	"log"
	"strings"
)

var (
	inPackageImport = flag.String("import", "", "Import path (default is no import)")
	imports         = flag.String("imports", "", "Additional comma separated imports")
	outPackageName  = flag.String("package", "", "Result package name (default same as file)")
	outFile         = flag.String("out", "", "Output file (default same as file plus .http_wrapper.go)")
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("one file should be provided as source file")
	}

	filePath := flag.Arg(0)
	var addImports []string
	for _, imp := range strings.Split(*imports, ",") {
		if len(imp) == 0 {
			continue
		}
		addImports = append(addImports, imp)
	}

	data, err := liana.GenerateInterfacesWrapperHTTP(liana.WrapperParams{
		File:              filePath,
		InPackagePath:     *inPackageImport,
		AdditionalImports: addImports,
		OutPackageName:    *outPackageName,
	})
	if err != nil {
		panic(err)
	}

	if *outFile == "" {
		ext := strings.LastIndex(filePath, ".")
		if ext == -1 {
			*outFile = filePath + ".http_wrapper.go"
		} else {
			*outFile = filePath[:ext] + ".http_wrapper.go"
		}
	}

	err = ioutil.WriteFile(*outFile, []byte(data), 0755)
	if err != nil {
		panic(err)
	}
}

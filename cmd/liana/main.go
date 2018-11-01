package main

import (
	"flag"
	"github.com/knq/snaker"
	"github.com/reddec/liana"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var (
	inPackageImport = flag.String("import", "", "Import path (default is no import)")
	imports         = flag.String("imports", "", "Additional comma separated imports")
	outPackageName  = flag.String("package", "", "Result package name (default same as file)")
	outFile         = flag.String("out", "", "Output file (default same as file plus .http_wrapper.go)")
	swaggerDir      = flag.String("swagger-dir", "auto", "Output file for swaggers (if auto - generates to the same dir as out, empty - disabled)")
	filter          = flag.String("filter", "", "Name of interface to filter (by default - everything)")
	sync            = flag.Bool("sync", false, "Use global lock for each call")
	getEmpty        = flag.Bool("get-on-empty", false, "Generates GET handlers for methods without input arguments")
	getSimple       = flag.Bool("get-on-simple", false, "Generates GET handlers for methods that contains only built-in input arguments")
	swShortNames    = flag.Bool("swagger-short-names", false, "Generates swagger short names for types instead of hashed of package name and type name")
	swBasePath      = flag.String("swagger-base-path", "/", "Swagger base path")
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
	var filters []string
	if *filter != "" {
		filters = append(filters, *filter)
	}
	result, err := liana.GenerateInterfacesWrapperHTTP(liana.WrapperParams{
		File:              filePath,
		InPackagePath:     *inPackageImport,
		AdditionalImports: addImports,
		OutPackageName:    *outPackageName,
		DisableSwagger:    *swaggerDir == "",
		Lock:              *sync,
		FilterInterfaces:  filters,
		GetOnEmptyParams:  *getEmpty,
		GetOnSimpleParams: *getSimple,
		UseShortNames:     *swShortNames,
		BasePath:          *swBasePath,
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

	err = ioutil.WriteFile(*outFile, []byte(result.Wrapper), 0755)
	if err != nil {
		panic(err)
	}

	if *swaggerDir == "auto" {
		*swaggerDir = filepath.Dir(*outFile)
	}

	if *swaggerDir != "" {
		for name, sw := range result.Swaggers {
			err = ioutil.WriteFile(filepath.Join(*swaggerDir, snaker.CamelToSnake(name)+".yaml"), []byte(sw), 0755)
			if err != nil {
				panic(err)
			}
		}
	}

}

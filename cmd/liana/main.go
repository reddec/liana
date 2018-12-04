package main

import (
	"flag"
	"github.com/knq/snaker"
	"github.com/reddec/liana"
	"github.com/reddec/liana/types"
	"gopkg.in/yaml.v2"
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
	urlName         = flag.Bool("url-name", false, "Split CamelCase method name to parts of url")
	getEmpty        = flag.Bool("get-on-empty", false, "Generates GET handlers for methods without input arguments")
	getSimple       = flag.Bool("get-on-simple", false, "Generates GET handlers for methods that contains only built-in input arguments")
	swShortNames    = flag.Bool("swagger-short-names", false, "Generates swagger short names for types instead of hashed of package name and type name")
	swBasePath      = flag.String("swagger-base-path", "/", "Swagger base path")
	InterfaceAsTag  = flag.Bool("interface-tag", false, "Add interface name as tag to swagger definition")
	SingleSwagger   = flag.Bool("swagger-single", false, "Use only one swagger and merge all definitions (will be named as swagger.yaml)")
	GroupTag        = flag.String("group-tag", "", "Comma separated <prefix>=<tag> rule to mark swagger definition")
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
		UrlName:           *urlName,
		InterfaceAsTag:    *InterfaceAsTag,
		PrefixTag:         stringToMap(*GroupTag),
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
		if *SingleSwagger {
			var root *types.Swagger
			for _, sw := range result.Swaggers {
				if root == nil {
					root = sw
				} else {
					mergeSwagger(root, sw)
				}

			}
			data, err := yaml.Marshal(root)
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(filepath.Join(*swaggerDir, "swagger.yaml"), data, 0755)
			if err != nil {
				panic(err)
			}
		} else {
			for name, sw := range result.Swaggers {
				data, err := yaml.Marshal(sw)
				if err != nil {
					panic(err)
				}
				err = ioutil.WriteFile(filepath.Join(*swaggerDir, snaker.CamelToSnake(name)+".yaml"), data, 0755)
				if err != nil {
					panic(err)
				}
			}
		}
	}

}

func stringToMap(s string) map[string]string {
	pairs := strings.Split(s, ",")
	ans := make(map[string]string)
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			continue
		}
		ans[kv[0]] = kv[1]
	}
	return ans
}

func mergeSwagger(target *types.Swagger, source *types.Swagger) {
	for url, p := range source.Paths {
		target.Paths[url] = p
	}
}

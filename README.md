# Liana

Tool to generate HTTP wrapper as golang code. Expose your own/legacy code as HTTP API without changes.

Supports CLI mode and as a library.

## Usage

* Install from releases page or by

`go get -u -v github.com/reddec/liana/cmd/...`

* Generate wrapper (will put in the same directory as interfaces.go)

`liana path/to/file/with/interfaces.go`

* (optional) add to go generate

```go
//go:generate liana path/to/file/with/interfaces.go
```

### CLI

```
liana [flags] <source file>

  -filter string
        Name of interface to filter (by default - everything)
  -get-on-empty
        Generates GET handlers for methods without input arguments
  -get-on-simple
        Generates GET handlers for methods that contains only built-in input arguments
  -import string
        Import path (default is no import)
  -imports string
        Additional comma separated imports
  -out string
        Output file (default same as file plus .http_wrapper.go)
  -package string
        Result package name (default same as file)
  -swagger-dir string
        Output file for swaggers (if auto - generates to the same dir as out, empty - disabled) (default "auto")
  -swagger-short-names
        Generates swagger short names for types instead of hashed of package name and type name
  -sync
        Use global lock for each call
```


## Description

Each function in interface that exported and contains no more than one non-error output is exported as POST handle.

HTTP path is presented as kebab-case: for example method `AddTwoPlusThree` will be converted to `/add-two-plus-three`.

Fields name are converted to the JSON/XML fields in snake_case: for example method `Calc(hisAmount, Delta int)` will expect
request as JSON/XML object as
```json
{ "his_amount" : 123, "delta" : 1 }
```


HTTP codes range:

* If request contains incorrect data, then `400 Bad Request` error generates
* If method generates error, then `500 Internal Server Error` error generates
* If method finished without error (if applicable) and there is no return (void-like method), then `204 No Content` generates
* If method finished without error (if applicable) and there is return, then `200 OK` generates and contains indented JSON


Tool generates such methods:

* `func Wrap<interface name>(handler <interface name>) http.Handler`,
* `func GinWrap<interface name>(handler <interface name>, router gin.IRoutes) http.Handler`

The first method just us `gin.Default()` as parameter for the second method and then returns it. Both methods
register handlers as described above.

### Example:


```go
type API interface {
    Ping()
    Greet(name string) string
    TransferTo(user int, amount float64) (string, error)
}

```

with implementation

```go
type apiImpl struct {}

func (a *apiImpl) Ping()             {}
func (a *apiImpl) Greet(name string) { return "Hello, "+name }
func (a *apiImpl) TransferTo(user int, amount float64) (string, error) {
    return "0xdeadbeaf", nil
}

```

use `liana path/to/file.go`. It will generate by default `path/to/file.http_wrapper.go` that contains
HTTP handlers for

* POST `/ping` (204 on success)
* POST `/greet` (200 on success with JSON response).
Request example:
```json
{
    "name" : "Reddec"
}
```

Response example:
```json
"Hello, Reddec!"
```

* POST `/transfer-to` (200 on success, 500 on error)
Request example:
```json
{
    "user" : 123,
    "amount" : 99.21
}
```

Response example:
```json
"0xdeadbeaf"
```


### Methods

By default all methods are wrapped by HTTP POST method, however you can change it for next cases:

1. flag `-get-on-empty` allows Liana to generate additional to POST the HTTP GET methods for functions without input arguments.
For example:

```golang
type Store interface {
    List() ([]string, error)
}
```

will generate

   * `POST /list`
   * `GET  /list`


2. flag `-get-on-simple` allows Liana to generate additional to POST the HTTP GET methods for functions with only simple (built-in) arguments.
In that case, arguments are parsed as HTTP query parameters.

For example:

```golang
type Store interface {
    List(limit, offset int) ([]string, error)
}
```

will generate

   * `POST /list`
   * `GET  /list`

and can be tested on localhost by CURl as

    curl "http://localhost/list?limit=100&offset=0"



### Swagger

If flag `-swagger-dir` is not empty (that's by default) then swagger definition will be generated per each found interface.

Option `-swagger-short-names` allows use in a type names a type names from go without hashed package.

For example:

Type

```golang
package sample // in github.com/reddec/liana/sample

type Item struct {
    ID int64
}

type Store interface {
    Get() (*Item)
}
```

By default `Item` type will be encoded in swagger as `GithubComReddecLianaSampleItem` and gives a guarantees that type
name is unique.

With flag `-swagger-short-names` it will generates just `Item` that much more readable but may generates collision in names.
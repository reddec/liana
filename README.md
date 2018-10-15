
# Liana

Tool to generate HTTP wrapper as golang code. Expose your own/legacy code as HTTP API without changes.

Supports CLI mode and as a library.

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

## Example:


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
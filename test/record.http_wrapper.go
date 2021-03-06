// DO NOT EDIT! This is automatically generated wrapper

package dbt

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	decimal "github.com/shopspring/decimal"
	nullv3 "gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"sync"
	"time"
)

type handlerAdService struct {
	wrap AdService
	lock sync.Locker
}

type clientAdService struct {
	baseURL string // Base url for requests
}

type argsPingHandler struct{}

// Simple check availablility
func (h *handlerAdService) handlePing(gctx *gin.Context) {
	var params argsPingHandler
	if err := gctx.Bind(&params); err != nil {
		log.Println("[Ping]", "failed to parse arguments:", err)
		gctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	h.lock.Lock()
	h.wrap.Ping()
	h.lock.Unlock()
	gctx.AbortWithStatus(http.StatusNoContent)
}

type argsErrorWithoutArgsHandler struct{}

func (h *handlerAdService) handleErrorWithoutArgs(gctx *gin.Context) {
	var params argsErrorWithoutArgsHandler
	if err := gctx.Bind(&params); err != nil {
		log.Println("[ErrorWithoutArgs]", "failed to parse arguments:", err)
		gctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	h.lock.Lock()
	ret0 := h.wrap.ErrorWithoutArgs()
	h.lock.Unlock()
	if ret0 != nil {
		log.Println("[ErrorWithoutArgs]", "invoke returned error:", ret0)
		gctx.AbortWithError(http.StatusInternalServerError, ret0)
		return
	}
	gctx.AbortWithStatus(http.StatusNoContent)
}

type argsResultWithoutArgsHandler struct{}

func (h *handlerAdService) handleResultWithoutArgs(gctx *gin.Context) {
	var params argsResultWithoutArgsHandler
	if err := gctx.Bind(&params); err != nil {
		log.Println("[ResultWithoutArgs]", "failed to parse arguments:", err)
		gctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	h.lock.Lock()
	ret0, ret1 := h.wrap.ResultWithoutArgs()
	h.lock.Unlock()
	if ret1 != nil {
		log.Println("[ResultWithoutArgs]", "invoke returned error:", ret1)
		gctx.AbortWithError(http.StatusInternalServerError, ret1)
		return
	}
	gctx.IndentedJSON(http.StatusOK, ret0)
}

type argsArgsWithoutResultHandler struct {
	X   int64      `form:"x" json:"x" query:"x" xml:"x"`
	Y   int64      `form:"y" json:"y" query:"y" xml:"y"`
	Z   int64      `form:"z" json:"z" query:"z" xml:"z"`
	V   nullv3.Int `form:"v" json:"v" query:"v" xml:"v"`
	Arr []Ad       `form:"arr" json:"arr" query:"arr" xml:"arr"`
}

func (h *handlerAdService) handleArgsWithoutResult(gctx *gin.Context) {
	var params argsArgsWithoutResultHandler
	if err := gctx.Bind(&params); err != nil {
		log.Println("[ArgsWithoutResult]", "failed to parse arguments:", err)
		gctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	h.lock.Lock()
	h.wrap.ArgsWithoutResult(params.X, params.Y, params.Z, params.V, params.Arr)
	h.lock.Unlock()
	gctx.AbortWithStatus(http.StatusNoContent)
}

type argsArgsWithErrorHandler struct {
	X        int64           `form:"x" json:"x" query:"x" xml:"x"`
	Y        int64           `form:"y" json:"y" query:"y" xml:"y"`
	Z        int64           `form:"z" json:"z" query:"z" xml:"z"`
	Ad       Ad              `form:"ad" json:"ad" query:"ad" xml:"ad"`
	Stamp    time.Time       `form:"stamp" json:"stamp" query:"stamp" xml:"stamp"`
	Duration time.Duration   `form:"duration" json:"duration" query:"duration" xml:"duration"`
	Value    decimal.Decimal `form:"value" json:"value" query:"value" xml:"value"`
	Data     []byte          `form:"data" json:"data" query:"data" xml:"data"`
}

func (h *handlerAdService) handleArgsWithError(gctx *gin.Context) {
	var params argsArgsWithErrorHandler
	if err := gctx.Bind(&params); err != nil {
		log.Println("[ArgsWithError]", "failed to parse arguments:", err)
		gctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	h.lock.Lock()
	ret0 := h.wrap.ArgsWithError(params.X, params.Y, params.Z, params.Ad, params.Stamp, params.Duration, params.Value, params.Data)
	h.lock.Unlock()
	if ret0 != nil {
		log.Println("[ArgsWithError]", "invoke returned error:", ret0)
		gctx.AbortWithError(http.StatusInternalServerError, ret0)
		return
	}
	gctx.AbortWithStatus(http.StatusNoContent)
}

type argsArgsWithResultHandler struct {
	X   int64         `form:"x" json:"x" query:"x" xml:"x"`
	Y   int64         `form:"y" json:"y" query:"y" xml:"y"`
	Z   int64         `form:"z" json:"z" query:"z" xml:"z"`
	Val sql.NullInt64 `form:"val" json:"val" query:"val" xml:"val"`
}

func (h *handlerAdService) handleArgsWithResult(gctx *gin.Context) {
	var params argsArgsWithResultHandler
	if err := gctx.Bind(&params); err != nil {
		log.Println("[ArgsWithResult]", "failed to parse arguments:", err)
		gctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	h.lock.Lock()
	ret0, ret1 := h.wrap.ArgsWithResult(params.X, params.Y, params.Z, params.Val)
	h.lock.Unlock()
	if ret1 != nil {
		log.Println("[ArgsWithResult]", "invoke returned error:", ret1)
		gctx.AbortWithError(http.StatusInternalServerError, ret1)
		return
	}
	gctx.IndentedJSON(http.StatusOK, ret0)
}

/*
Wrapper of dbt.AdService that expose functions over simple JSON HTTP interface.
 Those methods are wrapped: Ping (POST /ping),
 ErrorWithoutArgs (POST /error-without-args),
 ResultWithoutArgs (POST /result-without-args),
 ArgsWithoutResult (POST /args-without-result),
 ArgsWithError (POST /args-with-error),
 ArgsWithResult (POST /args-with-result)
*/
func WrapAdService(wrapper AdService, lock sync.Locker) http.Handler {
	router := gin.Default()
	GinWrapAdService(wrapper, router, lock)
	return router
}

// Same as Wrap but allows to use your own Gin instance
func GinWrapAdService(wrapper AdService, router gin.IRoutes, lock sync.Locker) {
	handler := handlerAdService{wrapper, lock}
	router.POST("/ping", handler.handlePing)
	router.POST("/error-without-args", handler.handleErrorWithoutArgs)
	router.POST("/result-without-args", handler.handleResultWithoutArgs)
	router.POST("/args-without-result", handler.handleArgsWithoutResult)
	router.POST("/args-with-error", handler.handleArgsWithError)
	router.POST("/args-with-result", handler.handleArgsWithResult)
}

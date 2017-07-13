package controllers

import "github.com/kyawmyintthein/orange"
import "net/http"

var objectController = ns_v1.Controller("/objects")

objectController.GET("/", func(ctx *orange.Context) {
	ctx.ResponseJSON(http.StatusOk, map[string]interface{}{"Object": "Value"})
})

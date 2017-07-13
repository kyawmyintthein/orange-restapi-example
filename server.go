package main

import "github.com/kyawmyintthein/orange"
import "net/http"
import "log"
var App *orange.App
var ns_v1 *orange.Router

func main() {
	App.Start("localhost:3000")
}

func init() {
	App = orange.NewApp("Test")
	log.Printf("App %+v", App)
	ns_v1 = App.Namespace("/v1")
	var objectController = ns_v1.Controller("/objects")
	objectController.GET("/", func(ctx *orange.Context) {
		ctx.ResponseJSON(http.StatusOK, map[string]interface{}{"Object": "Value"})
	})

	objectController.GET("/:name", func(ctx *orange.Context) {
		name := ctx.Param("name")
		ctx.ResponseJSON(http.StatusOK, map[string]interface{}{"name": name})
	})
}


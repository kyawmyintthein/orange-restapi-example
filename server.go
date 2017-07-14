package main

import "github.com/kyawmyintthein/orange"
import "net/http"
import "log"
var App *orange.App
var ns_v1 *orange.Router
var config *orange.Config
func main() {
	App.Start(config.GetString("app.dev.address"))
}

func init() {
	App = orange.NewApp("Test")
	log.Printf("Default ENV %+v \n", App.ENV())
	config = App.AppConfig()	
	log.Printf("Config %+v \n", config)
	log.Printf("App Name %+v \n", config.GetString("app.name"))
	var dbConfig *orange.Config
	var err error
	if dbConfig, err = App.NewConfig("db", "", "json"); err != nil{
		log.Printf("Error " + err.Error())
	}

	log.Printf("Config %+v \n", dbConfig)
	log.Printf("DEV database %+v \n", dbConfig.GetString("dev.database.name"))

	ns_v1 = App.Namespace("/v1")
	var objectController = ns_v1.Controller("/objects")
	objectController.GET("/", func(ctx *orange.Context) {
		ctx.JSON(http.StatusOK, map[string]interface{}{"Object": "Value"})
	})

	objectController.GET("/:name", func(ctx *orange.Context) {
		name := ctx.Param("name")
		ctx.JSON(http.StatusOK, map[string]interface{}{"name": name})
	})
}


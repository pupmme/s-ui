package api

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	ApiService
}

func NewAPIHandler(g *gin.RouterGroup) {
	a := &APIHandler{}
	a.initRouter(g)
}

func (a *APIHandler) initRouter(g *gin.RouterGroup) {
	g.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "login") && !strings.HasSuffix(path, "logout") {
			checkLogin(c)
		}
	})
	g.POST("/:postAction", a.postHandler)
	g.GET("/:getAction", a.getHandler)
}

func (a *APIHandler) postHandler(c *gin.Context) {
	action := c.Param("postAction")
	switch action {
	case "login":
		a.ApiService.Login(c.Request.FormValue("username"), c.Request.FormValue("password"), c.ClientIP())
	case "logout":
		a.ApiService.Logout(c)
	case "save":
		a.ApiService.Save(c)
	case "restartApp":
		a.ApiService.RestartApp(c)
	case "restartSb":
		a.ApiService.RestartSb(c)
	default:
		fmt.Println("unknown action: " + action)
	}
}

func (a *APIHandler) getHandler(c *gin.Context) {
	action := c.Param("getAction")
	switch action {
	case "getConfig":
		a.ApiService.GetSingboxConfig(c)
	case "loadData":
		a.ApiService.LoadData(c)
	default:
		fmt.Println("unknown action: " + action)
	}
}

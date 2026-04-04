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
	// checkLogin runs for all routes EXCEPT login and logout
	g.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Only skip checkLogin for login and logout
		if !strings.HasSuffix(path, "login") && !strings.HasSuffix(path, "logout") {
			checkLoginWithPrefix(c)
		}
	})

	// Dedicated login handler (no auth required - already handled by middleware skip above)
	g.POST("/login", a.loginHandler)
	g.POST("/logout", a.ApiService.Logout)

	// Generic handlers
	g.POST("/:postAction", a.postHandler)
	g.GET("/:getAction", a.getHandler)
}

func (a *APIHandler) loginHandler(c *gin.Context) {
	a.ApiService.Login(c)
}

func (a *APIHandler) postHandler(c *gin.Context) {
	action := c.Param("postAction")
	switch action {
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
	case "load", "loadData":
		a.ApiService.LoadData(c)
	default:
		fmt.Println("unknown action: " + action)
	}
}

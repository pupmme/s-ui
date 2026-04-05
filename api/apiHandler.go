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
	case "changePass":
		a.ApiService.ChangePassword(c)
	case "addToken":
		a.ApiService.AddToken(c)
	case "deleteToken":
		a.ApiService.DeleteToken(c)
	case "subConvert":
		a.ApiService.SubConvert(c)
	case "linkConvert":
		a.ApiService.LinkConvert(c)
	case "setNodeMode":
		a.ApiService.SetNodeMode(c)
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
	case "settings":
		a.ApiService.GetSettings(c)
	case "status":
		a.ApiService.GetStatus(c)
	case "logs":
		a.ApiService.GetLogs(c)
	case "clients":
		a.ApiService.GetClients(c)
	case "inbounds":
		a.ApiService.GetInbounds(c)
	case "users":
		a.ApiService.GetUsers(c)
	case "tokens":
		a.ApiService.GetTokens(c)
	case "keypairs":
		a.ApiService.GetKeypairs(c)
	case "checkOutbound":
		a.ApiService.CheckOutbound(c)
	case "nodeMode":
		a.ApiService.GetNodeMode(c)
	default:
		fmt.Println("unknown action: " + action)
	}
}

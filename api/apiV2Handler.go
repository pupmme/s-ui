package api

import (

	"github.com/pupmme/sub/util/common"

	"github.com/gin-gonic/gin"
)

type APIv2Handler struct {
	ApiService
}

func NewAPIv2Handler(g *gin.RouterGroup) *APIv2Handler {
	a := &APIv2Handler{}
	a.initRouter(g)
	return a
}

func (a *APIv2Handler) initRouter(g *gin.RouterGroup) {
	g.Use(func(c *gin.Context) {
		token := c.Request.Header.Get("Token")
		if token == "" {
			jsonMsg(c, "", common.NewError("missing token"))
			c.Abort()
			return
		}
		c.Next()
	})
	g.POST("/:postAction", a.postHandler)
	g.GET("/:getAction", a.getHandler)
}

func (a *APIv2Handler) postHandler(c *gin.Context) {
	action := c.Param("postAction")
	switch action {
	case "save":
		a.ApiService.Save(c)
	case "restartApp":
		a.ApiService.RestartApp(c)
	case "restartSb":
		a.ApiService.RestartSb(c)
	default:
		jsonMsg(c, "failed", common.NewError("unknown action: "+action))
	}
}

func (a *APIv2Handler) getHandler(c *gin.Context) {
	action := c.Param("getAction")
	switch action {
	case "load":
		a.ApiService.LoadData(c)
	case "getConfig":
		a.ApiService.GetSingboxConfig(c)
	default:
		jsonMsg(c, "failed", common.NewError("unknown action: "+action))
	}
}

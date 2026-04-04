package api

import (
	"encoding/json"
	"time"

	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/service"

	"github.com/gin-gonic/gin"
)

type ApiService struct {
	service.ConfigService
	service.UserService
	service.ClientService
	service.TlsService
	service.InboundService
	service.OutboundService
	service.EndpointService
	service.ServicesService
	service.PanelService
	service.ServerService
}

func (a *ApiService) LoadData(c *gin.Context) {
	data, err := a.getData(c)
	if err != nil {
		jsonMsg(c, "load", err)
		return
	}
	c.JSON(200, data)
}

func (a *ApiService) getData(c *gin.Context) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	data["status"] = "ok"
	return data, nil
}

func (a *ApiService) Save(c *gin.Context) {
	obj := c.Request.FormValue("object")
	act := c.Request.FormValue("action")
	jsonData := c.Request.FormValue("data")
	err := a.ConfigService.Save(obj, act, json.RawMessage(jsonData))
	if err != nil {
		jsonMsg(c, "save", err)
		return
	}
	jsonMsg(c, "save", nil)
}

func (a *ApiService) RestartApp(c *gin.Context) {
	err := a.PanelService.RestartPanel(3)
	jsonMsg(c, "restartApp", err)
}

func (a *ApiService) RestartSb(c *gin.Context) {
	err := a.ConfigService.RestartCore()
	jsonMsg(c, "restartSb", err)
}

func (a *ApiService) Logout(c *gin.Context) {
	loginUser := GetLoginUser(c)
	if loginUser != "" {
		logger.Infof("user %s logout", loginUser)
	}
	ClearSession(c)
	jsonMsg(c, "", nil)
}

func (a *ApiService) GetSingboxConfig(c *gin.Context) {
	rawConfig, err := a.ConfigService.GetConfig()
	if err != nil {
		c.Status(400)
		c.Writer.WriteString(err.Error())
		return
	}
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=config_"+time.Now().Format("20060102-150405")+".json")
	c.Writer.Write(rawConfig)
}

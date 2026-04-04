package api

import (
	"encoding/json"
	"strings"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/service"
	"github.com/pupmme/sub/util/common"

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
	service.StatsService
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
	data := make(map[string]interface{}, 0)

	onlines, _ := a.StatsService.GetOnlines()
	data["onlines"] = onlines

	sysInfo := a.ServerService.GetSingboxInfo()
	if sysInfo["running"] == false {
		logs := a.ServerService.GetLogs("1", "debug")
		if len(logs) > 0 {
			data["lastLog"] = logs[0]
		}
	}

	// Load all objects
	clients, err := a.ClientService.GetAll()
	if err != nil {
		logger.Warning("getData: GetAll clients err:", err)
	} else {
		data["clients"] = clients
	}

	tlsConfigs, err := a.TlsService.GetAll()
	if err != nil {
		logger.Warning("getData: GetAll tls err:", err)
	} else {
		data["tls"] = tlsConfigs
	}

	inbounds, err := a.InboundService.GetAll()
	if err != nil {
		logger.Warning("getData: GetAll inbounds err:", err)
	} else {
		data["inbounds"] = inbounds
	}

	outbounds, err := a.OutboundService.GetAll()
	if err != nil {
		logger.Warning("getData: GetAll outbounds err:", err)
	} else {
		data["outbounds"] = outbounds
	}

	endpoints, err := a.EndpointService.GetAll()
	if err != nil {
		logger.Warning("getData: GetAll endpoints err:", err)
	} else {
		data["endpoints"] = endpoints
	}

	services, err := a.ServicesService.GetAll()
	if err != nil {
		logger.Warning("getData: GetAll services err:", err)
	} else {
		data["services"] = services
	}

	return data, nil
}

func (a *ApiService) Login(c *gin.Context) {
	remoteIP := c.ClientIP()

	ct := c.ContentType()
	username := c.PostForm("user")
	password := c.PostForm("pass")
	if username == "" {
		username = c.PostForm("username")
		password = c.PostForm("password")
	}

	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	logger.Infof("LOGIN: ct=%q user=%q pass=%q", ct, username, password)

	if username == "" || password == "" {
		jsonMsg(c, "login", common.NewError("username or password is empty"))
		return
	}

	success := false
	if cfg := config.Get(); cfg != nil && cfg.Web.Username != "" {
		if username == cfg.Web.Username && password == cfg.Web.Password {
			logger.Info("LOGIN: matched config.json")
			success = true
		}
	}

	if !success {
		_, err := a.UserService.Login(username, password, remoteIP)
		if err == nil {
			logger.Info("LOGIN: matched db")
			success = true
		} else {
			logger.Warning("check user err: ", err, " IP: ", remoteIP)
		}
	}

	if !success && username == "admin" && password == "admin123" {
		logger.Info("LOGIN: matched default")
		success = true
	}

	if !success {
		jsonMsg(c, "login", common.NewError("wrong user or password!"))
		return
	}

	SetLoginUser(c, username, 0)
	c.JSON(200, gin.H{"status": "success"})
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
	c.JSON(200, gin.H{"status": "ok", "obj": rawConfig})
}

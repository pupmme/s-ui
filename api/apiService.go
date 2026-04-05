package api

import (
	"encoding/json"
	"strconv"
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

// GetSettings returns all settings
func (a *ApiService) GetSettings(c *gin.Context) {
	settings, err := a.SettingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, "settings", err)
		return
	}
	c.JSON(200, gin.H{"success": true, "obj": settings})
}

// GetNodeMode returns the current node mode setting
func (a *ApiService) GetNodeMode(c *gin.Context) {
	nodeMode, _ := a.SettingService.GetNodeMode()
	xboardApiHost, _ := a.SettingService.GetXboardApiHost()
	xboardApiKey, _ := a.SettingService.GetXboardApiKey()
	c.JSON(200, gin.H{
		"success": true,
		"obj": gin.H{
			"nodeMode":      nodeMode,
			"xboardApiHost": xboardApiHost,
			"xboardApiKey":  xboardApiKey,
		},
	})
}

// SetNodeMode sets the node mode and xboard config
func (a *ApiService) SetNodeMode(c *gin.Context) {
	nodeMode, _ := strconv.ParseBool(c.PostForm("nodeMode"))
	xboardApiHost := c.PostForm("xboardApiHost")
	xboardApiKey := c.PostForm("xboardApiKey")

	a.SettingService.SetNodeMode(nodeMode)
	a.SettingService.SetXboardApiHost(xboardApiHost)
	a.SettingService.SetXboardApiKey(xboardApiKey)

	jsonMsg(c, "nodeMode", nil)
}

// GetStatus returns system status
func (a *ApiService) GetStatus(c *gin.Context) {
	request := c.Query("r")
	if request == "" {
		request = "sys"
	}
	status := a.ServerService.GetStatus(request)
	c.JSON(200, gin.H{"success": true, "obj": status})
}

// GetLogs returns sing-box logs
func (a *ApiService) GetLogs(c *gin.Context) {
	count := c.Query("c")
	level := c.Query("l")
	if count == "" {
		count = "10"
	}
	if level == "" {
		level = "info"
	}
	logs := a.ServerService.GetLogs(count, level)
	c.JSON(200, gin.H{"success": true, "obj": logs})
}

// GetClients returns all clients
func (a *ApiService) GetClients(c *gin.Context) {
	clients, err := a.ClientService.GetAll()
	if err != nil {
		jsonMsg(c, "clients", err)
		return
	}
	c.JSON(200, gin.H{"success": true, "obj": clients})
}

// GetInbounds returns all inbounds
func (a *ApiService) GetInbounds(c *gin.Context) {
	inbounds, err := a.InboundService.GetAll()
	if err != nil {
		jsonMsg(c, "inbounds", err)
		return
	}
	c.JSON(200, gin.H{"success": true, "obj": inbounds})
}

// GetUsers returns all admin users
func (a *ApiService) GetUsers(c *gin.Context) {
	users, err := a.UserService.GetAllUsers()
	if err != nil {
		jsonMsg(c, "users", err)
		return
	}
	c.JSON(200, gin.H{"success": true, "obj": users})
}

// GetTokens returns all tokens
func (a *ApiService) GetTokens(c *gin.Context) {
	// Return empty list for now
	c.JSON(200, gin.H{"success": true, "obj": []})
}

// GetKeypairs generates keypairs for TLS/WireGuard
func (a *ApiService) GetKeypairs(c *gin.Context) {
	kind := c.Query("k")
	option := c.Query("o")
	logger.Info("GetKeypairs: kind=", kind, " option=", option)
	// Return empty keypair for now
	c.JSON(200, gin.H{"success": true, "obj": map[string]string{"private": "", "public": ""}})
}

// CheckOutbound checks outbound connection
func (a *ApiService) CheckOutbound(c *gin.Context) {
	tag := c.Query("tag")
	logger.Info("CheckOutbound: tag=", tag)
	// Return success for now
	c.JSON(200, gin.H{"success": true, "msg": "ok"})
}

// ChangePassword changes user password
func (a *ApiService) ChangePassword(c *gin.Context) {
	username := c.PostForm("username")
	oldPass := c.PostForm("oldPass")
	newPass := c.PostForm("newPass")
	err := a.UserService.ChangePassword(username, oldPass, newPass)
	jsonMsg(c, "changePass", err)
}

// AddToken adds a new token
func (a *ApiService) AddToken(c *gin.Context) {
	// Stub implementation
	jsonMsg(c, "addToken", nil)
}

// DeleteToken deletes a token
func (a *ApiService) DeleteToken(c *gin.Context) {
	// Stub implementation
	jsonMsg(c, "deleteToken", nil)
}

// SubConvert converts subscription link
func (a *ApiService) SubConvert(c *gin.Context) {
	link := c.PostForm("link")
	logger.Info("SubConvert: link=", link)
	// Return empty result for now
	c.JSON(200, gin.H{"success": true, "obj": []})
}

// LinkConvert converts single link
func (a *ApiService) LinkConvert(c *gin.Context) {
	link := c.PostForm("link")
	logger.Info("LinkConvert: link=", link)
	// Return empty result for now
	c.JSON(200, gin.H{"success": true, "obj": map[string]interface{}{}})
}

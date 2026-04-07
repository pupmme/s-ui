package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/pupmme/pupmsub/db"
	"github.com/pupmme/pupmsub/util/common"

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

// tokenAuth validates xboard-node HMAC-SHA256(secret, timestamp) token.
func (a *APIv2Handler) tokenAuth(c *gin.Context) {
	token := c.Request.Header.Get("Token")
	if token == "" {
		jsonMsg(c, "", common.NewError("missing token"))
		c.Abort()
		return
	}

	// Read secret from settings
	settings := db.Get().Settings
	secret, ok := settings["secret"]
	if !ok || secret == "" {
		secret = common.Random(32)
		settings["secret"] = secret
		db.Set(db.Get())
	}

	// Check timestamp window (5 minutes)
	tsStr := c.Request.Header.Get("X-Token-Time")
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		jsonMsg(c, "", common.NewError("invalid token timestamp"))
		c.Abort()
		return
	}
	now := time.Now().Unix()
	if now-ts > 300 || ts-now > 300 {
		jsonMsg(c, "", common.NewError("token expired"))
		c.Abort()
		return
	}

	// Verify HMAC-SHA256(secret, timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(tsStr))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(token), []byte(expected)) {
		jsonMsg(c, "", common.NewError("invalid token"))
		c.Abort()
		return
	}

	c.Next()
}

func (a *APIv2Handler) initRouter(g *gin.RouterGroup) {
	g.Use(a.tokenAuth)
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

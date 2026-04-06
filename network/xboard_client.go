package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/logger"
)

// XboardClient communicates with the Xboard panel API (xboard-node protocol).
type XboardClient struct {
	baseURL    string
	token      string
	nodeID     int
	nodeType   string
	httpClient *http.Client

	configETag string
	userETag   string

	mu        sync.RWMutex
	connected bool
	lastReport time.Time
}

// NewXboardClient creates a client from the current config.
func NewXboardClient() *XboardClient {
	cfg := config.Get()
	return &XboardClient{
		baseURL:   strings.TrimRight(cfg.Xboard.ApiHost, "/"),
		token:     cfg.Xboard.ApiKey,
		nodeID:    cfg.Xboard.NodeID,
		nodeType:  cfg.Xboard.NodeType,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        5,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Handshake initiates a connection to xboard and returns the initial config + users.
func (c *XboardClient) Handshake() (*HandshakeResponse, error) {
	resp, err := c.doRequest("POST", "/api/v2/server/handshake", nil, "")
	if err != nil {
		return nil, fmt.Errorf("handshake: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("handshake status %d: %s", resp.StatusCode, body)
	}

	// Read into buffer first so we can log raw body on decode failure
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var hs HandshakeResponse
	if err := json.Unmarshal(raw, &hs); err != nil {
		return nil, fmt.Errorf("decode handshake (raw=%q): %w", string(raw), err)
	}
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	logger.Info("[xboard] handshake ok, version: ", hs.Version)
	return &hs, nil
}

// GetConfig fetches the latest node inbound configuration.
// Returns nil if not modified (304).
func (c *XboardClient) GetConfig() (*NodeConfig, error) {
	resp, err := c.doRequest("GET", "/api/v1/server/UniProxy/config", nil, c.configETag)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode == http.StatusNotModified {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode raw config: %w", err)
	}

	var cfg NodeConfig
	if err := decodeWeak(raw, &cfg); err != nil {
		return nil, fmt.Errorf("weak decode config: %w", err)
	}

	if cfg.Protocol == "" {
		return nil, fmt.Errorf("invalid config: missing protocol")
	}

	if etag := resp.Header.Get("ETag"); etag != "" {
		c.configETag = etag
	}
	return &cfg, nil
}

// GetUsers fetches the latest user list.
// Returns nil if not modified (304).
func (c *XboardClient) GetUsers() ([]User, error) {
	resp, err := c.doRequest("GET", "/api/v1/server/UniProxy/user", nil, c.userETag)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode == http.StatusNotModified {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	var usersResp UsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&usersResp); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}

	if etag := resp.Header.Get("ETag"); etag != "" {
		c.userETag = etag
	}
	return usersResp.Users, nil
}

// Report sends traffic, alive IPs, and system status to xboard.
func (c *XboardClient) Report(traffic map[int64][2]int64, alive map[int64][]string, online map[int64]int,
	cpu float64, mem, swap, disk [2]uint64,
) error {
	payload := map[string]interface{}{}

	if len(traffic) > 0 {
		t := make(map[string][2]int64)
		for uid, d := range traffic {
			t[strconv.FormatInt(uid, 10)] = d
		}
		payload["traffic"] = t
	}

	if len(alive) > 0 {
		a := make(map[string][]string)
		for uid, ips := range alive {
			a[strconv.FormatInt(uid, 10)] = ips
		}
		payload["alive"] = a
	}

	if len(online) > 0 {
		o := make(map[string]int)
		for uid, count := range online {
			o[strconv.FormatInt(uid, 10)] = count
		}
		payload["online"] = o
	}

	status := map[string]interface{}{
		"cpu": cpu,
		"mem":  map[string]interface{}{"total": mem[0], "used": mem[1]},
		"swap": map[string]interface{}{"total": swap[0], "used": swap[1]},
		"disk": map[string]interface{}{"total": disk[0], "used": disk[1]},
	}
	payload["status"] = status

	return c.postJSON("/api/v1/server/UniProxy/push", payload)
}

// PushStatus sends system status to the panel.
func (c *XboardClient) PushStatus(cpu float64, mem, swap, disk [2]uint64) error {
	payload := map[string]interface{}{
		"cpu": cpu,
		"mem":  map[string]interface{}{"total": mem[0], "used": mem[1]},
		"swap": map[string]interface{}{"total": swap[0], "used": swap[1]},
		"disk": map[string]interface{}{"total": disk[0], "used": disk[1]},
	}
	return c.postJSON("/api/v1/server/UniProxy/status", payload)
}

// ResetETags clears cached ETags to force full refresh.
func (c *XboardClient) ResetETags() {
	c.configETag = ""
	c.userETag = ""
}

// IsConnected returns whether the node has successfully connected to xboard.
func (c *XboardClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// postJSON sends a POST request with auth fields to xboard.
func (c *XboardClient) postJSON(path string, payload map[string]interface{}) error {
	payload["token"] = c.token
	payload["node_id"] = c.nodeID
	if c.nodeType != "" {
		payload["node_type"] = c.nodeType
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	resp, err := c.doRequest("POST", path, body, "")
	if err != nil {
		return fmt.Errorf("post %s: %w", path, err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("status %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

func (c *XboardClient) doRequest(method, path string, body []byte, ifNoneMatch string) (*http.Response, error) {
	fullURL := c.baseURL + path

	var bodyReader io.Reader
	if method == "GET" {
		q := url.Values{}
		q.Set("token", c.token)
		q.Set("node_id", strconv.Itoa(c.nodeID))
		if c.nodeType != "" {
			q.Set("node_type", c.nodeType)
		}
		fullURL += "?" + q.Encode()
	} else if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		authOnly := map[string]interface{}{
			"token":   c.token,
			"node_id": c.nodeID,
		}
		if c.nodeType != "" {
			authOnly["node_type"] = c.nodeType
		}
		merged, _ := json.Marshal(authOnly)
		bodyReader = bytes.NewReader(merged)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	return resp, nil
}

// drainAndClose drains and closes the response body.
func drainAndClose(r io.ReadCloser) {
	if r == nil {
		return
	}
	io.Copy(io.Discard, r)
	r.Close()
}

// HandshakeResponse is returned by the /api/v2/server/handshake endpoint.
type HandshakeResponse struct {
	Version   string          `json:"version"`
	ExpiresIn int             `json:"expires_in"`
	Config    json.RawMessage `json:"config"`
	Users     json.RawMessage `json:"users"`
}

// NodeConfig represents the inbound configuration from xboard.
type NodeConfig struct {
	Protocol string          `json:"protocol"`
	Listen   string          `json:"listen"`
	Port     int             `json:"port"`
	Tag      string          `json:"tag"`
	Settings json.RawMessage `json:"settings"`
}

// User represents a user from xboard.
type User struct {
	ID         int64   `json:"id"`
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	Enable     bool    `json:"enable"`
	Flow       string  `json:"flow"`
	UUID       string  `json:"uuid"`
	Email      string  `json:"email"`
	Upload     int64   `json:"upload"`
	Download   int64   `json:"download"`
	Total      int64   `json:"total"`
	ExpiryTime int64   `json:"expiry_time"`
	UpSpeed    int64   `json:"up_speed"`
	DownSpeed  int64   `json:"down_speed"`
	Reset      int     `json:"reset"`
	IpLimit    int     `json:"ip_limit"`
	TgId       string  `json:"tg_id"`
	SubID      string  `json:"sub_id"`
	DelFlag    bool    `json:"del_flag"`
}

// UsersResponse wraps the user list.
type UsersResponse struct {
	Users []User `json:"users"`
}

// decodeWeak decodes a map into a struct with weak type conversion
// (string→int, bool→string, []interface{}→[]string, etc.)
func decodeWeak(input map[string]interface{}, output interface{}) error {
	outVal := reflect.ValueOf(output)
	if outVal.Kind() == reflect.Ptr {
		outVal = outVal.Elem()
	}
	if outVal.Kind() != reflect.Struct {
		return fmt.Errorf("output must be a struct")
	}
	for key, val := range input {
		field := findFieldByJSON(outVal, key)
		if !field.IsValid() {
			continue
		}
		setWeak(field, val)
	}
	return nil
}

// findFieldByJSON finds a struct field by its JSON tag name.
func findFieldByJSON(v reflect.Value, jsonTag string) reflect.Value {
	if v.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	// Strip ",string" or ",omitempty" suffixes for matching
	cleanTag := strings.Split(jsonTag, ",")[0]
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := field.Tag.Get("json")
		tagKey := strings.Split(tag, ",")[0]
		if tagKey == cleanTag {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

func setWeak(field reflect.Value, val interface{}) {
	if !field.IsValid() || !field.CanSet() {
		return
	}
	if val == nil {
		return
	}

	fieldType := field.Type()
	valType := reflect.TypeOf(val)

	// Same type
	if fieldType == valType {
		field.Set(reflect.ValueOf(val))
		return
	}

	switch fieldType.Kind() {
	case reflect.String:
		switch v := val.(type) {
		case string:
			field.SetString(v)
		case float64:
			field.SetString(strconv.FormatFloat(v, 'f', -1, 64))
		case int64:
			field.SetString(strconv.FormatInt(v, 10))
		case bool:
			field.SetString(strconv.FormatBool(v))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := val.(type) {
		case float64:
			field.SetInt(int64(v))
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				field.SetInt(i)
			}
		case int64:
			field.SetInt(v)
		}
	case reflect.Bool:
		switch v := val.(type) {
		case bool:
			field.SetBool(v)
		case string:
			field.SetBool(v == "true" || v == "1")
		case float64:
			field.SetBool(v != 0)
		}
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			if arr, ok := val.([]interface{}); ok {
				strs := make([]string, 0, len(arr))
				for _, x := range arr {
					if s, ok := x.(string); ok {
						strs = append(strs, s)
					}
				}
				field.Set(reflect.ValueOf(strs))
			}
		}
	}
}

package service

import (
	"encoding/json"
	"os"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util/common"
)

type UserService struct{}

// 面板管理员登录（保留）
func (s *UserService) GetFirstUser() (*model.User, error) {
	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateFirstUser(username string, password string) error {
	if username == "" {
		return common.NewError("username cannot be empty")
	} else if password == "" {
		return common.NewError("password cannot be empty")
	}
	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).First(user).Error
	if database.IsNotFound(err) {
		user.Username = username
		user.Password = password
		return db.Model(model.User{}).Create(user).Error
	} else if err != nil {
		return err
	}
	user.Username = username
	user.Password = password
	return db.Save(user).Error
}

func (s *UserService) Login(username string, password string, remoteIP string) (string, error) {
	user := s.CheckUser(username, password, remoteIP)
	if user == nil {
		return "", common.NewError("wrong user or password! IP: ", remoteIP)
	}
	return user.Username, nil
}

func (s *UserService) CheckUser(username string, password string, remoteIP string) *model.User {
	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).
		Where("username = ? and password = ?", username, password).
		First(user).Error
	if database.IsNotFound(err) {
		return nil
	} else if err != nil {
		logger.Warning("check user err:", err, " IP: ", remoteIP)
		return nil
	}
	return user
}

// GetUsers 从 sub 的 singbox.json 读取用户列表（只读）
// 保留面板登录用户，同时展示 sub 入站中的路由用户
func (s *UserService) GetUsers() (*[]model.User, error) {
	var users []model.User

	// 1. 面板管理员用户（从数据库读）
	db := database.GetDB()
	var adminUsers []model.User
	err := db.Model(model.User{}).Select("id,username,last_logins").Scan(&adminUsers).Error
	if err != nil && !database.IsNotFound(err) {
		return nil, err
	}
	users = append(users, adminUsers...)

	// 2. sub 入站路由用户（从 singbox.json 读）
	subUsers, err := s.GetSubUsers()
	if err == nil && len(subUsers) > 0 {
		users = append(users, subUsers...)
	}

	return &users, nil
}

// GetSubUsers 读取 /etc/sub/singbox.json 中的 inbound users
func (s *UserService) GetSubUsers() ([]model.User, error) {
	data, err := os.ReadFile("/etc/sub/singbox.json")
	if err != nil {
		return nil, err
	}

	var sb singboxRoot
	if err := json.Unmarshal(data, &sb); err != nil {
		return nil, err
	}

	var users []model.User
	uid := uint(1000) // 从 1000 开始避免和 admin id 冲突
	for _, inbound := range sb.Inbounds {
		for _, u := range inbound.Users {
			email := u.Email
			if email == "" {
				email = u.UUID
			}
			users = append(users, model.User{
				Id:       uid,
				Username: email,
			})
			uid++
		}
	}
	return users, nil
}

func (s *UserService) ChangePass(id string, oldPass string, newUser string, newPass string) error {
	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).Where("id = ? AND password = ?", id, oldPass).First(user).Error
	if err != nil || database.IsNotFound(err) {
		return err
	}
	user.Username = newUser
	user.Password = newPass
	return db.Save(user).Error
}

func (s *UserService) LoadTokens() ([]byte, error) {
	return []byte("[]"), nil // 订阅 Token 功能已禁用
}

func (s *UserService) GetUserTokens(username string) (*[]model.Tokens, error) {
	return &[]model.Tokens{}, nil
}

func (s *UserService) AddToken(username string, expiry int64, desc string) (string, error) {
	return "", common.NewError("token功能已禁用")
}

func (s *UserService) DeleteToken(id string) error {
	return common.NewError("token功能已禁用")
}

// sub singbox.json 结构（本地定义，不依赖 sing-box）
type singboxInbound struct {
	Tag   string           `json:"tag"`
	Users []singboxUser    `json:"users"`
}

type singboxUser struct {
	Email string `json:"email"`
	UUID  string `json:"uuid"`
	Flow  string `json:"flow,omitempty"`
}

type singboxRoot struct {
	Inbounds []singboxInbound `json:"inbounds"`
}

package service

import (
	"encoding/json"
	"os"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/db"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util/common"
)

type UserService struct{}

// GetFirstUser returns the first admin user.
func (s *UserService) GetFirstUser() (*model.User, error) {
	users := db.GetUsers()
	if len(users) == 0 {
		return nil, common.NewError("no users found")
	}
	return &model.User{
		Id:       users[0].Id,
		Username: users[0].Username,
		Password: users[0].Password,
	}, nil
}

// UpdateFirstUser updates the first admin user's username and password.
func (s *UserService) UpdateFirstUser(username string, password string) error {
	if username == "" {
		return common.NewError("username cannot be empty")
	} else if password == "" {
		return common.NewError("password cannot be empty")
	}
	cfg := db.Get()
	if len(cfg.Users) == 0 {
		cfg.Users = append(cfg.Users, db.User{Id: 1, Username: username, Password: password})
	} else {
		cfg.Users[0].Username = username
		cfg.Users[0].Password = password
	}
	db.Set(cfg)
	return database.SaveConfig()
}

func (s *UserService) Login(username string, password string, remoteIP string) (string, error) {
	user := s.CheckUser(username, password, remoteIP)
	if user == nil {
		return "", common.NewError("wrong user or password! IP: ", remoteIP)
	}
	return user.Username, nil
}

func (s *UserService) CheckUser(username string, password string, remoteIP string) *model.User {
	cfg := db.Get()
	for i := range cfg.Users {
		u := &cfg.Users[i]
		if u.Username == username && u.Password == password {
			return &model.User{
				Id:       u.Id,
				Username: u.Username,
				Password: u.Password,
			}
		}
	}
	logger.Warning("check user err: no match for username ", username, " IP: ", remoteIP)
	return nil
}

// GetUsers returns panel admin users plus sub inbound users.
func (s *UserService) GetUsers() (*[]model.User, error) {
	var users []model.User

	// 1. Panel admin users
	cfg := db.Get()
	for _, u := range cfg.Users {
		users = append(users, model.User{
			Id:         u.Id,
			Username:   u.Username,
			Password:   u.Password,
			LastLogins: u.LastLogins,
		})
	}

	// 2. Sub inbound users from singbox.json
	subUsers, err := s.GetSubUsers()
	if err == nil && len(subUsers) > 0 {
		users = append(users, subUsers...)
	}

	return &users, nil
}

// GetSubUsers reads inbound users from /etc/sub/singbox.json.
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
	uid := uint(1000)
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
	cfg := db.Get()
	for i := range cfg.Users {
		u := &cfg.Users[i]
		// Match by id string and old password
		if common.Itoa(int(u.Id)) == id && u.Password == oldPass {
			u.Username = newUser
			u.Password = newPass
			db.Set(cfg)
			return database.SaveConfig()
		}
	}
	return common.NewError("user not found or wrong password")
}

func (s *UserService) LoadTokens() ([]byte, error) {
	return []byte("[]"), nil // Token功能已禁用
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

// sub singbox.json structure (local definition)
type singboxInbound struct {
	Tag   string         `json:"tag"`
	Users []singboxUser  `json:"users"`
}

type singboxUser struct {
	Email string `json:"email"`
	UUID  string `json:"uuid"`
	Flow  string `json:"flow,omitempty"`
}

type singboxRoot struct {
	Inbounds []singboxInbound `json:"inbounds"`
}

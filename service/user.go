	package service

import (
	"github.com/pupmme/pupmsub/logger"
	"github.com/pupmme/pupmsub/util/common"
	"github.com/pupmme/pupmsub/db"
	"encoding/json"
	"os"

)

type UserService struct{}

// GetFirstUser returns the first admin user.
func (s *UserService) GetFirstUser() (*db.User, error) {
	users := db.GetUsers()
	if len(users) == 0 {
		return nil, common.NewError("no users found")
	}
	return &db.User{
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
	return db.SaveConfig()
}

func (s *UserService) Login(username string, password string, remoteIP string) (string, error) {
	user := s.CheckUser(username, password, remoteIP)
	if user == nil {
		return "", common.NewError("wrong user or password! IP: ", remoteIP)
	}
	return user.Username, nil
}

func (s *UserService) CheckUser(username string, password string, remoteIP string) *db.User {
	cfg := db.Get()
	for i := range cfg.Users {
		u := &cfg.Users[i]
		if u.Username == username && u.Password == password {
			return &db.User{
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
func (s *UserService) GetUsers() (*[]db.User, error) {
	var users []db.User

	// 1. Panel admin users
	cfg := db.Get()
	for _, u := range cfg.Users {
		users = append(users, db.User{
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
func (s *UserService) GetSubUsers() ([]db.User, error) {
	data, err := os.ReadFile("/etc/sub/singbox.json")
	if err != nil {
		return nil, err
	}

	var sb singboxRoot
	if err := json.Unmarshal(data, &sb); err != nil {
		return nil, err
	}

	var users []db.User
	uid := uint(1000)
	for _, inbound := range sb.Inbounds {
		for _, u := range inbound.Users {
			email := u.Email
			if email == "" {
				email = u.UUID
			}
			users = append(users, db.User{
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
			return db.SaveConfig()
		}
	}
	return common.NewError("user not found or wrong password")
}

// GetAllUsers returns all users (alias for GetUsers)
func (s *UserService) GetAllUsers() (*[]db.User, error) {
	return s.GetUsers()
}

// ChangePassword changes password for a user
func (s *UserService) ChangePassword(username string, oldPass string, newPass string) error {
	cfg := db.Get()
	for i := range cfg.Users {
		u := &cfg.Users[i]
		if u.Username == username && u.Password == oldPass {
			u.Password = newPass
			db.Set(cfg)
			return db.SaveConfig()
		}
	}
	return common.NewError("user not found or wrong password")
}

func (s *UserService) LoadTokens() ([]byte, error) {
	return []byte("[]"), nil // Token功能已禁用
}

func (s *UserService) GetUserTokens(username string) (*[]db.Tokens, error) {
	return &[]db.Tokens{}, nil
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

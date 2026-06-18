package service

import (
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// UserService 处理用户相关的业务逻辑。
// 业务校验、组合多次 DAO 调用、事务等都放在这一层。
type UserService struct {
	userDAO dao.UserDAO
}

// NewUserService 通过依赖注入构造 UserService。
func NewUserService(userDAO dao.UserDAO) *UserService {
	return &UserService{userDAO: userDAO}
}

// Create 创建用户。
func (s *UserService) Create(req *model.CreateUserRequest) (*model.User, error) {
	u := &model.User{
		Username: req.Username,
		Email:    req.Email,
	}
	if err := s.userDAO.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetByID 根据 ID 获取用户。
func (s *UserService) GetByID(id uint64) (*model.User, error) {
	return s.userDAO.GetByID(id)
}

// List 列出所有用户。
func (s *UserService) List() ([]*model.User, error) {
	return s.userDAO.List()
}

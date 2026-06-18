package dao

import (
	"errors"
	"sync"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// ErrNotFound 表示资源不存在。
var ErrNotFound = errors.New("record not found")

// UserDAO 抽象用户的持久化操作。
// 用接口隔离存储细节，后续接 gorm/redis 替换实现即可，service 层无感。
type UserDAO interface {
	Create(u *model.User) error
	GetByID(id uint64) (*model.User, error)
	List() ([]*model.User, error)
}

// memoryUserDAO 是基于内存的实现，仅用于脚手架阶段。
type memoryUserDAO struct {
	mu    sync.RWMutex
	seq   uint64
	users map[uint64]*model.User
}

// NewMemoryUserDAO 创建一个基于内存的 UserDAO 实现。
func NewMemoryUserDAO() UserDAO {
	return &memoryUserDAO{users: make(map[uint64]*model.User)}
}

func (d *memoryUserDAO) Create(u *model.User) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.seq++
	u.ID = d.seq
	u.CreatedAt = time.Now()
	d.users[u.ID] = u
	return nil
}

func (d *memoryUserDAO) GetByID(id uint64) (*model.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	u, ok := d.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (d *memoryUserDAO) List() ([]*model.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]*model.User, 0, len(d.users))
	for _, u := range d.users {
		list = append(list, u)
	}
	return list, nil
}

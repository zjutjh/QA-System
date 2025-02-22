package dao

import (
	"context"
	"time"

	"QA-System/internal/model"

	"github.com/jellydator/ttlcache/v3"
	"go.uber.org/zap"
)

// emailCache 用户邮箱缓存
var (
	emailCache *ttlcache.Cache[int, string]
	cacheTTL   = 30 * time.Minute
)

// InitializeCache 初始化用户邮箱缓存
func (d *Dao) InitializeCache() {
	emailCache = ttlcache.New(ttlcache.WithTTL[int, string](cacheTTL))
	go emailCache.Start()

	users := []model.User{}
	result := d.orm.Model(&model.User{}).Find(&users)

	if result.Error != nil {
		zap.L().Error("failed to cache user email", zap.Error(result.Error))
		return
	}

	for _, user := range users {
		emailCache.Set(user.ID, user.NotifyEmail, cacheTTL)
	}
}

// GetUserByUsername 根据用户名获取用户
func (d *Dao) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).First(&user)
	return &user, result.Error
}

// GetUserByID 根据用户ID获取用户
func (d *Dao) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).First(&user)
	return &user, result.Error
}

// CreateUser 创建新用户
func (d *Dao) CreateUser(ctx context.Context, user *model.User) error {
	result := d.orm.WithContext(ctx).Model(&model.User{}).Create(user)
	return result.Error
}

// UpdateUserPassword 更新用户密码
func (d *Dao) UpdateUserPassword(ctx context.Context, uid int, password string) error {
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("id = ?", uid).Update("password", password)
	return result.Error
}

// UpdateUserEmail 更新用户邮箱
func (d *Dao) UpdateUserEmail(ctx context.Context, uid int, email string) error {
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("id = ?", uid).Update("notify_email", email)
	if result.Error != nil {
		return result.Error
	}
	// 同步更新缓存
	emailCache.Set(uid, email, cacheTTL)
	return result.Error
}

// GetUserEmailByID 根据用户ID获取用户邮箱
func (d *Dao) GetUserEmailByID(ctx context.Context, uid int) (string, error) {
	// 尝试从缓存获取
	item := emailCache.Get(uid)
	if item != nil {
		return item.Value(), nil
	}

	// 缓存未命中，查询数据库
	user, err := d.GetUserByID(ctx, uid)
	if err != nil {
		return "", err
	}
	emailCache.Set(uid, user.NotifyEmail, cacheTTL)
	return user.NotifyEmail, nil
}

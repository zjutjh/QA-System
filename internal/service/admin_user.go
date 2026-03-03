package service

import (
	"QA-System/internal/model"
	"QA-System/internal/pkg/utils"
)

// GetAdminByUsername 根据用户名获取管理员
func GetAdminByUsername(username string) (*model.User, error) {
	user, err := d.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user.Password != "" {
		aesDecryptPassword(user)
	}
	return user, nil
}

// GetAdminByID 根据ID获取管理员
func GetAdminByID(id int) (*model.User, error) {
	user, err := d.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Password != "" {
		aesDecryptPassword(user)
	}
	return user, nil
}

// GetUserEmailByID 根据用户ID获取用户邮箱
func GetUserEmailByID(id int) (string, error) {
	email, err := d.GetUserEmailByID(ctx, id)
	if err != nil {
		return "", err
	}
	return email, nil
}

// IsAdminExist 判断管理员是否存在
func IsAdminExist(username string) error {
	_, err := d.GetUserByUsername(ctx, username)
	return err
}

// CreateAdmin 创建管理员
func CreateAdmin(user model.User) error {
	aesEncryptPassword(&user)
	err := d.CreateUser(ctx, &user)
	return err
}

// GetUserByName 根据用户名获取用户
func GetUserByName(username string) (*model.User, error) {
	user, err := d.GetUserByUsername(ctx, username)
	return user, err
}

// UpdateAdminPassword 更新管理员密码
func UpdateAdminPassword(id int, password string) error {
	encryptedPassword := utils.AesEncrypt(password)
	err := d.UpdateUserPassword(ctx, id, encryptedPassword)
	return err
}

// UpdateAdminEmail 更新管理员邮箱
func UpdateAdminEmail(id int, email string) error {
	err := d.UpdateUserEmail(ctx, id, email)
	return err
}

func aesDecryptPassword(user *model.User) {
	user.Password = utils.AesDecrypt(user.Password)
}

func aesEncryptPassword(user *model.User) {
	user.Password = utils.AesEncrypt(user.Password)
}

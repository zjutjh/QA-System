package service

import "strings"

// CreateQuestionPre 创建问题预先信息
func CreateQuestionPre(name string, value []string) error {
	// 将String[]类型转化为String,以逗号分隔
	pre := strings.Join(value, ",")
	err := d.CreateType(ctx, name, pre)
	return err
}

// GetQuestionPre 获取问题预先信息
func GetQuestionPre(name string) ([]string, error) {
	value, err := d.GetType(ctx, name)
	if err != nil {
		return nil, err
	}

	// 将预先信息转化为String[]类型
	pre := strings.Split(value, ",")
	return pre, nil
}

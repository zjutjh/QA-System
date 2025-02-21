package plugins

import (
	"errors"
	"fmt"

	"QA-System/internal/global/config"
	"QA-System/pkg/extension"

	"github.com/bytedance/gopkg/util/gopool"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// BetterEmailNotifier 插件需要的基本信息
type BetterEmailNotifier struct {
	smtpHost     string // SMTP服务器地址
	smtpPort     int    // SMTP服务器端口
	smtpUsername string // SMTP服务器用户名
	smtpPassword string // SMTP服务器密码
	from         string // 发件人地址
	workerNum    int    // 工作协程数量
}

var betterNotifier *BetterEmailNotifier

// init 注册插件
func init() {
	betterNotifier = &BetterEmailNotifier{
		workerNum: 20,
	}
	if err := betterNotifier.initialize(); err != nil {
		panic(fmt.Sprintf("Failed to initialize email_notifier: %v", err))
	}
	extension.GetDefaultManager().RegisterPlugin(betterNotifier)
}

// initialize 从配置文件中读取配置信息
func (p *BetterEmailNotifier) initialize() error {
	// 读取SMTP配置
	p.smtpHost = config.Config.GetString("email_notifier.smtp.host")
	p.smtpPort = config.Config.GetInt("email_notifier.smtp.port")
	p.smtpUsername = config.Config.GetString("email_notifier.smtp.username")
	p.smtpPassword = config.Config.GetString("email_notifier.smtp.password")
	p.from = config.Config.GetString("email_notifier.smtp.from")

	if p.smtpHost == "" || p.smtpUsername == "" || p.smtpPassword == "" || p.from == "" {
		return errors.New("invalid SMTP configuration, this may lead to email sending failure")
	}

	return nil
}

// GetMetadata 返回插件的元数据
func (p *BetterEmailNotifier) GetMetadata() extension.PluginMetadata {
	return extension.PluginMetadata{
		Name:        "BetterEmailNotifier",
		Version:     "0.1.0",
		Author:      "SituChengxiang, Copilot, Qwen2.5, DeepSeek",
		Description: "Send email notifications for new survey responses",
	}
}

// Execute 启动插件，现在只需要初始化
func (p *BetterEmailNotifier) Execute() error {
	zap.L().Info("BetterEmailNotifier started", zap.Int("workers", p.workerNum))
	return nil
}

func BetterEmailNotify(info any) error {
	// 初始化 gopool
	pool := gopool.NewPool("examplePool", 10, gopool.NewConfig())
	pool.Go(func() {
		// 从消息中提取数据
		recipient, ok := info.(map[string]any)["creator_email"].(string)
		if !ok {
			zap.L().Error("invalid creator_email type")
			return
		}

		title, ok := info.(map[string]any)["survey_title"].(string)
		if !ok {
			zap.L().Error("invalid survey_title type")
			return
		}
		// 准备邮件数据
		data := map[string]any{
			"title":     title,
			"recipient": recipient,
		}
		err := betterNotifier.SendEmail(data)
		if err != nil {
			zap.L().Error("Failed to send email", zap.Error(err))
		}
	})
	return nil
}

// SendEmail 发送邮件
func (p *BetterEmailNotifier) SendEmail(data map[string]any) error {
	// 检查收件人
	recipient, ok := data["recipient"].(string)
	if !ok || recipient == "" {
		zap.L().Info("Recipient email is empty, skip current sending email")
		return nil
	}

	title, ok := data["title"].(string)
	if !ok {
		return errors.New("invalid title type")
	}

	// 创建邮件
	m := gomail.NewMessage()
	m.SetHeader("From", p.from)
	m.SetHeader("To", recipient)
	m.SetAddressHeader("Cc", p.from, "QA-System")
	m.SetHeader("Subject", fmt.Sprintf("您的问卷\"%s\"收到了新回复", title))
	m.SetBody("text/plain", fmt.Sprintf("您的问卷\"%s\"收到了新回复，请及时查收。", title))

	// 发送邮件
	d := gomail.NewDialer(p.smtpHost, p.smtpPort, p.smtpUsername, p.smtpPassword)
	if err := d.DialAndSend(m); err != nil {
		zap.L().Error("Failed to send email",
			zap.String("recipient", recipient),
			zap.String("title", title),
			zap.Error(err))
		return err
	}

	zap.L().Info("Email sent successfully",
		zap.String("recipient", recipient),
		zap.String("title", title))
	return nil
}

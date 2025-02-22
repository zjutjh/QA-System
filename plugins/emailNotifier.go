package plugins

import (
	"context"
	"errors"
	"fmt"
	"time"

	"QA-System/internal/global/config"
	"QA-System/internal/pkg/redis"
	"QA-System/pkg/extension"

	redisv9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// EmailNotifier 插件需要的基本信息
type EmailNotifier struct {
	smtpHost     string                // SMTP服务器地址
	smtpPort     int                   // SMTP服务器端口
	smtpUsername string                // SMTP服务器用户名
	smtpPassword string                // SMTP服务器密码
	from         string                // 发件人地址
	streamName   string                // stream的名称
	groupName    string                // stream的消费者组名称
	consumerOld  string                // 处理pending消息的消费者
	consumerNew  string                // 处理新消息的消费者
	workerNum    int                   // 工作协程数量
	jobChan      chan redisv9.XMessage // 任务通道
}

// init 注册插件
func init() {
	notifier := &EmailNotifier{
		consumerOld: "consumerOld",
		consumerNew: "consumerNew",
	}
	if err := notifier.initialize(); err != nil {
		panic(fmt.Sprintf("Failed to initialize emailnotifier: %v", err))
	}
	extension.GetDefaultManager().RegisterPlugin(notifier)
}

// initialize 从配置文件中读取配置信息
func (p *EmailNotifier) initialize() error {
	// 读取SMTP配置
	p.smtpHost = config.Config.GetString("emailnotifier.smtp.host")
	p.smtpPort = config.Config.GetInt("emailnotifier.smtp.port")
	p.smtpUsername = config.Config.GetString("emailnotifier.smtp.username")
	p.smtpPassword = config.Config.GetString("emailnotifier.smtp.password")
	p.from = config.Config.GetString("emailnotifier.smtp.from")

	if p.smtpHost == "" || p.smtpUsername == "" || p.smtpPassword == "" || p.from == "" {
		return errors.New("invalid SMTP configuration, this may lead to email sending failure")
	}

	// 读取Stream配置（已准备弃用）
	p.streamName = config.Config.GetString("redis.stream.name")
	p.groupName = config.Config.GetString("redis.stream.group")

	if p.streamName == "" || p.groupName == "" {
		zap.L().Warn("Stream name or group name is empty, email notifier will not work")
	}

	// 读取工作协程配置
	p.workerNum = config.Config.GetInt("emailnotifier.worker.num")
	if p.workerNum <= 0 {
		p.workerNum = 3 // 默认3个工作协程
	}
	p.jobChan = make(chan redisv9.XMessage, p.workerNum*2)
	return nil
}

// GetMetadata 返回插件的元数据
func (p *EmailNotifier) GetMetadata() extension.PluginMetadata {
	_ = p
	return extension.PluginMetadata{
		Name:        "emailNotifier",
		Version:     "0.1.0",
		Author:      "SituChengxiang, Copilot, Qwen2.5, DeepSeek",
		Description: "Send email notifications for new survey responses",
	}
}

// Execute 启动消费者
func (p *EmailNotifier) Execute() error {
	fmt.Println("Another version of the email notifier has been released, you can change to that one as this one relies on redis stream")
	ctx := context.Background()
	zap.L().Info("Email notifier started", zap.Int("workers", p.workerNum))

	// 启动工作协程池
	for i := 0; i < p.workerNum; i++ {
		go p.startWorker(ctx, i)
	}

	// 启动两个消费者
	go p.consumeOld(ctx)
	go p.consumeNew(ctx)

	select {}
}

// 工作协程处理函数
func (p *EmailNotifier) startWorker(ctx context.Context, workerID int) {
	zap.L().Info("Worker started", zap.Int("workerID", workerID))

	for msg := range p.jobChan {
		// 处理消息
		if err := p.handleMessage(ctx, msg); err != nil {
			zap.L().Error("Failed to handle message",
				zap.Int("workerID", workerID),
				zap.String("messageID", msg.ID),
				zap.Error(err))
			continue
		}

		// 确认消息处理完成
		if err := redis.RedisClient.XAck(ctx, p.streamName, p.groupName, msg.ID).Err(); err != nil {
			zap.L().Error("Failed to ack message",
				zap.Int("workerID", workerID),
				zap.String("messageID", msg.ID),
				zap.Error(err))
		}
	}
}

// consumeOld 处理 pending 消息
func (p *EmailNotifier) consumeOld(ctx context.Context) {
	for {
		// 使用 XAUTOCLAIM 自动认领 pending 消息
		messages, _, err := redis.RedisClient.XAutoClaim(ctx, &redisv9.XAutoClaimArgs{
			Stream:   p.streamName,
			Group:    p.groupName,
			Consumer: p.consumerOld,
			MinIdle:  time.Minute * 5,
			Start:    "0-0",
			Count:    10,
		}).Result()
		if err != nil {
			zap.L().Error("Failed to auto claim messages in consumerOld", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		// 分发消息到工作协程
		for _, msg := range messages {
			p.jobChan <- msg
		}

		time.Sleep(time.Second)
	}
}

// consumeNew 处理新消息
func (p *EmailNotifier) consumeNew(ctx context.Context) {
	for {
		// 读取新消息
		streams, err := redis.RedisClient.XReadGroup(ctx, &redisv9.XReadGroupArgs{
			Group:    p.groupName,
			Consumer: p.consumerNew,
			Streams:  []string{p.streamName, ">"},
			Count:    10,
			Block:    time.Second * 2,
		}).Result()

		if err != nil && err != redisv9.Nil {
			zap.L().Error("Failed to read new messages in consumerNew", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if len(streams) > 0 && len(streams[0].Messages) > 0 {
			for _, msg := range streams[0].Messages {
				p.jobChan <- msg
			}
		}

		time.Sleep(time.Second)
	}
}

// handleMessage 处理消息，从信息里提取 title 和 creator_email
func (p *EmailNotifier) handleMessage(ctx context.Context, message redisv9.XMessage) error {
	_ = ctx
	// 从消息中提取数据
	title, ok := message.Values["survey_title"].(string)
	if !ok {
		return fmt.Errorf("invalid survey_title %s in message", message)
	}

	recipient, ok := message.Values["creator_email"].(string)
	if !ok {
		return fmt.Errorf("invalid creator_email %s in message", message)
	}

	// 准备邮件数据
	data := map[string]any{
		"title":     title,
		"recipient": recipient,
		"id":        message.ID,
	}

	zap.L().Info("Sending email",
		zap.String("recipient", recipient),
		zap.String("title", title))

	return p.sendEmail(data)
}

// sendEmail 发送邮件
func (p *EmailNotifier) sendEmail(data map[string]any) error {
	// 创建一个带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用channel来控制超时
	done := make(chan error, 1)

	go func() {
		// 创建邮件
		m := gomail.NewMessage()
		m.SetHeader("From", p.from)
		m.SetHeader("To", data["recipient"].(string))
		m.SetAddressHeader("Cc", p.from, "QA-System")
		m.SetHeader("Subject", fmt.Sprintf("您的问卷\"%s\"收到了新回复", data["title"].(string)))
		m.SetBody("text/plain", fmt.Sprintf("您的问卷\"%s\"收到了新回复，请及时查收。", data["title"].(string)))

		if data["recipient"].(string) == "" {
			zap.L().Info("Recipient email is empty, skip current sending email")
			done <- nil
		}

		// 发送邮件
		d := gomail.NewDialer(p.smtpHost, p.smtpPort, p.smtpUsername, p.smtpPassword)
		if err := d.DialAndSend(m); err != nil {
			zap.L().Error("Failed to send email", zap.Error(err))
			done <- err // 将错误发送到channel
			return
		}
		done <- nil // 发送nil表示成功
	}()

	// 等待邮件发送完成或超时
	select {
	case err := <-done:
		if err != nil {
			zap.L().Error("Failed to send email", zap.Error(err))
			return err
		}
		// // 发送成功后确认消息，用QA系统里的redis包打包里的AckMessage函数，参数少一点
		// err = redis.AckMessage(ctx, data["id"].(string))
		if err != nil {
			zap.L().Warn("Failed to ack message", zap.Error(err))
		}
		return nil
	case <-ctx.Done():
		return errors.New("send email timeout after 10 seconds")
	}
}

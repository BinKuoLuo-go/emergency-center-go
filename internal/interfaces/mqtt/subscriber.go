/**
@Time : 2026/09/17 15:41
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package mqtt

import (
	"context"
	"encoding/json"
	alarmapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/alarm"
	devicestatusapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/devicestatus"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/alarm"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/devicestatus"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/config"
	applog "github.com/BinKuoLuo-go/emergency-center-go/pkg/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

const (
	defaultAlarmTopic  = "emergency/alarm/#"
	defaultStatusTopic = "emergency/status/#"
)

// Subscriber MQTT 订阅器：告警 + 设备状态。
type Subscriber struct {
	cfg             config.MQTTConfig
	alarmSvc        *alarmapp.Service
	deviceStatusSvc *devicestatusapp.Service
	client          mqtt.Client
}

func NewSubscriber(cfg config.MQTTConfig, alarmSvc *alarmapp.Service, deviceStatusSvc *devicestatusapp.Service) *Subscriber {
	return &Subscriber{cfg: cfg, alarmSvc: alarmSvc, deviceStatusSvc: deviceStatusSvc}
}

// Start 连接 Broker 并订阅主题。未启用时直接返回 nil。
func (s *Subscriber) Start() error {
	if !s.cfg.Enabled {
		applog.Info("MQTT 订阅未启用，跳过")
		return nil
	}
	if s.cfg.Broker == "" {
		applog.Warn("MQTT 已启用但未配置 broker，跳过")
		return nil
	}

	alarmTopic := s.cfg.Topic
	if alarmTopic == "" {
		alarmTopic = defaultAlarmTopic
	}
	statusTopic := s.cfg.StatusTopic
	if statusTopic == "" {
		statusTopic = defaultStatusTopic
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(s.cfg.Broker)
	opts.SetClientID(s.cfg.ClientID)
	if s.cfg.Username != "" {
		opts.SetUsername(s.cfg.Username)
	}
	if s.cfg.Password != "" {
		opts.SetPassword(s.cfg.Password)
	}
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		if tok := c.Subscribe(alarmTopic, 1, s.onAlarmMessage); tok.Wait() && tok.Error() != nil {
			applog.Error("MQTT 订阅失败", "topic", alarmTopic, "err", tok.Error())
		} else {
			applog.Info("MQTT 订阅成功", "topic", alarmTopic)
		}
		if tok := c.Subscribe(statusTopic, 1, s.onStatusMessage); tok.Wait() && tok.Error() != nil {
			applog.Error("MQTT 订阅失败", "topic", statusTopic, "err", tok.Error())
		} else {
			applog.Info("MQTT 订阅成功", "topic", statusTopic)
		}
	})

	s.client = mqtt.NewClient(opts)

	// 连接（带有限重试，避免启动即失败导致进程退出）
	var err error
	for i := 0; i < 5; i++ {
		tok := s.client.Connect()
		tok.Wait()
		if err = tok.Error(); err == nil {
			break
		}
		applog.Warn("MQTT 连接失败，重试中", "attempt", i+1, "err", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return err
	}
	applog.Info("MQTT 已连接", "broker", s.cfg.Broker, "clientId", s.cfg.ClientID)
	return nil
}

// Stop 断开连接（优雅退出）。
func (s *Subscriber) Stop() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(500)
	}
}

// onAlarmMessage 处理告警消息。
// 未超员周期上报(normal) 同时刷新在线状态(Redis)并落 MySQL；报警/销警(alarm/cancel) 落 MySQL。
func (s *Subscriber) onAlarmMessage(_ mqtt.Client, msg mqtt.Message) {
	var a alarm.Alarm
	if err := json.Unmarshal(msg.Payload(), &a); err != nil {
		applog.Warn("MQTT 告警消息解析失败", "topic", msg.Topic(), "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 未超员周期上报：同时刷新在线状态（Redis），并照常落 MySQL
	if a.ReportType == "normal" {
		if err := s.deviceStatusSvc.MarkOnline(ctx, a.DeviceID, a.CompanyCode, a.CompanyName); err != nil {
			applog.Warn("在线状态更新失败", "deviceId", a.DeviceID, "err", err)
		}
	}

	// 所有告警消息（心跳/报警/销警）均落 MySQL
	if err := s.alarmSvc.Record(ctx, &a); err != nil {
		applog.Error("告警入库失败", "id", a.ID, "err", err)
		return
	}
	applog.Info("告警入库成功",
		"id", a.ID,
		"companyCode", a.CompanyCode,
		"deviceId", a.DeviceID,
		"status", a.AlarmStatus,
		"reportType", a.ReportType,
	)
}

// onStatusMessage 处理设备状态消息：覆盖写最新资源状态到 Redis（不落 MySQL）。
func (s *Subscriber) onStatusMessage(_ mqtt.Client, msg mqtt.Message) {
	var d devicestatus.DeviceStatus
	if err := json.Unmarshal(msg.Payload(), &d); err != nil {
		applog.Warn("MQTT 设备状态消息解析失败", "topic", msg.Topic(), "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.deviceStatusSvc.MarkStatus(ctx, &d); err != nil {
		applog.Error("设备状态更新失败", "deviceId", d.DeviceID, "err", err)
		return
	}
	applog.Info("设备状态已更新",
		"deviceId", d.DeviceID,
		"cpuPercent", d.CpuPercent,
		"memPercent", d.MemPercent,
		"diskPercent", d.DiskPercent,
	)
}

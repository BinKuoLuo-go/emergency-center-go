/**
@Time : 2026/09/24 11:17
@Author: FangYao( 方少、)
@Description: 上报省平台 客户端
@Email: fy20030315@163.com
*/

package province

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/province"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/config"
	applog "github.com/BinKuoLuo-go/emergency-center-go/pkg/log"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client 省平台上报客户端。
type Client struct {
	cfg           config.ProvinceConfig
	tokenProvider province.TokenProvider
	hc            *http.Client
}

// NewClient 构建省平台客户端。tokenProvider 提供 token + AES 密钥/IV。
func NewClient(cfg config.ProvinceConfig, tokenProvider province.TokenProvider) *Client {
	timeout := cfg.TimeoutSec
	if timeout <= 0 {
		timeout = 10
	}
	return &Client{
		cfg:           cfg,
		tokenProvider: tokenProvider,
		hc:            &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Report 上报到省平台
func (c *Client) Report(ctx context.Context, batchID string, items []province.GatherAlarm) error {
	if len(items) == 0 {
		return nil
	}

	// 业务数据
	plain, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("序列化业务数据失败: %w", err)
	}

	// 获取凭据
	cred, err := c.tokenProvider.Get(ctx)
	if err != nil {
		return fmt.Errorf("获取省平台凭据失败: %w", err)
	}

	// AES-GCM 加密
	enc, err := encryptAESGCM(plain, cred.AESKey, cred.AESIV)
	if err != nil {
		return err
	}

	// 外层报文
	outer := struct {
		BatchID string `json:"batchId"`
		Data    string `json:"data"`
	}{BatchID: batchID, Data: enc}
	body, err := json.Marshal(outer)
	if err != nil {
		return fmt.Errorf("序列化外层报文失败: %w", err)
	}

	// 组装请求
	url := c.cfg.BaseURL + c.cfg.Endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=utf8")
	req.Header.Set("Authorization", cred.Token)

	// 发送
	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("省平台上报请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("省平台上报返回非 2xx: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	applog.Info("省平台上报成功",
		"batchId", batchID,
		"count", len(items),
		"status", resp.StatusCode,
	)
	return nil
}

// 数据AES加密 gcm模式
func encryptAESGCM(jsonBytes []byte, keyHex, ivHex string) (string, error) {
	if keyHex == "" || ivHex == "" {
		applog.Error("省平台AES密钥和IV未配置")
		return "", fmt.Errorf("省平台AES密钥和IV未配置")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		applog.Error("AES Key解码失败", "err", err)
		return "", fmt.Errorf("AES密钥解码失败:%w", err)
	}
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		applog.Error("AES IV解码失败", "err", err)
		return "", fmt.Errorf("AES IV解码失败:%w", err)
	}
	if len(key) != 32 {
		applog.Error("AES-256 密钥长度错误", "want", 32, "got", len(key))
		return "", fmt.Errorf("AES-256 密钥必须为 32 字节，当前 %d 字节", len(key))
	}
	if len(iv) != 16 {
		applog.Error("AES IV 长度错误", "want", 16, "got", len(iv))
		return "", fmt.Errorf("AES IV 必须为 16 字节，当前 %d 字节", len(iv))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		applog.Error("构造AES Cipher失败", "err", err)
		return "", fmt.Errorf("构造AES Cipher失败:%w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		applog.Errorf("构造GCM失败", "err", err)
		return "", fmt.Errorf("构造GCM失败:%w", err)
	}
	sealed := gcm.Seal(nil, iv, jsonBytes, []byte{})
	return base64.StdEncoding.EncodeToString(sealed), nil

}

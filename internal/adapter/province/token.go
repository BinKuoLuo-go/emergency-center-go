/**
@Time : 2026/09/24 14:21
@Author: FangYao( 方少、)
@Description: 省平台访问凭据token + AES 密钥/IV提供器。
@Email: fy20030315@163.com
*/

package province

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/province"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/config"
	applog "github.com/BinKuoLuo-go/emergency-center-go/pkg/log"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RemoteTokenProvider 通过省平台 token 接口动态获取凭据（token + AES 密钥/IV）。
// TODO: 待省平台提供 token 获取接口规范后，按实际路径/请求参数/响应字段补齐 requestBody 与 parseResp。
// 当前实现为通用占位：POST token_url，从响应 JSON 中解析 token / aesKey / aesIv / expireAt 字段。
type RemoteTokenProvider struct {
	url string
	key string // 配置的 AES key（接口未返回时回退用）
	iv  string // 配置的 AES IV（接口未返回时回退用）
	hc  *http.Client

	mu     sync.Mutex
	cached *province.Credential
}

func NewRemoteTokenProvider(cfg config.ProvinceConfig) *RemoteTokenProvider {
	timeout := cfg.TimeoutSec
	if timeout <= 0 {
		timeout = 10
	}
	return &RemoteTokenProvider{
		url: strings.TrimSpace(cfg.TokenURL),
		key: cfg.AESKey,
		iv:  cfg.AESIV,
		hc:  &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Get 获取凭据，带缓存：未到过期时间直接返回缓存，否则重新请求。
func (p *RemoteTokenProvider) Get(ctx context.Context) (*province.Credential, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	// 提前 5 分钟刷新，避免临界过期；ExpireAt 零值视为长期有效，不刷新
	if p.cached != nil {
		if p.cached.ExpireAt.IsZero() || now.Add(5*time.Minute).Before(p.cached.ExpireAt) {
			return p.cached, nil
		}
	}

	if p.url == "" {
		return nil, fmt.Errorf("省平台 token 接口未配置")
	}

	// TODO: 待省平台提供接口规范后，按实际请求参数构造 requestBody
	reqBody := strings.NewReader("")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("构造 token 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取省平台 token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("获取省平台 token 返回非 2xx: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	cred, err := parseTokenResp(body)
	if err != nil {
		return nil, err
	}
	// 接口未返回 AES 密钥/IV 时回退用配置值
	if cred.AESKey == "" {
		cred.AESKey = p.key
	}
	if cred.AESIV == "" {
		cred.AESIV = p.iv
	}
	p.cached = cred
	applog.Info("省平台 token 获取成功", "expireAt", cred.ExpireAt.Format(time.RFC3339))
	return cred, nil
}

// parseTokenResp 解析 token 接口响应（字段名待省平台确认，此处使用常见命名占位）。
// TODO: 按省平台实际响应结构调整字段名。
func parseTokenResp(body []byte) (*province.Credential, error) {
	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token    string `json:"token"`
			AESKey   string `json:"aesKey"`
			AESIV    string `json:"aesIv"`
			ExpireAt string `json:"expireAt"` // 过期时间，支持 RFC3339 或毫秒时间戳
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析 token 响应失败: %w", err)
	}
	if raw.Data.Token == "" {
		return nil, fmt.Errorf("token 响应缺少 token 字段: %s", strings.TrimSpace(string(body)))
	}

	cred := &province.Credential{
		Token:  raw.Data.Token,
		AESKey: raw.Data.AESKey,
		AESIV:  raw.Data.AESIV,
	}
	if raw.Data.ExpireAt != "" {
		cred.ExpireAt = parseExpireAt(raw.Data.ExpireAt)
	}
	return cred, nil
}

// parseExpireAt 解析过期时间：优先 RFC3339，其次毫秒时间戳。
func parseExpireAt(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	var ms int64
	if _, err := fmt.Sscanf(s, "%d", &ms); err == nil {
		return time.UnixMilli(ms)
	}
	return time.Time{}
}

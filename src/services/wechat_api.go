package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// WeChatAPIService provides WeChat API integration
type WeChatAPIService struct {
	baseURL    string
	appID      string
	appSecret  string
	httpClient *http.Client
}

// NewWeChatAPIService creates a new WeChat API service
func NewWeChatAPIService(appID, appSecret string) *WeChatAPIService {
	return &WeChatAPIService{
		baseURL:   "https://api.weixin.qq.com",
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccessToken 获取微信访问令牌
func (w *WeChatAPIService) GetAccessToken() (string, error) {
	url := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		w.baseURL, w.appID, w.appSecret)

	resp, err := w.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("WeChat API error: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	log.Printf("Successfully obtained WeChat access token")
	return tokenResp.AccessToken, nil
}

// SendTemplateMessage 发送模板消息
func (w *WeChatAPIService) SendTemplateMessage(accessToken, openID, templateID string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/cgi-bin/message/template/send?access_token=%s", w.baseURL, accessToken)

	message := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"data":        data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := w.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send template message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var sendResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &sendResp); err != nil {
		return fmt.Errorf("failed to parse send response: %w", err)
	}

	if sendResp.ErrCode != 0 {
		return fmt.Errorf("WeChat API error: %d - %s", sendResp.ErrCode, sendResp.ErrMsg)
	}

	log.Printf("Successfully sent template message to %s", openID)
	return nil
}

// GetUserInfo 获取用户信息
func (w *WeChatAPIService) GetUserInfo(accessToken, openID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/cgi-bin/user/info?access_token=%s&openid=%s", w.baseURL, accessToken, openID)

	resp, err := w.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	log.Printf("Successfully got user info for %s", openID)
	return userInfo, nil
}

// ValidateAPIToken 验证API令牌是否有效
func (w *WeChatAPIService) ValidateAPIToken() error {
	_, err := w.GetAccessToken()
	if err != nil {
		return fmt.Errorf("WeChat API token validation failed: %w", err)
	}

	log.Println("WeChat API token validation successful")
	return nil
}

// GetBaseURL returns the base URL for WeChat API
func (w *WeChatAPIService) GetBaseURL() string {
	return w.baseURL
}

// GetAppID returns the app ID
func (w *WeChatAPIService) GetAppID() string {
	return w.appID
}

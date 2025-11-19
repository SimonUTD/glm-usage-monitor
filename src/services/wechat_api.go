package services

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// WeChatConfig represents WeChat API configuration
type WeChatConfig struct {
	AppID          string `json:"app_id"`
	AppSecret      string `json:"app_secret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	AccessToken    string `json:"access_token"`
	ExpiresAt      int64  `json:"expires_at"`
}

// WeChatUserService handles WeChat user operations
type WeChatUserService struct {
	config *WeChatConfig
	client *http.Client
}

// WeChatOAuthService handles WeChat OAuth operations
type WeChatOAuthService struct {
	config *WeChatConfig
	client *http.Client
}

// WeChatAPIService provides comprehensive WeChat API integration
type WeChatAPIService struct {
	config              *WeChatConfig
	userService         *WeChatUserService
	oauthService        *WeChatOAuthService
	client              *http.Client
	notificationService *NotificationService
}

// WeChatUser represents WeChat user information
type WeChatUser struct {
	Subscribe     int    `json:"subscribe"`
	OpenID        string `json:"openid"`
	Nickname      string `json:"nickname"`
	Sex           int    `json:"sex"`
	Language      string `json:"language"`
	City          string `json:"city"`
	Province      string `json:"province"`
	Country       string `json:"country"`
	HeadImgURL    string `json:"headimgurl"`
	SubscribeTime int64  `json:"subscribe_time"`
	UnionID       string `json:"unionid"`
	Remark        string `json:"remark"`
	GroupID       int    `json:"groupid"`
}

// WeChatOAuthResponse represents OAuth response
type WeChatOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

// WeChatUserInfo represents user info from OAuth
type WeChatUserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

// WeChatMessage represents WeChat message
type WeChatMessage struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	MsgID        int64  `xml:"MsgId"`
}

// WeChatReplyMessage represents WeChat reply message
type WeChatReplyMessage struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
}

// AccessTokenResponse represents access token response
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// ErrorResponse represents WeChat API error response
type ErrorResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// NewWeChatAPIService creates a new WeChat API service
func NewWeChatAPIService(config *WeChatConfig, notificationService *NotificationService) *WeChatAPIService {
	if config == nil {
		config = &WeChatConfig{}
	}

	service := &WeChatAPIService{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		notificationService: notificationService,
	}

	service.userService = &WeChatUserService{
		config: config,
		client: service.client,
	}

	service.oauthService = &WeChatOAuthService{
		config: config,
		client: service.client,
	}

	return service
}

// SetConfig updates WeChat configuration
func (w *WeChatAPIService) SetConfig(config *WeChatConfig) {
	w.config = config
	w.userService.config = config
	w.oauthService.config = config
}

// GetAccessToken retrieves access token from WeChat API
func (w *WeChatAPIService) GetAccessToken() (string, error) {
	// Check if current token is still valid
	if w.config.AccessToken != "" && w.config.ExpiresAt > time.Now().Unix() {
		return w.config.AccessToken, nil
	}

	// Request new access token
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		w.config.AppID, w.config.AppSecret)

	resp, err := w.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var tokenResp AccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse access token response: %w", err)
	}

	// Check for error
	if tokenResp.AccessToken == "" {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.ErrCode != 0 {
			return "", fmt.Errorf("WeChat API error: %d - %s", errorResp.ErrCode, errorResp.ErrMsg)
		}
		return "", fmt.Errorf("invalid access token response")
	}

	// Update config
	w.config.AccessToken = tokenResp.AccessToken
	w.config.ExpiresAt = time.Now().Unix() + int64(tokenResp.ExpiresIn) - 300 // 5 minutes buffer

	log.Printf("WeChat access token refreshed, expires at: %d", w.config.ExpiresAt)
	return tokenResp.AccessToken, nil
}

// GetUserInfo retrieves user information by OpenID
func (w *WeChatAPIService) GetUserInfo(openID string) (*WeChatUser, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN",
		accessToken, openID)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var user WeChatUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	// Check for error
	if user.OpenID == "" {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.ErrCode != 0 {
			return nil, fmt.Errorf("WeChat API error: %d - %s", errorResp.ErrCode, errorResp.ErrMsg)
		}
		return nil, fmt.Errorf("invalid user info response")
	}

	return &user, nil
}

// GetOAuthURL generates OAuth URL for user authorization
func (w *WeChatAPIService) GetOAuthURL(redirectURI, state string) string {
	baseURL := "https://open.weixin.qq.com/connect/oauth2/authorize"
	params := url.Values{}
	params.Add("appid", w.config.AppID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "snsapi_userinfo")
	params.Add("state", state)

	return fmt.Sprintf("%s?%s#wechat_redirect", baseURL, params.Encode())
}

// GetOAuthAccessToken exchanges code for access token
func (w *WeChatAPIService) GetOAuthAccessToken(code string) (*WeChatOAuthResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		w.config.AppID, w.config.AppSecret, code)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var oauthResp WeChatOAuthResponse
	if err := json.Unmarshal(body, &oauthResp); err != nil {
		return nil, fmt.Errorf("failed to parse OAuth response: %w", err)
	}

	// Check for error
	if oauthResp.AccessToken == "" {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.ErrCode != 0 {
			return nil, fmt.Errorf("WeChat OAuth error: %d - %s", errorResp.ErrCode, errorResp.ErrMsg)
		}
		return nil, fmt.Errorf("invalid OAuth response")
	}

	return &oauthResp, nil
}

// GetOAuthUserInfo retrieves user info using OAuth access token
func (w *WeChatAPIService) GetOAuthUserInfo(accessToken, openID string) (*WeChatUserInfo, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		accessToken, openID)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo WeChatUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse OAuth user info response: %w", err)
	}

	// Check for error
	if userInfo.OpenID == "" {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.ErrCode != 0 {
			return nil, fmt.Errorf("WeChat OAuth error: %d - %s", errorResp.ErrCode, errorResp.ErrMsg)
		}
		return nil, fmt.Errorf("invalid OAuth user info response")
	}

	return &userInfo, nil
}

// SendMessage sends template message to user
func (w *WeChatAPIService) SendMessage(openID, templateID string, data map[string]interface{}) error {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return err
	}

	message := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"data":        data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)
	resp, err := w.client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse send message response: %w", err)
	}

	// Check for error
	if errCode, ok := result["errcode"].(float64); ok && errCode != 0 {
		errMsg, _ := result["errmsg"].(string)
		return fmt.Errorf("WeChat send message error: %.0f - %s", errCode, errMsg)
	}

	log.Printf("Message sent successfully to user %s", openID)
	return nil
}

// ValidateSignature validates WeChat signature for webhook
func (w *WeChatAPIService) ValidateSignature(signature, timestamp, nonce string) bool {
	if w.config.Token == "" {
		return false
	}

	// Sort parameters
	params := []string{w.config.Token, timestamp, nonce}
	sort.Strings(params)

	// Generate signature
	str := strings.Join(params, "")
	h := hmac.New(sha1.New, []byte(w.config.Token))
	h.Write([]byte(str))
	expectedSignature := fmt.Sprintf("%x", h.Sum(nil))

	return expectedSignature == signature
}

// ParseMessage parses incoming WeChat message
func (w *WeChatAPIService) ParseMessage(body []byte) (*WeChatMessage, error) {
	var message WeChatMessage
	if err := xml.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("failed to parse WeChat message: %w", err)
	}
	return &message, nil
}

// CreateReplyMessage creates reply message
func (w *WeChatAPIService) CreateReplyMessage(toUser, fromUser, content string) *WeChatReplyMessage {
	return &WeChatReplyMessage{
		ToUserName:   toUser,
		FromUserName: fromUser,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      content,
	}
}

// SendSyncNotification sends sync completion notification via WeChat
func (w *WeChatAPIService) SendSyncNotification(openID, billingMonth string, syncedCount, totalCount int, success bool) error {
	if w.notificationService != nil {
		title := "账单同步通知"
		var message string
		var notificationType NotificationType

		if success {
			message = fmt.Sprintf("账单月份 %s 同步完成，成功同步 %d/%d 条记录", billingMonth, syncedCount, totalCount)
			notificationType = NotificationTypeSuccess
		} else {
			message = fmt.Sprintf("账单月份 %s 同步失败，请检查配置后重试", billingMonth)
			notificationType = NotificationTypeError
		}

		w.notificationService.AddNotification(notificationType, title, message, map[string]interface{}{
			"open_id":       openID,
			"billing_month": billingMonth,
			"synced_count":  syncedCount,
			"total_count":   totalCount,
			"success":       success,
			"type":          "wechat_sync_notification",
		})
	}

	// Send WeChat template message if template ID is configured
	// This would require a template ID configuration
	// templateID := w.config.SyncNotificationTemplateID
	// if templateID != "" {
	//     data := map[string]interface{}{
	//         "first": map[string]interface{}{"value": "账单同步完成", "color": "#173177"},
	//         "keyword1": map[string]interface{}{"value": billingMonth, "color": "#173177"},
	//         "keyword2": map[string]interface{}{"value": fmt.Sprintf("%d/%d", syncedCount, totalCount), "color": "#173177"},
	//         "remark": map[string]interface{}{"value": "感谢您的使用", "color": "#173177"},
	//     }
	//     return w.SendMessage(openID, templateID, data)
	// }

	return nil
}

// IsConfigured checks if WeChat API is properly configured
func (w *WeChatAPIService) IsConfigured() bool {
	return w.config != nil && w.config.AppID != "" && w.config.AppSecret != ""
}

// GetConfig returns current WeChat configuration (without sensitive data)
func (w *WeChatAPIService) GetConfig() map[string]interface{} {
	if w.config == nil {
		return map[string]interface{}{
			"configured": false,
		}
	}

	return map[string]interface{}{
		"configured":    w.IsConfigured(),
		"app_id":        w.config.AppID,
		"token":         w.config.Token,
		"has_token":     w.config.AccessToken != "",
		"token_expires": w.config.ExpiresAt,
	}
}

/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: miniapp_auth.go
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-06-02
 */

package app

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	errMiniAppAuthNotConfigured = errors.New("miniapp wechat auth not configured")
	errMiniAppCodeMissing       = errors.New("miniapp login code is required")
	errMiniAppCodeExchange      = errors.New("miniapp code exchange failed")
	errMiniAppPhoneCodeExchange = errors.New("miniapp phone code exchange failed")
	errMiniAppPhoneDataDecrypt  = errors.New("miniapp encrypted phone data decrypt failed")
	errMiniAppUserNotBound      = errors.New("wechat account is not bound to any operator")
)

type wechatMiniAppSession struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type wechatAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type wechatPhoneNumberResponse struct {
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	} `json:"phone_info"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type wechatEncryptedPhoneData struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

func (c Config) hasMiniAppWechatAuth() bool {
	return strings.TrimSpace(c.WechatMiniAppAppID) != "" && strings.TrimSpace(c.WechatMiniAppAppSecret) != ""
}

func (a *App) authenticateMiniAppUser(ctx context.Context, req miniAppLoginRequest) (UserAccount, error) {
	if a.Config.hasMiniAppWechatAuth() {
		if strings.TrimSpace(req.Code) == "" {
			return UserAccount{}, errMiniAppCodeMissing
		}

		session, err := a.exchangeMiniAppCode(ctx, req.Code)
		if err != nil {
			return UserAccount{}, err
		}

		if user, ok := a.store.findUserByWechatOpenID(session.OpenID); ok {
			return user, nil
		}

		phone := strings.TrimSpace(req.Profile.Phone)
		if strings.TrimSpace(req.PhoneCode) != "" {
			wechatPhone, err := a.exchangeMiniAppPhoneCode(ctx, req.PhoneCode)
			if err != nil {
				return UserAccount{}, err
			}
			phone = wechatPhone
		} else if strings.TrimSpace(req.PhoneEncryptedData) != "" || strings.TrimSpace(req.PhoneIV) != "" {
			wechatPhone, err := decryptMiniAppPhoneData(session.SessionKey, req.PhoneEncryptedData, req.PhoneIV)
			if err != nil {
				return UserAccount{}, err
			}
			phone = wechatPhone
		}

		if user, ok := a.store.bindMiniAppUserByProfile(session, phone, req.Profile.RoleKey); ok {
			return user, nil
		}

		a.store.recordPendingMiniAppBinding(session, req.Profile.Name, phone, req.Profile.RoleKey, req.Profile.StoreName, req.Profile.StoreCode)

		if a.Config.MiniAppAllowMockLogin {
			return a.store.mockMiniAppUserForRole(strings.TrimSpace(req.Profile.RoleKey))
		}

		return UserAccount{}, fmt.Errorf("%w: %s", errMiniAppUserNotBound, session.OpenID)
	}

	if a.Config.MiniAppAllowMockLogin {
		return a.store.mockMiniAppUserForRole(strings.TrimSpace(req.Profile.RoleKey))
	}

	return UserAccount{}, errMiniAppAuthNotConfigured
}

func (a *App) exchangeMiniAppCode(ctx context.Context, code string) (wechatMiniAppSession, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(a.Config.WechatAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.weixin.qq.com"
	}

	requestURL, err := url.Parse(baseURL + "/sns/jscode2session")
	if err != nil {
		return wechatMiniAppSession{}, fmt.Errorf("%w: invalid api base url", errMiniAppCodeExchange)
	}

	query := requestURL.Query()
	query.Set("appid", strings.TrimSpace(a.Config.WechatMiniAppAppID))
	query.Set("secret", strings.TrimSpace(a.Config.WechatMiniAppAppSecret))
	query.Set("js_code", strings.TrimSpace(code))
	query.Set("grant_type", "authorization_code")
	requestURL.RawQuery = query.Encode()

	requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return wechatMiniAppSession{}, fmt.Errorf("%w: %v", errMiniAppCodeExchange, err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return wechatMiniAppSession{}, fmt.Errorf("%w: %v", errMiniAppCodeExchange, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return wechatMiniAppSession{}, fmt.Errorf("%w: http %d", errMiniAppCodeExchange, resp.StatusCode)
	}

	var result wechatMiniAppSession
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return wechatMiniAppSession{}, fmt.Errorf("%w: decode response: %v", errMiniAppCodeExchange, err)
	}
	if result.ErrCode != 0 {
		return wechatMiniAppSession{}, fmt.Errorf("%w: %d %s", errMiniAppCodeExchange, result.ErrCode, strings.TrimSpace(result.ErrMsg))
	}
	if strings.TrimSpace(result.OpenID) == "" {
		return wechatMiniAppSession{}, fmt.Errorf("%w: missing openid", errMiniAppCodeExchange)
	}
	return result, nil
}

func (a *App) exchangeMiniAppPhoneCode(ctx context.Context, phoneCode string) (string, error) {
	accessToken, err := a.getWechatAccessToken(ctx)
	if err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(strings.TrimSpace(a.Config.WechatAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.weixin.qq.com"
	}

	requestURL, err := url.Parse(baseURL + "/wxa/business/getuserphonenumber")
	if err != nil {
		return "", fmt.Errorf("%w: invalid api base url", errMiniAppPhoneCodeExchange)
	}
	query := requestURL.Query()
	query.Set("access_token", accessToken)
	requestURL.RawQuery = query.Encode()

	requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	payload := strings.NewReader(fmt.Sprintf(`{"code":%q}`, strings.TrimSpace(phoneCode)))
	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodPost, requestURL.String(), payload)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errMiniAppPhoneCodeExchange, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errMiniAppPhoneCodeExchange, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: http %d", errMiniAppPhoneCodeExchange, resp.StatusCode)
	}

	var result wechatPhoneNumberResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: decode response: %v", errMiniAppPhoneCodeExchange, err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("%w: %d %s", errMiniAppPhoneCodeExchange, result.ErrCode, strings.TrimSpace(result.ErrMsg))
	}

	phone := strings.TrimSpace(result.PhoneInfo.PurePhoneNumber)
	if phone == "" {
		phone = strings.TrimSpace(result.PhoneInfo.PhoneNumber)
	}
	if phone == "" {
		return "", fmt.Errorf("%w: missing phone number", errMiniAppPhoneCodeExchange)
	}
	return phone, nil
}

func decryptMiniAppPhoneData(sessionKey, encryptedData, iv string) (string, error) {
	sessionKeyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(sessionKey))
	if err != nil {
		return "", fmt.Errorf("%w: invalid session key", errMiniAppPhoneDataDecrypt)
	}
	encryptedBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encryptedData))
	if err != nil {
		return "", fmt.Errorf("%w: invalid encrypted data", errMiniAppPhoneDataDecrypt)
	}
	ivBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(iv))
	if err != nil {
		return "", fmt.Errorf("%w: invalid iv", errMiniAppPhoneDataDecrypt)
	}

	block, err := aes.NewCipher(sessionKeyBytes)
	if err != nil {
		return "", fmt.Errorf("%w: invalid aes key", errMiniAppPhoneDataDecrypt)
	}
	if len(ivBytes) != block.BlockSize() {
		return "", fmt.Errorf("%w: invalid iv size", errMiniAppPhoneDataDecrypt)
	}
	if len(encryptedBytes) == 0 || len(encryptedBytes)%block.BlockSize() != 0 {
		return "", fmt.Errorf("%w: invalid encrypted data size", errMiniAppPhoneDataDecrypt)
	}

	plainBytes := make([]byte, len(encryptedBytes))
	cipher.NewCBCDecrypter(block, ivBytes).CryptBlocks(plainBytes, encryptedBytes)

	padding := int(plainBytes[len(plainBytes)-1])
	if padding < 1 || padding > block.BlockSize() || padding > len(plainBytes) {
		return "", fmt.Errorf("%w: invalid padding", errMiniAppPhoneDataDecrypt)
	}
	for _, value := range plainBytes[len(plainBytes)-padding:] {
		if int(value) != padding {
			return "", fmt.Errorf("%w: invalid padding bytes", errMiniAppPhoneDataDecrypt)
		}
	}
	plainBytes = plainBytes[:len(plainBytes)-padding]

	var result wechatEncryptedPhoneData
	if err := json.Unmarshal(plainBytes, &result); err != nil {
		return "", fmt.Errorf("%w: decode phone data: %v", errMiniAppPhoneDataDecrypt, err)
	}

	phone := strings.TrimSpace(result.PurePhoneNumber)
	if phone == "" {
		phone = strings.TrimSpace(result.PhoneNumber)
	}
	if phone == "" {
		return "", fmt.Errorf("%w: missing phone number", errMiniAppPhoneDataDecrypt)
	}
	return phone, nil
}

func (a *App) getWechatAccessToken(ctx context.Context) (string, error) {
	a.wechatAccessTokenMu.Lock()
	defer a.wechatAccessTokenMu.Unlock()

	if strings.TrimSpace(a.wechatAccessToken) != "" && time.Now().Before(a.wechatAccessTokenExpiresAt.Add(-5*time.Minute)) {
		return a.wechatAccessToken, nil
	}

	baseURL := strings.TrimRight(strings.TrimSpace(a.Config.WechatAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.weixin.qq.com"
	}

	requestURL, err := url.Parse(baseURL + "/cgi-bin/token")
	if err != nil {
		return "", fmt.Errorf("%w: invalid api base url", errMiniAppPhoneCodeExchange)
	}
	query := requestURL.Query()
	query.Set("grant_type", "client_credential")
	query.Set("appid", strings.TrimSpace(a.Config.WechatMiniAppAppID))
	query.Set("secret", strings.TrimSpace(a.Config.WechatMiniAppAppSecret))
	requestURL.RawQuery = query.Encode()

	requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errMiniAppPhoneCodeExchange, err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errMiniAppPhoneCodeExchange, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: token http %d", errMiniAppPhoneCodeExchange, resp.StatusCode)
	}

	var result wechatAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: decode token response: %v", errMiniAppPhoneCodeExchange, err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("%w: token %d %s", errMiniAppPhoneCodeExchange, result.ErrCode, strings.TrimSpace(result.ErrMsg))
	}
	if strings.TrimSpace(result.AccessToken) == "" {
		return "", fmt.Errorf("%w: missing access_token", errMiniAppPhoneCodeExchange)
	}

	expiresIn := result.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7200
	}
	a.wechatAccessToken = strings.TrimSpace(result.AccessToken)
	a.wechatAccessTokenExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	return a.wechatAccessToken, nil
}

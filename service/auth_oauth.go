package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tigerowo/infinite-canvas/config"
	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
	"github.com/google/uuid"
)

type oauthConfig struct {
	clientID     string
	clientSecret string
	authorizeURL string
	tokenURL     string
	userInfoURL  string
	scope        string
	redirectURI  string
	prompt       string
}

type oauthState struct {
	Value    string
	Redirect string
}

func GoogleAuthorizeURL(r *http.Request, redirect string) (string, *http.Cookie, error) {
	return oauthAuthorizeURL(r, "google", redirect, oauthConfig{
		clientID: config.Cfg.GoogleClientID, clientSecret: config.Cfg.GoogleClientSecret,
		authorizeURL: config.Cfg.GoogleAuthorizeURL, tokenURL: config.Cfg.GoogleTokenURL,
		userInfoURL: config.Cfg.GoogleUserInfoURL, scope: config.Cfg.GoogleScope,
		redirectURI: config.Cfg.GoogleRedirectURI,
		prompt:      "select_account",
	})
}

func DoingFBAuthorizeURL(r *http.Request, redirect string) (string, *http.Cookie, error) {
	return oauthAuthorizeURL(r, "doingfb", redirect, oauthConfig{
		clientID: config.Cfg.DoingFBClientID, clientSecret: config.Cfg.DoingFBClientSecret,
		authorizeURL: config.Cfg.DoingFBAuthorizeURL, tokenURL: config.Cfg.DoingFBTokenURL,
		userInfoURL: config.Cfg.DoingFBUserInfoURL, scope: config.Cfg.DoingFBScope,
		redirectURI: config.Cfg.DoingFBRedirectURI,
		prompt:      "login",
	})
}

func LoginWithGoogle(r *http.Request, code string, state string) (model.AuthSession, string, error) {
	return loginWithOAuth(r, "google", code, state, oauthConfig{
		clientID: config.Cfg.GoogleClientID, clientSecret: config.Cfg.GoogleClientSecret,
		tokenURL: config.Cfg.GoogleTokenURL, userInfoURL: config.Cfg.GoogleUserInfoURL,
		redirectURI: config.Cfg.GoogleRedirectURI,
	})
}

func LoginWithDoingFB(r *http.Request, code string, state string) (model.AuthSession, string, error) {
	return loginWithOAuth(r, "doingfb", code, state, oauthConfig{
		clientID: config.Cfg.DoingFBClientID, clientSecret: config.Cfg.DoingFBClientSecret,
		tokenURL: config.Cfg.DoingFBTokenURL, userInfoURL: config.Cfg.DoingFBUserInfoURL,
		redirectURI: config.Cfg.DoingFBRedirectURI,
	})
}

func oauthAuthorizeURL(r *http.Request, provider string, redirect string, setting oauthConfig) (string, *http.Cookie, error) {
	if strings.TrimSpace(setting.clientID) == "" || strings.TrimSpace(setting.clientSecret) == "" {
		return "", nil, safeMessageError{message: provider + "_oauth_not_configured"}
	}
	state := oauthState{Value: uuid.NewString(), Redirect: safeRedirectPath(redirect)}
	rawState, _ := json.Marshal(state)
	encodedState := base64.RawURLEncoding.EncodeToString(rawState)
	values := url.Values{}
	values.Set("client_id", setting.clientID)
	values.Set("redirect_uri", oauthRedirectURI(r, provider, setting.redirectURI))
	values.Set("response_type", "code")
	values.Set("scope", firstNonEmpty(setting.scope, "openid email profile"))
	values.Set("state", encodedState)
	if strings.TrimSpace(setting.prompt) != "" {
		values.Set("prompt", strings.TrimSpace(setting.prompt))
	}
	return setting.authorizeURL + "?" + values.Encode(), OAuthStateCookie(r, provider, encodedState, 600), nil
}

func loginWithOAuth(r *http.Request, provider string, code string, state string, setting oauthConfig) (model.AuthSession, string, error) {
	redirect, err := validateOAuthState(r, provider, state)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	if strings.TrimSpace(code) == "" {
		return model.AuthSession{}, redirect, safeMessageError{message: provider + "_oauth_code_missing"}
	}
	if strings.TrimSpace(setting.clientID) == "" || strings.TrimSpace(setting.clientSecret) == "" {
		return model.AuthSession{}, redirect, safeMessageError{message: provider + "_oauth_not_configured"}
	}
	token, err := exchangeOAuthToken(r, code, setting, provider)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	profile, err := fetchOAuthProfile(token, setting.userInfoURL, provider)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	user, err := upsertOAuthUser(provider, profile)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	user.LastLoginAt = now()
	user.UpdatedAt = now()
	user, err = repository.SaveUser(user)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	session, err := newSession(user)
	return session, redirect, err
}

func OAuthStateCookie(r *http.Request, provider string, value string, maxAge int) *http.Cookie {
	secure := strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
	return &http.Cookie{
		Name:     "infinite_canvas_" + provider + "_oauth_state",
		Value:    value,
		Path:     "/api/auth/" + provider + "/callback",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func validateOAuthState(r *http.Request, provider string, state string) (string, error) {
	cookie, err := r.Cookie("infinite_canvas_" + provider + "_oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != state {
		return "/", safeMessageError{message: provider + "_oauth_state_invalid"}
	}
	data, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "/", safeMessageError{message: provider + "_oauth_state_invalid"}
	}
	var value oauthState
	if json.Unmarshal(data, &value) != nil || value.Value == "" {
		return "/", safeMessageError{message: provider + "_oauth_state_invalid"}
	}
	return safeRedirectPath(value.Redirect), nil
}

func oauthRedirectURI(r *http.Request, provider string, configured string) string {
	if strings.TrimSpace(configured) != "" {
		return strings.TrimSpace(configured)
	}
	return RequestOrigin(r) + "/api/auth/" + provider + "/callback"
}

func exchangeOAuthToken(r *http.Request, code string, setting oauthConfig, provider string) (string, error) {
	values := url.Values{}
	values.Set("client_id", setting.clientID)
	values.Set("client_secret", setting.clientSecret)
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", oauthRedirectURI(r, provider, setting.redirectURI))
	req, err := http.NewRequest(http.MethodPost, setting.tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var payload map[string]any
	if err := doOAuthJSON(req, &payload); err != nil {
		return "", safeMessageError{message: provider + "_token_failed"}
	}
	token, _ := payload["access_token"].(string)
	if strings.TrimSpace(token) == "" {
		return "", safeMessageError{message: provider + "_token_failed"}
	}
	return token, nil
}

func fetchOAuthProfile(token string, endpoint string, provider string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	var payload map[string]any
	if err := doOAuthJSON(req, &payload); err != nil {
		return nil, safeMessageError{message: provider + "_userinfo_failed"}
	}
	return payload, nil
}

func doOAuthJSON(req *http.Request, payload any) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("oauth request failed")
	}
	return json.NewDecoder(bytes.NewReader(body)).Decode(payload)
}

func upsertOAuthUser(provider string, profile map[string]any) (model.User, error) {
	providerID := oauthString(profile, "sub", "id")
	if providerID == "" {
		return model.User{}, safeMessageError{message: provider + "_user_missing_id"}
	}
	email := normalizeEmail(oauthString(profile, "email"))
	verified := oauthBool(profile, "email_verified", "verified_email")
	providerUser := model.User{}
	var ok bool
	var err error
	if provider == "google" {
		providerUser, ok, err = repository.GetUserByGoogleID(providerID)
	} else {
		providerUser, ok, err = repository.GetUserByDoingFBID(providerID)
	}
	if err != nil {
		return model.User{}, err
	}
	if !ok && verified && email != "" {
		providerUser, ok, err = repository.GetUserByEmail(email)
		if err != nil {
			return model.User{}, err
		}
	}
	if !ok {
		settings, settingsErr := repository.GetSettings()
		if settingsErr != nil {
			return model.User{}, settingsErr
		}
		settings = normalizeSettings(settings)
		if settings.Public.Auth.AllowRegister != nil && !*settings.Public.Auth.AllowRegister {
			return model.User{}, safeMessageError{message: "registration_disabled"}
		}
		username := oauthUsername(oauthString(profile, "username"), email, providerID)
		providerUser = model.User{
			ID:          newID("user"),
			Username:    username,
			Email:       emailIfVerified(email, verified),
			DisplayName: firstNonEmpty(oauthString(profile, "name"), username),
			AvatarURL:   oauthString(profile, "picture", "avatar_url"),
			Role:        model.UserRoleUser,
			AffCode:     newAffCode(),
			Status:      model.UserStatusActive,
			CreatedAt:   now(),
			EmailVerifiedAt: emailVerifiedAt(verified, email),
		}
	} else if providerUser.Status == model.UserStatusBan {
		return model.User{}, safeMessageError{message: "璐﹀彿宸茶绂佺敤"}
	}
	providerUser.DisplayName = firstNonEmpty(oauthString(profile, "name"), providerUser.DisplayName)
	providerUser.AvatarURL = firstNonEmpty(oauthString(profile, "picture", "avatar_url"), providerUser.AvatarURL)
	if email != "" && verified && providerUser.Email == "" {
		providerUser.Email = email
		providerUser.EmailVerifiedAt = firstNonEmpty(providerUser.EmailVerifiedAt, now())
	}
	if provider == "google" {
		providerUser.GoogleID = providerID
	} else {
		providerUser.DoingFBID = providerID
	}
	return providerUser, nil
}

func oauthString(profile map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := profile[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
		if value, ok := profile[key].(float64); ok {
			return fmt.Sprintf("%.0f", value)
		}
	}
	return ""
}

func oauthBool(profile map[string]any, keys ...string) bool {
	for _, key := range keys {
		if value, ok := profile[key].(bool); ok {
			return value
		}
	}
	return false
}

func emailIfVerified(email string, verified bool) string {
	if verified {
		return email
	}
	return ""
}

func emailVerifiedAt(verified bool, email string) string {
	if verified && email != "" {
		return now()
	}
	return ""
}

func oauthUsername(username string, email string, providerID string) string {
	base := strings.TrimSpace(username)
	if base == "" && email != "" {
		base = strings.Split(email, "@")[0]
	}
	if base == "" {
		base = "oauth-" + providerID
	}
	base = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			return '-'
		}
		return r
	}, base)
	if _, ok, err := repository.GetUserByUsername(base); err != nil || !ok {
		return base
	}
	return base + "-" + providerID
}

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"time"

	"github.com/tigerowo/infinite-canvas/config"
	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
	"github.com/google/uuid"
)

const (
	verificationTokenTTL = 24 * time.Hour
	resetTokenTTL        = 30 * time.Minute
)

type VerificationResult struct {
	EmailVerificationRequired bool `json:"emailVerificationRequired"`
	VerificationEmailSent     bool `json:"verificationEmailSent"`
}

func issueVerificationEmail(user model.User) (bool, error) {
	token := uuid.NewString() + uuid.NewString()
	user.EmailVerificationTokenHash = hashAuthToken(token)
	user.EmailVerificationExpiresAt = nowAfter(verificationTokenTTL)
	if _, err := repository.SaveUser(user); err != nil {
		return false, err
	}
	link := authLink("verify_token", token)
	err := sendAuthMail(user.Email, "Infinite Canvas email verification", authMailBody(user.DisplayName, link, "完成邮箱验证"))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func ResendVerificationEmail(email string) (VerificationResult, error) {
	email = normalizeEmail(email)
	if email == "" {
		return VerificationResult{}, safeMessageError{message: "email_required"}
	}
	user, ok, err := repository.GetUserByEmail(email)
	if err != nil {
		return VerificationResult{}, err
	}
	result := VerificationResult{EmailVerificationRequired: smtpConfigured()}
	if !ok || user.EmailVerifiedAt != "" {
		return result, nil
	}
	if !smtpConfigured() {
		user.EmailVerifiedAt = now()
		user.UpdatedAt = now()
		_, err = repository.SaveUser(user)
		return result, err
	}
	result.VerificationEmailSent, err = issueVerificationEmail(user)
	return result, err
}

func VerifyEmailToken(token string) error {
	_, ok, err := repository.VerifyEmail(hashAuthToken(token), now())
	if err != nil {
		return err
	}
	if !ok {
		return safeMessageError{message: "invalid_verification_token"}
	}
	return nil
}

func RequestPasswordReset(email string) error {
	email = normalizeEmail(email)
	if email == "" {
		return safeMessageError{message: "email_required"}
	}
	if !smtpConfigured() {
		return safeMessageError{message: "email_not_configured"}
	}
	user, ok, err := repository.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if !ok || user.EmailVerifiedAt == "" || user.Password == "" {
		return nil
	}
	token := uuid.NewString() + uuid.NewString()
	user.PasswordResetTokenHash = hashAuthToken(token)
	user.PasswordResetExpiresAt = nowAfter(resetTokenTTL)
	if _, err := repository.SaveUser(user); err != nil {
		return err
	}
	return sendAuthMail(user.Email, "Infinite Canvas password reset", authMailBody(user.DisplayName, authLink("reset_token", token), "重置密码"))
}

func ResetPassword(token string, password string) (model.AuthSession, error) {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(password) == "" {
		return model.AuthSession{}, safeMessageError{message: "invalid_reset_token"}
	}
	hash, err := hashPassword(password)
	if err != nil {
		return model.AuthSession{}, err
	}
	user, ok, err := repository.ResetPassword(hashAuthToken(token), hash, now())
	if err != nil {
		return model.AuthSession{}, err
	}
	if !ok {
		return model.AuthSession{}, safeMessageError{message: "invalid_reset_token"}
	}
	return newSession(user)
}

func authLink(key string, token string) string {
	base := strings.TrimRight(strings.TrimSpace(config.Cfg.PublicBaseURL), "/")
	if base == "" {
		base = "http://127.0.0.1:3000"
	}
	values := url.Values{}
	values.Set(key, token)
	return base + "/login?" + values.Encode()
}

func hashAuthToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func nowAfter(duration time.Duration) string {
	return time.Now().Add(duration).Format(time.RFC3339)
}

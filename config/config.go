package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                string `env:"PORT" envDefault:"8080"`
	AdminUsername       string `env:"ADMIN_USERNAME" envDefault:"admin"`
	AdminPassword       string `env:"ADMIN_PASSWORD" envDefault:"infinite-canvas"`
	JWTSecret           string `env:"JWT_SECRET" envDefault:"infinite-canvas"`
	JWTExpireHours      int    `env:"JWT_EXPIRE_HOURS" envDefault:"168"`
	StorageDriver       string `env:"STORAGE_DRIVER" envDefault:"sqlite"`
	DatabaseDSN         string `env:"DATABASE_DSN" envDefault:"data/infinite-canvas.db"`
	PublicBaseURL       string `env:"PUBLIC_BASE_URL"`
	LinuxDoAuthorizeURL string `env:"LINUX_DO_AUTHORIZE_URL" envDefault:"https://connect.linux.do/oauth2/authorize"`
	LinuxDoTokenURL     string `env:"LINUX_DO_TOKEN_URL" envDefault:"https://connect.linux.do/oauth2/token"`
	LinuxDoUserInfoURL  string `env:"LINUX_DO_USERINFO_URL" envDefault:"https://connect.linux.do/api/user"`
	DoingFBClientID     string `env:"DOINGFB_CLIENT_ID"`
	DoingFBClientSecret string `env:"DOINGFB_CLIENT_SECRET"`
	DoingFBAuthorizeURL string `env:"DOINGFB_AUTHORIZATION_URL" envDefault:"https://doingfb.com/oauth2/authorize"`
	DoingFBTokenURL     string `env:"DOINGFB_TOKEN_URL" envDefault:"https://doingfb.com/oauth2/token"`
	DoingFBUserInfoURL  string `env:"DOINGFB_USERINFO_URL" envDefault:"https://doingfb.com/api/oauth/user"`
	DoingFBScope        string `env:"DOINGFB_SCOPE" envDefault:"user.read user.email"`
	DoingFBRedirectURI  string `env:"DOINGFB_REDIRECT_URI"`
	GoogleClientID      string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret  string `env:"GOOGLE_CLIENT_SECRET"`
	GoogleAuthorizeURL  string `env:"GOOGLE_AUTHORIZATION_URL" envDefault:"https://accounts.google.com/o/oauth2/v2/auth"`
	GoogleTokenURL      string `env:"GOOGLE_TOKEN_URL" envDefault:"https://oauth2.googleapis.com/token"`
	GoogleUserInfoURL   string `env:"GOOGLE_USERINFO_URL" envDefault:"https://openidconnect.googleapis.com/v1/userinfo"`
	GoogleScope         string `env:"GOOGLE_SCOPE" envDefault:"openid email profile"`
	GoogleRedirectURI   string `env:"GOOGLE_REDIRECT_URI"`
	SMTPHost            string `env:"SMTP_HOST"`
	SMTPPort            int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPSecure          bool   `env:"SMTP_SECURE"`
	SMTPUser            string `env:"SMTP_USER"`
	SMTPPass            string `env:"SMTP_PASS"`
	SMTPFrom            string `env:"SMTP_FROM"`
	SMTPReplyTo         string `env:"SMTP_REPLY_TO"`
	AILogDir            string `env:"AI_LOG_DIR" envDefault:"data/logs/ai-calls"`
}

var Cfg Config

func Load() error {
	_ = godotenv.Load()
	if err := env.Parse(&Cfg); err != nil {
		return err
	}
	normalizeDockerSQLiteDSN("/app/data")
	if strings.TrimSpace(Cfg.JWTSecret) == "" || Cfg.JWTSecret == "infinite-canvas" {
		secret, err := randomSecret()
		if err != nil {
			return err
		}
		Cfg.JWTSecret = secret
	}
	return nil
}

func normalizeDockerSQLiteDSN(appDataDir string) {
	driver := strings.ToLower(strings.TrimSpace(Cfg.StorageDriver))
	if driver != "" && driver != "sqlite" {
		return
	}
	dsn := strings.TrimSpace(Cfg.DatabaseDSN)
	if dsn == "" || dsn == ":memory:" || strings.HasPrefix(dsn, "file:") {
		return
	}
	pathPart, suffix := dsn, ""
	if index := strings.Index(dsn, "?"); index >= 0 {
		pathPart = dsn[:index]
		suffix = dsn[index:]
	}
	if filepath.IsAbs(pathPart) {
		return
	}
	slashPath := filepath.ToSlash(pathPart)
	if slashPath != "data" && !strings.HasPrefix(slashPath, "data/") {
		return
	}
	if _, err := os.Stat(appDataDir); err != nil {
		return
	}
	Cfg.DatabaseDSN = filepath.Join(filepath.Dir(appDataDir), filepath.FromSlash(slashPath)) + suffix
}

func randomSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

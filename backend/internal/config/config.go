package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	CORS     CORSConfig     `mapstructure:"cors"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	WeChat   WeChatConfig   `mapstructure:"wechat"`
	Log      LogConfig      `mapstructure:"log"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	Issuer      string `mapstructure:"issuer"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type WeChatConfig struct {
	AppID          string `mapstructure:"app_id"`
	AppSecret      string `mapstructure:"app_secret"`
	APIBase        string `mapstructure:"api_base"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

func Load() (*Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	for _, path := range []string{"configs", "../configs", "../../configs"} {
		v.AddConfigPath(path)
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := bindEnvs(v); err != nil {
		return nil, fmt.Errorf("bind env failed: %w", err)
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config failed: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	normalizeConfig(&cfg)
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "gin-backend")
	v.SetDefault("app.env", "development")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 15)
	v.SetDefault("server.write_timeout", 15)
	v.SetDefault("server.idle_timeout", 60)

	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "agent_system")
	v.SetDefault("database.charset", "utf8mb4")
	v.SetDefault("database.parse_time", true)
	v.SetDefault("database.loc", "Local")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", 3600)

	v.SetDefault("cors.allow_origins", []string{"http://localhost:5173"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	v.SetDefault("cors.expose_headers", []string{"Content-Length"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 43200)

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.issuer", "gin-backend")
	v.SetDefault("jwt.expire_hours", 72)

	v.SetDefault("wechat.api_base", "https://api.weixin.qq.com")
	v.SetDefault("wechat.timeout_seconds", 5)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "console")
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c DatabaseConfig) DSN() string {
	location := c.Loc
	if location == "" {
		location = "Local"
	}

	// If password is empty, don't include it in DSN
	if c.Password == "" {
		return fmt.Sprintf(
			"%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=%t&loc=%s&multiStatements=true",
			c.User,
			c.Host,
			c.Port,
			c.Name,
			c.ParseTime,
			url.QueryEscape(location),
		)
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=%t&loc=%s&multiStatements=true",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.ParseTime,
		url.QueryEscape(location),
	)
}

func bindEnvs(v *viper.Viper) error {
	if err := v.BindEnv("app.env", "APP_ENV", "SERVER_MODE"); err != nil {
		return err
	}
	if err := v.BindEnv("server.host", "SERVER_HOST"); err != nil {
		return err
	}
	if err := v.BindEnv("server.port", "SERVER_PORT"); err != nil {
		return err
	}
	if err := v.BindEnv("database.host", "DB_HOST"); err != nil {
		return err
	}
	if err := v.BindEnv("database.port", "DB_PORT"); err != nil {
		return err
	}
	if err := v.BindEnv("database.user", "DB_USER"); err != nil {
		return err
	}
	if err := v.BindEnv("database.password", "DB_PASSWORD"); err != nil {
		return err
	}
	if err := v.BindEnv("database.name", "DB_NAME"); err != nil {
		return err
	}
	if err := v.BindEnv("cors.allow_origins", "CORS_ALLOW_ORIGINS"); err != nil {
		return err
	}
	if err := v.BindEnv("cors.allow_methods", "CORS_ALLOW_METHODS"); err != nil {
		return err
	}
	if err := v.BindEnv("cors.allow_headers", "CORS_ALLOW_HEADERS"); err != nil {
		return err
	}
	if err := v.BindEnv("cors.expose_headers", "CORS_EXPOSE_HEADERS"); err != nil {
		return err
	}
	if err := v.BindEnv("cors.allow_credentials", "CORS_ALLOW_CREDENTIALS"); err != nil {
		return err
	}
	if err := v.BindEnv("cors.max_age", "CORS_MAX_AGE"); err != nil {
		return err
	}
	if err := v.BindEnv("jwt.secret", "JWT_SECRET"); err != nil {
		return err
	}
	if err := v.BindEnv("jwt.expire_hours", "JWT_EXPIRE_HOURS"); err != nil {
		return err
	}
	if err := v.BindEnv("wechat.app_id", "WECHAT_APPID"); err != nil {
		return err
	}
	if err := v.BindEnv("wechat.app_secret", "WECHAT_SECRET"); err != nil {
		return err
	}
	if err := v.BindEnv("log.level", "LOG_LEVEL"); err != nil {
		return err
	}
	return nil
}

func normalizeConfig(cfg *Config) {
	cfg.App.Env = normalizeEnv(cfg.App.Env)
	cfg.Database.Host = strings.TrimSpace(cfg.Database.Host)
	cfg.Database.User = strings.TrimSpace(cfg.Database.User)
	cfg.Database.Name = strings.TrimSpace(cfg.Database.Name)
	cfg.JWT.Secret = strings.TrimSpace(cfg.JWT.Secret)
	cfg.WeChat.AppID = strings.TrimSpace(cfg.WeChat.AppID)
	cfg.WeChat.AppSecret = strings.TrimSpace(cfg.WeChat.AppSecret)
	cfg.CORS.AllowOrigins = normalizeStringSlice(cfg.CORS.AllowOrigins)
	cfg.CORS.AllowMethods = normalizeStringSlice(cfg.CORS.AllowMethods)
	cfg.CORS.AllowHeaders = normalizeStringSlice(cfg.CORS.AllowHeaders)
	cfg.CORS.ExposeHeaders = normalizeStringSlice(cfg.CORS.ExposeHeaders)
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
}

func normalizeEnv(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "dev", "debug", "development":
		return "development"
	case "prod", "production", "release":
		return "production"
	case "test", "testing":
		return "test"
	default:
		return strings.TrimSpace(value)
	}
}

func validateConfig(cfg *Config) error {
	if cfg.Server.Port <= 0 {
		return fmt.Errorf("server.port must be greater than 0")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if len(cfg.CORS.AllowOrigins) == 0 {
		return fmt.Errorf("cors.allow_origins must not be empty")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if cfg.App.Env != "development" {
		if len(cfg.JWT.Secret) < 32 {
			return fmt.Errorf("jwt.secret must be at least 32 characters in production")
		}
		insecureSecrets := []string{"change-me-in-production", "your-super-secret-key", "secret", "password", "dev-secret-key"}
		for _, insecure := range insecureSecrets {
			if strings.Contains(strings.ToLower(cfg.JWT.Secret), insecure) {
				return fmt.Errorf("jwt.secret contains insecure pattern")
			}
		}
	}
	if cfg.JWT.ExpireHours <= 0 {
		return fmt.Errorf("jwt.expire_hours must be greater than 0")
	}

	if cfg.App.Env != "development" {
		if isPlaceholder(cfg.WeChat.AppID) || isPlaceholder(cfg.WeChat.AppSecret) {
			return fmt.Errorf("wechat credentials must be configured in non-development environments")
		}
	}

	return nil
}

func normalizeStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func isPlaceholder(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return true
	}
	placeholders := []string{"your_wechat_appid", "your_wechat_secret", "placeholder", "changeme"}
	for _, p := range placeholders {
		if normalized == p || strings.Contains(normalized, "your_") || strings.Contains(normalized, "your-") {
			return true
		}
	}
	return false
}

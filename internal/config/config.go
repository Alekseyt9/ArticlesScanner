package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultTimezone   = "UTC"
	configPathEnv     = "ARTICLE_SCANNER_CONFIG"
	databaseDSNEnv    = "DATABASE_DSN"
	chatGPTAPIKeyEnv  = "CHATGPT_API_KEY"
	chatGPTModelEnv   = "CHATGPT_MODEL"
	telegramTokenEnv  = "TELEGRAM_BOT_TOKEN"
	telegramChatIDEnv = "TELEGRAM_CHAT_ID"
)

// Config holds high-level settings required across the application.
type Config struct {
	Database      DatabaseConfig     `yaml:"database"`
	Scheduler     SchedulerConfig    `yaml:"scheduler"`
	Providers     ProviderConfig     `yaml:"providers"`
	Notifications NotificationConfig `yaml:"notifications"`
	ML            MLConfig           `yaml:"ml"`
	ChatGPT       ChatGPTConfig      `yaml:"chatgpt"`
	Sites         []SiteConfig       `yaml:"sites"`
}

// DatabaseConfig describes Postgres connection details.
type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

// SchedulerConfig defines when the parser should run.
type SchedulerConfig struct {
	CronExpression string         `yaml:"cronExpression"`
	Timezone       string         `yaml:"timezone"`
	location       *time.Location `yaml:"-"`
}

// Location resolves the scheduler timezone string to a time.Location.
func (s SchedulerConfig) Location() *time.Location {
	if s.location != nil {
		return s.location
	}
	loc, _ := time.LoadLocation(defaultTimezone)
	return loc
}

// ProviderConfig groups settings for article sources.
type ProviderConfig struct {
	ArticleAPIURL string `yaml:"articleApiUrl"`
}

// NotificationConfig encapsulates outbound channels (Telegram, etc.).
type NotificationConfig struct {
	Telegram TelegramConfig `yaml:"telegram"`
}

// TelegramConfig wires all data required to send messages.
type TelegramConfig struct {
	BotToken string `yaml:"botToken"`
	ChatID   string `yaml:"chatId"`
}

// MLConfig describes neural-service integration parameters.
type MLConfig struct {
	InferenceURL string `yaml:"inferenceUrl"`
	APIKey       string `yaml:"apiKey"`
}

// ChatGPTConfig defines how to contact the ChatGPT API.
type ChatGPTConfig struct {
	Endpoint     string `yaml:"endpoint"`
	Model        string `yaml:"model"`
	APIKey       string `yaml:"apiKey"`
	SystemPrompt string `yaml:"systemPrompt"`
}

// SiteConfig describes a single site with its scanner strategy.
type SiteConfig struct {
	Name       string            `yaml:"name"`
	Scanner    string            `yaml:"scanner"`
	Categories []CategoryConfig  `yaml:"categories"`
	Options    map[string]string `yaml:"options"`
}

// CategoryConfig holds the concrete endpoints to crawl (e.g., Arxiv category URLs).
type CategoryConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Load reads YAML configuration (if present) and applies environment overrides.
func Load() Config {
	cfg := defaultConfig()

	if path := os.Getenv(configPathEnv); path != "" {
		if raw, err := os.ReadFile(path); err != nil {
			log.Printf("config: cannot read %s: %v (falling back to defaults)", path, err)
		} else {
			var fileCfg Config
			if err := yaml.Unmarshal(raw, &fileCfg); err != nil {
				log.Printf("config: cannot parse %s: %v (falling back to defaults)", path, err)
			} else {
				cfg = mergeConfig(cfg, fileCfg)
			}
		}
	}

	cfg.applyEnvOverrides()
	cfg.bindTimezone()

	if len(cfg.Sites) == 0 {
		cfg.Sites = defaultConfig().Sites
	}

	return cfg
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv(databaseDSNEnv); v != "" {
		c.Database.DSN = v
	}

	if v := os.Getenv(telegramTokenEnv); v != "" {
		c.Notifications.Telegram.BotToken = v
	}

	if v := os.Getenv(telegramChatIDEnv); v != "" {
		c.Notifications.Telegram.ChatID = v
	}

	if v := os.Getenv(chatGPTAPIKeyEnv); v != "" {
		c.ChatGPT.APIKey = v
	}

	if v := os.Getenv(chatGPTModelEnv); v != "" {
		c.ChatGPT.Model = v
	}
}

func (c *Config) bindTimezone() {
	tz := c.Scheduler.Timezone
	if tz == "" {
		tz = defaultTimezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("config: unknown timezone %s, reverting to %s", tz, defaultTimezone)
		loc, _ = time.LoadLocation(defaultTimezone)
	}
	c.Scheduler.location = loc
}

func mergeConfig(base, override Config) Config {
	if override.Database.DSN != "" {
		base.Database = override.Database
	}

	if override.Scheduler.CronExpression != "" {
		base.Scheduler.CronExpression = override.Scheduler.CronExpression
	}
	if override.Scheduler.Timezone != "" {
		base.Scheduler.Timezone = override.Scheduler.Timezone
	}

	if override.Providers.ArticleAPIURL != "" {
		base.Providers = override.Providers
	}

	if override.Notifications.Telegram.BotToken != "" {
		base.Notifications.Telegram.BotToken = override.Notifications.Telegram.BotToken
	}
	if override.Notifications.Telegram.ChatID != "" {
		base.Notifications.Telegram.ChatID = override.Notifications.Telegram.ChatID
	}

	if override.ML.InferenceURL != "" {
		base.ML.InferenceURL = override.ML.InferenceURL
	}
	if override.ML.APIKey != "" {
		base.ML.APIKey = override.ML.APIKey
	}

	if override.ChatGPT.Endpoint != "" {
		base.ChatGPT.Endpoint = override.ChatGPT.Endpoint
	}
	if override.ChatGPT.Model != "" {
		base.ChatGPT.Model = override.ChatGPT.Model
	}
	if override.ChatGPT.APIKey != "" {
		base.ChatGPT.APIKey = override.ChatGPT.APIKey
	}
	if override.ChatGPT.SystemPrompt != "" {
		base.ChatGPT.SystemPrompt = override.ChatGPT.SystemPrompt
	}

	if len(override.Sites) > 0 {
		base.Sites = override.Sites
	}

	return base
}

func defaultConfig() Config {
	tz, _ := time.LoadLocation(defaultTimezone)
	return Config{
		Database:  DatabaseConfig{DSN: "postgres://user:pass@localhost:5432/articles"},
		Scheduler: SchedulerConfig{CronExpression: "0 6 * * *", Timezone: defaultTimezone, location: tz},
		Providers: ProviderConfig{ArticleAPIURL: "https://api.example.org/articles"},
		Notifications: NotificationConfig{
			Telegram: TelegramConfig{BotToken: "", ChatID: ""},
		},
		ML: MLConfig{InferenceURL: "https://ml.example.org/infer", APIKey: ""},
		ChatGPT: ChatGPTConfig{
			Endpoint:     "https://api.openai.com/v1/chat/completions",
			Model:        "gpt-4o-mini",
			APIKey:       "",
			SystemPrompt: "You summarize scientific articles.",
		},
		Sites: []SiteConfig{
			{
				Name:    "arxiv-default",
				Scanner: "arxiv",
				Categories: []CategoryConfig{
					{Name: "cs.AI", URL: "https://export.arxiv.org/list/cs.AI/pastweek"},
				},
			},
		},
	}
}

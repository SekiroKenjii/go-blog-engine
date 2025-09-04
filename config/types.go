package config

// #region Env

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type LogConfig struct {
	Level       string `mapstructure:"level"`
	FileName    string `mapstructure:"file_name"`
	MaxFileSize int    `mapstructure:"max_file_size"`
	MaxBackups  int    `mapstructure:"max_backups"`
	MaxAge      int    `mapstructure:"max_age"`
	Compressed  bool   `mapstructure:"is_compressed"`
}

type PostgresConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type JWTConfig struct {
	SecretKey  string `mapstructure:"secret_key"`
	RefreshKey string `mapstructure:"refresh_key"`
}

type CorsConfig struct {
	AllowOrigins     string `mapstructure:"allow_origins"`
	AllowMethods     string `mapstructure:"allow_methods"`
	AllowHeaders     string `mapstructure:"allow_headers"`
	ExposeHeaders    string `mapstructure:"expose_headers"`
	AllowCredentials bool   `mapstructure:"allow_credentials"`
	MaxAge           int    `mapstructure:"max_age"`
}

type SecurityConfig struct {
	Jwt  JWTConfig  `mapstructure:"jwt"`
	Cors CorsConfig `mapstructure:"cors"`
}

type RedisConfig struct {
	Host          string `mapstructure:"host"`
	SentinelPorts string `mapstructure:"sentinel_ports"`
	Database      int    `mapstructure:"database"`
	MasterName    string `mapstructure:"master_name"`
	Password      string `mapstructure:"password"`
}

type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type EmailWorkerConfig struct {
	WorkerCount int `mapstructure:"worker_count"`
	QueueSize   int `mapstructure:"queue_size"`
	MaxRetries  int `mapstructure:"max_retries"`
}

type EmailConfig struct {
	SMTPHost    string            `mapstructure:"smtp_host"`
	SMTPPort    int               `mapstructure:"smtp_port"`
	Username    string            `mapstructure:"username"`
	Password    string            `mapstructure:"password"`
	FromEmail   string            `mapstructure:"from_email"`
	FromName    string            `mapstructure:"from_name"`
	TemplateDir string            `mapstructure:"template_dir"`
	Worker      EmailWorkerConfig `mapstructure:"worker"`
}

// #endregion

// #region Rules
// #endregion

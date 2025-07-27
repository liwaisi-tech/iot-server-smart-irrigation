package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Logger   LoggerConfig   `yaml:"logger"`
	MQTT     MQTTConfig     `yaml:"mqtt"`
	Database DatabaseConfig `yaml:"database"`
	Handlers HandlersConfig `yaml:"handlers"`
}

type MQTTConfig struct {
	Broker          BrokerConfig          `yaml:"broker"`
	Client          ClientConfig          `yaml:"client"`
	Auth            MQTTAuthConfig        `yaml:"auth"`
	TLS             MQTTTLSConfig         `yaml:"tls"`
	Connection      MQTTConnectionConfig  `yaml:"connection"`
	Subscriptions   []SubscriptionConfig  `yaml:"subscriptions"`
	QualityOfService QoSConfig            `yaml:"quality_of_service"`
	MessageHandling MessageHandlingConfig `yaml:"message_handling"`
}

type BrokerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
}

type ClientConfig struct {
	ClientID       string `yaml:"client_id"`
	CleanSession   bool   `yaml:"clean_session"`
	AutoReconnect  bool   `yaml:"auto_reconnect"`
	KeepAlive      int    `yaml:"keep_alive"`
	PingTimeout    string `yaml:"ping_timeout"`
	ConnectTimeout string `yaml:"connect_timeout"`
	WriteTimeout   string `yaml:"write_timeout"`
	ResumeSubs     bool   `yaml:"resume_subs"`
}

type MQTTAuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type MQTTTLSConfig struct {
	Enabled            bool   `yaml:"enabled"`
	CAFile             string `yaml:"ca_file"`
	CertFile           string `yaml:"cert_file"`
	KeyFile            string `yaml:"key_file"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	ServerName         string `yaml:"server_name"`
}

type MQTTConnectionConfig struct {
	MaxReconnectInterval  string `yaml:"max_reconnect_interval"`
	ReconnectBackoff      string `yaml:"reconnect_backoff"`
	MaxReconnectBackoff   string `yaml:"max_reconnect_backoff"`
	ConnectRetry          bool   `yaml:"connect_retry"`
	ConnectRetryInterval  string `yaml:"connect_retry_interval"`
}

type SubscriptionConfig struct {
	Topic   string `yaml:"topic"`
	QoS     int    `yaml:"qos"`
	Handler string `yaml:"handler"`
}

type QoSConfig struct {
	DefaultQoS       int  `yaml:"default_qos"`
	MaxQoS           int  `yaml:"max_qos"`
	RetainAvailable  bool `yaml:"retain_available"`
	RetainHandling   int  `yaml:"retain_handling"`
}

type MessageHandlingConfig struct {
	MaxInflight          int    `yaml:"max_inflight"`
	MessageChannelDepth  int    `yaml:"message_channel_depth"`
	ErrorHandler         string `yaml:"error_handler"`
}


type DatabaseConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	User              string `yaml:"user"`
	Password          string `yaml:"password"`
	DBName            string `yaml:"dbname"`
	SSLMode           string `yaml:"sslmode"`
	MaxConnections    int    `yaml:"max_connections"`
	MaxIdleConnections int    `yaml:"max_idle_connections"`
}

type HandlersConfig struct {
	SensorData      HandlerConfig `yaml:"sensor_data"`
	CommandResponse HandlerConfig `yaml:"command_response"`
	HealthStatus    HandlerConfig `yaml:"health_status"`
}

type HandlerConfig struct {
	Workers    int `yaml:"workers"`
	BufferSize int `yaml:"buffer_size"`
}

type LoggerConfig struct {
	Level       string   `yaml:"level"`
	Environment string   `yaml:"environment"`
	OutputPaths []string `yaml:"output_paths,omitempty"`
	Encoding    string   `yaml:"encoding"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

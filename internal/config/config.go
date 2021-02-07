package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var c *Config
var once sync.Once

const (
	envKeyList           = "LIST_"
	envKeyToken          = "TOKEN_"
	envKeyMaxConnections = "MAXCON_"
)

// List struct
type List struct {
	Token          string `mapstructure:"token"`
	URL            string `mapstructure:"url"`
	MaxConnections int    `mapstructure:"maxConnections"`
}

// App struct
type App struct {
	EncryptionKey string `mapstructure:"encryptionKey"`
	URL           string `mapstructure:"url"`
}

// Server struct
type Server struct {
	Port                         int           `mapstructure:"port"`
	WriteTimeout                 time.Duration `mapstructure:"writeTimeout"`
	ReadTimeout                  time.Duration `mapstructure:"readTimeout"`
	IdleTimeout                  time.Duration `mapstructure:"idleTimeout"`
	WaitForConnectionSlotTimeout time.Duration `mapstructure:"waitForConnectionSlotTimeout"`
}

// Client struct
type Client struct {
	DialTimeout           time.Duration `mapstructure:"dialTimeout"`
	DialKeepalive         time.Duration `mapstructure:"dialKeepalive"`
	TLSHandshakeTimeout   time.Duration `mapstructure:"tlsHandshakeTimeout"`
	ResponseHeaderTimeout time.Duration `mapstructure:"responseHeaderTimeout"`
	ExpectContinueTimeout time.Duration `mapstructure:"expectContinueTimeout"`
	Timeout               time.Duration `mapstructure:"timeout"`
}

// Config struct
type Config struct {
	Lists  map[string]List `mapstructure:"lists"`
	App    App             `mapstructure:"app"`
	Server Server          `mapstructure:"server"`
	Client Client          `mapstructure:"client"`
}

// GetConfig returns initialized config struct
func GetConfig() Config {
	once.Do(func() {
		initConfig()
	})

	return *c
}

// GetListURL returns url for playlist and error if list doesnt exist
func GetListURL(k string) (string, error) {
	list, err := GetListFromConfig(k)
	return list.URL, err
}

// GetListToken returns token for list
func GetListToken(k string) (string, error) {
	list, err := GetListFromConfig(k)
	return list.Token, err
}

// GetListMaxConnectios returns max clinet simultaneous conncetions for playlist
func GetListMaxConnectios(k string) (int, error) {
	list, err := GetListFromConfig(k)
	return list.MaxConnections, err
}

// GetListFromConfig find list in config
func GetListFromConfig(name string) (List, error) {
	c := GetConfig()
	for k, l := range c.Lists {
		if k == name {
			return l, nil
		}
	}
	return List{}, errors.New("List " + name + " doesn`t exist")
}

func initConfig() {
	config := Config{}

	viper.SetDefault("server.port", 1338)
	viper.SetDefault("app.url", "http://127.0.0.1:1338")
	viper.SetDefault("app.encryptionKey", "some_key")
	viper.SetDefault("client.dialTimeout", "1m")
	viper.SetDefault("client.dialKeepalive", "5m")
	viper.SetDefault("client.tlsHandshakeTimeout", "30s")
	viper.SetDefault("client.responseHeaderTimeout", "30s")
	viper.SetDefault("client.expectContinueTimeout", "5s")
	viper.SetDefault("client.timeout", "5m")
	viper.SetDefault("server.writeTimeout", "5m")
	viper.SetDefault("server.readTimeout", "5m")
	viper.SetDefault("server.idleTimeout", "5m")
	viper.SetDefault("server.waitForConnectionSlotTimeout", "1s")

	path := "."
	viper.AddConfigPath(path)
	viper.SetConfigName("iptvproxy_config")
	viper.SetConfigType("yaml")

	viper.SafeWriteConfig()

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	if config.Lists == nil {
		config.Lists = make(map[string]List)
	}
	addEnvPlaylists(&config)

	c = &config
}

func addEnvPlaylists(c *Config) {
	for _, e := range os.Environ() {
		if envKeyList == e[:len(envKeyList)] {
			pair := strings.SplitN(e, "=", 2)
			name := pair[0][len(envKeyList):]
			url := pair[1]

			if name == "" || url == "" {
				continue
			}
			token := getEnv(envKeyToken+name, "")
			maxConVal := getEnv(envKeyMaxConnections+name, "1")
			maxCon, _ := strconv.Atoi(maxConVal)

			c.Lists[name] = List{URL: url, Token: token, MaxConnections: maxCon}
		}
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

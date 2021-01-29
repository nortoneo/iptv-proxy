package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var c *Config
var once sync.Once

// Config struct
type Config struct {
	ListenAddress               string
	AppURL                      string
	EncryptionKey               string
	ClientDialTimeout           time.Duration
	ClientDialKeepAlive         time.Duration
	ClientTLSHandshakeTimeout   time.Duration
	ClientResponseHeaderTimeout time.Duration
	ClientExpectContinueTimeout time.Duration
	ClientTimeout               time.Duration
	ServWriteTimeout            time.Duration
	ServReadTimeout             time.Duration
	ServIdleTimeout             time.Duration
}

// GetConfig returns initialized config struct
func GetConfig() Config {
	once.Do(func() {
		initConfig()
		log.Println("Config initialized")
	})

	return *c
}

// GetListURL returns url for playlist and error if list doesnt exist
func GetListURL(k string) (string, error) {
	urlString := getEnv("LIST_"+k, "")
	if urlString == "" {
		return urlString, errors.New("List " + k + " doesn`t exist")
	}
	return urlString, nil
}

// GetListToken returns token for list
func GetListToken(k string) (string, error) {
	token := getEnv("TOKEN_"+k, "")
	if token == "" {
		return token, errors.New("Token for list " + k + " doesn`t exist")
	}
	return token, nil
}

func initConfig() {
	config := Config{}
	//general settings
	config.AppURL = getEnv("APP_URL", "http://localhost:1338")
	config.ListenAddress = getEnv("LISTEN_ADDRESS", ":1338")
	//http client timeouts
	config.ClientDialTimeout = getEnvSeconds("C_DIAL_TIMEOUT", 60)
	config.ClientDialKeepAlive = getEnvSeconds("C_DIAL_KEEPALIVE", 5*60)
	config.ClientTLSHandshakeTimeout = getEnvSeconds("C_TLS_HANDSHAKE_TIMEOUT", 30)
	config.ClientResponseHeaderTimeout = getEnvSeconds("C_RESPONSE_HEADER_TIMEOUT", 30)
	config.ClientExpectContinueTimeout = getEnvSeconds("C_EXPECT_CONTINUE_TIMEOUT", 5)
	config.ClientTimeout = getEnvSeconds("C_TIMEOUT", 5*60)
	//http server timeouts
	config.ServWriteTimeout = getEnvSeconds("S_WRITE_TIMEOUT", 5*60)
	config.ServReadTimeout = getEnvSeconds("S_READ_TIMEOUT", 5*60)
	config.ServIdleTimeout = getEnvSeconds("S_IDLE_TIMEOUT", 5*60)

	setEncryptionKey(&config)

	c = &config
}

func setEncryptionKey(config *Config) {
	key := getEnv("ENCRYPTION_KEY", "")
	if key == "" {
		//TODO: Consider generating random password and store it somewhere to avoid key change on app restart
		log.Print("\n---------------\nWARNING!\nEnvironment variable ENCRYPTION_KEY is not defined!\nUsing default passphrase.\n---------------\n")
		key = "Encryption key is not defined!"
	}
	config.EncryptionKey = key
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvSeconds(key string, fallback int) time.Duration {
	intVal := getEnvInt(key, fallback)
	d := time.Duration(intVal) * time.Second
	return d
}

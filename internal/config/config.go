package config

import (
	"github.com/spf13/viper"
	"go.bug.st/serial"
	"log"
)

type AppConfig struct {
	Debug     bool   `long:"debug" env:"DEBUG" default:"false"`
	Version   string `long:"version" env:"VERSION" default:"unversioned"`
	Commit    string `long:"commit" env:"COMMIT"`
	BuildDate string `long:"build-date" env:"BUILD_DATE"`
	Name      string `long:"name" env:"NAME" default:"lazydocker"`

	Config Config
}

type Config struct {
	SerialConfig     SerialConfig
	ShowHex          bool
	messageFavorites []*MessageFavorite
}

type SerialConfig struct {
	SerialPort string
	SerialMode serial.Mode
}

type MessageFavorite struct {
	Name    string
	Message string
}

const (
	DefaultPort     string = "/dev/ttyUSB0"
	DefaultBaudrate int    = 57600
)

const (
	FlagPort               string = "port"
	FlagBaudrate           string = "baudrate"
	FlagConfig             string = "config"
	ConfigMessageFavorites string = "message_favorites"
)

func NewConfig() Config {
	var messageFavorites []*MessageFavorite

	err := viper.UnmarshalKey(ConfigMessageFavorites, &messageFavorites)
	//err := viper.Unmarshal(&C)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
		//return Model{}
	}

	return Config{
		SerialConfig: SerialConfig{
			SerialPort: DefaultPort,
			SerialMode: serial.Mode{
				BaudRate: DefaultBaudrate,
			},
		},
		ShowHex:          false,
		messageFavorites: messageFavorites,
	}
}

func NewMessageFavorite(name string, message string) *MessageFavorite {
	return &MessageFavorite{name, message}
}

func SaveConfig() {
	err := viper.SafeWriteConfig()
	if err != nil {
		err = viper.WriteConfig()
		if err != nil {
			log.Println("Could not save config", err)
			return
		}
	}
	log.Printf("AppConfig saved")
}

func (fav *MessageFavorite) IsValid() bool {
	return len(fav.Message) > 0 && len(fav.Name) > 0
}

func (c *Config) GetMessageFavorites() []*MessageFavorite {
	return c.messageFavorites
}

func (c *Config) AddMessageFavorite(fav *MessageFavorite, saveConfigToFile bool) {
	c.messageFavorites = append(c.messageFavorites, fav)

	if saveConfigToFile {
		viper.Set(ConfigMessageFavorites, c.messageFavorites)
		SaveConfig()
	}
}

func (c *Config) RemoveMessageFavorite(name string, message string) {
	for i, favorite := range c.messageFavorites {
		if favorite.Name == name && favorite.Message == message {
			c.messageFavorites = append(c.messageFavorites[:i], c.messageFavorites[i+1:]...)
			viper.Set(ConfigMessageFavorites, c.messageFavorites)
			SaveConfig()
			return
		}
	}
}

func (c *Config) GetMessageFavoriteIndex(favorite *MessageFavorite) int {
	for i, messageFavorite := range c.messageFavorites {
		if messageFavorite == favorite {
			return i
		}
	}
	return 0
}

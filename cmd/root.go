package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"yesbotics/ysm/internal/config"
	"yesbotics/ysm/internal/gui"
)

type Model struct {
	AppConfig config.AppConfig
}

var (
	// Used for flags.
	cfgFile  string
	baudrate uint
	port     string

	rootCmd = &cobra.Command{
		Use:   "ysm",
		Short: "Yesbotics Serial Monitor is a very fast static site generator",
		Long: `A fast and lightweight serial monitor with some helping functions to work with the 
Yesbotics Simple Serial Protocol (https://github.com/yesbotics/simple-serial-protocol-docs)`,
		Run: run,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	configUsage := "AppConfig file path (default is $HOME/.ysm.yaml)"
	baudrateUsage := fmt.Sprintf("Serial connection baudrate (default is %d)", config.DefaultBaudrate)
	portUsage := fmt.Sprintf("Serial connection port (default is %s)", config.DefaultPort)

	rootCmd.PersistentFlags().StringVar(&cfgFile, config.FlagConfig, "", configUsage)
	rootCmd.PersistentFlags().UintVarP(&baudrate, config.FlagBaudrate, "b", 0, baudrateUsage)
	rootCmd.PersistentFlags().StringVarP(&port, config.FlagPort, "p", "", portUsage)
	_ = viper.BindPFlag(config.FlagBaudrate, rootCmd.PersistentFlags().Lookup(config.FlagBaudrate))
	_ = viper.BindPFlag(config.FlagPort, rootCmd.PersistentFlags().Lookup(config.FlagPort))
	viper.SetDefault(config.FlagBaudrate, config.DefaultBaudrate)
	viper.SetDefault(config.FlagPort, config.DefaultPort)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ysm")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func run(cmd *cobra.Command, args []string) {
	model := newModel()
	_, err := gui.New(model.AppConfig)
	if err != nil {
		slog.Error(err.Error())
	}
}

func newModel() Model {
	m := Model{
		AppConfig: config.AppConfig{
			Config: config.NewConfig(),
		},
	}

	return m
}

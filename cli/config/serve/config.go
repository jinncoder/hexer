package config_serve

import (
	"fmt"
	"os"

	"github.com/archimoebius/hexer/cli/config"
	"github.com/archimoebius/hexer/util"
	"github.com/fatih/structs"
	"github.com/leebenson/conform"
	"github.com/sanity-io/litter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Setting is a global config object
var Setting *setting

// initial settings (defaults)
var initial = &setting{
	IP:           "127.0.0.1",
	Port:         "2222",
	DatabaseIP:   "127.0.0.1",
	DatabasePort: "8090",
	Local:        false,
	Theme:        config.DefaultTheme(),
	HostKey:      "",
	StoragePath:  "./storage",
}

// Create private data struct to hold setting options.
// `mapstructure` => viper tags
// `struct` => fatih structs tag
// `env` => environment variable name
type setting struct {
	Port         string        `mapstructure:"port" structs:"port" env:"HEXER_PORT"`
	IP           string        `mapstructure:"ip" structs:"ip" env:"HEXER_IP"`
	DatabasePort string        `mapstructure:"database_port" structs:"database_port" env:"HEXER_DATABASE_PORT"`
	DatabaseIP   string        `mapstructure:"database_ip" structs:"database_ip" env:"HEXER_DATABASE_IP"`
	Local        bool          `mapstructure:"local" structs:"local" env:"HEXER_LOCAL"`
	Theme        *config.Theme `mapstructure:"theme" structs:"theme"`
	HostKey      string        `mapstructure:"hostkey" structs:"hostkey"`
	StoragePath  string        `mapstructure:"storage_path" structs:"storage_path" env:"HEXER_STORAGE_PATH"`
}

func Load() {
	// Priority of configuration options
	// 1: CLI Parameters
	// 2: environment
	// 3: config.yaml
	// 4: defaults

	defaultsAsMap := structs.Map(initial)

	// Set defaults
	for key, value := range defaultsAsMap {
		viper.SetDefault(key, value)
	}
	// Read config from file
	viper.SetConfigType("yaml")

	viper.SetConfigName(".hexer.yaml")

	viper.AddConfigPath(".")
	local_config_path, err := util.ExpandTilde("~/.config/hexer/")
	if err == nil {
		viper.AddConfigPath(local_config_path)
	}
	viper.AddConfigPath("/etc/hexer/")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read configuration %s\n", viper.ConfigFileUsed())
	}

	// Unmarshal config into struct
	Setting = &setting{}
	err = viper.Unmarshal(Setting)
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
		os.Exit(1)
	}
}

// configInit must be called from the packages' init() func
func CommandInit(command *cobra.Command) error {
	// Keep cli parameters in sync with the config struct
	command.PersistentFlags().String("port", initial.Port, "The port to listen on for SSH connections")
	command.PersistentFlags().String("ip", initial.IP, "The IP to listen on for SSH connections")
	command.PersistentFlags().String("database_port", initial.DatabasePort, "The port to listen on for Database connections")
	command.PersistentFlags().String("database_ip", initial.DatabaseIP, "The IP to listen on for Database connections")
	command.PersistentFlags().String("storage_path", initial.StoragePath, "The folder to store the database in")
	command.PersistentFlags().Bool("local", false, "Don't start an SSH handler for the TUI, just run the TUI locally")

	for _, field := range structs.Fields(&setting{}) {
		// Get the struct tag values
		key := field.Tag("structs")

		if key == "" {
			continue
		}

		env := field.Tag("env")

		// Bind cobra flags to viper
		err := viper.BindPFlag(key, command.PersistentFlags().Lookup(key))
		if err != nil {
			if key == "theme" || key == "hostkey" {
				continue
			}
			return err
		}
		err = viper.BindEnv(key, env)
		if err != nil {
			return err
		}
	}

	return nil
}

// Print the config object
// but remove sensitive data
func (c *setting) Print() {
	cp := *c
	_ = conform.Strings(&cp)
	litter.Dump(cp)
}

// String the config object
// but remove sensitive data
func (c *setting) String() string {
	cp := *c
	_ = conform.Strings(&cp)
	return litter.Sdump(cp)
}

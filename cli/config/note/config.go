package config_note

import (
	"fmt"
	"os"

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
	IP:        "127.0.0.1",
	Port:      "2222",
	Local:     false,
	Export:    false,
	ProjectId: "",
	SSHKey:    "",
	SSHUser:   "",
}

// Create private data struct to hold setting options.
// `mapstructure` => viper tags
// `struct` => fatih structs tag
// `env` => environment variable name
type setting struct {
	Port      string `mapstructure:"port" structs:"port" env:"hexer_PORT"`
	IP        string `mapstructure:"ip" structs:"ip" env:"hexer_IP"`
	ProjectId string `mapstructure:"project-id" structs:"project-id" env:"hexer_note_PROJECT_ID"`
	Local     bool   `mapstructure:"local" structs:"local" env:"hexer_LOCAL"`
	Export    bool   `mapstructure:"export" structs:"export" env:"hexer_note_EXPORT"`
	SSHKey    string `mapstructure:"ssh-key" structs:"ssh-key" env:"hexer_note_SSHKEY"`
	SSHUser   string `mapstructure:"ssh-user" structs:"ssh-user" env:"hexer_note_SSHUSER"`
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

	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	// }

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
	command.PersistentFlags().String("port", initial.Port, "The port to listen on for SSH connections - if not set, will bind to a random high port (if not local)")
	command.PersistentFlags().String("ip", initial.IP, "The IP to listen on for SSH connections - if not set, will bind to 127.0.0.1 (if not local)")
	command.PersistentFlags().Bool("local", initial.Local, "Don't start an SSH handler for the TUI, just run the TUI locally")
	command.PersistentFlags().String("ssh-key", initial.SSHKey, "The private SSH key to leverage to connecting (if not local)")
	command.PersistentFlags().String("ssh-user", initial.SSHUser, "The private SSH user to leverage to connecting (if not local)")
	command.PersistentFlags().String("project-id", initial.ProjectId, "The project ID to record notes under")
	command.PersistentFlags().Bool("export", initial.Export, "Instead of uploading notes - download them")

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

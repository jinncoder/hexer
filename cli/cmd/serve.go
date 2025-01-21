package cli

import (
	"log"

	"github.com/archimoebius/hexer/app"
	rootConfig "github.com/archimoebius/hexer/cli/config/root"
	config "github.com/archimoebius/hexer/cli/config/serve"
	"github.com/archimoebius/hexer/util"
	"github.com/leebenson/conform"
	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server on the desired ip/port",
	Long:  `Leveraging the goodness of Golang and SSH - setup a local server for recording HTB credentials & flags`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		CallPersistentPreRun(cmd, args)
		config.Load()

		if rootConfig.Setting.Debug {
			util.Logger.SetReportCaller(true)
			config.Setting.Print()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := app.NewApplication().Start()

		if err != nil {
			util.Logger.Error(err)
		}
	},
}

// init is called before main
func init() {
	// A custom sanitizer to redact sensitive data by defining a struct tag= named "redact".
	conform.AddSanitizer("redact", func(_ string) string { return "*****" })

	// Initialize the config and panic on failure
	if err := config.CommandInit(ServeCmd); err != nil {
		log.Fatal(err)
	}
}

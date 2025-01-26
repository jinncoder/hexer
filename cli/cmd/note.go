package cli

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/archimoebius/hexer/app"
	config "github.com/archimoebius/hexer/cli/config/note"
	rootConfig "github.com/archimoebius/hexer/cli/config/root"
	"github.com/archimoebius/hexer/util"
	"github.com/archimoebius/hexer/util/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/leebenson/conform"
	"github.com/spf13/cobra"
)

var NoteCmd = &cobra.Command{
	Use:   "note",
	Short: "Record history as notes for a project",
	Long:  `Having used the 'execute' command to record execution, record the results to a project`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		CallPersistentPreRun(cmd, args)
		config.Load()

		if rootConfig.Setting.Debug {
			util.Logger.SetReportCaller(true)
			config.Setting.Print()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		if !config.Setting.Local {
			if config.Setting.SSHKey == "" || config.Setting.SSHUser == "" || config.Setting.IP == "" || config.Setting.Port == "" {
				util.Logger.Error("When not in local mode - sshkey, sshuser, ip, and port are all required")
				os.Exit(1)
			}

			if config.Setting.SSHUser != "" {
				usernamePlain, _, projectId := util.GetUsernameProjectIfPresent(config.Setting.SSHUser)
				config.Setting.SSHUser = usernamePlain

				if config.Setting.ProjectId == "" { // enable the default project-id in the config to be override by command line if present
					config.Setting.ProjectId = projectId
				}
			}
		}

		if config.Setting.ProjectId == "" {
			util.Logger.Error("project id is required")
			os.Exit(2)
		}

		if config.Setting.Export {

			if config.Setting.Local {
				_ = app.SetupPocketbase(config.Setting.StoragePath)

				data, err := database.GetNotesAsMarkdown(config.Setting.ProjectId)
				if err != nil {
					util.Logger.Error(fmt.Sprintf("error reading note: %v", err))
					os.Exit(3)
				}

				fmt.Println(string(data))

			} else {
				err := util.ReadNote(config.Setting.SSHUser, config.Setting.SSHKey, config.Setting.ProjectId, config.Setting.IP, config.Setting.Port)
				if err != nil {
					util.Logger.Error(fmt.Sprintf("error reading note: %v", err))
					os.Exit(3)
				}
			}
		} else {
			db, err := util.CreateSQLite3Database()
			if db != nil {
				defer db.Close()
			}
			if err != nil {
				util.Logger.Error(fmt.Sprintf("error creating database: %v", err))
				os.Exit(3)
			}

			historySelection, err := util.GetHistory(db, 20)
			if err != nil {
				util.Logger.Error(fmt.Sprintf("error querying database: %v", err))
				os.Exit(3)
			}

			var options []huh.Option[util.HistorySelection]
			var selectedHistoryItems []util.HistorySelection
			for _, cmd := range historySelection {
				options = append(options, huh.NewOption(cmd.Command, cmd))
			}
			var noteTitle = ""
			var noteComment = ""

			_ = tea.ClearScreen()

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Title").
						Placeholder("The title for the note").
						Value(&noteTitle).
						Description("The contents of the following commands/arguments + output if any will be used to create the note"),
					huh.NewText().
						Title("Comment").
						Placeholder("Any context you wish to add for the command('s) that were executed").
						CharLimit(5000).
						Value(&noteComment).
						WithHeight(5),
					huh.NewMultiSelect[util.HistorySelection]().
						Title("Select 1..N to create a new note").
						Description("If a command starts with `[!]` - its exitcode was none 0").
						Height(25).
						Value(&selectedHistoryItems).
						Options(options...),
				),
			).WithTheme(huh.ThemeCharm())

			if err := form.Run(); err != nil {
				if form.State == huh.StateAborted {
					util.Logger.Error("aborting note selection/upload")
					os.Exit(0)
				}

				util.Logger.Error(fmt.Sprintf("error querying database: %v", err))
				os.Exit(4)
			}

			sort.Slice(selectedHistoryItems, func(x, y int) bool {
				return selectedHistoryItems[x].Id > selectedHistoryItems[y].Id
			})

			var ids []int
			for _, c := range selectedHistoryItems {
				ids = append(ids, c.Id)
			}

			note, err := util.GetNote(db, ids, noteTitle, noteComment)
			if err != nil {
				util.Logger.Error(fmt.Sprintf("error querying database: %v", err))
				os.Exit(4)
			}

			if config.Setting.Local {
				_ = app.SetupPocketbase(config.Setting.StoragePath)

				err := database.AddNote(config.Setting.ProjectId, noteTitle, note)

				if err != nil {
					util.Logger.Error(fmt.Sprintf("error saving note: %v", err))
					os.Exit(5)
				}
			} else {

				note = fmt.Sprintf("%s\r\n%s", noteTitle, note)

				err = util.SendNote(note, config.Setting.SSHUser, config.Setting.SSHKey, config.Setting.ProjectId, config.Setting.IP, config.Setting.Port)
				if err != nil {
					util.Logger.Error(fmt.Sprintf("error sending note: %v", err))
					os.Exit(5)
				}
			}
		}
	},
}

// init is called before main
func init() {
	// A custom sanitizer to redact sensitive data by defining a struct tag= named "redact".
	conform.AddSanitizer("redact", func(_ string) string { return "*****" })

	// Initialize the config and panic on failure
	if err := config.CommandInit(NoteCmd); err != nil {
		log.Fatal(err)
	}
}

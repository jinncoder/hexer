package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/archimoebius/hexer/util"
	_ "github.com/mattn/go-sqlite3"

	"github.com/spf13/cobra"
)

var ExecuteCmd = &cobra.Command{
	Use:                "execute",
	Short:              "Wrap and collect executing command and resulting output",
	Long:               "Provided with a command to execute, record it and the returned output from execution (including exitcode)",
	DisableFlagParsing: true,
	Run: func(_ *cobra.Command, args []string) {
		executable := args[0]

		db, err := util.CreateSQLite3Database()
		if db != nil {
			defer db.Close()
		}
		if err != nil {
			fmt.Printf("error creating database: %v", err)
			return
		}

		arguments := strings.Join(args[1:], " ")

		envVars := os.Environ()
		environment := strings.Join(envVars, ":")

		insertSQL := `INSERT INTO history (executable, arguments, environment) VALUES (?, ?, ?)`

		result, err := db.Exec(insertSQL, executable, arguments, environment)
		if err != nil {
			fmt.Printf("error inserting record: %v", err)
			return
		}
		historyId, err := result.LastInsertId()
		if err != nil {
			fmt.Printf("error obtaining historyId: %v", err)
			return
		}

		cmd := exec.Command(executable, args[1:]...) // #nosec G204 - indeed, we do not care

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println("Error creating stdout pipe:", err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Println("Error creating stderr pipe:", err)
			return
		}

		if err := cmd.Start(); err != nil {
			fmt.Println("Error starting command:", err)
			return
		}

		var stdoutOutput, stderrOutput string

		stdoutScan := bufio.NewScanner(stdout)
		for stdoutScan.Scan() {
			data := stdoutScan.Text()
			fmt.Println(data)
			stdoutOutput += data + "\n"
		}

		stderrScan := bufio.NewScanner(stderr)
		for stderrScan.Scan() {
			data := stderrScan.Text()
			fmt.Println(data)
			stderrOutput += data + "\n"
		}

		var exitcode = -1

		if err := cmd.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitcode = exitError.ExitCode()
			} else {
				err := fmt.Sprintf("Error waiting for command: %v", err)
				fmt.Println(err)
				stderrOutput += err
				exitcode = -2
			}
		} else {
			exitcode = 0
		}

		insertSQL = `INSERT INTO output (history_id, stdout, stderr, exitcode) VALUES (?, ?, ?, ?)`

		_, err = db.Exec(insertSQL, historyId, stdoutOutput, stderrOutput, exitcode)
		if err != nil {
			fmt.Printf("error inserting record: %v", err)
			return
		}
	},
}

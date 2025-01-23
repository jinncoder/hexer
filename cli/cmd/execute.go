package cli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/archimoebius/hexer/util"
	"github.com/creack/pty"
	_ "github.com/mattn/go-sqlite3"

	"github.com/spf13/cobra"
)

const SQL_INSERT_OUTPUT = `INSERT INTO output (history_id, stdout, stderr, exitcode) VALUES (?, ?, ?, ?)`
const SQL_UPSERT_OUTPUT = `INSERT INTO output (history_id, stdout, stderr, exitcode) VALUES (?, ?, ?, ?) ON CONFLICT(history_id) DO UPDATE SET stdout = excluded.stdout, stderr = excluded.stderr, exitcode = excluded.exitcode;`
const SQL_INSERT_HISTORY = `INSERT INTO history (executable, arguments, environment) VALUES (?, ?, ?)`

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

		result, err := db.Exec(SQL_INSERT_HISTORY, executable, arguments, environment)
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

		// Create pipes for stdout and stderr
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Failed to create stdout pipe: %v", err)
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			log.Fatalf("Failed to create stderr pipe: %v", err)
		}

		// Create a pseudo-terminal
		ptmx, err := pty.Start(cmd)
		if err != nil {
			log.Fatalf("Failed to start pty: %v", err)
		}
		defer ptmx.Close()

		var stdoutOutput, stderrOutput string

		// Read from the PTY in a separate goroutine

		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := ptmx.Read(buf)
				if err != nil {
					break
				}
				// Print the output from the PTY
				os.Stdout.Write(buf[:n])
				stdoutOutput += string(buf[:n])
			}
		}()

		go func() {
			// Create a scanner to read from standard input
			scanner := bufio.NewScanner(os.Stdin)
			// Read input line by line
			for scanner.Scan() {
				// Insert previous output into the database
				_, err = db.Exec(SQL_UPSERT_OUTPUT, historyId, stdoutOutput, stderrOutput, 99)
				if err != nil {
					fmt.Printf("error inserting record: %v", err)
					return
				}

				stdoutOutput = ""
				stderrOutput = ""

				commandLine := scanner.Text() // Get the current line
				args := strings.Split(commandLine, " ")
				result, err := db.Exec(SQL_INSERT_HISTORY, args[0], strings.Join(args[1:], " "), "")
				if err != nil {
					fmt.Printf("error inserting record: %v", err)
					return
				}
				historyId, err = result.LastInsertId()
				if err != nil {
					fmt.Printf("error obtaining historyId: %v", err)
					return
				}

				// Write the line to the command's stdin (PTY)
				if _, err := io.WriteString(ptmx, commandLine+"\n"); err != nil {
					log.Fatalf("Failed to write to pty: %v", err)
				}
			}
		}()

		go func() {
			stdoutScan := bufio.NewScanner(stdoutPipe)
			for stdoutScan.Scan() {
				data := stdoutScan.Text() + "\n"
				os.Stdout.WriteString(data)
				stdoutOutput += data
			}
		}()

		go func() {
			stderrScan := bufio.NewScanner(stderrPipe)
			for stderrScan.Scan() {
				data := stderrScan.Text() + "\n"
				os.Stderr.WriteString(data)
				stderrOutput += data
			}
		}()

		var exitcode = -1

		if err := cmd.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitcode = exitError.ExitCode()
				fmt.Fprintln(os.Stderr, exitError.Stderr)
			} else {
				err := fmt.Sprintf("Error waiting for command: %v", err)
				fmt.Fprintln(os.Stderr, err)
				stderrOutput += err
				exitcode = cmd.ProcessState.ExitCode()
			}
		} else {
			exitcode = cmd.ProcessState.ExitCode()
		}

		_, err = db.Exec(SQL_UPSERT_OUTPUT, historyId, stdoutOutput, stderrOutput, exitcode)
		if err != nil {
			fmt.Printf("error inserting record: %v", err)
			return
		}
	},
}

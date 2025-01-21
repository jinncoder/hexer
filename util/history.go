package util

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryOutput represents the combined result of history and output tables
type HistoryOutput struct {
	Executable  string
	Command     string
	Environment string
	ExitCode    int
	Stdout      string
	Stderr      string
}

type HistorySelection struct {
	Id      int
	Command string
}

// ExpandTilde expands the tilde in the given path to the user's home directory.
func ExpandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		// Replace the tilde with the home directory
		return filepath.Join(homeDir, path[1:]), nil
	}
	return path, nil
}

func CreateSQLite3Database() (*sql.DB, error) {
	databaseFilepath, err := ExpandTilde("~/.config/hexer/hexer-execute.db")
	if err != nil {
		return nil, fmt.Errorf("failed to expand directory: %v", err)
	}

	dirPath := filepath.Dir(databaseFilepath)

	err = os.MkdirAll(dirPath, 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Open a database connection
	db, err := sql.Open("sqlite3", databaseFilepath)
	if err != nil {
		return db, err
	}

	// Create the history table if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	executable TEXT,
	arguments TEXT,
	environment TEXT
);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return db, err
	}

	// Create the output table if it doesn't exist
	createTableSQL = `CREATE TABLE IF NOT EXISTS output (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	history_id INTEGER,
	exitcode INTEGER,
	stdout TEXT,
	stderr TEXT
);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return db, err
	}

	return db, nil
}

func GetHistory(db *sql.DB, limit int) ([]HistorySelection, error) {
	query := "SELECT history.id, o.exitcode, executable, arguments FROM history JOIN output o ON history.id == o.history_id LIMIT ?"

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var historySelections []HistorySelection

	for rows.Next() {
		var id, exitcode int
		var executable, arguments string

		if err := rows.Scan(&id, &exitcode, &executable, &arguments); err != nil {
			log.Fatal(err)
		}

		var start = "[+] "

		if exitcode != 0 {
			start = "[!] "
		}

		command := fmt.Sprintf("%s%s %s", start, executable, strings.TrimSpace(arguments))

		historySelections = append(historySelections, HistorySelection{
			Id:      id,
			Command: command,
		})
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return historySelections, nil
}

func convertToInterfaceSlice(ids []int) []interface{} {
	iface := make([]interface{}, len(ids))
	for i, v := range ids {
		iface[i] = v
	}
	return iface
}

func GetNote(db *sql.DB, historyIDs []int, title string, comment string) (string, error) {

	placeholders := make([]string, len(historyIDs))
	for i := range historyIDs {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`
		SELECT 
			h.executable,
			h.arguments,
			h.environment,
			o.exitcode,
			o.stdout,
			o.stderr
		FROM 
			history h
		JOIN 
			output o ON h.id = o.history_id
		WHERE 
			h.id IN (%s)
	`, strings.Join(placeholders, ", ")) // #nosec G201 - meh?
	rows, err := db.Query(query, convertToInterfaceSlice(historyIDs)...)
	if err != nil {
		log.Fatalf("Failed to execute query |%s|: %v", query, err)
	}
	defer rows.Close()

	var historyOutputs []HistoryOutput

	// Iterate through the result set
	for rows.Next() {
		var ho HistoryOutput
		var executable, arguments string

		// Scan the result into variables
		if err := rows.Scan(&executable, &arguments, &ho.Environment, &ho.ExitCode, &ho.Stdout, &ho.Stderr); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		// Construct the Command by joining Executable and Arguments
		if strings.TrimSpace(arguments) != "" {
			ho.Command = fmt.Sprintf("%s %s", executable, strings.TrimSpace(arguments))
		} else {
			ho.Command = executable
		}

		ho.Executable = executable

		// Append to the slice
		historyOutputs = append(historyOutputs, ho)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Fatalf("Error occurred during row iteration: %v", err)
	}

	var note = fmt.Sprintf("%s\n\n", comment)

	for _, output := range historyOutputs {
		start := "(+)"

		if output.ExitCode != 0 {
			start = "(!)"
		}

		// TODO: add datetime?
		// TODO: include environment?
		note += fmt.Sprintf("## %s Command: %s\n\n```bash\n> %s\n\n%s\n%s```\n\n", start, output.Executable, output.Command, output.Stdout, output.Stderr)
	}

	return note, nil
}

func SendNote(note string, username string, pk string, projectId string, ip string, port string) error {

	pkData, err := os.ReadFile(pk) // #nosec G304 - user supplied value, we do not care
	if err == nil {
		pk = string(pkData)
	} else {
		pk = string(FixSSHKeyData(pk))
	}

	config := SFTPConfig{
		Username:   fmt.Sprintf("%s-%s", username, projectId),
		Password:   "",
		PrivateKey: string(pk),
		Server:     fmt.Sprintf("%s:%s", ip, port),
		Timeout:    time.Second * 30,
	}

	client, err := NewSFTPClientWrapper(config)
	if err != nil {
		return err
	}
	defer client.Close()

	destination, err := client.Create("tmp/note.txt")
	if err != nil {
		return err
	}
	defer destination.Close()

	if err := client.Upload(strings.NewReader(note), destination, 1000000); err != nil {
		return err
	}

	return nil
}

func ReadNote(username string, pk string, projectId string, ip string, port string) error {

	pkData, err := os.ReadFile(pk) // #nosec G304 - user supplied value, we do not care
	if err == nil {
		pk = string(pkData)
	} else {
		pk = string(FixSSHKeyData(pk))
	}

	config := SFTPConfig{
		Username:   fmt.Sprintf("%s-%s", username, projectId),
		Password:   "",
		PrivateKey: string(pk),
		Server:     fmt.Sprintf("%s:%s", ip, port),
		Timeout:    time.Second * 30,
	}

	client, err := NewSFTPClientWrapper(config)
	if err != nil {
		return err
	}
	defer client.Close()

	file, err := client.Download("notes.md")
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(data))

	return nil
}

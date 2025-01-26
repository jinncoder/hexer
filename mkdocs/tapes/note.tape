Output mkdocs/docs/images/note.gif

Set FontSize 14
Set Width 1980
Set Height 1024

Sleep 1s
Type "# Let's start by creating a project"
Enter
Sleep 2s

Sleep 1s
Type "go run github.com/archimoebius/hexer@latest serve"
Enter
Sleep 4s
Type "?"
Sleep 5s
Type "a"
Sleep 1s
Type "Example Project"
Sleep 1s
Tab
Sleep 1s
Type "This is an example project to highlight note keeping abilities..."
Sleep 1s
Tab
Sleep 1s
Left
Sleep 1s
Enter
Sleep 3s
Ctrl+d
Sleep 1s

Type "# Now that we've a project - let's run some commands and record them!"
Enter
Sleep 2s

Type "go run github.com/archimoebius/hexer@latest execute ls -trhal /tmp/"
Enter
Sleep 3s
Type "go run github.com/archimoebius/hexer@latest execute curl qrenco.de/hexer"
Enter
Sleep 4s
Hide
Sleep 1s
Type "PROJECT_ID=`sqlite3 storage/data.db 'select id from project order by id desc limit 1;' | tr -d '\n'`"
Enter
Sleep 2s
Show
Ctrl+l
Sleep 1s
Type "go run github.com/archimoebius/hexer@latest note --project-id $PROJECT_ID"
Enter
Sleep 2s
Type "This is an example note!"
Sleep 1s
Tab
Type "This comment goes below the header (note title) when markdown is generated."
Sleep 2s
Tab
Space
Sleep 1s
Down
Space
Sleep 2s
Enter
Sleep 3s

Type "# Let's go check that these notes made it to the server"
Enter
Sleep 2s

Type "go run github.com/archimoebius/hexer@latest serve"
Enter
Sleep 4s
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 2s
Enter
Sleep 2s
Right
Sleep 500ms
Right
Sleep 500ms
Right
Sleep 500ms
Right
Sleep 3s
Ctrl+p
Sleep 3s
PageDown
PageDown
Sleep 4s
Escape
Sleep 500ms
Escape
Sleep 500ms
Escape
Sleep 500ms
Ctrl+d
Sleep 1s

Type "# Finally - let's export our notes to markdown..."
Enter
Sleep 5s

Type "go run github.com/archimoebius/hexer@latest note --project-id $PROJECT_ID --export | go run github.com/charmbracelet/glow@latest -p"
Enter
Sleep 5s
PageDown
PageDown
PageDown
Sleep 5s
Type "q"
Sleep 1s

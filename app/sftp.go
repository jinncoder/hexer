package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/archimoebius/hexer/app/cache"
	"github.com/archimoebius/hexer/util"
	"github.com/archimoebius/hexer/util/database"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
)

type sftpHandler struct {
	fileSystem *InMemoryFileSystem
	root       string
	session    ssh.Session
}

var (
	_ sftp.FileLister = &sftpHandler{}
	_ sftp.FileReader = &sftpHandler{}
	_ sftp.FileWriter = &sftpHandler{}
	_ sftp.FileCmder  = &sftpHandler{}
)

func (s *sftpHandler) Filecmd(request *sftp.Request) error {
	return sftp.ErrSSHFxOpUnsupported
}

func (s *sftpHandler) setupNotes(request *sftp.Request) error {

	_, _, projectId := util.GetUsernameProjectIfPresent(s.session.User())

	if projectId == "" {
		return sftp.ErrSSHFxOpUnsupported
	}

	if filepath.Base(request.Filepath) != "notes.md" {
		return sftp.ErrSSHFxNoSuchFile
	}

	data, err := database.GetNotesAsMarkdown(projectId)
	if err != nil {
		return sftp.ErrSSHFxFailure
	}

	s.fileSystem.AddFile("notes.md", []byte(data))

	return nil
}

// Fileread implements sftp.FileReader.
func (s *sftpHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	_, _, projectId := util.GetUsernameProjectIfPresent(s.session.User())

	if projectId == "" {
		return nil, sftp.ErrSSHFxOpUnsupported
	}

	if filepath.Base(r.Filepath) != "notes.md" {
		return nil, sftp.ErrSSHFxNoSuchFile
	}

	var data = s.fileSystem.files["notes.md"]
	if data == nil {
		err := s.setupNotes(r)
		if err != nil {
			return nil, err
		}

		data = s.fileSystem.files["notes.md"]
		if data == nil {
			return nil, sftp.ErrSSHFxNoSuchFile
		}
	}

	return strings.NewReader(string(data.content)), nil
}

// Filelist implements sftp.FileLister.
func (s *sftpHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {

	if r.Method != "Stat" {
		return nil, sftp.ErrSSHFxFailure
	}

	entry := s.fileSystem.files["notes.md"]
	if entry == nil {
		err := s.setupNotes(r)
		if err != nil {
			return nil, err
		}

		entry = s.fileSystem.files["notes.md"]
		if entry == nil {
			return nil, sftp.ErrSSHFxNoSuchFile
		}
	}

	var fileInfos []os.FileInfo
	fileInfos = append(fileInfos, &InMemoryFileInfo{name: entry.name, size: int64(len(entry.content)), mode: 0644, modTime: time.Now()})

	return ListerAt(fileInfos), nil
}

// Filewrite implements sftp.FileWriter.
func (s *sftpHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	_, _, projectId := util.GetUsernameProjectIfPresent(s.session.User())

	if projectId == "" {
		return nil, sftp.ErrSSHFxOpUnsupported
	}

	uuid4, err := uuid.NewRandom()
	if err != nil {
		return nil, sftp.ErrInternalInconsistency
	}

	filePath := filepath.Join(s.root, uuid4.String())

	file, err := os.Create(filePath) // #nosec
	if err != nil {
		return nil, sftp.ErrSSHFxFailure
	}

	s.session.Context().SetValue(util.ContextKeyFilepath, filePath)

	return file, nil
}

// InMemoryFile represents an in-memory file.
type InMemoryFile struct {
	name    string
	content []byte
	offset  int64
}

// NewInMemoryFile creates a new in-memory file.
func NewInMemoryFile(name string, content []byte) *InMemoryFile {
	return &InMemoryFile{name: name, content: content}
}

// Read reads data from the in-memory file.
func (f *InMemoryFile) Read(b []byte) (n int, err error) {
	if f.offset >= int64(len(f.content)) {
		return 0, io.EOF
	}
	n = copy(b, f.content[f.offset:])
	f.offset += int64(n)
	return n, nil
}

// Stat returns the file's information.
func (f *InMemoryFile) Stat() (os.FileInfo, error) {
	return &InMemoryFileInfo{name: f.name, size: int64(len(f.content)), mode: 0644, modTime: time.Now()}, nil
}

// InMemoryFileInfo implements os.FileInfo for in-memory files.
type InMemoryFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name returns the name of the file.
func (fi *InMemoryFileInfo) Name() string {
	return fi.name
}

// Size returns the size of the file.
func (fi *InMemoryFileInfo) Size() int64 {
	return fi.size
}

// Mode returns the file's mode.
func (fi *InMemoryFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime returns the modification time of the file.
func (fi *InMemoryFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns whether the file is a directory.
func (fi *InMemoryFileInfo) IsDir() bool {
	return false
}

// Sys returns the underlying data source.
func (fi *InMemoryFileInfo) Sys() interface{} {
	return nil
}

// InMemoryFileSystem implements the sftp.FileSystem interface.
type InMemoryFileSystem struct {
	files map[string]*InMemoryFile
}

// NewInMemoryFileSystem creates a new in-memory file system.
func NewInMemoryFileSystem() *InMemoryFileSystem {
	return &InMemoryFileSystem{files: make(map[string]*InMemoryFile)}
}

// AddFile adds a new file to the in-memory file system.
func (fs *InMemoryFileSystem) AddFile(name string, content []byte) {
	fs.files[name] = NewInMemoryFile(name, content)
}

type ListerAt []os.FileInfo

func (l ListerAt) ListAt(f []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l)) {
		return 0, io.EOF
	}

	if n := copy(f, l[offset:]); n < len(f) {
		return n, io.EOF
	} else {
		return n, nil
	}
}

func (fs *InMemoryFileSystem) List(path string) (sftp.ListerAt, error) {
	var fileInfos []os.FileInfo
	for _, file := range fs.files {
		fileInfos = append(fileInfos, &InMemoryFileInfo{name: file.name, size: int64(len(file.content)), mode: 0644, modTime: time.Now()})
	}
	return ListerAt(fileInfos), nil
}

func sftpSubsystem(root string) ssh.SubsystemHandler {
	return func(s ssh.Session) {

		if !validateSessionUser(s) {
			wish.Fatalln(s, "You're account is not verified - please contact your administrator")
			return
		}

		_, usernameHash, projectId := util.GetUsernameProjectIfPresent(s.User())

		if projectId == "" {
			wish.Fatalln(s, "Project ID required - RTFM")
			return
		}

		userSSHPublicKey, err := cache.GetUserPublicSSHKeyFromCache(usernameHash)

		if err != nil {
			wish.Fatalln(s, err.Error())
			return
		}

		if !ssh.KeysEqual(s.PublicKey(), userSSHPublicKey) {
			wish.Fatalln(s, "You're account is not verified - please contact your administrator")
			return
		}

		fs := &sftpHandler{
			fileSystem: NewInMemoryFileSystem(),
			root:       root,
			session:    s,
		}

		srv := sftp.NewRequestServer(s, sftp.Handlers{
			FileList: fs,
			FileGet:  fs,
			FilePut:  fs,
			FileCmd:  fs,
		})

		if err := srv.Serve(); err == io.EOF {
			if err := srv.Close(); err != nil {
				wish.Fatalln(s, "sftp:", err)
				return
			}
		} else if err != nil {
			wish.Fatalln(s, "sftp:", err)
			return
		}

		var filepath = s.Context().Value(util.ContextKeyFilepath)
		if filepath == nil {
			wish.Fatalln(s, "sftp: upload failed - filepath was nil")
			return
		}

		err = database.AddNoteFromFilepath(projectId, fmt.Sprintf("%v", filepath))
		if err == nil {
			wish.Fatalln(s, fmt.Sprintf("sftp: upload failed %v", err))
			return
		}
	}
}

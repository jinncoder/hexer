package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/archimoebius/hexer/app/cache"
	serveConfig "github.com/archimoebius/hexer/cli/config/serve"
	"github.com/archimoebius/hexer/tui"
	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util"
	"github.com/archimoebius/hexer/util/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/scp"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"golang.org/x/term"

	_ "github.com/archimoebius/hexer/migrations"
)

// Application is the interface for the application
type Application interface {
	Start() error
}

// app is the implementation of the application
type app struct {
}

// NewApplication creates a new application
func NewApplication() Application {
	return &app{}
}

var pocketbaseApplication *pocketbase.PocketBase = nil

func SetupPocketbase(storagePath string) *pocketbase.PocketBase {
	if pocketbaseApplication == nil {
		// loosely check if it was executed using "go run"
		isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

		pocketbaseApplication = pocketbase.NewWithConfig(pocketbase.Config{
			DefaultDataDir:  storagePath,
			DefaultDev:      false,
			HideStartBanner: !isGoRun,
		})

		migratecmd.MustRegister(pocketbaseApplication, pocketbaseApplication.RootCmd, migratecmd.Config{
			// enable auto creation of migration files when making collection changes in the Dashboard
			// (the isGoRun check is to enable it only during development)
			Automigrate: isGoRun,
		})

		pocketbaseApplication.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
			fmt.Println("Exiting application derpyderp")
			return e.Next()
		})

		err := pocketbaseApplication.Bootstrap()
		if err != nil {
			log.Error("Could not start database server", "error", err)
			os.Exit(1)
		}

		var db = pocketbaseApplication.DB()

		err = database.SetDatabaseInstance(&db)
		if err != nil {
			log.Error("Could not start database server", "error", err)
			os.Exit(1)
		}
	}

	return pocketbaseApplication
}

func (a *app) Start() error {
	app := SetupPocketbase(serveConfig.Setting.StoragePath)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := apis.Serve(app, apis.ServeConfig{
			HttpAddr:        fmt.Sprintf("%s:%s", serveConfig.Setting.DatabaseIP, serveConfig.Setting.DatabasePort),
			ShowStartBanner: app.IsDev(),
			AllowedOrigins:  []string{"*"},
		})

		if errors.Is(err, http.ErrServerClosed) {
			log.Error("Could not start database server", "error", err)
			done <- nil
		}
	}()

	if serveConfig.Setting.Local {
		fd := int(os.Stdin.Fd())
		width, height, err := term.GetSize(fd)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting terminal size: %v - defaulting\n", err)
			width = 100
			height = 100
		}

		model, err := tui.NewModel(
			tui.NewInput(
				nil,
				constant.ModeListProject,
				constant.NewListProjectInput(width, height),
			),
		)
		if err != nil {
			log.Error("Could not start database server", "error", err)
			done <- nil
		}

		_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
		if err != nil {
			log.Error("Could not start new program with tea", "error", err)
			done <- nil
		}
	} else {
		users, err := database.GetUsers()

		if err != nil {
			log.Error("Could not start database server - no users?", "error", err)
			done <- nil
		}

		for _, user := range users {
			cache.AddUserKeyToCache(user.Name, user.SSHPublicKey)
		}

		rootSFTPFilepath, _ := filepath.Abs(fmt.Sprintf("%s/sftp", serveConfig.Setting.StoragePath))

		err = os.MkdirAll(rootSFTPFilepath, 0750)
		if err != nil {
			log.Error("failed to create directory: %v", err)
			done <- nil
		}

		handlerSFTP := scp.NewFileSystemHandler(rootSFTPFilepath)

		svr, err := wish.NewServer(
			wish.WithAddress(net.JoinHostPort(serveConfig.Setting.IP, serveConfig.Setting.Port)),
			wish.WithHostKeyPEM(util.FixSSHKeyData(serveConfig.Setting.HostKey)),
			wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
				usernamePlain, usernameHash, projectId := util.GetUsernameProjectIfPresent(ctx.User())

				if projectId != "" {
					ctx.SetValue(util.ContextKeyProjectId, projectId)
				}

				if strings.EqualFold(strings.ToLower(usernamePlain), "register") {
					return true
				}

				userSSHPublicKey, err := cache.GetUserPublicSSHKeyFromCache(usernameHash)

				if err != nil {
					return false
				}

				return ssh.KeysEqual(key, userSSHPublicKey)
			}),
			// setup the sftp subsystem
			wish.WithSubsystem("sftp", sftpSubsystem(rootSFTPFilepath)),
			wish.WithMiddleware(
				hexerMiddleware(),
				scp.Middleware(handlerSFTP, handlerSFTP),
				onlyVerifiedUsersMiddleware(),
				activeterm.Middleware(),
				noSSHCommandMiddleware(),
				logging.Middleware(),
			),
		)
		if err != nil {
			log.Error("Could not create server", "error", err)
			done <- nil
		} else {

			log.Info("Starting SSH server", "host", serveConfig.Setting.IP, "port", serveConfig.Setting.Port)
			go func() {
				if err = svr.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
					log.Error("Could not start server", "error", err)
					done <- nil
				}
			}()
		}

		<-done
		log.Info("Stopping SSH server")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer func() { cancel() }()
		if err := svr.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not stop server", "error", err)
		}
	}

	return nil
}

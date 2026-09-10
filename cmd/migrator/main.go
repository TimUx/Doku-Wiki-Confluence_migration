package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/config"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/server"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/store"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && args[0] == "--version" {
		fmt.Println(version)
		return nil
	}
	check := len(args) > 0 && args[0] == "check-config"
	if check {
		args = args[1:]
	}
	fs := flag.NewFlagSet("dokuwiki-confluence-migrator", flag.ContinueOnError)
	path := fs.String("config", "config.yaml", "path to config.yaml")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load(*path)
	if err != nil {
		return err
	}
	if check {
		fmt.Println("configuration valid")
		return nil
	}
	db, err := store.Open(cfg.Storage.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	app := server.New(cfg, db)
	httpServer := &http.Server{Addr: cfg.Server.Listen, Handler: app.Handler(), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdown)
	}()
	slog.Info("server started", "listen", cfg.Server.Listen)
	err = httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

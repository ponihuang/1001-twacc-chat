package app

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"1001-twacc-chat/internal/auth"
	"1001-twacc-chat/internal/chat"
	"1001-twacc-chat/internal/config"
	"1001-twacc-chat/internal/erp"
	"1001-twacc-chat/internal/httpx"
	storemysql "1001-twacc-chat/internal/store/mysql"
)

// Run starts the HTTP server for the chat system scaffold.
func Run() error {
	cfg := config.Load()

	var integrationHandler *erp.Handler
	var chatHandler *chat.Handler
	realtimeHub := chat.NewHub()
	if cfg.EnableMySQL {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dbStore, err := storemysql.Open(ctx, cfg.MySQL)
		if err != nil {
			return err
		}
		defer dbStore.Close()

		if cfg.AutoMigrate {
			migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer migrateCancel()

			applied, err := storemysql.ApplyMigrations(migrateCtx, dbStore.DB(), cfg.MigrationsDir)
			if err != nil {
				return err
			}
			if applied > 0 {
				fmt.Printf("applied %d migrations from %s\n", applied, cfg.MigrationsDir)
			}
		}

		sessionService := auth.NewService(storemysql.NewSessionRepository(dbStore.DB()), cfg.LoginTokenTTL)
		integrationRepo := storemysql.NewIntegrationRepository(dbStore.DB())
		integrationService := erp.NewService(integrationRepo, sessionService)
		integrationService.SetRequireTrustedDevice(cfg.RequireTrustedDevice)
		integrationHandler = erp.NewHandler(
			integrationService,
			sessionService,
			cfg.IntegrationSharedToken,
			cfg.IntegrationSignatureSecret,
			cfg.IntegrationTimestampTolerance,
		)

		chatRepo := storemysql.NewChatRepository(dbStore.DB())
		chatService := chat.NewService(chatRepo, realtimeHub)
		chatHandler = chat.NewHandler(chatService, sessionService, realtimeHub, chat.NewLocalFileStore(filepath.Join("web", "uploads")))
	} else {
		integrationHandler = erp.NewHandler(
			nil,
			nil,
			cfg.IntegrationSharedToken,
			cfg.IntegrationSignatureSecret,
			cfg.IntegrationTimestampTolerance,
		)
		chatHandler = chat.NewHandler(nil, nil, nil, nil)
	}

	server := &http.Server{
		Addr: cfg.ListenHost + ":" + cfg.Port,
		Handler: httpx.NewRouter(integrationHandler, chatHandler, httpx.UIConfig{
			IntegrationSharedToken: cfg.IntegrationSharedToken,
		}),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	fmt.Printf("server listening on http://%s:%s\n", cfg.ListenHost, cfg.Port)
	return server.ListenAndServe()
}

package app

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"1001-twacc-chat/internal/adminauth"
	"1001-twacc-chat/internal/auth"
	"1001-twacc-chat/internal/chat"
	"1001-twacc-chat/internal/config"
	"1001-twacc-chat/internal/erp"
	"1001-twacc-chat/internal/httpx"
	"1001-twacc-chat/internal/mailer"
	storemysql "1001-twacc-chat/internal/store/mysql"
)

// Run starts the HTTP server for the chat system scaffold.
func Run() error {
	cfg := config.Load()

	var integrationHandler *erp.Handler
	var chatHandler *chat.Handler
	realtimeHub := chat.NewHub()
	appCtx, stopApp := context.WithCancel(context.Background())
	defer stopApp()
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
		adminSessionService := adminauth.NewService(storemysql.NewAdminSessionRepository(dbStore.DB()), cfg.LoginTokenTTL)
		integrationRepo := storemysql.NewIntegrationRepository(dbStore.DB())
		integrationService := erp.NewService(integrationRepo, sessionService)
		integrationService.SetAdminSessions(adminSessionService)
		integrationService.SetRequireTrustedDevice(cfg.RequireTrustedDevice)
		integrationService.SetInvitationTTL(cfg.UserInvitationTTL)
		integrationService.SetInvitationBaseURL(cfg.AppURL)
		var smtpMailer *mailer.SMTPMailer
		if cfg.Mail.Enabled() {
			smtpMailer = mailer.NewSMTPMailer(cfg.Mail)
			integrationService.SetInvitationMailer(smtpMailer)
		}
		integrationHandler = erp.NewHandler(
			integrationService,
			sessionService,
			adminSessionService,
			cfg.IntegrationSharedToken,
			cfg.IntegrationSignatureSecret,
			cfg.IntegrationTimestampTolerance,
		)

		chatRepo := storemysql.NewChatRepository(dbStore.DB())
		chatService := chat.NewService(chatRepo, realtimeHub)
		if smtpMailer != nil {
			go chat.NewEmailNotificationWorker(chatRepo, smtpMailer).Start(appCtx)
		}
		chatHandler = chat.NewHandler(chatService, sessionService, realtimeHub, chat.NewLocalFileStore(filepath.Join("web", "uploads")))
	} else {
		integrationHandler = erp.NewHandler(
			nil,
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

	if cfg.Server.TLSCertFile != "" || cfg.Server.TLSKeyFile != "" {
		if cfg.Server.TLSCertFile == "" || cfg.Server.TLSKeyFile == "" {
			return fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE must be set together")
		}
		fmt.Printf("server listening on https://%s:%s\n", cfg.ListenHost, cfg.Port)
		return server.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile)
	}

	fmt.Printf("server listening on http://%s:%s\n", cfg.ListenHost, cfg.Port)
	return server.ListenAndServe()
}

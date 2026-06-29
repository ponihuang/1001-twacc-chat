package app

import (
	"context"
	"flag"
	"fmt"
	"time"

	"1001-twacc-chat/internal/auth"
	"1001-twacc-chat/internal/config"
	"1001-twacc-chat/internal/erp"
	storemysql "1001-twacc-chat/internal/store/mysql"
)

// SeedAdmin bootstraps the first office system-admin account.
func SeedAdmin(args []string) error {
	flags := flag.NewFlagSet("seed-admin", flag.ContinueOnError)
	account := flags.String("account", "", "管理員帳號")
	password := flags.String("password", "", "管理員密碼")
	name := flags.String("name", "", "管理員名稱")
	status := flags.String("status", "active", "帳號狀態: active 或 inactive")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if *account == "" {
		return fmt.Errorf("missing required flag: --account")
	}
	if *password == "" {
		return fmt.Errorf("missing required flag: --password")
	}
	if *name == "" {
		*name = *account
	}

	cfg := config.Load()
	if !cfg.MySQL.Enabled() {
		return fmt.Errorf("mysql config is incomplete")
	}

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

	repo := storemysql.NewIntegrationRepository(dbStore.DB())
	sessionService := auth.NewService(storemysql.NewSessionRepository(dbStore.DB()), cfg.LoginTokenTTL)
	service := erp.NewService(repo, sessionService)

	result, err := service.BootstrapSystemAdmin(erp.SystemAdminCreateRequest{
		ExternalUserID: *account,
		Password:       *password,
		DisplayName:    *name,
		Status:         *status,
	})
	if err != nil {
		return err
	}

	fmt.Printf("system admin ready: account=%s source=office user_created=%t admin_created=%t\n",
		result.User.ExternalUserID,
		result.UserCreated,
		result.AdminCreated,
	)
	return nil
}

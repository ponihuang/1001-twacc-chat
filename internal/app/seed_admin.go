package app

import (
	"context"
	"flag"
	"fmt"
	"time"

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
	resetPassword := flags.Bool("reset-password", false, "帳號已存在時重設密碼")

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
	service := erp.NewService(repo, nil)

	result, err := service.BootstrapSystemAdmin(erp.SystemAdminCreateRequest{
		ExternalUserID: *account,
		Password:       *password,
		DisplayName:    *name,
		Status:         *status,
	})
	if err != nil {
		return err
	}

	passwordReset := false
	if !result.AdminCreated && *resetPassword {
		resp, _, err := service.UpdateSystemAdmin(result.Admin.AdminUserID, erp.SystemAdminCreateRequest{
			DisplayName: result.Admin.DisplayName,
			Password:    *password,
			Role:        result.Admin.Role,
			Status:      result.Admin.Status,
		}, erp.AdminSessionPrincipal{AdminUserID: result.Admin.AdminUserID})
		if err != nil {
			return err
		}
		if admin, ok := resp.Data.(erp.SystemAdminSummary); ok {
			result.Admin = admin
		}
		passwordReset = true
	}

	fmt.Printf("system admin ready: account=%s admin_created=%t\n",
		result.Admin.ExternalUserID,
		result.AdminCreated,
	)
	if passwordReset {
		fmt.Printf("system admin password reset: account=%s\n", result.Admin.ExternalUserID)
	}
	return nil
}

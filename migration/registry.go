// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package migration

import (
	"database/sql"
	"fmt"
	"strings"
)

// detectDriver attempts to determine the database driver type
func detectDriver(db *sql.DB) string {
	// Check SQLite
	_, err := db.Exec("SELECT sqlite_version()")
	if err == nil {
		return "sqlite"
	}

	// Check PostgreSQL
	_, err = db.Exec("SELECT version()")
	if err == nil {
		var version string
		db.QueryRow("SELECT version()").Scan(&version)
		if strings.Contains(strings.ToLower(version), "postgresql") {
			return "postgres"
		}
	}

	// Unknown database driver
	return "unknown"
}

// GetAll returns all registered migrations (basic tables + users + sessions only)
func GetAll() []Migration {
	return []Migration{
		{
			Version:     "20250101000003",
			Description: "Create options table",
			Up:          createOptionsTable,
			Down:        dropOptionsTable,
		},
		{
			Version:     "20250101000004",
			Description: "Create users table",
			Up:          createUsersTable,
			Down:        dropUsersTable,
		},
		{
			Version:     "20250101000005",
			Description: "Create users_meta table",
			Up:          createUsersMetaTable,
			Down:        dropUsersMetaTable,
		},
		{
			Version:     "20250101000006",
			Description: "Create user_sessions table",
			Up:          createUserSessionsTable,
			Down:        dropUserSessionsTable,
		}, {
			Version:     "20250101000007",
			Description: "Create user_api_keys table",
			Up:          createUserAPIKeysTable,
			Down:        dropUserAPIKeysTable,
		},
		{
			Version:     "20250101000009",
			Description: "Create workspaces table",
			Up:          createWorkspacesTable,
			Down:        dropWorkspacesTable,
		},
		{
			Version:     "20250101000010",
			Description: "Create user_invites table",
			Up:          createUserInvitesTable,
			Down:        dropUserInvitesTable,
		},
		{
			Version:     "20250101000011",
			Description: "Create workspaces_meta table",
			Up:          createWorkspacesMetaTable,
			Down:        dropWorkspacesMetaTable,
		},
		{
			Version:     "20250101000012",
			Description: "Create workspace_users table",
			Up:          createWorkspaceUsersTable,
			Down:        dropWorkspaceUsersTable,
		},
		{
			Version:     "20250101000013",
			Description: "Create workspace_ssh_keys table",
			Up:          createWorkspaceSSHKeysTable,
			Down:        dropWorkspaceSSHKeysTable,
		},
		{
			Version:     "20250101000013a",
			Description: "Create workspace_credentials table",
			Up:          createWorkspaceCredentialsTable,
			Down:        dropWorkspaceCredentialsTable,
		},
		{
			Version:     "20250101000014",
			Description: "Create workspace_repositories table",
			Up:          createWorkspaceRepositoriesTable,
			Down:        dropWorkspaceRepositoriesTable,
		},
		{
			Version:     "20250101000015",
			Description: "Create workspace_integrations table",
			Up:          createWorkspaceIntegrationsTable,
			Down:        dropWorkspaceIntegrationsTable,
		},
		{
			Version:     "20250101000016",
			Description: "Create runners table",
			Up:          createRunnersTable,
			Down:        dropRunnersTable,
		},
		{
			Version:     "20250101000017",
			Description: "Create workspace_inventories table",
			Up:          createWorkspaceInventoriesTable,
			Down:        dropWorkspaceInventoriesTable,
		},
		{
			Version:     "20250101000018",
			Description: "Create workspace_environments table",
			Up:          createWorkspaceEnvironmentsTable,
			Down:        dropWorkspaceEnvironmentsTable,
		},
		{
			Version:     "20250101000020",
			Description: "Create system_integrations table",
			Up:          createSystemIntegrationsTable,
			Down:        dropSystemIntegrationsTable,
		},
		{
			Version:     "20250101000021",
			Description: "Create password_reset_tokens table",
			Up:          createPasswordResetTokensTable,
			Down:        dropPasswordResetTokensTable,
		},
		{
			Version:     "20250101000022",
			Description: "Create workspace_events table",
			Up:          createWorkspaceEventsTable,
			Down:        dropWorkspaceEventsTable,
		},
		{
			Version:     "20250101000023",
			Description: "Create system_events table",
			Up:          createSystemEventsTable,
			Down:        dropSystemEventsTable,
		},
		{
			Version:     "20250101000024",
			Description: "Create workspace_backend table",
			Up:          createWorkspaceBackendTable,
			Down:        dropWorkspaceBackendTable,
		},
		{
			Version:     "20250101000025",
			Description: "Create workspace_backend_state table",
			Up:          createWorkspaceBackendStateTable,
			Down:        dropWorkspaceBackendStateTable,
		},
		{
			Version:     "20250101000026",
			Description: "Create workspace_backend_lock table",
			Up:          createWorkspaceBackendLockTable,
			Down:        dropWorkspaceBackendLockTable,
		},
		{
			Version:     "20250101000028",
			Description: "Create user_ai_messages table",
			Up:          createUserAIMessagesTable,
			Down:        dropUserAIMessagesTable,
		},
		{
			Version:     "20250101000029",
			Description: "Create workspace_repository_clone_tasks table",
			Up:          createWorkspaceRepositoryCloneTasksTable,
			Down:        dropWorkspaceRepositoryCloneTasksTable,
		},
		{
			Version:     "20250101000030",
			Description: "Create workspace_events_deliverables table",
			Up:          createWorkspaceEventsDeliverablesTable,
			Down:        dropWorkspaceEventsDeliverablesTable,
		},
		{
			Version:     "20250101000031",
			Description: "Create system_events_deliverables table",
			Up:          createSystemEventsDeliverablesTable,
			Down:        dropSystemEventsDeliverablesTable,
		},
		{
			Version:     "20250101000032",
			Description: "Create workspace_variable_groups table",
			Up:          createWorkspaceVariableGroupsTable,
			Down:        dropWorkspaceVariableGroupsTable,
		},
	}
}

// createOptionsTable creates the options table
func createOptionsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE options (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key VARCHAR(128) NOT NULL UNIQUE,
			value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE options (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			key VARCHAR(128) NOT NULL UNIQUE,
			value TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_key ON options(key)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropOptionsTable drops the options table
func dropOptionsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS options")
	return err
}

// createUsersTable creates the users table
func createUsersTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		// role is admin, user
		query = `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(191) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'regular',
			is_active BOOLEAN DEFAULT 1,
			last_login_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			email VARCHAR(191) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'regular',
			is_active BOOLEAN DEFAULT true,
			last_login_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_email ON users(email)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUsersTable drops the users table
func dropUsersTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	return err
}

// createUsersMetaTable creates the users_meta table
func createUsersMetaTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE users_meta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key VARCHAR(128) NOT NULL,
			value TEXT,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(user_id, key)
		)`
	case "postgres":
		query = `
		CREATE TABLE users_meta (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			key VARCHAR(128) NOT NULL,
			value TEXT,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (user_id, key)
		);
		CREATE INDEX idx_user_id ON users_meta(user_id);
		CREATE INDEX idx_key ON users_meta(key)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUsersMetaTable drops the users_meta table
func dropUsersMetaTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS users_meta")
	return err
}

// createUserSessionsTable creates the user_sessions table
func createUserSessionsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE user_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token VARCHAR(255) NOT NULL UNIQUE,
			user_id INTEGER NOT NULL,
			ip_address VARCHAR(45),
			user_agent VARCHAR(255),
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE user_sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			token VARCHAR(255) NOT NULL UNIQUE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			ip_address VARCHAR(45),
			user_agent VARCHAR(255),
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_token ON user_sessions(token);
		CREATE INDEX idx_user_id ON user_sessions(user_id);
		CREATE INDEX idx_expires_at ON user_sessions(expires_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUserSessionsTable drops the user_sessions table
func dropUserSessionsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS user_sessions")
	return err
}

// createUserAPIKeysTable creates the user_api_keys table
func createUserAPIKeysTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE user_api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name VARCHAR(100),
			key VARCHAR(512) NOT NULL,
			expires_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE user_api_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(100),
			key VARCHAR(512) NOT NULL,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_user_api_keys_user_id ON user_api_keys(user_id);
		CREATE INDEX idx_user_api_keys_expires_at ON user_api_keys(expires_at);
		CREATE INDEX idx_user_api_keys_key ON user_api_keys(key)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUserAPIKeysTable drops the user_api_keys table
func dropUserAPIKeysTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS user_api_keys")
	return err
}

// createUserInvitesTable creates the user_invites table
func createUserInvitesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE user_invites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email VARCHAR(191) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'regular',
			token VARCHAR(255) NOT NULL UNIQUE,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			inviter_user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			accepted_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (inviter_user_id) REFERENCES users(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE user_invites (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(191) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'regular',
			token VARCHAR(255) NOT NULL UNIQUE,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			inviter_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at TIMESTAMP NOT NULL,
			accepted_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_user_invites_email ON user_invites(email);
		CREATE INDEX idx_user_invites_token ON user_invites(token);
		CREATE INDEX idx_user_invites_expires_at ON user_invites(expires_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUserInvitesTable drops the user_invites table.
func dropUserInvitesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS user_invites")
	return err
}

// createWorkspacesTable creates the workspaces table
func createWorkspacesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE workspaces (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspacesTable drops the workspaces table.
func dropWorkspacesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspaces")
	return err
}

// createWorkspacesMetaTable creates the workspaces_meta table (key-value per workspace)
func createWorkspacesMetaTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspaces_meta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			key VARCHAR(128) NOT NULL,
			value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			UNIQUE(workspace_id, key)
		)`
	case "postgres":
		query = `
		CREATE TABLE workspaces_meta (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			key VARCHAR(128) NOT NULL,
			value TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (workspace_id, key)
		);
		CREATE INDEX idx_workspaces_meta_workspace_id ON workspaces_meta(workspace_id);
		CREATE INDEX idx_workspaces_meta_key ON workspaces_meta(key)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspacesMetaTable drops the workspaces_meta table.
func dropWorkspacesMetaTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspaces_meta")
	return err
}

// createWorkspaceUsersTable creates the workspace_users table (user-workspace membership and role)
func createWorkspaceUsersTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'regular',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(workspace_id, user_id)
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL DEFAULT 'regular',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (workspace_id, user_id)
		);
		CREATE INDEX idx_workspace_users_workspace_id ON workspace_users(workspace_id);
		CREATE INDEX idx_workspace_users_user_id ON workspace_users(user_id)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceUsersTable drops the workspace_users table.
func dropWorkspaceUsersTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_users")
	return err
}

// createWorkspaceSSHKeysTable creates the workspace_ssh_keys table (SSH keys attached to workspaces)
func createWorkspaceSSHKeysTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_ssh_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(100) NOT NULL,
			committer_name VARCHAR(100) NOT NULL,
			committer_email VARCHAR(100) NOT NULL,
			private_key TEXT,
			public_key TEXT,
			passphrase VARCHAR(200) NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_ssh_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			committer_name VARCHAR(100) NOT NULL,
			committer_email VARCHAR(100) NOT NULL,
			private_key TEXT,
			public_key TEXT,
			passphrase VARCHAR(200) NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_ssh_keys_workspace_id ON workspace_ssh_keys(workspace_id)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceSSHKeysTable drops the workspace_ssh_keys table.
func dropWorkspaceSSHKeysTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_ssh_keys")
	return err
}

// createWorkspaceCredentialsTable creates the workspace_credentials table (git or user/password credentials per workspace).
func createWorkspaceCredentialsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'git',
			committer_name VARCHAR(100),
			committer_email VARCHAR(100),
			username VARCHAR(100),
			password_or_token TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_credentials (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'git',
			committer_name VARCHAR(100),
			committer_email VARCHAR(100),
			username VARCHAR(100),
			password_or_token TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_credentials_workspace_id ON workspace_credentials(workspace_id);
		CREATE INDEX idx_workspace_credentials_type ON workspace_credentials(type)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceCredentialsTable drops the workspace_credentials table.
func dropWorkspaceCredentialsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_credentials")
	return err
}

// createWorkspaceRepositoriesTable creates the workspace_repositories table
func createWorkspaceRepositoriesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(100) NOT NULL,
			git_url VARCHAR(500) NOT NULL,
			git_branch VARCHAR(100) NULL,
			git_tag VARCHAR(100) NULL,
			ssh_key_id INTEGER NULL,
			credentials_id INTEGER NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY (ssh_key_id) REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			FOREIGN KEY (credentials_id) REFERENCES workspace_credentials(id) ON DELETE SET NULL
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_repositories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			git_url VARCHAR(500) NOT NULL,
			git_branch VARCHAR(100) NULL,
			git_tag VARCHAR(100) NULL,
			ssh_key_id UUID NULL REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			credentials_id UUID NULL REFERENCES workspace_credentials(id) ON DELETE SET NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_repositories_workspace_id ON workspace_repositories(workspace_id);
		CREATE INDEX idx_workspace_repositories_ssh_key_id ON workspace_repositories(ssh_key_id);
		CREATE INDEX idx_workspace_repositories_credentials_id ON workspace_repositories(credentials_id)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceRepositoriesTable drops the workspace_repositories table.
func dropWorkspaceRepositoriesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_repositories")
	return err
}

// createWorkspaceRepositoryCloneTasksTable creates the workspace_repository_clone_tasks table (clone tasks per repository).
func createWorkspaceRepositoryCloneTasksTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_repository_clone_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repository_id INTEGER NOT NULL,
			runner_id INTEGER NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			runner_task_id VARCHAR(64) NULL,
			result TEXT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (repository_id) REFERENCES workspace_repositories(id) ON DELETE CASCADE,
			FOREIGN KEY (runner_id) REFERENCES runners(id) ON DELETE SET NULL
		);
		CREATE INDEX idx_workspace_repository_clone_tasks_repository_id ON workspace_repository_clone_tasks(repository_id);
		CREATE INDEX idx_workspace_repository_clone_tasks_runner_id ON workspace_repository_clone_tasks(runner_id);
		CREATE INDEX idx_workspace_repository_clone_tasks_status ON workspace_repository_clone_tasks(status)`
	case "postgres":
		query = `
		CREATE TABLE workspace_repository_clone_tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			repository_id UUID NOT NULL REFERENCES workspace_repositories(id) ON DELETE CASCADE,
			runner_id UUID NULL REFERENCES runners(id) ON DELETE SET NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			runner_task_id VARCHAR(64) NULL,
			result TEXT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_repository_clone_tasks_repository_id ON workspace_repository_clone_tasks(repository_id);
		CREATE INDEX idx_workspace_repository_clone_tasks_runner_id ON workspace_repository_clone_tasks(runner_id);
		CREATE INDEX idx_workspace_repository_clone_tasks_status ON workspace_repository_clone_tasks(status)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceRepositoryCloneTasksTable drops the workspace_repository_clone_tasks table.
func dropWorkspaceRepositoryCloneTasksTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_repository_clone_tasks")
	return err
}

// createWorkspaceIntegrationsTable creates the integrations table
func createWorkspaceIntegrationsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_integrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'webhook',
			name VARCHAR(100) NOT NULL,
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_integrations_workspace_id ON workspace_integrations(workspace_id);
		CREATE INDEX idx_integrations_type ON workspace_integrations(type)`
	case "postgres":
		query = `
		CREATE TABLE workspace_integrations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL DEFAULT 'webhook',
			name VARCHAR(100) NOT NULL,
			config TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_integrations_workspace_id ON workspace_integrations(workspace_id);
		CREATE INDEX idx_integrations_type ON workspace_integrations(type)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceIntegrationsTable drops the integrations table.
func dropWorkspaceIntegrationsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_integrations")
	return err
}

// createSystemIntegrationsTable creates the system_integrations table
func createSystemIntegrationsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE system_integrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type VARCHAR(50) NOT NULL DEFAULT 'webhook',
			name VARCHAR(100) NOT NULL,
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_system_integrations_type ON system_integrations(type)`
	case "postgres":
		query = `
		CREATE TABLE system_integrations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type VARCHAR(50) NOT NULL DEFAULT 'webhook',
			name VARCHAR(100) NOT NULL,
			config TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_system_integrations_type ON system_integrations(type)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropSystemIntegrationsTable drops the system_integrations table.
func dropSystemIntegrationsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS system_integrations")
	return err
}

// createPasswordResetTokensTable creates the password_reset_tokens table
func createPasswordResetTokensTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string
	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE password_reset_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
		CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at)`
	case "postgres":
		query = `
		CREATE TABLE password_reset_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
		CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}
	_, err := db.Exec(query)
	return err
}

// dropPasswordResetTokensTable drops the password_reset_tokens table.
func dropPasswordResetTokensTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS password_reset_tokens")
	return err
}

// createWorkspaceEventsTable creates the workspace_events table (workspace-scoped event log)
func createWorkspaceEventsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			type VARCHAR(50) NOT NULL,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_workspace_events_workspace_id ON workspace_events(workspace_id);
		CREATE INDEX idx_workspace_events_type ON workspace_events(type);
		CREATE INDEX idx_workspace_events_created_at ON workspace_events(created_at)`
	case "postgres":
		query = `
		CREATE TABLE workspace_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			data TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_events_workspace_id ON workspace_events(workspace_id);
		CREATE INDEX idx_workspace_events_type ON workspace_events(type);
		CREATE INDEX idx_workspace_events_created_at ON workspace_events(created_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceEventsTable drops the workspace_events table.
func dropWorkspaceEventsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_events")
	return err
}

// createSystemEventsTable creates the system_events table (runner_added, user_invited, etc.)
func createSystemEventsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE system_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type VARCHAR(50) NOT NULL,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_system_events_type ON system_events(type);
		CREATE INDEX idx_system_events_created_at ON system_events(created_at)`
	case "postgres":
		query = `
		CREATE TABLE system_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type VARCHAR(50) NOT NULL,
			data TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_system_events_type ON system_events(type);
		CREATE INDEX idx_system_events_created_at ON system_events(created_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropSystemEventsTable drops the system_events table.
func dropSystemEventsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS system_events")
	return err
}

// createRunnersTable creates the runners table
func createRunnersTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE runners (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(100),
			tags VARCHAR(128),
			max_parallel_tasks INTEGER DEFAULT 10,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			url VARCHAR(255),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE runners (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			token VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(100),
			tags VARCHAR(128),
			max_parallel_tasks INTEGER DEFAULT 10,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			url VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_runners_token ON runners(token);
		CREATE INDEX idx_runners_status ON runners(status)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropRunnersTable drops the runners table.
func dropRunnersTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS runners")
	return err
}

// createWorkspaceInventoriesTable creates the workspace_inventories table (Ansible/Terraform/Tofu inventory per workspace)
func createWorkspaceInventoriesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_inventories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(100) NOT NULL,
			inventory TEXT,
			inventory_path VARCHAR(512),
			type VARCHAR(50) NOT NULL DEFAULT 'static',
			ssh_key_id INTEGER NULL,
			become_key_id INTEGER NULL,
			repository_id INTEGER NULL,
			runner_tag VARCHAR(128),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY (ssh_key_id) REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			FOREIGN KEY (become_key_id) REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			FOREIGN KEY (repository_id) REFERENCES workspace_repositories(id) ON DELETE SET NULL
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_inventories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			inventory TEXT,
			inventory_path VARCHAR(512),
			type VARCHAR(50) NOT NULL DEFAULT 'static',
			ssh_key_id UUID NULL REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			become_key_id UUID NULL REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			repository_id UUID NULL REFERENCES workspace_repositories(id) ON DELETE SET NULL,
			runner_tag VARCHAR(128),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_inventories_workspace_id ON workspace_inventories(workspace_id);
		CREATE INDEX idx_workspace_inventories_type ON workspace_inventories(type)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceInventoriesTable drops the workspace_inventories table.
func dropWorkspaceInventoriesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_inventories")
	return err
}

// createWorkspaceVariableGroupsTable creates the workspace_variable_groups table
func createWorkspaceVariableGroupsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_variable_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(100) NOT NULL,
			extra_variables TEXT,
			environment_variables TEXT,
			secrets TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_variable_groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			extra_variables TEXT,
			environment_variables TEXT,
			secrets TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_variable_groups_workspace_id ON workspace_variable_groups(workspace_id);
		CREATE INDEX idx_workspace_variable_groups_name ON workspace_variable_groups(name)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceVariableGroupsTable drops the workspace_variable_groups table.
func dropWorkspaceVariableGroupsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_variable_groups")
	return err
}

// createWorkspaceEnvironmentsTable creates the workspace_environments table (extra vars / env for Ansible per workspace)
func createWorkspaceEnvironmentsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_environments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(100) NOT NULL,
			password TEXT,
			json TEXT,
			env TEXT,
			secret_storage_key_prefix VARCHAR(128),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_environments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			password TEXT,
			json TEXT,
			env TEXT,
			secret_storage_key_prefix VARCHAR(128),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_environments_workspace_id ON workspace_environments(workspace_id);
		CREATE INDEX idx_workspace_environments_name ON workspace_environments(name)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceEnvironmentsTable drops the workspace_environments table.
func dropWorkspaceEnvironmentsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_environments")
	return err
}

// createWorkspaceBackendTable creates the backend table
func createWorkspaceBackendTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_backend (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			username VARCHAR(255),
			password TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			UNIQUE(workspace_id, name),
			UNIQUE(workspace_id, slug)
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_backend (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			username VARCHAR(255),
			password TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(workspace_id, name),
			UNIQUE(workspace_id, slug)
		);
		CREATE INDEX idx_workspace_backend_workspace_id ON workspace_backend(workspace_id)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceBackendTable drops the backend table.
func dropWorkspaceBackendTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_backend")
	return err
}

// createWorkspaceBackendStateTable creates the state table (TF state JSON per backend; name unique per backend).
func createWorkspaceBackendStateTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_backend_state (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			backend_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			value TEXT,
			checksum VARCHAR(64),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (backend_id) REFERENCES workspace_backend(id) ON DELETE CASCADE,
			UNIQUE(backend_id, name)
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_backend_state (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			backend_id UUID NOT NULL REFERENCES workspace_backend(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			value TEXT,
			checksum VARCHAR(64),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(backend_id, name)
		);
		CREATE INDEX idx_workspace_backend_state_backend_id ON workspace_backend_state(backend_id)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceBackendStateTable drops the state table.
func dropWorkspaceBackendStateTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_backend_state")
	return err
}

// createWorkspaceBackendLockTable creates the lock table (TF state lock per backend; is_active for current lock).
func createWorkspaceBackendLockTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_backend_lock (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			backend_id INTEGER NOT NULL,
			operation VARCHAR(64),
			info TEXT,
			who VARCHAR(255),
			version VARCHAR(64),
			path VARCHAR(512),
			lock_id VARCHAR(128),
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (backend_id) REFERENCES workspace_backend(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_backend_lock (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			backend_id UUID NOT NULL REFERENCES workspace_backend(id) ON DELETE CASCADE,
			operation VARCHAR(64),
			info TEXT,
			who VARCHAR(255),
			version VARCHAR(64),
			path VARCHAR(512),
			lock_id VARCHAR(128),
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_backend_lock_backend_id ON workspace_backend_lock(backend_id);
		CREATE INDEX idx_workspace_backend_lock_is_active ON workspace_backend_lock(is_active)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceBackendLockTable drops the lock table.
func dropWorkspaceBackendLockTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_backend_lock")
	return err
}

// createAIMessagesTable creates the user_ai_messages table (user chat history with the AI assistant).
func createUserAIMessagesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE user_ai_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			role VARCHAR(32) NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE user_ai_messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(32) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_user_ai_messages_user_id ON user_ai_messages(user_id);
		CREATE INDEX idx_user_ai_messages_created_at ON user_ai_messages(created_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUserAIMessagesTable drops the user_ai_messages table.
func dropUserAIMessagesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS user_ai_messages")
	return err
}

// createWorkspaceEventsDeliverablesTable creates the workspace_events_deliverables table (events to deliver to workspace integrations).
func createWorkspaceEventsDeliverablesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_events_deliverables (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id INTEGER NOT NULL,
			integration_id INTEGER NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			delivered_at DATETIME NULL,
			error_message TEXT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (event_id) REFERENCES workspace_events(id) ON DELETE CASCADE,
			FOREIGN KEY (integration_id) REFERENCES workspace_integrations(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_workspace_events_deliverables_event_id ON workspace_events_deliverables(event_id);
		CREATE INDEX idx_workspace_events_deliverables_integration_id ON workspace_events_deliverables(integration_id);
		CREATE INDEX idx_workspace_events_deliverables_status ON workspace_events_deliverables(status);
		CREATE INDEX idx_workspace_events_deliverables_created_at ON workspace_events_deliverables(created_at)`
	case "postgres":
		query = `
		CREATE TABLE workspace_events_deliverables (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event_id UUID NOT NULL REFERENCES workspace_events(id) ON DELETE CASCADE,
			integration_id UUID NOT NULL REFERENCES workspace_integrations(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			delivered_at TIMESTAMP NULL,
			error_message TEXT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_events_deliverables_event_id ON workspace_events_deliverables(event_id);
		CREATE INDEX idx_workspace_events_deliverables_integration_id ON workspace_events_deliverables(integration_id);
		CREATE INDEX idx_workspace_events_deliverables_status ON workspace_events_deliverables(status);
		CREATE INDEX idx_workspace_events_deliverables_created_at ON workspace_events_deliverables(created_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropWorkspaceEventsDeliverablesTable drops the workspace_events_deliverables table.
func dropWorkspaceEventsDeliverablesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_events_deliverables")
	return err
}

// createSystemEventsDeliverablesTable creates the system_events_deliverables table (events to deliver to system integrations).
func createSystemEventsDeliverablesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE system_events_deliverables (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id INTEGER NOT NULL,
			integration_id INTEGER NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			delivered_at DATETIME NULL,
			error_message TEXT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (event_id) REFERENCES system_events(id) ON DELETE CASCADE,
			FOREIGN KEY (integration_id) REFERENCES system_integrations(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_system_events_deliverables_event_id ON system_events_deliverables(event_id);
		CREATE INDEX idx_system_events_deliverables_integration_id ON system_events_deliverables(integration_id);
		CREATE INDEX idx_system_events_deliverables_status ON system_events_deliverables(status);
		CREATE INDEX idx_system_events_deliverables_created_at ON system_events_deliverables(created_at)`
	case "postgres":
		query = `
		CREATE TABLE system_events_deliverables (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event_id UUID NOT NULL REFERENCES system_events(id) ON DELETE CASCADE,
			integration_id UUID NOT NULL REFERENCES system_integrations(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			delivered_at TIMESTAMP NULL,
			error_message TEXT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_system_events_deliverables_event_id ON system_events_deliverables(event_id);
		CREATE INDEX idx_system_events_deliverables_integration_id ON system_events_deliverables(integration_id);
		CREATE INDEX idx_system_events_deliverables_status ON system_events_deliverables(status);
		CREATE INDEX idx_system_events_deliverables_created_at ON system_events_deliverables(created_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropSystemEventsDeliverablesTable drops the system_events_deliverables table.
func dropSystemEventsDeliverablesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS system_events_deliverables")
	return err
}

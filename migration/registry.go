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
			Version:     "20250101000014",
			Description: "Create workspace_repositories table",
			Up:          createWorkspaceRepositoriesTable,
			Down:        dropWorkspaceRepositoriesTable,
		},
		{
			Version:     "20250101000015",
			Description: "Create integrations table",
			Up:          createIntegrationsTable,
			Down:        dropIntegrationsTable,
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
			Version:     "20250101000019",
			Description: "Create workspace_tf_backends table",
			Up:          createWorkspaceTFBackendsTable,
			Down:        dropWorkspaceTFBackendsTable,
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
			key VARCHAR(255) NOT NULL UNIQUE,
			value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE options (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			key VARCHAR(255) NOT NULL UNIQUE,
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
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			is_active BOOLEAN DEFAULT 1,
			last_login_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
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
			key VARCHAR(255) NOT NULL,
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
			key VARCHAR(255) NOT NULL,
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
			user_agent VARCHAR(500),
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
			user_agent VARCHAR(500),
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
			name VARCHAR(255),
			key_hash VARCHAR(255) NOT NULL,
			expires_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE user_api_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255),
			key_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_user_api_keys_user_id ON user_api_keys(user_id);
		CREATE INDEX idx_user_api_keys_expires_at ON user_api_keys(expires_at);
		CREATE INDEX idx_user_api_keys_key_hash ON user_api_keys(key_hash)`
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

// createUserInvitesTable creates the user_invites table (invite users by email to a workspace)
func createUserInvitesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE user_invites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'guest',
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
			email VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'guest',
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
			name VARCHAR(255) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE workspaces (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

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
			key VARCHAR(255) NOT NULL,
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
			key VARCHAR(255) NOT NULL,
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
			role VARCHAR(50) NOT NULL DEFAULT 'guest',
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
			role VARCHAR(50) NOT NULL DEFAULT 'guest',
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
			name VARCHAR(255) NOT NULL,
			private_key TEXT,
			public_key TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_ssh_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			private_key TEXT,
			public_key TEXT,
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

func dropWorkspaceSSHKeysTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_ssh_keys")
	return err
}

// createWorkspaceRepositoriesTable creates the workspace_repositories table (repos attached to workspaces; optional ssh_key_id)
func createWorkspaceRepositoriesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			git_url VARCHAR(500) NOT NULL,
			git_branch VARCHAR(255) DEFAULT 'main',
			ssh_key_id INTEGER NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY (ssh_key_id) REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_repositories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			git_url VARCHAR(500) NOT NULL,
			git_branch VARCHAR(255) DEFAULT 'main',
			ssh_key_id UUID NULL REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_repositories_workspace_id ON workspace_repositories(workspace_id);
		CREATE INDEX idx_workspace_repositories_ssh_key_id ON workspace_repositories(ssh_key_id)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

func dropWorkspaceRepositoriesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_repositories")
	return err
}

// createIntegrationsTable creates the integrations table (webhooks/CI triggers per workspace)
func createIntegrationsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE integrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE integrations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

func dropIntegrationsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS integrations")
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
			name VARCHAR(255),
			tag VARCHAR(255),
			max_parallel_tasks INTEGER DEFAULT 10,
			active BOOLEAN DEFAULT 1,
			public_key TEXT,
			url VARCHAR(255),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE runners (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			token VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255),
			tag VARCHAR(255),
			max_parallel_tasks INTEGER DEFAULT 10,
			active BOOLEAN DEFAULT true,
			public_key TEXT,
			url VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_runners_token ON runners(token);
		CREATE INDEX idx_runners_active ON runners(active)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

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
			name VARCHAR(255) NOT NULL,
			inventory TEXT,
			type VARCHAR(50) NOT NULL DEFAULT 'static',
			ssh_key_id INTEGER NULL,
			become_key_id INTEGER NULL,
			repository_id INTEGER NULL,
			runner_tag VARCHAR(255),
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
			name VARCHAR(255) NOT NULL,
			inventory TEXT,
			type VARCHAR(50) NOT NULL DEFAULT 'static',
			ssh_key_id UUID NULL REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			become_key_id UUID NULL REFERENCES workspace_ssh_keys(id) ON DELETE SET NULL,
			repository_id UUID NULL REFERENCES workspace_repositories(id) ON DELETE SET NULL,
			runner_tag VARCHAR(255),
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

func dropWorkspaceInventoriesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_inventories")
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
			name VARCHAR(255) NOT NULL,
			password TEXT,
			json TEXT,
			env TEXT,
			secret_storage_key_prefix VARCHAR(255),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_environments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			password TEXT,
			json TEXT,
			env TEXT,
			secret_storage_key_prefix VARCHAR(255),
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

func dropWorkspaceEnvironmentsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_environments")
	return err
}

// createWorkspaceTFBackendsTable creates the workspace_tf_backends table (Terraform/OpenTofu backend config per workspace)
func createWorkspaceTFBackendsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE workspace_tf_backends (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'local',
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE workspace_tf_backends (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'local',
			config TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_workspace_tf_backends_workspace_id ON workspace_tf_backends(workspace_id);
		CREATE INDEX idx_workspace_tf_backends_type ON workspace_tf_backends(type)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

func dropWorkspaceTFBackendsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS workspace_tf_backends")
	return err
}

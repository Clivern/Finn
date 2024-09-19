## Met Database Schema

### options

| Column     | Type         | Constraints                    |
|------------|--------------|--------------------------------|
| id         | PK (auto)     |                                |
| key        | VARCHAR(255) | NOT NULL UNIQUE                |
| value      | TEXT          |                                |
| created_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |
| updated_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |

**Indexes:** `key`

---

### users

| Column        | Type         | Constraints                    |
|---------------|--------------|--------------------------------|
| id            | PK (auto)     |                                |
| name          | VARCHAR(255) | NOT NULL                       |
| email         | VARCHAR(255) | NOT NULL UNIQUE                |
| password_hash | VARCHAR(255) | NOT NULL                       |
| role          | VARCHAR(50)  | NOT NULL DEFAULT 'user'        |
| is_active     | boolean      | DEFAULT true                   |
| last_login_at | datetime     | NULL                           |
| created_at    | datetime     | DEFAULT CURRENT_TIMESTAMP      |
| updated_at    | datetime     | DEFAULT CURRENT_TIMESTAMP      |

**Indexes:** `email`

---

## users_meta

Key-value metadata per user.

| Column     | Type         | Constraints                    |
|------------|--------------|--------------------------------|
| id         | PK (auto)     |                                |
| key        | VARCHAR(255) | NOT NULL                       |
| value      | TEXT          |                                |
| user_id    | FK: users(id)| NOT NULL, ON DELETE CASCADE    |
| created_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |
| updated_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |

**Unique:** (user_id, key)
**Indexes:** `user_id` (user_id), `key` (key)

---

## user_sessions

| Column     | Type         | Constraints                    |
|------------|--------------|--------------------------------|
| id         | PK (auto)     |                                |
| token      | VARCHAR(255) | NOT NULL UNIQUE                |
| user_id    | FK: users(id)| NOT NULL, ON DELETE CASCADE    |
| ip_address | VARCHAR(45)  |                                |
| user_agent | VARCHAR(500) |                                |
| expires_at | datetime     | NOT NULL                       |
| created_at | datetime     | DEFAULT CURRENT_TIMESTAMP      |
| updated_at | datetime     | DEFAULT CURRENT_TIMESTAMP      |

**Indexes:** `idx_token` (token), `idx_user_id` (user_id), `idx_expires_at` (expires_at)

---

## user_api_keys

| Column     | Type         | Constraints                    |
|------------|--------------|--------------------------------|
| id         | PK (auto)     |                                |
| user_id    | FK: users(id)| NOT NULL, ON DELETE CASCADE    |
| name       | VARCHAR(255) |                                |
| key_hash   | VARCHAR(255) | NOT NULL                       |
| expires_at | datetime     | NULL                           |
| created_at | datetime     | DEFAULT CURRENT_TIMESTAMP      |

**Indexes:** `idx_user_api_keys_user_id`, `idx_user_api_keys_expires_at`, `idx_user_api_keys_key_hash`

---

## projects

| Column     | Type         | Constraints                    |
|------------|--------------|--------------------------------|
| id         | PK (auto)     |                                |
| name       | VARCHAR(255) | NOT NULL                       |
| created_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |
| updated_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |

---

## user_invites

| Column          | Type         | Constraints                    |
|-----------------|--------------|--------------------------------|
| id              | PK (auto)     |                                |
| email           | VARCHAR(255) | NOT NULL                       |
| role            | VARCHAR(50)  | NOT NULL DEFAULT 'guest'       |
| token           | VARCHAR(255) | NOT NULL UNIQUE                |
| status          | VARCHAR(50)  | NOT NULL DEFAULT 'pending'    |
| inviter_user_id | FK: users(id)| NOT NULL, ON DELETE CASCADE    |
| expires_at      | datetime     | NOT NULL                       |
| accepted_at     | datetime     | NULL                           |
| created_at      | datetime     | DEFAULT CURRENT_TIMESTAMP      |
| updated_at      | datetime     | DEFAULT CURRENT_TIMESTAMP      |

**Indexes:** `idx_user_invites_email`, `idx_user_invites_token`, `idx_user_invites_expires_at`

---

## projects_meta

| Column     | Type            | Constraints                     |
|------------|-----------------|---------------------------------|
| id         | PK (auto)        |                                 |
| project_id | FK: projects(id)| NOT NULL, ON DELETE CASCADE     |
| key        | VARCHAR(255)    | NOT NULL                        |
| value      | TEXT            |                                 |
| created_at | datetime        | DEFAULT CURRENT_TIMESTAMP       |
| updated_at | datetime        | DEFAULT CURRENT_TIMESTAMP       |

**Unique:** (project_id, key)
**Indexes:** `idx_projects_meta_project_id`, `idx_projects_meta_key`

---

## project_users

| Column     | Type            | Constraints                     |
|------------|-----------------|---------------------------------|
| id         | PK (auto)        |                                 |
| project_id | FK: projects(id)| NOT NULL, ON DELETE CASCADE     |
| user_id    | FK: users(id)   | NOT NULL, ON DELETE CASCADE     |
| role       | VARCHAR(50)     | NOT NULL DEFAULT 'guest'       |
| created_at | datetime        | DEFAULT CURRENT_TIMESTAMP       |
| updated_at | datetime        | DEFAULT CURRENT_TIMESTAMP       |

**Unique:** (project_id, user_id)
**Indexes:** `idx_project_users_project_id`, `idx_project_users_user_id`

---

## project_ssh_keys

| Column      | Type            | Constraints                     |
|-------------|-----------------|---------------------------------|
| id          | PK (auto)        |                                 |
| project_id  | FK: projects(id)| NOT NULL, ON DELETE CASCADE     |
| name        | VARCHAR(255)    | NOT NULL                        |
| private_key | TEXT            |                                 |
| public_key  | TEXT            |                                 |
| created_at  | datetime        | DEFAULT CURRENT_TIMESTAMP       |
| updated_at  | datetime        | DEFAULT CURRENT_TIMESTAMP       |

**Indexes:** `idx_project_ssh_keys_project_id`

---

## project_repositories

| Column      | Type                    | Constraints                     |
|-------------|-------------------------|---------------------------------|
| id          | PK (auto)               |                                 |
| project_id  | FK: projects(id)        | NOT NULL, ON DELETE CASCADE     |
| name        | VARCHAR(255)           | NOT NULL                        |
| git_url     | VARCHAR(500)            | NOT NULL                        |
| git_branch  | VARCHAR(255)            | DEFAULT 'main'                  |
| ssh_key_id  | FK: project_ssh_keys(id)| NULL, ON DELETE SET NULL        |
| created_at  | datetime                | DEFAULT CURRENT_TIMESTAMP       |
| updated_at  | datetime                | DEFAULT CURRENT_TIMESTAMP       |

**Indexes:** `idx_project_repositories_project_id`, `idx_project_repositories_ssh_key_id`

---

## integrations

| Column     | Type         | Constraints                    |
|------------|--------------|--------------------------------|
| id         | PK (auto)     |                                |
| name       | VARCHAR(255) | NOT NULL                       |
| created_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |
| updated_at | datetime      | DEFAULT CURRENT_TIMESTAMP      |

---

## runners

| Column           | Type         | Constraints                    |
|------------------|--------------|--------------------------------|
| id               | PK (auto)     |                                |
| token            | VARCHAR(255) | NOT NULL UNIQUE                |
| name             | VARCHAR(255) |                                |
| tag              | VARCHAR(255) |                                |
| max_parallel_tasks | INTEGER    | DEFAULT 10                     |
| active           | boolean      | DEFAULT true                   |
| public_key       | TEXT         |                                |
| url              | VARCHAR(255) |                                |
| created_at       | datetime     | DEFAULT CURRENT_TIMESTAMP      |
| updated_at       | datetime     | DEFAULT CURRENT_TIMESTAMP      |

**Note:** SQLite migration has `FOREIGN KEY (project_id)` but no `project_id` column; add column here if runners are per-project.

**Indexes:** `idx_runners_token`, `idx_runners_active`

---

## project_inventories

| Column        | Type                    | Constraints                     |
|---------------|-------------------------|---------------------------------|
| id            | PK (auto)               |                                 |
| project_id    | FK: projects(id)        | NOT NULL, ON DELETE CASCADE     |
| name          | VARCHAR(255)            | NOT NULL                        |
| inventory     | TEXT                    |                                 |
| type          | VARCHAR(50)              | NOT NULL DEFAULT 'static'      |
| ssh_key_id    | FK: project_ssh_keys(id)| NULL, ON DELETE SET NULL        |
| become_key_id | FK: project_ssh_keys(id)| NULL, ON DELETE SET NULL        |
| repository_id | FK: project_repositories(id)| NULL, ON DELETE SET NULL    |
| runner_tag    | VARCHAR(255)            |                                 |
| created_at    | datetime                | DEFAULT CURRENT_TIMESTAMP       |
| updated_at    | datetime                | DEFAULT CURRENT_TIMESTAMP       |

**Indexes:** `idx_project_inventories_project_id`, `idx_project_inventories_type`

---

## project_environments

| Column                   | Type            | Constraints                     |
|--------------------------|-----------------|---------------------------------|
| id                       | PK (auto)       |                                 |
| project_id               | FK: projects(id)| NOT NULL, ON DELETE CASCADE     |
| name                     | VARCHAR(255)    | NOT NULL                        |
| password                 | TEXT            |                                 |
| json                     | TEXT            |                                 |
| env                      | TEXT            |                                 |
| secret_storage_key_prefix| VARCHAR(255)    |                                 |
| created_at               | datetime        | DEFAULT CURRENT_TIMESTAMP       |
| updated_at               | datetime        | DEFAULT CURRENT_TIMESTAMP       |

**Indexes:** `idx_project_environments_project_id`, `idx_project_environments_name`

---

## project_tf_backends

| Column     | Type            | Constraints                     |
|------------|-----------------|---------------------------------|
| id         | PK (auto)        |                                 |
| project_id | FK: projects(id)| NOT NULL, ON DELETE CASCADE     |
| name       | VARCHAR(255)    | NOT NULL                        |
| type       | VARCHAR(50)     | NOT NULL DEFAULT 'local'       |
| config     | TEXT            |                                 |
| created_at | datetime        | DEFAULT CURRENT_TIMESTAMP      |
| updated_at | datetime        | DEFAULT CURRENT_TIMESTAMP      |

**Indexes:** `idx_project_tf_backends_project_id`, `idx_project_tf_backends_type`

---

## Migration order (dependency order)

1. options
2. users
3. users_meta
4. user_sessions
5. user_api_keys
6. projects
7. user_invites
8. projects_meta
9. project_users
10. project_ssh_keys
11. project_repositories
12. integrations
13. runners
14. project_inventories
15. project_environments
16. project_tf_backends

// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"fmt"

	"github.com/clivern/finn/db"
	"github.com/clivern/finn/pkg/util"
)

// Template bundles repositories used for templates.
type Template struct {
	Templates   db.WorkspaceTemplateRepository
	Repos       db.WorkspaceRepoRepository
	SSHKeys     db.WorkspaceSSHKeyRepository
	Creds       db.WorkspaceCredRepository
	GVars       db.WorkspaceVariableGroupRepository
	Inventories db.WorkspaceInventoryRepository
}

// PayloadByID builds a nested template payload by template ID.
func (r Template) PayloadByID(templateID db.ID) (map[string]interface{}, error) {
	t, err := r.Templates.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("template json payload: get template: %w", err)
	}
	if t == nil {
		return nil, fmt.Errorf("template json payload: template not found")
	}

	repo, err := r.Repos.GetByIDAndWorkspace(t.RepositoryID, t.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("template json payload: get repository: %w", err)
	}
	if repo == nil {
		return nil, fmt.Errorf("template json payload: repository not found")
	}

	var repoKey *db.WorkspaceSSHKey
	if repo.SSHKeyID != nil {
		k, err := r.SSHKeys.GetByIDAndWorkspace(*repo.SSHKeyID, t.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("template json payload: get repo ssh key: %w", err)
		}
		repoKey = k
	}

	var repoCred *db.WorkspaceCred
	if repo.GitCredentialsID != nil {
		c, err := r.Creds.GetByIDAndWorkspace(*repo.GitCredentialsID, t.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("template json payload: get repo git credential: %w", err)
		}
		repoCred = c
	}

	var gv *db.WorkspaceVariableGroup
	if t.GroupVarsID != nil {
		g, err := r.GVars.GetByIDAndWorkspace(*t.GroupVarsID, t.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("template json payload: get variable group: %w", err)
		}
		gv = g
	}

	ansibleMap := decodeJSONStringToMap(t.AnsibleConfig)
	var inv *db.WorkspaceInventory
	var invKey *db.WorkspaceSSHKey
	var invCred *db.WorkspaceCred
	var becomeKey *db.WorkspaceSSHKey
	var becomeCred *db.WorkspaceCred
	var invRepo *db.WorkspaceRepo
	var invRepoKey *db.WorkspaceSSHKey
	var invRepoCred *db.WorkspaceCred

	if invID := inventoryIDFromAnsibleMap(ansibleMap); invID != "" {
		i, err := r.Inventories.GetByIDAndWorkspace(db.ID(invID), t.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("template json payload: get inventory: %w", err)
		}
		inv = i

		if inv != nil {
			if inv.SSHKeyID != nil {
				k, err := r.SSHKeys.GetByIDAndWorkspace(*inv.SSHKeyID, t.WorkspaceID)
				if err != nil {
					return nil, fmt.Errorf("template json payload: get inventory ssh key: %w", err)
				}
				invKey = k
			}
			if inv.CredentialID != nil {
				c, err := r.Creds.GetByIDAndWorkspace(*inv.CredentialID, t.WorkspaceID)
				if err != nil {
					return nil, fmt.Errorf("template json payload: get inventory credential: %w", err)
				}
				invCred = c
			}
			if inv.BecomeKeyID != nil {
				k, err := r.SSHKeys.GetByIDAndWorkspace(*inv.BecomeKeyID, t.WorkspaceID)
				if err != nil {
					return nil, fmt.Errorf("template json payload: get inventory become ssh key: %w", err)
				}
				becomeKey = k
			}
			if inv.BecomeCredentialID != nil {
				c, err := r.Creds.GetByIDAndWorkspace(*inv.BecomeCredentialID, t.WorkspaceID)
				if err != nil {
					return nil, fmt.Errorf("template json payload: get inventory become credential: %w", err)
				}
				becomeCred = c
			}
			if inv.RepositoryID != nil {
				ir, err := r.Repos.GetByIDAndWorkspace(*inv.RepositoryID, t.WorkspaceID)
				if err != nil {
					return nil, fmt.Errorf("template json payload: get inventory repository: %w", err)
				}
				invRepo = ir

				if invRepo != nil {
					if invRepo.SSHKeyID != nil {
						k, err := r.SSHKeys.GetByIDAndWorkspace(*invRepo.SSHKeyID, t.WorkspaceID)
						if err != nil {
							return nil, fmt.Errorf("template json payload: get inventory repo ssh key: %w", err)
						}
						invRepoKey = k
					}
					if invRepo.GitCredentialsID != nil {
						c, err := r.Creds.GetByIDAndWorkspace(*invRepo.GitCredentialsID, t.WorkspaceID)
						if err != nil {
							return nil, fmt.Errorf("template json payload: get inventory repo git credential: %w", err)
						}
						invRepoCred = c
					}
				}
			}
		}
	}

	payload := r.templateToPayloadFull(
		t,
		repo,
		decryptSSHKeyPayload(repoKey),
		decryptCredPayload(repoCred),
		gv,
		inv,
		decryptSSHKeyPayload(invKey),
		decryptCredPayload(invCred),
		decryptSSHKeyPayload(becomeKey),
		decryptCredPayload(becomeCred),
		invRepo,
		decryptSSHKeyPayload(invRepoKey),
		decryptCredPayload(invRepoCred),
	)

	// ansible_config.inventory_id is an implementation detail for DB joins; hide it from the display payload.
	if ac, ok := payload["ansible_config"].(map[string]interface{}); ok {
		delete(ac, "inventory_id")
		delete(ac, "inventoryId")
		delete(ac, "inventoryID")
	}

	return payload, nil
}

func (Template) templateToPayloadFull(
	t *db.WorkspaceTemplate,
	repo *db.WorkspaceRepo,
	repoSSHKey map[string]interface{},
	repoCred map[string]interface{},
	groupVars *db.WorkspaceVariableGroup,
	ansibleInventory *db.WorkspaceInventory,
	invSSHKey map[string]interface{},
	invCred map[string]interface{},
	becomeSSHKey map[string]interface{},
	becomeCred map[string]interface{},
	invRepo *db.WorkspaceRepo,
	invRepoSSHKey map[string]interface{},
	invRepoCred map[string]interface{},
) map[string]interface{} {
	if t == nil {
		return map[string]interface{}{}
	}

	payload := map[string]interface{}{
		"id":                    t.ID.String(),
		"name":                  t.Name,
		"description":           t.Description,
		"engine":                t.Engine,
		"working_dir":           t.WorkingDirectory,
		"git_branch_override":   t.GitBranchOverride,
		"git_tag_override":      t.GitTagOverride,
		"allow_concurrent_runs": t.AllowConcurrentRuns,
		"max_concurrent_runs":   t.MaxConcurrentRuns,
		"approval_required":     t.ApprovalRequired,
		"extra_args":            t.ExtraArgs,
	}

	if m := decodeJSONStringToMap(t.RunnerTags); m != nil {
		payload["runner_tags"] = m
	}
	if m := decodeJSONStringToMap(t.ScriptConfig); m != nil {
		payload["script_config"] = m
	}
	if m := decodeJSONStringToMap(t.IaCConfig); m != nil {
		payload["iac_config"] = m
	}

	var ansibleCfg map[string]interface{}
	if m := decodeJSONStringToMap(t.AnsibleConfig); m != nil {
		ansibleCfg = m
		payload["ansible_config"] = ansibleCfg
	}

	if repo != nil {
		r := map[string]interface{}{
			"id":         repo.ID.String(),
			"name":       repo.Name,
			"git_url":    repo.GitURL,
			"git_branch": repo.GitBranch,
			"git_tag":    repo.GitTag,
		}
		if len(repoSSHKey) > 0 {
			r["ssh_key"] = repoSSHKey
		}
		if len(repoCred) > 0 {
			r["git_credential"] = repoCred
		}
		payload["repository"] = r
	}

	if groupVars != nil {
		payload["group_vars"] = map[string]interface{}{
			"id":                           groupVars.ID.String(),
			"name":                         groupVars.Name,
			"extra_variables":              groupVars.ExtraVariables,
			"environment_variables":        groupVars.EnvironmentVariables,
			"secret_extra_variables":       groupVars.SecretExtraVariables,
			"secret_environment_variables": groupVars.SecretEnvironmentVariables,
		}
	}

	if ansibleCfg != nil && ansibleInventory != nil {
		inv := map[string]interface{}{
			"id":             ansibleInventory.ID.String(),
			"name":           ansibleInventory.Name,
			"type":           ansibleInventory.Type,
			"inventory":      ansibleInventory.Inventory,
			"inventory_path": ansibleInventory.InventoryPath,
		}
		if len(invSSHKey) > 0 {
			inv["ssh_key"] = invSSHKey
		}
		if len(invCred) > 0 {
			inv["git_credential"] = invCred
		}
		if len(becomeSSHKey) > 0 {
			inv["become_ssh_key"] = becomeSSHKey
		}
		if len(becomeCred) > 0 {
			inv["become_git_credential"] = becomeCred
		}
		if invRepo != nil {
			invR := map[string]interface{}{
				"id":         invRepo.ID.String(),
				"name":       invRepo.Name,
				"git_url":    invRepo.GitURL,
				"git_branch": invRepo.GitBranch,
				"git_tag":    invRepo.GitTag,
			}
			if len(invRepoSSHKey) > 0 {
				invR["ssh_key"] = invRepoSSHKey
			}
			if len(invRepoCred) > 0 {
				invR["git_credential"] = invRepoCred
			}
			inv["repository"] = invR
		}
		ansibleCfg["inventory"] = inv
	}

	return payload
}

func decodeJSONStringToMap(s string) map[string]interface{} {
	if s == "" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func inventoryIDFromAnsibleMap(m map[string]interface{}) string {
	if m == nil {
		return ""
	}
	if s := stringish(m["inventory_id"]); s != "" {
		return s
	}
	if s := stringish(m["inventoryId"]); s != "" {
		return s
	}
	if s := stringish(m["inventoryID"]); s != "" {
		return s
	}
	return ""
}

func stringish(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

func decryptSSHKeyPayload(k *db.WorkspaceSSHKey) map[string]interface{} {
	if k == nil {
		return nil
	}
	out := map[string]interface{}{
		"id":              k.ID.String(),
		"name":            k.Name,
		"committer_name":  k.CommitterName,
		"committer_email": k.CommitterEmail,
		"private_key": util.DecryptFromBase64String(
			k.PrivateKey,
			util.GetEncryptionKey(),
		),
		"passphrase": util.DecryptFromBase64String(
			k.Passphrase,
			util.GetEncryptionKey(),
		),
	}
	if k.PublicKey != nil {
		out["public_key"] = *k.PublicKey
	}
	return out
}

func decryptCredPayload(c *db.WorkspaceCred) map[string]interface{} {
	if c == nil {
		return nil
	}
	out := map[string]interface{}{
		"id":              c.ID.String(),
		"name":            c.Name,
		"committer_name":  c.CommitterName,
		"committer_email": c.CommitterEmail,
		"username":        c.Username,
		"password_or_token": util.DecryptFromBase64String(
			c.PasswordOrToken,
			util.GetEncryptionKey(),
		),
	}
	if c.Type != "" {
		out["type"] = c.Type
	}
	return out
}

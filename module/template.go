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

	payload := db.TemplateToPayloadFull(
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

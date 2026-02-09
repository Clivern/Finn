// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package opa

/*
ctx := context.Background()

modules := map[string]string{
	"terraform.rego": `
package terraform.policy

default allow := true

deny contains "Diff opens ingress to the world (0.0.0.0/0)" if {
  contains(input.diff_text, "0.0.0.0/0")
}

allow if count(deny) == 0
`,
}

client := opa.NewClient(modules)

input := map[string]interface{}{
	"diff_text": `
diff --git a/main.tf b/main.tf
index abc123..def456 100644
--- a/main.tf
+++ b/main.tf
@@ -10,7 +10,7 @@ resource "aws_security_group_rule" "ssh" {
-  cidr_blocks = ["10.0.0.0/8"]
+  cidr_blocks = ["0.0.0.0/0"]
}
`,
}

allowed, violations, err := client.EvaluateChange(ctx, DataPathTerraform, input)
if err != nil {
	panic(err)
}

fmt.Println("allowed:", allowed)
fmt.Println("violations:", violations)
*/

package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"hal/internal/global"
)

func TestPQCVersionPrerequisite(t *testing.T) {
	for _, v := range []string{"2.2.0+ent", "2.2.0+ent.hsm", "3.0.0+ent"} {
		if e := pqcCheckVersion(v); e != nil {
			t.Errorf("%s: %v", v, e)
		}
	}
	for _, v := range []string{"2.1.9+ent", "2.2.0", "2.2.0-rc1+ent", "garbage"} {
		if e := pqcCheckVersion(v); e == nil {
			t.Errorf("accepted unsupported %s", v)
		}
	}
}
func pqcFlagCommand() *cobra.Command {
	c := &cobra.Command{}
	vaultPKICmd.Flags().VisitAll(func(f *pflag.Flag) { copy := *f; copy.Changed = false; c.Flags().AddFlag(&copy) })
	return c
}

func TestPQCFlagsBeforeMutation(t *testing.T) {
	savedK8s, savedDomain, savedImage := pkiK8s, pqcDomain, pqcImage
	t.Cleanup(func() {
		pkiK8s, pqcDomain, pqcImage = savedK8s, savedDomain, savedImage
	})
	pqcDomain, pqcImage = pqcDefaultDomain, pqcDefaultImage
	pkiK8s = false
	for _, action := range []string{"enable", "update", "disable", "status"} {
		if e := validatePQCFlags(pqcFlagCommand(), action); e != nil {
			t.Fatalf("%s without --k8s: %v", action, e)
		}
	}
	if e := validatePQCFlags(pqcFlagCommand(), "status"); e != nil {
		t.Fatal(e)
	}
	pkiK8s = true
	for _, f := range []string{"hsm", "no-hsm", "acme", "force", "root-mount", "int-mount"} {
		c := pqcFlagCommand()
		c.Flags().Lookup(f).Changed = true
		if e := validatePQCFlags(c, "update"); e == nil {
			t.Errorf("accepted --%s", f)
		}
	}
	pqcImage = ""
	if e := validatePQCFlags(pqcFlagCommand(), "enable"); e == nil {
		t.Fatal("empty NGINX image accepted")
	}

}
func TestPQCDryRunNoToolsOrState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	t.Setenv("VAULT_ADDR", "http://invalid.invalid")
	oldDry, oldK8s := global.DryRun, pkiK8s
	global.DryRun, pkiK8s = true, false
	t.Cleanup(func() { global.DryRun, pkiK8s = oldDry, oldK8s })
	if e := runPQC(pqcFlagCommand(), []string{"enable"}); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(home, ".hal")); !os.IsNotExist(e) {
		t.Fatalf("dry-run created local state: %v", e)
	}
}
func TestPQCPrerequisitesReadOnly(t *testing.T) {
	routes := map[string]fakeVaultResponse{
		healthKey:                    healthBody("2.2.0+ent.hsm"),
		"GET /v1/sys/license/status": ok(`{"data":{"autoloaded":{"expiration_time":"2999-01-01T00:00:00Z"}}}`),
		mountsKey:                    ok(`{"data":{"existing/":{"type":"pki"}}}`),
		"GET /v1/existing/root/generate/internal": ok(`{"help":"parameter_set: ml-dsa-44, ml-dsa-65, ml-dsa-87"}`),
	}
	fv, c := newFakeVault(t, routes)
	if e := pqcPrerequisites(c); e != nil {
		t.Fatal(e)
	}
	for _, call := range fv.calls {
		if !strings.HasPrefix(call.Key, "GET ") {
			t.Errorf("prerequisite mutated Vault: %s", call.Key)
		}
	}
	routes[mountsKey] = ok(`{"data":{}}`)
	if e := pqcPrerequisites(c); e != nil {
		t.Fatalf("fresh Vault rejected: %v", e)
	}
	routes["GET /v1/sys/license/status"] = ok(`{"data":{"autoloaded":{"expiration_time":"2020-01-01T00:00:00Z"}}}`)
	if e := pqcPrerequisites(c); e == nil {
		t.Fatal("expired license accepted")
	}
}
func TestPQCMountOwnership(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, m := range []*vaultapi.MountOutput{nil, {Type: "pki", Description: "user mount"}, {Type: "kv", Description: pqcMountDescription}} {
		if pqcOwnedMount(m) {
			t.Fatal("adopted unowned mount")
		}
	}
	if !pqcOwnedMount(&vaultapi.MountOutput{Type: "pki", Description: pqcMountDescription}) {
		t.Fatal("did not recognize owned mount")
	}
	t.Setenv("PATH", t.TempDir())
	fv, c := newFakeVault(t, map[string]fakeVaultResponse{mountsKey: ok(`{"data":{"hal-pqc-root/":{"type":"pki","description":"user-owned"}}}`)})
	if e := pqcDisable(c); e == nil {
		t.Fatal("disable accepted unowned mount")
	}
	for _, call := range fv.calls {
		if !strings.HasPrefix(call.Key, "GET ") {
			t.Errorf("disable mutated unowned resources: %s", call.Key)
		}
	}
}

func TestPQCDisableOwnedOrder(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	fv, c := newFakeVault(t, map[string]fakeVaultResponse{mountsKey: ok(`{"data":{"hal-pqc-root/":{"type":"pki","description":"HAL post-quantum PKI lab (hal-vault-pqc-pki)"},"hal-pqc-int/":{"type":"pki","description":"HAL post-quantum PKI lab (hal-vault-pqc-pki)"},"pki-root/":{"type":"pki","description":"conventional"}}}`)})
	if e := pqcDisable(c); e != nil {
		t.Fatal(e)
	}
	var deleted []string
	for _, call := range fv.calls {
		if strings.HasPrefix(call.Key, "DELETE ") {
			deleted = append(deleted, call.Key)
		}
	}
	want := []string{"DELETE /v1/sys/mounts/hal-pqc-int", "DELETE /v1/sys/mounts/hal-pqc-root"}
	if strings.Join(deleted, ",") != strings.Join(want, ",") {
		t.Fatalf("teardown order/scope = %v", deleted)
	}
}

func TestPQCRejectsIncompatibleImageBeforeMutation(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	t.Setenv("VAULT_TOKEN", "test")
	oldDry, oldK8s := global.DryRun, pkiK8s
	global.DryRun, pkiK8s = false, true
	t.Cleanup(func() { global.DryRun, pkiK8s = oldDry, oldK8s })
	for name, body := range map[string]string{
		"docker": "case \"$1\" in\ninfo|image) exit 0;;\nrun) echo 'nginx version: nginx/1.28.0'; echo 'built with OpenSSL 3.4.0'; exit 0;;\nesac\nexit 1\n",
		"kind":   "exit 1\n", "kubectl": "exit 0\n", "openssl": "exit 0\n",
	} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	fv, c := newFakeVault(t, map[string]fakeVaultResponse{
		healthKey:                    healthBody("2.2.0+ent.hsm"),
		"GET /v1/sys/license/status": ok(`{"data":{"autoloaded":{"expiration_time":"2999-01-01T00:00:00Z"}}}`),
		mountsKey:                    ok(`{"data":{"existing/":{"type":"pki"}}}`),
		"GET /v1/existing/root/generate/internal": ok(`{"help":"parameter_set ml-dsa-65"}`),
	})
	t.Setenv("VAULT_ADDR", c.Address())
	if err := runPQC(pqcFlagCommand(), []string{"enable"}); err == nil || !strings.Contains(err.Error(), "needs NGINX") {
		t.Fatalf("unsupported image error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hal")); !os.IsNotExist(err) {
		t.Fatalf("image rejection created local state: %v", err)
	}
	for _, call := range fv.calls {
		if !strings.HasPrefix(call.Key, "GET ") {
			t.Errorf("image rejection mutated Vault: %s", call.Key)
		}
	}
}

// Exercise deployment rendering with the embedded assets and a nondefault DNS zone.
func TestPQCEmbeddedDeployment(t *testing.T) {
	dir, bin := t.TempDir(), t.TempDir()
	capture := filepath.Join(dir, "resources.json")
	t.Setenv("HAL_PQC_CAPTURE", capture)
	t.Setenv("PATH", bin)
	script := "#!/bin/sh\n/bin/cat >> \"$HAL_PQC_CAPTURE\"\nprintf '\\n' >> \"$HAL_PQC_CAPTURE\"\n"
	if err := os.WriteFile(filepath.Join(bin, "kubectl"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"frontend-chain.pem", "frontend.key", "backend-chain.pem", "backend.key", "browser.pem", "browser.key", "root.pem"} {
		if err := os.WriteFile(filepath.Join(dir, file), []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	oldDomain := pqcDomain
	pqcDomain = "test.internal"
	t.Cleanup(func() { pqcDomain = oldDomain })
	if err := pqcDeploy(dir, "test-image@sha256:fixture"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	configs, deployments := 0, 0
	for {
		var resource struct {
			Kind     string
			Metadata struct{ Name, Namespace string }
			Data     map[string]string
			Spec     struct {
				Template struct {
					Spec struct {
						Containers []struct {
							Image        string
							Args         []string
							VolumeMounts []struct{ Name, MountPath string }
						}
						Volumes []struct {
							Name      string
							ConfigMap struct{ Name string }
						}
					}
				}
			}
		}
		if err := decoder.Decode(&resource); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		switch resource.Kind {
		case "ConfigMap":
			if !strings.HasSuffix(resource.Metadata.Name, "-config") {
				continue
			}
			configs++
			name := strings.TrimSuffix(resource.Metadata.Name, "-config")
			expected, err := pqcAssets.ReadFile("pki-pqc/" + name + ".conf")
			if err != nil {
				t.Fatal(err)
			}
			if resource.Data["nginx.conf"] != strings.ReplaceAll(string(expected), "__CLUSTER_DOMAIN__", pqcDomain) {
				t.Fatal("deployed config differs from embedded asset")
			}
			for key, value := range resource.Data {
				if key == "nginx.conf" {
					continue
				}
				expected, err := pqcAssets.ReadFile("pki-pqc/" + key)
				if err != nil || value != string(expected) {
					t.Fatalf("invalid embedded %s: %v", key, err)
				}
			}
			if resource.Data["evidence.js"] == "" || (name == pqcFrontend && resource.Data["index.html"] == "") {
				t.Fatal("missing runtime asset")
			}
		case "Deployment":
			deployments++
			pod := resource.Spec.Template.Spec
			if len(pod.Containers) != 1 || pod.Containers[0].Image != "test-image@sha256:fixture" || strings.Join(pod.Containers[0].Args, " ") != "-c /app/nginx.conf -g daemon off;" {
				t.Fatal("incorrect NGINX invocation")
			}
			found := false
			for _, volume := range pod.Volumes {
				if volume.Name == "app" && volume.ConfigMap.Name == resource.Metadata.Name+"-config" {
					found = true
				}
			}
			mounted := false
			for _, mount := range pod.Containers[0].VolumeMounts {
				if mount.Name == "app" && mount.MountPath == "/app" {
					mounted = true
				}
			}
			if !found || !mounted {
				t.Fatal("embedded assets are not mounted at /app")
			}
		}
	}
	if configs != 2 || deployments != 2 {
		t.Fatalf("got %d configs, %d deployments", configs, deployments)
	}
}

func TestPQCIssuanceSchemaRollback(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint("existing=", existing), func(t *testing.T) {
			mounts := `{"data":{}}`
			if existing {
				mounts = `{"data":{"hal-pqc-root/":{"type":"pki","description":"HAL post-quantum PKI lab (hal-vault-pqc-pki)"}}}`
			}
			fv, c := newFakeVault(t, map[string]fakeVaultResponse{
				mountsKey: ok(mounts),
				"GET /v1/hal-pqc-root/root/generate/internal": ok(`{"help":"parameter_set ml-dsa-65"}`),
				"GET /v1/hal-pqc-int/root/generate/internal":  ok(`{"help":"rsa ec"}`),
			})
			if err := pqcIssue(c, t.TempDir()); err == nil || !strings.Contains(err.Error(), "does not advertise") {
				t.Fatalf("schema error = %v", err)
			}
			var deleted []string
			for _, call := range fv.calls {
				if strings.HasPrefix(call.Key, "DELETE ") {
					deleted = append(deleted, call.Key)
				}
				if strings.Contains(call.Key, "generate") && !strings.HasPrefix(call.Key, "GET ") {
					t.Fatal("generated keys before checking both schemas")
				}
			}
			want := "DELETE /v1/sys/mounts/hal-pqc-int"
			if !existing {
				want += ",DELETE /v1/sys/mounts/hal-pqc-root"
			}
			if strings.Join(deleted, ",") != want {
				t.Fatalf("cleanup = %v", deleted)
			}
		})
	}
}

func TestPQCIssuanceFreshMounts(t *testing.T) {
	routes := map[string]fakeVaultResponse{mountsKey: ok(`{"data":{}}`)}
	for _, mount := range []string{pqcRootMount, pqcIntMount} {
		routes["GET /v1/"+mount+"/root/generate/internal"] = ok(`{"help":"parameter_set ml-dsa-65"}`)
	}
	routes["PUT /v1/hal-pqc-root/root/generate/internal"] = ok(`{"data":{"certificate":"root"}}`)
	routes["PUT /v1/hal-pqc-int/intermediate/generate/internal"] = ok(`{"data":{"csr":"csr"}}`)
	routes["PUT /v1/hal-pqc-root/root/sign-intermediate"] = ok(`{"data":{"certificate":"intermediate"}}`)
	for _, name := range []string{pqcFrontend, pqcBackend} {
		routes["PUT /v1/hal-pqc-int/issue/"+name] = ok(`{"data":{"certificate":"leaf","private_key":"key"}}`)
	}
	fv, c := newFakeVault(t, routes)
	if err := pqcIssue(c, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	for _, call := range fv.calls {
		if strings.HasPrefix(call.Key, "DELETE ") {
			t.Fatal("removed successfully provisioned mount")
		}
	}
}

func TestPQCHostPortMapping(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		mapped     bool
	}{
		{"mapped", "echo '0.0.0.0:8093'", true},
		{"older-cluster", "exit 1", false},
		{"different-port", "echo '127.0.0.1:18093'", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := filepath.Join(t.TempDir(), "engine")
			if err := os.WriteFile(engine, []byte("#!/bin/sh\n"+tc.body+"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			if got := pqcHostPortMapped(engine); got != tc.mapped {
				t.Fatalf("mapped=%t, want %t", got, tc.mapped)
			}
		})
	}
}

package k8s

import (
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFindKubeconfigUsesCurrentDirectoryWhenAvailable(t *testing.T) {
	oldKubeconfig := os.Getenv("KUBECONFIG")
	if err := os.Unsetenv("KUBECONFIG"); err != nil {
		t.Fatalf("failed to unset KUBECONFIG: %v", err)
	}
	defer func() {
		if oldKubeconfig == "" {
			_ = os.Unsetenv("KUBECONFIG")
			return
		}
		_ = os.Setenv("KUBECONFIG", oldKubeconfig)
	}()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "kubeconfig")
	if err := os.WriteFile(path, []byte("apiVersion: v1\nkind: Config\n"), 0o600); err != nil {
		t.Fatalf("failed to create kubeconfig: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to switch working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	if got := FindKubeconfig(log.New()); got != path {
		t.Fatalf("expected kubeconfig from current directory %q, got %q", path, got)
	}
}

func TestCheckUnstructuredForReadyStateIgnoresCompletedContainerReadySignals(t *testing.T) {
	resource := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      "completed-pod",
			"namespace": "default",
		},
		"status": map[string]interface{}{
			"phase": "Running",
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "False",
				},
			},
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":  "done",
					"ready": false,
					"state": map[string]interface{}{
						"terminated": map[string]interface{}{
							"reason": "Completed",
						},
					},
				},
			},
		},
	}}

	ready, err := CheckUnstructuredForReadyState(log.New(), resource)
	if err != nil {
		t.Fatalf("expected completed pod to be considered ready, got error: %v", err)
	}
	if !ready {
		t.Fatal("expected completed pod to be ready")
	}
}

func TestCheckUnstructuredForReadyStateRejectsMalformedConditionValues(t *testing.T) {
	resource := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name": "broken-pod",
		},
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   42,
					"status": "True",
				},
			},
		},
	}}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CheckUnstructuredForReadyState should not panic on malformed condition values: %v", r)
		}
	}()

	if ready, err := CheckUnstructuredForReadyState(log.New(), resource); err == nil || ready {
		t.Fatalf("expected malformed resource to fail validation, got ready=%v err=%v", ready, err)
	}
}

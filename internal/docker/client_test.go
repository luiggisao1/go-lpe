package docker_test

import (
	"context"
	"testing"
	"time"

	"github.com/luiggisao/auto-golpe/internal/docker"
)

func TestDockerClientPing(t *testing.T) {
	cli, err := docker.NewClient()
	if err != nil {
		t.Fatalf("failed to create docker client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cli.Ping(ctx); err != nil {
		t.Fatalf("docker ping failed: %v", err)
	}
}

func TestSandboxLifecycle(t *testing.T) {
	cli, err := docker.NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	containerID, err := cli.CreateSandbox(ctx, "alpine:latest")
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}
	t.Logf("Sandbox created with ID: %s", containerID)

	res, err := cli.ExecInSandbox(ctx, containerID, []string{"echo", "Hello from inside the sandbox!"})
	if err != nil {
		t.Fatalf("failed to exec in sandbox: %v", err)
	}
	t.Logf("Command exit code: %d", res.ExitCode)
	t.Logf("Command stdout: %s", res.Stdout)
	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}

	defer func() {
		if err := cli.RemoveSandbox(context.Background(), containerID); err != nil {
			t.Errorf("failed to remove sandbox: %v", err)
		} else {
			t.Logf("Sandbox removed successfully")
		}
	}()
	if containerID == "" {
		t.Fatal("expected non-empty container ID")
	}
}

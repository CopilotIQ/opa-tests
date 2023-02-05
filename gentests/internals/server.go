package internals

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// This package contains all the code necessary to run a self-contained OPA
// server that loads the policies bundle, and can then be used to execute
// tests against.
//
// It uses TestContainers to run OPA in a docker container.

const (
	// OpaImage is the Docker image to be used for tests.
	// TODO: make the OPA version a configuration value
	OpaImage     = "openpolicyagent/opa:0.48.0"
	OpaPort      = "8181"
	OpaBundleDir = "/etc/opa/bundles"
)

// OpaServer represents the running server.
type OpaServer struct {
	Address        string
	BundleFilepath string
	Container      testcontainers.Container
}

func (s *OpaServer) GetEndpoint(endpoint string) (*http.Response, error) {
	return http.Get(fmt.Sprintf("http://%s%s", s.Address, endpoint))
}

func (s *OpaServer) IsHealthy() bool {
	res, err := s.GetEndpoint("/health")
	return err == nil && res.StatusCode == http.StatusOK
}

func (s *OpaServer) WaitHealthy(ctx context.Context, pollInterval time.Duration) error {
	d, ok := ctx.Deadline()
	for {
		if s.IsHealthy() {
			return nil
		}
		// First order: if we got canceled, we bail
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			// if ok was false, no deadline was set, we carry on forever
			if ok && d.Before(time.Now()) {
				return context.DeadlineExceeded
			}
			time.Sleep(pollInterval)
		}
	}
}

// ContainerHealthyStrategy fixes an issue with TestContainers' HealthStrategy which unreferences a null
// Health pointer, when trying to check on the health of a container.
// See: https://github.com/testcontainers/testcontainers-go/issues/801
type ContainerHealthyStrategy struct {
	Strategy *wait.HealthStrategy
}

func (ws *ContainerHealthyStrategy) WaitUntilReady(ctx context.Context, target wait.StrategyTarget) (err error) {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			state, err := target.State(ctx)
			if err != nil {
				return err
			}
			if state.Health != nil && state.Health.Status != "healthy" {
				time.Sleep(ws.Strategy.PollInterval)
				continue
			}
			return nil
		}
	}
}

func NewOpaContainer(ctx context.Context, bundlePath string) (*OpaServer, error) {
	// Note that Docker will only mount the full path of the directory that contains the bundle
	bundleDir, err := filepath.Abs(filepath.Dir(bundlePath))
	if err != nil {
		return nil, err
	}
	bundle := filepath.Base(bundlePath)
	req := testcontainers.ContainerRequest{
		Image:        OpaImage,
		ExposedPorts: []string{OpaPort},
		Binds: []string{
			strings.Join([]string{bundleDir, OpaBundleDir}, ":"),
		},

		Cmd: []string{"run", "--server",
			"--addr", fmt.Sprintf(":%s", OpaPort),
			filepath.Join(OpaBundleDir, bundle)},

		WaitingFor: &ContainerHealthyStrategy{
			Strategy: wait.NewHealthStrategy().WithPollInterval(1 * time.Second)},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, OpaPort)
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return &OpaServer{Container: container, Address: address, BundleFilepath: bundlePath}, nil
}

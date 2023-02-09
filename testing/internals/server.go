package internals

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"path/filepath"
	"strings"
)

// This package contains all the code necessary to run a self-contained OPA
// server that loads the policies bundle, and can then be used to execute
// tests against.
//
// It uses TestContainers to run OPA in a docker container.

const (
	// OpaImage is the Docker image to be used for tests.
	// TODO: make the OPA version a configuration value
	OpaImage     = "openpolicyagent/opa:0.47.4"
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
		Cmd: []string{"run", "--server", "--addr", fmt.Sprintf(":%s", OpaPort),
			filepath.Join(OpaBundleDir, bundle)},
		WaitingFor: wait.ForHTTP("/health").WithPort(OpaPort),
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

package main

import (
	"context"
	"fmt"
)

type Alpine struct {
	Version  string
	Packages []string
}

func (m *Alpine) WithVersion(ctx context.Context, version string) (*Alpine, error) {
	m.Version = version
	return m, nil
}

func (m *Alpine) WithPackage(ctx context.Context, name string) (*Alpine, error) {
	m.Packages = append(m.Packages, name)
	return m, nil
}

func (m *Alpine) Container(ctx context.Context) (*Container, error) {
	version := "latest"
	if m.Version != "" {
		version = m.Version
	}

	ctr := dag.
		Container().
		From(fmt.Sprintf("alpine:%s", version))

	for _, pkg := range m.Packages {
		ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", pkg})
	}
	return ctr, nil
}

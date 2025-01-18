// Scan for secrets in a directory using Trufflehog
package main

import (
	"context"
	"dagger/trufflehog/internal/dagger"
)

// Scan for secrets in a directory using Trufflehog
type Trufflehog struct {
}

func (m *Trufflehog) base() *dagger.Container {
	return dag.Container().
		From("trufflesecurity/trufflehog")
}

// Find credentials in git repositories. Returns the JSON output of the scan
func (m *Trufflehog) Git(
	ctx context.Context,
	// Git repository URL. https:// or ssh:// schema expected
	uri string,
) (string, error) {
	return m.base().WithExec([]string{"trufflehog", "--json", "git", uri}).Stdout(ctx)
}

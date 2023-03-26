//go:build e2evscode

package cmd

import (
	"github.com/cwxstat/go-pod-launch-run/pkg"
	"testing"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{
			name: "Simple test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg.Run("dev2", "default",
				"aws-cli",
				"default",
				true, nil,
				"resultTest.pod")
		})
	}
}

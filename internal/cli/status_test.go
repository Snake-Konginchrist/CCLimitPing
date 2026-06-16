package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/wavever/CCLimitPing/internal/provider"
	"github.com/wavever/CCLimitPing/internal/usage"
)

func TestRunStatusPrintsProgressBeforeReadUsage(t *testing.T) {
	var out bytes.Buffer
	var progress bytes.Buffer

	p := fakeStatusProvider{
		name:  "codex",
		usage: &usage.Usage{Provider: "codex"},
		onRead: func() {
			if !strings.Contains(progress.String(), "Fetching codex usage...\n") {
				t.Fatalf("progress output before ReadUsage = %q, want fetching message", progress.String())
			}
		},
	}

	if err := runStatus(context.Background(), &out, &progress, enText, []provider.Provider{p}, false); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}
	if !strings.Contains(out.String(), "codex\n") {
		t.Fatalf("status output = %q, want provider usage", out.String())
	}
}

type fakeStatusProvider struct {
	name   string
	usage  *usage.Usage
	err    error
	onRead func()
}

func (f fakeStatusProvider) Name() string {
	return f.name
}

func (f fakeStatusProvider) ReadUsage(context.Context) (*usage.Usage, error) {
	if f.onRead != nil {
		f.onRead()
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.usage, nil
}

func (f fakeStatusProvider) Trigger(context.Context, bool) (*provider.TriggerResult, error) {
	return nil, nil
}

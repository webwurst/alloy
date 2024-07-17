package discovery

import (
	"fmt"

	"github.com/go-kit/log"
	"go.opentelemetry.io/otel/trace"
)

type DiscoverFn func() ([]string, error)

type Options struct {
	JoinPeers     []string
	DiscoverPeers string
	DefaultPort   int
	// Logger to surface extra information to the user. Required.
	Logger log.Logger
	// Tracer to emit spans. Required.
	Tracer trace.TracerProvider
}

func NewPeerDiscoveryFn(opts Options) (DiscoverFn, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("logger is required, got nil")
	}
	if opts.Tracer == nil {
		return nil, fmt.Errorf("tracer is required, got nil")
	}
	if len(opts.JoinPeers) > 0 && opts.DiscoverPeers != "" {
		return nil, fmt.Errorf("at most one of join peers and discover peers may be set, "+
			"got join peers %q and discover peers %q", opts.JoinPeers, opts.DiscoverPeers)
	}
	switch {
	case len(opts.JoinPeers) > 0:
		return newStaticDiscovery(opts.JoinPeers, opts.DefaultPort, opts.Logger), nil
	case opts.DiscoverPeers != "":
		return newDynamicDiscovery(opts.Logger, opts.DiscoverPeers, opts.DefaultPort)
	default:
		// Here, both JoinPeers and DiscoverPeers are empty. This is desirable when
		// starting a seed node that other nodes connect to, so we don't require
		// one of the fields to be set.
		return nil, nil
	}
}

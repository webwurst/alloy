package discovery

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/go-kit/log"
)

func TestPeerDiscovery(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stdout)
	tracer := noop.NewTracerProvider()
	tests := []struct {
		name                     string
		args                     Options
		expected                 []string
		expectedErrContain       string
		expectedCreateErrContain string
	}{
		{
			name: "no logger",
			args: Options{
				JoinPeers:   []string{"host:1234"},
				DefaultPort: 8888,
				Tracer:      tracer,
			},
			expectedCreateErrContain: "logger is required, got nil",
		},
		{
			name: "no tracer",
			args: Options{
				JoinPeers:   []string{"host:1234"},
				DefaultPort: 8888,
				Logger:      logger,
			},
			expectedCreateErrContain: "tracer is required, got nil",
		},
		//TODO(thampiotr): there is an inconsistency here: when given host:port, we resolve to it without looking
		// up the IP addresses. But when give a host only, we look up the IP addresses with the DNS.
		{
			name: "both join and discover peers given",
			args: Options{
				JoinPeers:     []string{"host:1234"},
				DiscoverPeers: "some.service:something",
				Logger:        logger,
				Tracer:        tracer,
			},
			expectedCreateErrContain: "at most one of join peers and discover peers may be set",
		},
		{
			name: "static host:port",
			args: Options{
				JoinPeers:   []string{"host:1234"},
				DefaultPort: 8888,
				Logger:      logger,
				Tracer:      tracer,
			},
			expected: []string{"host:1234"},
		},
		//TODO(thampiotr): this returns only one right now, but I think it should return multiple
		{
			name: "multiple static host:ports given",
			args: Options{
				JoinPeers:   []string{"host1:1234", "host2:1234"},
				DefaultPort: 8888,
				Logger:      logger,
				Tracer:      tracer,
			},
			expected: []string{"host1:1234"},
		},
		{
			name: "static ip address with port",
			args: Options{
				JoinPeers:   []string{"10.10.10.10:8888"},
				DefaultPort: 12345,
				Logger:      logger,
				Tracer:      tracer,
			},
			expected: []string{"10.10.10.10:8888"},
		},
		{
			name: "static ip address with default port",
			args: Options{
				JoinPeers:   []string{"10.10.10.10"},
				DefaultPort: 12345,
				Logger:      logger,
				Tracer:      tracer,
			},
			expected: []string{"10.10.10.10:12345"},
		},
		{
			name: "invalid ip address",
			args: Options{
				JoinPeers:   []string{"10.301.10.10"},
				DefaultPort: 12345,
				Logger:      logger,
				Tracer:      tracer,
			},
			//TODO(thampiotr): the error message is not very informative in this case
			expectedErrContain: "lookup 10.301.10.10: no such host",
		},
		//TODO(thampiotr): should we support multiple?
		{
			name: "multiple ip addresses",
			args: Options{
				JoinPeers:   []string{"10.10.10.10", "11.11.11.11"},
				DefaultPort: 12345,
				Logger:      logger,
				Tracer:      tracer,
			},
			expected: []string{"10.10.10.10:12345"},
		},
		{
			name: "multiple ip addresses with some invalid",
			args: Options{
				JoinPeers:   []string{"10.10.10.10", "11.311.11.11", "22.22.22.22"},
				DefaultPort: 12345,
				Logger:      logger,
				Tracer:      tracer,
			},
			expected: []string{"10.10.10.10:12345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, err := NewPeerDiscoveryFn(tt.args)
			if tt.expectedCreateErrContain != "" {
				require.ErrorContains(t, err, tt.expectedCreateErrContain)
				return
			} else {
				require.NoError(t, err)
			}

			actual, err := fn()
			if tt.expectedErrContain != "" {
				require.ErrorContains(t, err, tt.expectedErrContain)
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected, actual)
		})
	}
}

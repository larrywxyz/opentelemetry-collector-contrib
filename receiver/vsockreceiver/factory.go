package vsockreceiver

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr = "vsockreceiver"
)

// NewFactory creates a factory for vsockreceiver.
func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesReceiver(createTracesReceiver))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
		Addr:             ":5001",
	}
}

// createTracesReceiver creates a trace receiver based on provided config.
func createTracesReceiver(
	_ context.Context,
	set component.ReceiverCreateSettings,
	cfg config.Receiver,
	nextConsumer consumer.Traces,
) (component.TracesReceiver, error) {
	r := receivers.GetOrAdd(cfg, func() component.Component {
		return newVsockReceiver(cfg.(*Config), set)
	})

	if err := r.Unwrap().(*vsockReceiver).registerTraceConsumer(nextConsumer); err != nil {
		return nil, err
	}
	return r, nil
}

var receivers = sharedcomponent.NewSharedComponents()

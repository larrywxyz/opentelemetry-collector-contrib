package vsockreceiver

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/grpc"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

type vsockReceiver struct {
	cfg        *Config
	serverGRPC *grpc.Server

	traceReceiver *trace.Receiver
	shutdownWG    sync.WaitGroup

	settings component.ReceiverCreateSettings
}

// newVsockReceiver just creates the OpenTelemetry receiver services. It is the caller's
// responsibility to invoke the respective Start*Reception methods as well
// as the various Stop*Reception methods to end it.
func newVsockReceiver(cfg *Config, settings component.ReceiverCreateSettings) *vsockReceiver {
	r := &vsockReceiver{
		cfg:      cfg,
		settings: settings,
	}

	return r
}

func (r *vsockReceiver) startGRPCServer(cfg *configgrpc.GRPCServerSettings, host component.Host) error {
	r.settings.Logger.Info("Starting GRPC server on endpoint " + cfg.NetAddr.Endpoint)

	gln, err := cfg.ToListener()
	if err != nil {
		return err
	}
	r.shutdownWG.Add(1)
	go func() {
		defer r.shutdownWG.Done()

		if errGrpc := r.serverGRPC.Serve(gln); errGrpc != nil && !errors.Is(errGrpc, grpc.ErrServerStopped) {
			host.ReportFatalError(errGrpc)
		}
	}()
	return nil
}

func (r *vsockReceiver) startProtocolServers(host component.Host) error {
	var err error
	var opts []grpc.ServerOption
	opts, err = r.cfg.GRPC.ToServerOption(host, r.settings.TelemetrySettings)
	if err != nil {
		return err
	}
	r.serverGRPC = grpc.NewServer(opts...)

	if r.traceReceiver != nil {
		ptraceotlp.RegisterServer(r.serverGRPC, r.traceReceiver)
	}

	err = r.startGRPCServer(r.cfg.GRPC, host)
	if err != nil {
		return err
	}

	return err
}

// Start runs the trace receiver on the gRPC server. Currently
// it also enables the metrics receiver too.
func (r *vsockReceiver) Start(_ context.Context, host component.Host) error {
	return r.startProtocolServers(host)
}

// Shutdown is a method to turn off receiving.
func (r *vsockReceiver) Shutdown(ctx context.Context) error {
	var err error

	if r.serverGRPC != nil {
		r.serverGRPC.GracefulStop()
	}

	r.shutdownWG.Wait()
	return err
}

func (r *vsockReceiver) registerTraceConsumer(tc consumer.Traces) error {
	if tc == nil {
		return component.ErrNilNextConsumer
	}
	r.traceReceiver = trace.New(r.cfg.ID(), tc, r.settings)
	return nil
}

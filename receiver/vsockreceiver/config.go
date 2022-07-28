package vsockreceiver

import (
	"fmt"
	"strings"

	"github.com/linuxkit/virtsock/pkg/hvsock"
	"go.opentelemetry.io/collector/config"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	config.ReceiverSettings `mapstructure:",squash"`
	Addr                    string `mapstructure:"addr"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	split := strings.Split(cfg.Addr, ":")

	if len(split) != 2 {
		return fmt.Errorf("Addr must be in format of VMID:ServiceID")
	}

	_, err := hvsock.GUIDFromString(split[0])
	if err != nil {
		return err
	}

	_, err = hvsock.GUIDFromString(split[1])
	if err != nil {
		return err
	}

	return nil
}

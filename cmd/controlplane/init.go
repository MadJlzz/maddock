package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/MadJlzz/maddock/internal/pki"
	"github.com/spf13/cobra"
)

const (
	DefaultRootStateDir = "/var/lib/maddock-controlplane"
	DefaultUserStateDir = ".local/share/maddock-controlplane"

	ControlPlaneCertName = "controlplane.crt"
	ControlPlaneKeyName  = "controlplane.key"

	caCommonName = "maddock-ca"
	certValidity = 10 * 365 * 24 * time.Hour
)

func defaultStateDir() (string, error) {
	if os.Geteuid() == 0 {
		return DefaultRootStateDir, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(homeDir, DefaultUserStateDir), nil
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := cmd.Flags().GetString("state-dir")
			if err != nil {
				return err
			}
			return runInit(stateDir)
		},
	}
	return cmd
}

func runInit(stateDir string) error {
	caPath := filepath.Join(stateDir, pki.CertificateAuthorityCertName)
	if _, err := os.Stat(caPath); err == nil {
		return fmt.Errorf("state directory already initialized at %s: delete it manually if you really want to re-init", stateDir)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat CA certificate: %w", err)
	}

	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	ca, err := pki.GenerateCA(caCommonName)
	if err != nil {
		return fmt.Errorf("generate CA: %w", err)
	}
	if err := ca.Save(stateDir); err != nil {
		return fmt.Errorf("save CA: %w", err)
	}

	cpKey, err := pki.GenerateKey()
	if err != nil {
		return fmt.Errorf("generate control plane key: %w", err)
	}

	cpDER, err := ca.SignControlPlaneCert(cpKey.Public(), certValidity)
	if err != nil {
		return fmt.Errorf("sign control plane cert: %w", err)
	}

	cpCertPEM := pki.EncodeCertPEM(cpDER)
	cpKeyPEM, err := pki.EncodeKeyPEM(cpKey)
	if err != nil {
		return fmt.Errorf("encode control plane key: %w", err)
	}

	if err := os.WriteFile(filepath.Join(stateDir, ControlPlaneCertName), cpCertPEM, 0644); err != nil {
		return fmt.Errorf("write control plane cert: %w", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, ControlPlaneKeyName), cpKeyPEM, 0600); err != nil {
		return fmt.Errorf("write control plane key: %w", err)
	}

	fmt.Printf("initialized control plane at %s\n", stateDir)
	return nil
}

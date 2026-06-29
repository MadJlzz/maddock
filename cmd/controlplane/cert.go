package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MadJlzz/maddock/internal/pki"
	"github.com/spf13/cobra"
)

func newCertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert",
		Short: "Handle certificates for agents",
	}
	cmd.AddCommand(newCertIssueCmd())
	return cmd
}

func newCertIssueCmd() *cobra.Command {
	var (
		hostname string
		ttl      time.Duration
		output   string
	)
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue certificates for agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := cmd.Flags().GetString("state-dir")
			if err != nil {
				return err
			}

			ca, err := pki.Load(stateDir)
			if err != nil {
				return err
			}
			pkey, err := pki.GenerateKey()
			if err != nil {
				return err
			}

			signedCert, err := ca.SignAgentCert(hostname, pkey.Public(), ttl)
			if err != nil {
				return err
			}

			if err = os.MkdirAll(output, 0700); err != nil {
				return err
			}

			caFilepath := filepath.Join(output, "ca.crt")
			if err = os.WriteFile(caFilepath, pki.EncodeCertPEM(ca.Cert.Raw), 0600); err != nil {
				return err
			}

			agentCertFilepath := filepath.Join(output, fmt.Sprintf("%s.crt", hostname))
			if err = os.WriteFile(agentCertFilepath, pki.EncodeCertPEM(signedCert), 0600); err != nil {
				return err
			}

			agentPrivateKeyFilepath := filepath.Join(output, fmt.Sprintf("%s.key", hostname))
			encodedPkey, err := pki.EncodeKeyPEM(pkey)
			if err != nil {
				return err
			}
			if err = os.WriteFile(agentPrivateKeyFilepath, encodedPkey, 0600); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "agent", "agent hostname; used as the cert CN/SAN and output filenames")
	cmd.Flags().DurationVar(&ttl, "ttl", time.Hour*24, "certificate validity duration")
	cmd.Flags().StringVar(&output, "output", "./", "directory to write ca.crt, <hostname>.crt and <hostname>.key into")
	return cmd
}

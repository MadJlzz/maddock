package main

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"

	"github.com/MadJlzz/maddock/internal/pki"
	"github.com/MadJlzz/maddock/internal/transport"
	"github.com/MadJlzz/maddock/internal/transport/proto"
	"google.golang.org/grpc/credentials"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func newServeCmd() *cobra.Command {
	var (
		listen string
		caCert string
		cert   string
		key    string
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the agent as a gRPC server, accepting catalogs pushed by maddock-controlplane",
		RunE: func(cmd *cobra.Command, args []string) error {
			if caCert == "" || cert == "" || key == "" {
				return fmt.Errorf("--ca-cert, --cert and --key are all required")
			}

			agentCert, err := tls.LoadX509KeyPair(cert, key)
			if err != nil {
				return fmt.Errorf("loading agent keypair: %w", err)
			}
			caPool, err := pki.LoadCertPool(caCert)
			if err != nil {
				return fmt.Errorf("loading CA cert: %w", err)
			}

			lis, err := net.Listen("tcp", listen)
			if err != nil {
				return fmt.Errorf("listening on %s: %w", listen, err)
			}

			tlsCfg := &tls.Config{
				Certificates: []tls.Certificate{agentCert},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    caPool,
				MinVersion:   tls.VersionTLS13,
			}

			grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsCfg)))
			proto.RegisterAgentServiceServer(grpcServer, &transport.AgentServer{Version: Version})

			slog.Info("maddock agent listening", "address", listen, "version", Version)
			return grpcServer.Serve(lis)
		},
	}
	cmd.Flags().StringVar(&listen, "listen", ":9600", "address to listen on (host:port)")
	cmd.Flags().StringVar(&caCert, "ca-cert", "", "path to the CA cert (trust anchor for the control plane's client cert)")
	cmd.Flags().StringVar(&cert, "cert", "", "path to the agent's own server cert")
	cmd.Flags().StringVar(&key, "key", "", "path to the agent's own server key")
	return cmd
}

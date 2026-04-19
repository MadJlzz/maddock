package main

import (
	"fmt"
	"net"

	"github.com/MadJlzz/maddock/internal/transport"
	"github.com/MadJlzz/maddock/internal/transport/proto"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func newServeCmd() *cobra.Command {
	var listen string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the agent as a gRPC server, accepting catalogs pushed by maddock-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			lis, err := net.Listen("tcp", listen)
			if err != nil {
				return fmt.Errorf("listening on %s: %w", listen, err)
			}

			grpcServer := grpc.NewServer()
			proto.RegisterAgentServiceServer(grpcServer, &transport.Server{Version: Version})

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Maddock agent listening on %s\n", listen)
			return grpcServer.Serve(lis)
		},
	}
	cmd.Flags().StringVar(&listen, "listen", ":9600", "address to listen on (host:port)")
	return cmd
}

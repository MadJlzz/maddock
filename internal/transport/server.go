package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/engine"
	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/transport/proto"
	"google.golang.org/grpc"
)

var stateToProto = map[resource.State]proto.State{
	resource.Ok:      proto.State_STATE_OK,
	resource.Changed: proto.State_STATE_CHANGED,
	resource.Failed:  proto.State_STATE_FAILED,
	resource.Skipped: proto.State_STATE_SKIPPED,
}

type Server struct {
	proto.UnimplementedAgentServiceServer
	Version string
}

func (s *Server) ApplyCatalog(request *proto.CatalogRequest, g grpc.ServerStreamingServer[proto.ResourceReportMsg]) error {
	slog.Info("received ApplyCatalog",
		"manifest", request.GetManifestName(),
		"resources", len(request.GetResources()),
		"dry_run", request.GetDryRun())

	var catalogResources []resource.Resource
	for _, rr := range request.GetResources() {
		var attrs map[string]any
		if err := json.Unmarshal(rr.Attributes, &attrs); err != nil {
			return err
		}
		r, err := resource.Parse(rr.Type, rr.Name, attrs)
		if err != nil {
			return err
		}
		catalogResources = append(catalogResources, r)
	}
	c := catalog.Catalog{
		Name:      request.GetManifestName(),
		Resources: catalogResources,
	}
	report := engine.Run(g.Context(), &c, request.DryRun)

	for _, rr := range report.ResourceReports {
		changes := make([]*proto.DifferenceMsg, len(rr.Differences))
		for i, d := range rr.Differences {
			changes[i] = &proto.DifferenceMsg{
				Attribute: d.Attribute,
				Current:   d.Current,
				Desired:   d.Desired,
			}
		}
		var errStr string
		if rr.Error != nil {
			errStr = rr.Error.Error()
		}
		prrm := proto.ResourceReportMsg{
			Type:       rr.Type,
			Name:       rr.Name,
			State:      stateToProto[rr.State],
			Changes:    changes,
			Error:      errStr,
			DurationMs: rr.Duration.Milliseconds(),
		}
		err := g.Send(&prrm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Ping(_ context.Context, _ *proto.PingRequest) (*proto.PingResponse, error) {
	name, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &proto.PingResponse{
		Hostname:     name,
		AgentVersion: s.Version,
	}, nil
}

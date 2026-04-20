package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/MadJlzz/maddock/internal/catalog"
	"github.com/MadJlzz/maddock/internal/report"
	"github.com/MadJlzz/maddock/internal/resource"
	"github.com/MadJlzz/maddock/internal/transport/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var protoToState = map[proto.State]resource.State{
	proto.State_STATE_OK:      resource.Ok,
	proto.State_STATE_CHANGED: resource.Changed,
	proto.State_STATE_FAILED:  resource.Failed,
	proto.State_STATE_SKIPPED: resource.Skipped,
}

// Client is a gRPC client for the agent service.
type Client struct {
	conn  *grpc.ClientConn
	agent proto.AgentServiceClient
}

// NewClient dials the agent at the given address.
func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dialing %s: %w", address, err)
	}
	return &Client{
		conn:  conn,
		agent: proto.NewAgentServiceClient(conn),
	}, nil
}

// Close shuts down the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Ping returns the agent's hostname and version.
func (c *Client) Ping(ctx context.Context) (*proto.PingResponse, error) {
	return c.agent.Ping(ctx, &proto.PingRequest{})
}

// ApplyCatalog pushes the raw catalog to the agent and collects streamed
// per-resource reports into a single Report.
func (c *Client) ApplyCatalog(ctx context.Context, rc *catalog.RawCatalog, dryRun bool) (*report.Report, error) {
	resources := make([]*proto.ResourceMsg, 0, len(rc.Resources))
	for _, rr := range rc.Resources {
		attrs, err := json.Marshal(rr.Attributes)
		if err != nil {
			return nil, fmt.Errorf("marshaling attributes for %s:%s: %w", rr.Type, rr.Name, err)
		}
		resources = append(resources, &proto.ResourceMsg{
			Type:       rr.Type,
			Name:       rr.Name,
			Attributes: attrs,
		})
	}

	req := &proto.CatalogRequest{
		ManifestName: rc.Name,
		DryRun:       dryRun,
		Resources:    resources,
	}

	stream, err := c.agent.ApplyCatalog(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("starting ApplyCatalog stream: %w", err)
	}

	r := &report.Report{Name: rc.Name}
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("receiving report: %w", err)
		}
		r.ResourceReports = append(r.ResourceReports, resourceReportFromProto(msg))
	}
	return r, nil
}

func resourceReportFromProto(msg *proto.ResourceReportMsg) report.ResourceReport {
	diffs := make([]resource.Difference, len(msg.Changes))
	for i, d := range msg.Changes {
		diffs[i] = resource.Difference{
			Attribute: d.Attribute,
			Current:   d.Current,
			Desired:   d.Desired,
		}
	}
	var err error
	if msg.Error != "" {
		err = errors.New(msg.Error)
	}
	return report.ResourceReport{
		Type:        msg.Type,
		Name:        msg.Name,
		State:       protoToState[msg.State],
		Differences: diffs,
		Error:       err,
		Duration:    time.Duration(msg.DurationMs) * time.Millisecond,
	}
}

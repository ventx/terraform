package plugin

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform/plugin/proto"
	"github.com/hashicorp/terraform/terraform"
	"google.golang.org/grpc"
)

// terraform.ResourceProvider grpc implementation
type GRPCResourceProvider struct {
	conn   *grpc.ClientConn
	client proto.ProviderClient
}

func (p *GRPCResourceProvider) Stop() error {
	_, err := p.client.Stop(context.TODO(), nil)
	return err
}

func (p *GRPCResourceProvider) GetSchema(req *terraform.ProviderSchemaRequest) (*terraform.ProviderSchema, error) {
	resp, err := p.client.GetSchema(context.TODO(), &proto.GetSchemaRequest{
		ResourceTypes: req.ResourceTypes,
		DataSources:   req.DataSources,
	})

	if err != nil {
		return nil, err
	}

	var s terraform.ProviderSchema
	if err := json.Unmarshal(resp.Schema, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *GRPCResourceProvider) Input(input terraform.UIInput, c *terraform.ResourceConfig) (*terraform.ResourceConfig, error) {
	req := &proto.InputRequest{
		ResourceConfig: proto.NewResourceConfig(c),
	}

	// Open an input stream with the plugin
	inputClient, err := p.client.Input(context.TODO())
	if err != nil {
		return nil, err
	}

	if err := inputClient.Send(req); err != nil {
		return nil, err
	}

	// remove the config, since further requests are input replies
	req.ResourceConfig = nil

	// accept input until we get a ResourceConfig
	var rc *terraform.ResourceConfig
	for {
		resp, err := inputClient.Recv()
		if err != nil {
			return nil, err
		}

		if resp.ResourceConfig != nil {
			rc = resp.ResourceConfig.TFResourceConfig()
			break
		}

		opts := &terraform.InputOpts{
			Id:          resp.Id,
			Query:       resp.Query,
			Description: resp.Description,
			Default:     resp.Default,
		}

		userInput, err := input.Input(opts)
		if err != nil {
			return nil, err
		}

		req.Reply = userInput
		if err := inputClient.Send(req); err != nil {
			return nil, err
		}
	}

	return rc, nil
}

func (p *GRPCResourceProvider) Validate(c *terraform.ResourceConfig) ([]string, []error) {
	req := &proto.ValidateRequest{
		Config: proto.NewResourceConfig(c),
	}
	resp, err := p.client.Validate(context.TODO(), req)
	if err != nil {
		return nil, []error{err}
	}

	return resp.Warnings, resp.ErrorList()
}

func (p *GRPCResourceProvider) ValidateResource(t string, c *terraform.ResourceConfig) ([]string, []error) {

	req := &proto.ValidateResourceRequest{
		Type:   t,
		Config: proto.NewResourceConfig(c),
	}

	resp, err := p.client.ValidateResource(context.TODO(), req)
	if err != nil {
		return nil, []error{err}
	}

	return resp.Warnings, resp.ErrorList()
}

func (p *GRPCResourceProvider) Configure(c *terraform.ResourceConfig) error {
	req := &proto.ConfigureRequest{
		ResourceConfig: proto.NewResourceConfig(c),
	}

	_, err := p.client.Configure(context.TODO(), req)
	return err
}

func (p *GRPCResourceProvider) Apply(info *terraform.InstanceInfo, s *terraform.InstanceState, d *terraform.InstanceDiff) (*terraform.InstanceState, error) {
	req := &proto.ApplyRequest{
		Info:  proto.NewInstanceInfo(info),
		State: proto.NewInstanceState(s),
		Diff:  proto.NewInstanceDiff(d),
	}

	resp, err := p.client.Apply(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	return resp.State.TFInstanceState(), nil
}

func (p *GRPCResourceProvider) Diff(info *terraform.InstanceInfo, s *terraform.InstanceState, c *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
	req := &proto.DiffRequest{
		Info:   proto.NewInstanceInfo(info),
		State:  proto.NewInstanceState(s),
		Config: proto.NewResourceConfig(c),
	}

	resp, err := p.client.Diff(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	return resp.Diff.TFInstanceDiff(), nil
}

func (p *GRPCResourceProvider) ValidateDataSource(t string, c *terraform.ResourceConfig) ([]string, []error) {
	req := &proto.ValidateDataSourceRequest{
		Type:   t,
		Config: proto.NewResourceConfig(c),
	}

	resp, err := p.client.ValidateDataSource(context.TODO(), req)
	if err != nil {
		return nil, []error{err}
	}

	return resp.Warnings, resp.ErrorList()
}

func (p *GRPCResourceProvider) Refresh(info *terraform.InstanceInfo, s *terraform.InstanceState) (*terraform.InstanceState, error) {
	req := &proto.RefreshRequest{
		Info:  proto.NewInstanceInfo(info),
		State: proto.NewInstanceState(s),
	}

	resp, err := p.client.Refresh(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	return resp.State.TFInstanceState(), nil
}

func (p *GRPCResourceProvider) ImportState(info *terraform.InstanceInfo, id string) ([]*terraform.InstanceState, error) {
	req := &proto.ImportStateRequest{
		Id:   id,
		Info: proto.NewInstanceInfo(info),
	}

	resp, err := p.client.ImportState(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	return resp.TFInstanceStates(), nil
}

func (p *GRPCResourceProvider) Resources() []terraform.ResourceType {
	resp, err := p.client.Resources(context.TODO(), nil)
	if err != nil {
		log.Println("[ERROR]", err)
		return nil
	}

	return resp.TFResources()
}

func (p *GRPCResourceProvider) ReadDataDiff(info *terraform.InstanceInfo, c *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
	req := &proto.ReadDataDiffRequest{
		Info:   proto.NewInstanceInfo(info),
		Config: proto.NewResourceConfig(c),
	}

	resp, err := p.client.ReadDataDiff(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	return resp.Diff.TFInstanceDiff(), nil
}

func (p *GRPCResourceProvider) ReadDataApply(info *terraform.InstanceInfo, d *terraform.InstanceDiff) (*terraform.InstanceState, error) {
	req := &proto.ReadDataApplyRequest{
		Info: proto.NewInstanceInfo(info),
		Diff: proto.NewInstanceDiff(d),
	}

	resp, err := p.client.ReadDataApply(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	return resp.State.TFInstanceState(), nil
}

func (p *GRPCResourceProvider) DataSources() []terraform.DataSource {
	resp, err := p.client.DataSources(context.TODO(), nil)
	if err != nil {
		log.Println("[ERROR]", err)
	}

	return resp.TFDataSources()
}

func (p *GRPCResourceProvider) Close() error {
	return nil
	return p.conn.Close()
}

type GRPCResourceProviderServer struct {
	provider terraform.ResourceProvider
}

func (s *GRPCResourceProviderServer) Stop(_ context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return new(proto.Empty), s.provider.Stop()
}

func (s *GRPCResourceProviderServer) GetSchema(_ context.Context, req *proto.GetSchemaRequest) (*proto.GetSchemaResponse, error) {
	psr := &terraform.ProviderSchemaRequest{
		ResourceTypes: req.ResourceTypes,
		DataSources:   req.DataSources,
	}

	ps, err := s.provider.GetSchema(psr)
	if err != nil {
		return nil, err
	}

	js, err := json.Marshal(ps)
	if err != nil {
		return nil, err
	}

	return &proto.GetSchemaResponse{Schema: js}, nil

}

func (s *GRPCResourceProviderServer) Input(server proto.Provider_InputServer) error {
	req, err := server.Recv()
	if err != nil {
		return err
	}

	rc := req.ResourceConfig.TFResourceConfig()
	rc, err = s.provider.Input(&grpcInputServer{server: server}, rc)
	if err != nil {
		return err
	}

	return server.Send(&proto.InputResponse{ResourceConfig: proto.NewResourceConfig(rc)})
}

type grpcInputServer struct {
	server         proto.Provider_InputServer
	resourceConfig *terraform.ResourceConfig
}

func (s *grpcInputServer) Input(opts *terraform.InputOpts) (string, error) {
	resp := &proto.InputResponse{
		Id:          opts.Id,
		Query:       opts.Query,
		Description: opts.Description,
		Default:     opts.Default,
	}

	if err := s.server.Send(resp); err != nil {
		return "", err
	}

	req, err := s.server.Recv()
	if err != nil {
		return "", err
	}

	return req.Reply, nil
}

func (s *GRPCResourceProviderServer) Validate(_ context.Context, req *proto.ValidateRequest) (*proto.ValidateResponse, error) {
	w, e := s.provider.Validate(req.Config.TFResourceConfig())
	return proto.NewValidateResponse(w, e), nil
}

func (s *GRPCResourceProviderServer) ValidateResource(_ context.Context, req *proto.ValidateResourceRequest) (*proto.ValidateResponse, error) {
	w, e := s.provider.ValidateResource(req.Type, req.Config.TFResourceConfig())
	return proto.NewValidateResponse(w, e), nil
}

func (s *GRPCResourceProviderServer) Configure(_ context.Context, req *proto.ConfigureRequest) (*proto.Empty, error) {
	err := s.provider.Configure(req.ResourceConfig.TFResourceConfig())
	return new(proto.Empty), err
}

func (s *GRPCResourceProviderServer) Apply(_ context.Context, req *proto.ApplyRequest) (*proto.ApplyResponse, error) {
	is, err := s.provider.Apply(req.Info.TFInstanceInfo(), req.State.TFInstanceState(), req.Diff.TFInstanceDiff())
	if err != nil {
		return nil, err
	}

	return &proto.ApplyResponse{State: proto.NewInstanceState(is)}, nil
}

func (s *GRPCResourceProviderServer) Diff(_ context.Context, req *proto.DiffRequest) (*proto.DiffResponse, error) {
	d, err := s.provider.Diff(req.Info.TFInstanceInfo(), req.State.TFInstanceState(), req.Config.TFResourceConfig())
	if err != nil {
		return nil, err
	}
	return &proto.DiffResponse{Diff: proto.NewInstanceDiff(d)}, nil
}

func (s *GRPCResourceProviderServer) Refresh(_ context.Context, req *proto.RefreshRequest) (*proto.RefreshResponse, error) {
	is, err := s.provider.Refresh(req.Info.TFInstanceInfo(), req.State.TFInstanceState())
	if err != nil {
		return nil, err
	}
	return &proto.RefreshResponse{State: proto.NewInstanceState(is)}, nil
}

func (s *GRPCResourceProviderServer) ImportState(_ context.Context, req *proto.ImportStateRequest) (*proto.ImportStateResponse, error) {
	states, err := s.provider.ImportState(req.Info.TFInstanceInfo(), req.Id)
	if err != nil {
		return nil, err
	}

	return proto.NewImportStateResponse(states), nil
}

func (s *GRPCResourceProviderServer) Resources(_ context.Context, _ *proto.Empty) (*proto.ResourcesResponse, error) {
	return proto.NewResourcesResponse(s.provider.Resources()), nil
}

func (s *GRPCResourceProviderServer) ValidateDataSource(_ context.Context, req *proto.ValidateDataSourceRequest) (*proto.ValidateResponse, error) {
	w, e := s.provider.ValidateDataSource(req.Type, req.Config.TFResourceConfig())
	return proto.NewValidateResponse(w, e), nil
}

func (s *GRPCResourceProviderServer) ReadDataDiff(_ context.Context, req *proto.ReadDataDiffRequest) (*proto.ReadDataDiffResponse, error) {
	diff, err := s.provider.ReadDataDiff(req.Info.TFInstanceInfo(), req.Config.TFResourceConfig())
	if err != nil {
		return nil, err
	}

	return &proto.ReadDataDiffResponse{Diff: proto.NewInstanceDiff(diff)}, nil
}

func (s *GRPCResourceProviderServer) ReadDataApply(_ context.Context, req *proto.ReadDataApplyRequest) (*proto.ReadDataApplyResponse, error) {
	state, err := s.provider.ReadDataApply(req.Info.TFInstanceInfo(), req.Diff.TFInstanceDiff())
	if err != nil {
		return nil, err
	}
	return &proto.ReadDataApplyResponse{State: proto.NewInstanceState(state)}, nil
}

func (s *GRPCResourceProviderServer) DataSources(_ context.Context, _ *proto.Empty) (*proto.DataSourcesResponse, error) {
	return proto.NewDataSourcesResponse(s.provider.DataSources()), nil
}

// terraform.ResourceProvider grpc implementation
type GRPCResourceProvisioner struct {
	conn   *grpc.ClientConn
	client proto.ProvisionerClient
}

func (p *GRPCResourceProvisioner) Validate(c *terraform.ResourceConfig) ([]string, []error) {
	req := &proto.ValidateRequest{
		Config: proto.NewResourceConfig(c),
	}
	resp, err := p.client.Validate(context.TODO(), req)
	if err != nil {
		return nil, []error{err}
	}

	return resp.Warnings, resp.ErrorList()
}

func (p *GRPCResourceProvisioner) Apply(out terraform.UIOutput, s *terraform.InstanceState, c *terraform.ResourceConfig) error {
	req := &proto.ProvisionerApplyRequest{
		State:  proto.NewInstanceState(s),
		Config: proto.NewResourceConfig(c),
	}

	outputClient, err := p.client.Apply(context.TODO(), req)
	if err != nil {
		return err
	}

	for {
		output, err := outputClient.Recv()
		if err != nil {
			return err
		}
		out.Output(output.Message)
	}

	return nil
}

func (p *GRPCResourceProvisioner) Stop() error {
	_, err := p.client.Stop(context.TODO(), nil)
	return err
}

type GRPCResourceProvisionerServer struct {
	provisioner terraform.ResourceProvisioner
}

func (s *GRPCResourceProvisionerServer) Validate(_ context.Context, req *proto.ValidateRequest) (*proto.ValidateResponse, error) {
	w, e := s.provisioner.Validate(req.Config.TFResourceConfig())
	return proto.NewValidateResponse(w, e), nil
}

func (s *GRPCResourceProvisionerServer) Apply(req *proto.ProvisionerApplyRequest, server proto.Provisioner_ApplyServer) error {
	return s.provisioner.Apply(&grpcOutputServer{server: server}, req.State.TFInstanceState(), req.Config.TFResourceConfig())
}

func (s *GRPCResourceProvisionerServer) Stop(_ context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return new(proto.Empty), s.provisioner.Stop()
}

type grpcOutputServer struct {
	server proto.Provisioner_ApplyServer
}

func (s *grpcOutputServer) Output(msg string) {
	s.server.Send(&proto.UIOutput{Message: msg})
}

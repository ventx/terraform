package proto

import (
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform/terraform"
)

// This file defines new functions and methods to translate the protobuf types
// to and from terraform types.
// New* function take a terraform type and return a protobuf type
// TF* methods return a terraform type from a protobuf type.

// marshalMap marshals the interface{}s into json to ensure that the payload is
// serializable over grpc
func marshalMap(m map[string]interface{}) map[string][]byte {
	if m == nil {
		return nil
	}

	n := make(map[string][]byte)
	for k, v := range m {
		js, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		n[k] = js
	}

	return n
}

// unmarshalMap unmarshals the json data into an empty interface{} for use in
// the terraform package
func unmarshalMap(m map[string][]byte) map[string]interface{} {
	if m == nil {
		return nil
	}

	n := make(map[string]interface{})

	for k, v := range m {
		var i interface{}
		err := json.Unmarshal(v, &i)
		if err != nil {
			panic(err)
		}
		n[k] = i
	}

	return n
}

func NewResourceConfig(c *terraform.ResourceConfig) *ResourceConfig {
	if c == nil {
		return nil
	}

	return &ResourceConfig{
		ComputedKeys: c.ComputedKeys,
		Raw:          marshalMap(c.Raw),
		Config:       marshalMap(c.Config),
	}
}

func (c *ResourceConfig) TFResourceConfig() *terraform.ResourceConfig {
	return &terraform.ResourceConfig{
		ComputedKeys: c.ComputedKeys,
		Raw:          unmarshalMap(c.Raw),
		Config:       unmarshalMap(c.Config),
	}
}

func NewInstanceInfo(i *terraform.InstanceInfo) *InstanceInfo {
	if i == nil {
		return nil
	}
	return &InstanceInfo{
		Id:         i.Id,
		Type:       i.Type,
		ModulePath: i.ModulePath,
	}
}

func (i *InstanceInfo) TFInstanceInfo() *terraform.InstanceInfo {
	return &terraform.InstanceInfo{
		Id:         i.Id,
		Type:       i.Type,
		ModulePath: i.ModulePath,
	}
}

func NewEphemeralState(s terraform.EphemeralState) *EphemeralState {
	return &EphemeralState{
		Type:     s.Type,
		ConnInfo: s.ConnInfo,
	}
}

func (s *EphemeralState) TFEphemeralState() terraform.EphemeralState {
	es := terraform.EphemeralState{}
	if s != nil {
		es.Type = s.Type
		es.ConnInfo = s.ConnInfo
	}

	return es
}

func NewInstanceState(s *terraform.InstanceState) *InstanceState {
	if s == nil {
		return nil
	}

	attrs := make(map[string][]byte)

	// these may hold more info in the future, but for now they are just strings
	for k, v := range s.Attributes {
		attrs[k] = []byte(v)
	}

	meta, err := json.Marshal(s.Meta)
	if err != nil {
		panic(err)
	}

	return &InstanceState{
		Id:         s.ID,
		Attributes: attrs,
		Ephemeral:  NewEphemeralState(s.Ephemeral),
		Meta:       meta,
		Tainted:    s.Tainted,
	}
}

func (s *InstanceState) TFInstanceState() *terraform.InstanceState {
	if s == nil {
		return nil
	}

	var attrs map[string]string
	if s.Attributes != nil {
		attrs = make(map[string]string)
	}

	for k, v := range s.Attributes {
		attrs[k] = string(v)
	}

	var meta map[string]interface{}
	if s.Meta != nil {
		err := json.Unmarshal(s.Meta, &meta)
		if err != nil {
			panic(err)
		}
	}

	return &terraform.InstanceState{
		ID:         s.Id,
		Attributes: attrs,
		Ephemeral:  s.Ephemeral.TFEphemeralState(),
		Meta:       meta,
		Tainted:    s.Tainted,
	}
}

func NewResourceAttrDiff(d *terraform.ResourceAttrDiff) *ResourceAttrDiff {
	if d == nil {
		return nil
	}

	newExtra, err := json.Marshal(d.NewExtra)
	if err != nil {
		panic(err)
	}

	return &ResourceAttrDiff{
		Old:         d.Old,
		New:         d.New,
		NewComputed: d.NewComputed,
		NewRemoved:  d.NewRemoved,
		NewExtra:    newExtra,
		RequiresNew: d.RequiresNew,
		Sensitive:   d.Sensitive,
		Type:        DiffAttrType(d.Type),
	}
}

func (d *ResourceAttrDiff) TFResourceAttrDiff() *terraform.ResourceAttrDiff {
	if d == nil {
		return nil
	}

	var newExtra interface{}
	if d.NewExtra != nil {
		err := json.Unmarshal(d.NewExtra, &newExtra)
		if err != nil {
			panic(err)
		}
	}

	return &terraform.ResourceAttrDiff{
		Old:         d.Old,
		New:         d.New,
		NewComputed: d.NewComputed,
		NewRemoved:  d.NewRemoved,
		NewExtra:    newExtra,
		RequiresNew: d.RequiresNew,
		Sensitive:   d.Sensitive,
		Type:        terraform.DiffAttrType(d.Type),
	}
}

func NewInstanceDiff(d *terraform.InstanceDiff) *InstanceDiff {
	if d == nil {
		return nil
	}

	// make sure nil is conveyed
	var attrs map[string]*ResourceAttrDiff
	if d.Attributes != nil {
		attrs = make(map[string]*ResourceAttrDiff)
	}
	for k, attr := range d.Attributes {
		attrs[k] = NewResourceAttrDiff(attr)
	}

	meta, err := json.Marshal(d.Meta)
	if err != nil {
		panic(err)
	}

	return &InstanceDiff{
		Attributes:     attrs,
		Destroy:        d.Destroy,
		DestroyDeposed: d.DestroyDeposed,
		DestroyTainted: d.DestroyTainted,
		Meta:           meta,
	}
}

func (d *InstanceDiff) TFInstanceDiff() *terraform.InstanceDiff {
	if d == nil {
		return nil
	}

	// make sure nil is conveyed
	var attrs map[string]*terraform.ResourceAttrDiff
	if d.Attributes != nil {
		attrs = make(map[string]*terraform.ResourceAttrDiff)
	}
	for k, attr := range d.Attributes {
		attrs[k] = attr.TFResourceAttrDiff()
	}

	var meta map[string]interface{}
	if d.Meta != nil {
		err := json.Unmarshal(d.Meta, &meta)
		if err != nil {
			panic(err)
		}
	}

	return &terraform.InstanceDiff{
		Attributes:     attrs,
		Destroy:        d.Destroy,
		DestroyDeposed: d.DestroyDeposed,
		DestroyTainted: d.DestroyTainted,
		Meta:           meta,
	}
}

func NewImportStateResponse(s []*terraform.InstanceState) *ImportStateResponse {
	r := &ImportStateResponse{}
	for _, state := range s {
		r.State = append(r.State, NewInstanceState(state))
	}

	return r
}

func (r *ImportStateResponse) TFInstanceStates() []*terraform.InstanceState {
	var states []*terraform.InstanceState

	for _, s := range r.State {
		states = append(states, s.TFInstanceState())
	}

	return states
}

func NewDataSourcesResponse(ds []terraform.DataSource) *DataSourcesResponse {
	resp := &DataSourcesResponse{}
	for _, d := range ds {
		resp.DataSources = append(resp.DataSources, d.Name)
	}
	return resp
}

func (r *DataSourcesResponse) TFDataSources() []terraform.DataSource {
	var ds []terraform.DataSource
	for _, d := range r.DataSources {
		ds = append(ds, terraform.DataSource{Name: d})
	}
	return ds
}

func NewResourceType(rs terraform.ResourceType) *ResourceType {
	return &ResourceType{
		Name:       rs.Name,
		Importable: rs.Importable,
	}
}

func (r *ResourceType) TFResourceType() terraform.ResourceType {
	rt := terraform.ResourceType{}
	if r != nil {
		rt.Name = r.Name
		rt.Importable = r.Importable
	}
	return rt
}

func NewResourcesResponse(rs []terraform.ResourceType) *ResourcesResponse {
	resp := &ResourcesResponse{}
	for _, r := range rs {
		resp.Resources = append(resp.Resources, NewResourceType(r))
	}
	return resp
}

func (r *ResourcesResponse) TFResources() []terraform.ResourceType {
	var rs []terraform.ResourceType
	for _, rt := range r.Resources {
		rs = append(rs, rt.TFResourceType())
	}
	return rs
}

func NewValidateResponse(w []string, e []error) *ValidateResponse {
	resp := &ValidateResponse{
		Warnings: w,
	}
	for _, err := range e {
		resp.Errors = append(resp.Errors, err.Error())
	}
	return resp
}

// ErrorList is a convenience method to convert the array of protobuf Error
// messages to a Go []error.
func (r *ValidateResponse) ErrorList() []error {
	if r == nil || len(r.Errors) == 0 {
		return nil
	}

	errs := make([]error, len(r.Errors))
	for i := range r.Errors {
		errs[i] = errors.New(r.Errors[i])
	}
	return errs
}

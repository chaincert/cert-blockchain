package types

import (
	"context"

	"google.golang.org/grpc"
)

// MsgServer defines the attestation module's gRPC message service
type MsgServer interface {
	// RegisterSchema registers a new attestation schema
	RegisterSchema(context.Context, *MsgRegisterSchema) (*MsgRegisterSchemaResponse, error)

	// Attest creates a new public attestation
	Attest(context.Context, *MsgAttest) (*MsgAttestResponse, error)

	// Revoke revokes an existing attestation
	Revoke(context.Context, *MsgRevoke) (*MsgRevokeResponse, error)

	// CreateEncryptedAttestation creates a new encrypted attestation
	CreateEncryptedAttestation(context.Context, *MsgCreateEncryptedAttestation) (*MsgCreateEncryptedAttestationResponse, error)
}

// MsgRegisterSchemaResponse is the response for MsgRegisterSchema
type MsgRegisterSchemaResponse struct {
	Uid string `json:"uid"`
}

// MsgAttestResponse is the response for MsgAttest
type MsgAttestResponse struct {
	Uid string `json:"uid"`
}

// MsgRevokeResponse is the response for MsgRevoke
type MsgRevokeResponse struct{}

// MsgCreateEncryptedAttestationResponse is the response for MsgCreateEncryptedAttestation
type MsgCreateEncryptedAttestationResponse struct {
	Uid string `json:"uid"`
}

// QueryServer defines the attestation module's gRPC query service
type QueryServer interface {
	// Schema queries a schema by UID
	Schema(context.Context, *QuerySchemaRequest) (*QuerySchemaResponse, error)

	// Attestation queries an attestation by UID
	Attestation(context.Context, *QueryAttestationRequest) (*QueryAttestationResponse, error)

	// AttestationsByAttester queries all attestations by an attester
	AttestationsByAttester(context.Context, *QueryAttestationsByAttesterRequest) (*QueryAttestationsByAttesterResponse, error)

	// AttestationsByRecipient queries all attestations for a recipient
	AttestationsByRecipient(context.Context, *QueryAttestationsByRecipientRequest) (*QueryAttestationsByRecipientResponse, error)

	// EncryptedAttestation queries an encrypted attestation
	EncryptedAttestation(context.Context, *QueryEncryptedAttestationRequest) (*QueryEncryptedAttestationResponse, error)

	// Stats returns attestation statistics
	Stats(context.Context, *QueryStatsRequest) (*QueryStatsResponse, error)
}

// Query request/response types

// QuerySchemaRequest is the request type for the Query/Schema RPC method
type QuerySchemaRequest struct {
	Uid string `json:"uid" protobuf:"bytes,1,opt,name=uid,proto3"`
}

func (m *QuerySchemaRequest) Reset()         { *m = QuerySchemaRequest{} }
func (m *QuerySchemaRequest) String() string { return m.Uid }
func (m *QuerySchemaRequest) ProtoMessage()  {}

// QuerySchemaResponse is the response type for the Query/Schema RPC method
type QuerySchemaResponse struct {
	Schema *Schema `json:"schema" protobuf:"bytes,1,opt,name=schema,proto3"`
}

func (m *QuerySchemaResponse) Reset()         { *m = QuerySchemaResponse{} }
func (m *QuerySchemaResponse) String() string { return "QuerySchemaResponse" }
func (m *QuerySchemaResponse) ProtoMessage()  {}

// QueryAttestationRequest is the request type for the Query/Attestation RPC method
type QueryAttestationRequest struct {
	Uid string `json:"uid" protobuf:"bytes,1,opt,name=uid,proto3"`
}

func (m *QueryAttestationRequest) Reset()         { *m = QueryAttestationRequest{} }
func (m *QueryAttestationRequest) String() string { return m.Uid }
func (m *QueryAttestationRequest) ProtoMessage()  {}

// QueryAttestationResponse is the response type for the Query/Attestation RPC method
type QueryAttestationResponse struct {
	Attestation *Attestation `json:"attestation" protobuf:"bytes,1,opt,name=attestation,proto3"`
}

func (m *QueryAttestationResponse) Reset()         { *m = QueryAttestationResponse{} }
func (m *QueryAttestationResponse) String() string { return "QueryAttestationResponse" }
func (m *QueryAttestationResponse) ProtoMessage()  {}

// QueryAttestationsByAttesterRequest is the request type for Query/AttestationsByAttester
type QueryAttestationsByAttesterRequest struct {
	Attester string `json:"attester" protobuf:"bytes,1,opt,name=attester,proto3"`
}

func (m *QueryAttestationsByAttesterRequest) Reset()         { *m = QueryAttestationsByAttesterRequest{} }
func (m *QueryAttestationsByAttesterRequest) String() string { return m.Attester }
func (m *QueryAttestationsByAttesterRequest) ProtoMessage()  {}

// QueryAttestationsByAttesterResponse is the response type for Query/AttestationsByAttester
type QueryAttestationsByAttesterResponse struct {
	Attestations []Attestation `json:"attestations" protobuf:"bytes,1,rep,name=attestations,proto3"`
}

func (m *QueryAttestationsByAttesterResponse) Reset() { *m = QueryAttestationsByAttesterResponse{} }
func (m *QueryAttestationsByAttesterResponse) String() string {
	return "QueryAttestationsByAttesterResponse"
}
func (m *QueryAttestationsByAttesterResponse) ProtoMessage() {}

// QueryAttestationsByRecipientRequest is the request type for Query/AttestationsByRecipient
type QueryAttestationsByRecipientRequest struct {
	Recipient string `json:"recipient" protobuf:"bytes,1,opt,name=recipient,proto3"`
}

func (m *QueryAttestationsByRecipientRequest) Reset()         { *m = QueryAttestationsByRecipientRequest{} }
func (m *QueryAttestationsByRecipientRequest) String() string { return m.Recipient }
func (m *QueryAttestationsByRecipientRequest) ProtoMessage()  {}

// QueryAttestationsByRecipientResponse is the response type for Query/AttestationsByRecipient
type QueryAttestationsByRecipientResponse struct {
	Attestations []Attestation `json:"attestations" protobuf:"bytes,1,rep,name=attestations,proto3"`
}

func (m *QueryAttestationsByRecipientResponse) Reset() { *m = QueryAttestationsByRecipientResponse{} }
func (m *QueryAttestationsByRecipientResponse) String() string {
	return "QueryAttestationsByRecipientResponse"
}
func (m *QueryAttestationsByRecipientResponse) ProtoMessage() {}

// QueryEncryptedAttestationRequest is the request type for Query/EncryptedAttestation
type QueryEncryptedAttestationRequest struct {
	Uid       string `json:"uid" protobuf:"bytes,1,opt,name=uid,proto3"`
	Requester string `json:"requester" protobuf:"bytes,2,opt,name=requester,proto3"` // For access control verification
}

func (m *QueryEncryptedAttestationRequest) Reset()         { *m = QueryEncryptedAttestationRequest{} }
func (m *QueryEncryptedAttestationRequest) String() string { return m.Uid }
func (m *QueryEncryptedAttestationRequest) ProtoMessage()  {}

// QueryEncryptedAttestationResponse is the response type for Query/EncryptedAttestation
type QueryEncryptedAttestationResponse struct {
	Attestation  *EncryptedAttestation `json:"attestation" protobuf:"bytes,1,opt,name=attestation,proto3"`
	EncryptedKey string                `json:"encrypted_key,omitempty" protobuf:"bytes,2,opt,name=encrypted_key,proto3"` // Only if authorized
	Authorized   bool                  `json:"authorized" protobuf:"varint,3,opt,name=authorized,proto3"`
}

func (m *QueryEncryptedAttestationResponse) Reset() { *m = QueryEncryptedAttestationResponse{} }
func (m *QueryEncryptedAttestationResponse) String() string {
	return "QueryEncryptedAttestationResponse"
}
func (m *QueryEncryptedAttestationResponse) ProtoMessage() {}

// QueryStatsRequest is the request type for Query/Stats
type QueryStatsRequest struct{}

func (m *QueryStatsRequest) Reset()         { *m = QueryStatsRequest{} }
func (m *QueryStatsRequest) String() string { return "QueryStatsRequest" }
func (m *QueryStatsRequest) ProtoMessage()  {}

// QueryStatsResponse is the response type for Query/Stats
type QueryStatsResponse struct {
	TotalAttestations          uint64 `json:"total_attestations" protobuf:"varint,1,opt,name=total_attestations,proto3"`
	TotalEncryptedAttestations uint64 `json:"total_encrypted_attestations" protobuf:"varint,2,opt,name=total_encrypted_attestations,proto3"`
	TotalSchemas               uint64 `json:"total_schemas" protobuf:"varint,3,opt,name=total_schemas,proto3"`
	TotalRevocations           uint64 `json:"total_revocations" protobuf:"varint,4,opt,name=total_revocations,proto3"`
}

func (m *QueryStatsResponse) Reset()         { *m = QueryStatsResponse{} }
func (m *QueryStatsResponse) String() string { return "QueryStatsResponse" }
func (m *QueryStatsResponse) ProtoMessage()  {}

// RegisterMsgServer registers the MsgServer implementation with the gRPC server
func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

// RegisterQueryServer registers the QueryServer implementation with the gRPC server
func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

// Service descriptors for gRPC registration
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cert.attestation.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterSchema",
			Handler:    _Msg_RegisterSchema_Handler,
		},
		{
			MethodName: "Attest",
			Handler:    _Msg_Attest_Handler,
		},
		{
			MethodName: "Revoke",
			Handler:    _Msg_Revoke_Handler,
		},
		{
			MethodName: "CreateEncryptedAttestation",
			Handler:    _Msg_CreateEncryptedAttestation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cert/attestation/v1/tx.proto",
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cert.attestation.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "cert/attestation/v1/query.proto",
}

// gRPC method handlers for Msg service
func _Msg_RegisterSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRegisterSchema)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RegisterSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cert.attestation.v1.Msg/RegisterSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RegisterSchema(ctx, req.(*MsgRegisterSchema))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Attest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAttest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Attest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cert.attestation.v1.Msg/Attest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Attest(ctx, req.(*MsgAttest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Revoke_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevoke)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Revoke(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cert.attestation.v1.Msg/Revoke",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Revoke(ctx, req.(*MsgRevoke))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateEncryptedAttestation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateEncryptedAttestation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateEncryptedAttestation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cert.attestation.v1.Msg/CreateEncryptedAttestation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateEncryptedAttestation(ctx, req.(*MsgCreateEncryptedAttestation))
	}
	return interceptor(ctx, in, info, handler)
}

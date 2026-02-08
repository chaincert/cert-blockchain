package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
)

// MsgServer is the server API for Msg service
type MsgServer interface {
	CreateProfile(context.Context, *MsgCreateProfile) (*MsgCreateProfileResponse, error)
	UpdateProfile(context.Context, *MsgUpdateProfile) (*MsgUpdateProfileResponse, error)
	AddCredential(context.Context, *MsgAddCredential) (*MsgAddCredentialResponse, error)
	RemoveCredential(context.Context, *MsgRemoveCredential) (*MsgRemoveCredentialResponse, error)
	VerifySocial(context.Context, *MsgVerifySocial) (*MsgVerifySocialResponse, error)
	RequestVerification(context.Context, *MsgRequestVerification) (*MsgRequestVerificationResponse, error)
	AwardBadge(context.Context, *MsgAwardBadge) (*MsgAwardBadgeResponse, error)
	RevokeBadge(context.Context, *MsgRevokeBadge) (*MsgRevokeBadgeResponse, error)
	UpdateTrustScore(context.Context, *MsgUpdateTrustScore) (*MsgUpdateTrustScoreResponse, error)
	SetVerificationStatus(context.Context, *MsgSetVerificationStatus) (*MsgSetVerificationStatusResponse, error)
	AuthorizeOracle(context.Context, *MsgAuthorizeOracle) (*MsgAuthorizeOracleResponse, error)
	RevokeOracle(context.Context, *MsgRevokeOracle) (*MsgRevokeOracleResponse, error)
	RegisterHandle(context.Context, *MsgRegisterHandle) (*MsgRegisterHandleResponse, error)
}

func RegisterMsgServer(s gogogrpc.Server, srv MsgServer) {
	if s == nil {
		return
	}
	s.RegisterService(&_Msg_serviceDesc, srv)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cert.certid.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateProfile", Handler: _Msg_CreateProfile_Handler},
		{MethodName: "UpdateProfile", Handler: _Msg_UpdateProfile_Handler},
		{MethodName: "AddCredential", Handler: _Msg_AddCredential_Handler},
		{MethodName: "RemoveCredential", Handler: _Msg_RemoveCredential_Handler},
		{MethodName: "VerifySocial", Handler: _Msg_VerifySocial_Handler},
		{MethodName: "RequestVerification", Handler: _Msg_RequestVerification_Handler},
		{MethodName: "AwardBadge", Handler: _Msg_AwardBadge_Handler},
		{MethodName: "RevokeBadge", Handler: _Msg_RevokeBadge_Handler},
		{MethodName: "UpdateTrustScore", Handler: _Msg_UpdateTrustScore_Handler},
		{MethodName: "SetVerificationStatus", Handler: _Msg_SetVerificationStatus_Handler},
		{MethodName: "AuthorizeOracle", Handler: _Msg_AuthorizeOracle_Handler},
		{MethodName: "RevokeOracle", Handler: _Msg_RevokeOracle_Handler},
		{MethodName: "RegisterHandle", Handler: _Msg_RegisterHandle_Handler},
	},
	Metadata: "cert/certid/v1/tx.proto",
}

func _Msg_CreateProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateProfile)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cert.certid.v1.Msg/CreateProfile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateProfile(ctx, req.(*MsgCreateProfile))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateProfile)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).UpdateProfile(ctx, in)
}
func _Msg_AddCredential_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAddCredential)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).AddCredential(ctx, in)
}
func _Msg_RemoveCredential_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveCredential)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).RemoveCredential(ctx, in)
}
func _Msg_VerifySocial_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgVerifySocial)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).VerifySocial(ctx, in)
}
func _Msg_RequestVerification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRequestVerification)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).RequestVerification(ctx, in)
}
func _Msg_AwardBadge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAwardBadge)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).AwardBadge(ctx, in)
}
func _Msg_RevokeBadge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevokeBadge)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).RevokeBadge(ctx, in)
}
func _Msg_UpdateTrustScore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateTrustScore)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).UpdateTrustScore(ctx, in)
}
func _Msg_SetVerificationStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSetVerificationStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).SetVerificationStatus(ctx, in)
}
func _Msg_AuthorizeOracle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAuthorizeOracle)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).AuthorizeOracle(ctx, in)
}
func _Msg_RevokeOracle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevokeOracle)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).RevokeOracle(ctx, in)
}
func _Msg_RegisterHandle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRegisterHandle)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(MsgServer).RegisterHandle(ctx, in)
}

type QueryServer interface {
	Profile(context.Context, *QueryProfileRequest) (*QueryProfileResponse, error)
	ProfileByHandle(context.Context, *QueryProfileByHandleRequest) (*QueryProfileByHandleResponse, error)
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
}

func RegisterQueryServer(s gogogrpc.Server, srv QueryServer) {
	if s == nil {
		return
	}
	s.RegisterService(&_Query_serviceDesc, srv)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cert.certid.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Profile", Handler: _Query_Profile_Handler},
		{MethodName: "ProfileByHandle", Handler: _Query_ProfileByHandle_Handler},
		{MethodName: "Params", Handler: _Query_Params_Handler},
	},
	Metadata: "cert/certid/v1/query.proto",
}

func _Query_Profile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	if interceptor == nil {
		return srv.(QueryServer).Profile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cert.certid.v1.Query/Profile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Profile(ctx, req.(*QueryProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Query_ProfileByHandle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProfileByHandleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(QueryServer).ProfileByHandle(ctx, in)
}
func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, nil
	}
	return srv.(QueryServer).Params(ctx, in)
}

type QueryClient interface {
	Profile(ctx context.Context, in *QueryProfileRequest, opts ...grpc.CallOption) (*QueryProfileResponse, error)
	ProfileByHandle(ctx context.Context, in *QueryProfileByHandleRequest, opts ...grpc.CallOption) (*QueryProfileByHandleResponse, error)
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryClient(cc grpc.ClientConnInterface) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Profile(ctx context.Context, in *QueryProfileRequest, opts ...grpc.CallOption) (*QueryProfileResponse, error) {
	out := new(QueryProfileResponse)
	err := c.cc.Invoke(ctx, "/cert.certid.v1.Query/Profile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
func (c *queryClient) ProfileByHandle(ctx context.Context, in *QueryProfileByHandleRequest, opts ...grpc.CallOption) (*QueryProfileByHandleResponse, error) {
	out := new(QueryProfileByHandleResponse)
	err := c.cc.Invoke(ctx, "/cert.certid.v1.Query/ProfileByHandle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/cert.certid.v1.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Codec & Interface registration
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateProfile{}, "certid/CreateProfile", nil)
	cdc.RegisterConcrete(&MsgUpdateProfile{}, "certid/UpdateProfile", nil)
	cdc.RegisterConcrete(&MsgAddCredential{}, "certid/AddCredential", nil)
	cdc.RegisterConcrete(&MsgRemoveCredential{}, "certid/RemoveCredential", nil)
	cdc.RegisterConcrete(&MsgVerifySocial{}, "certid/VerifySocial", nil)
	cdc.RegisterConcrete(&MsgRequestVerification{}, "certid/RequestVerification", nil)
	cdc.RegisterConcrete(&MsgAwardBadge{}, "certid/AwardBadge", nil)
	cdc.RegisterConcrete(&MsgRevokeBadge{}, "certid/RevokeBadge", nil)
	cdc.RegisterConcrete(&MsgUpdateTrustScore{}, "certid/UpdateTrustScore", nil)
	cdc.RegisterConcrete(&MsgSetVerificationStatus{}, "certid/SetVerificationStatus", nil)
	cdc.RegisterConcrete(&MsgAuthorizeOracle{}, "certid/AuthorizeOracle", nil)
	cdc.RegisterConcrete(&MsgRevokeOracle{}, "certid/RevokeOracle", nil)
	cdc.RegisterConcrete(&MsgRegisterHandle{}, "certid/RegisterHandle", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateProfile{},
		&MsgUpdateProfile{},
		&MsgAddCredential{},
		&MsgRemoveCredential{},
		&MsgVerifySocial{},
		&MsgRequestVerification{},
		&MsgAwardBadge{},
		&MsgRevokeBadge{},
		&MsgUpdateTrustScore{},
		&MsgSetVerificationStatus{},
		&MsgAuthorizeOracle{},
		&MsgRevokeOracle{},
		&MsgRegisterHandle{},
	)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
	Amino.Seal()
}

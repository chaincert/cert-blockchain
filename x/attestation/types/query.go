package types

import (
	"context"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
)

// QueryClient is the client API for Query service
type QueryClient interface {
	Schema(ctx context.Context, in *QuerySchemaRequest, opts ...grpc.CallOption) (*QuerySchemaResponse, error)
	Attestation(ctx context.Context, in *QueryAttestationRequest, opts ...grpc.CallOption) (*QueryAttestationResponse, error)
	AttestationsByAttester(ctx context.Context, in *QueryAttestationsByAttesterRequest, opts ...grpc.CallOption) (*QueryAttestationsByAttesterResponse, error)
	AttestationsByRecipient(ctx context.Context, in *QueryAttestationsByRecipientRequest, opts ...grpc.CallOption) (*QueryAttestationsByRecipientResponse, error)
	EncryptedAttestation(ctx context.Context, in *QueryEncryptedAttestationRequest, opts ...grpc.CallOption) (*QueryEncryptedAttestationResponse, error)
	Stats(ctx context.Context, in *QueryStatsRequest, opts ...grpc.CallOption) (*QueryStatsResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

// NewQueryClient creates a new Query client
func NewQueryClient(cc client.Context) QueryClient {
	return &queryClient{cc}
}

// Schema queries a schema by UID
func (c *queryClient) Schema(ctx context.Context, in *QuerySchemaRequest, opts ...grpc.CallOption) (*QuerySchemaResponse, error) {
	out := new(QuerySchemaResponse)
	err := c.cc.Invoke(ctx, "/cert.attestation.v1.Query/Schema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Attestation queries an attestation by UID
func (c *queryClient) Attestation(ctx context.Context, in *QueryAttestationRequest, opts ...grpc.CallOption) (*QueryAttestationResponse, error) {
	out := new(QueryAttestationResponse)
	err := c.cc.Invoke(ctx, "/cert.attestation.v1.Query/Attestation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AttestationsByAttester queries attestations by attester address
func (c *queryClient) AttestationsByAttester(ctx context.Context, in *QueryAttestationsByAttesterRequest, opts ...grpc.CallOption) (*QueryAttestationsByAttesterResponse, error) {
	out := new(QueryAttestationsByAttesterResponse)
	err := c.cc.Invoke(ctx, "/cert.attestation.v1.Query/AttestationsByAttester", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AttestationsByRecipient queries attestations by recipient address
func (c *queryClient) AttestationsByRecipient(ctx context.Context, in *QueryAttestationsByRecipientRequest, opts ...grpc.CallOption) (*QueryAttestationsByRecipientResponse, error) {
	out := new(QueryAttestationsByRecipientResponse)
	err := c.cc.Invoke(ctx, "/cert.attestation.v1.Query/AttestationsByRecipient", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EncryptedAttestation queries an encrypted attestation with access control
func (c *queryClient) EncryptedAttestation(ctx context.Context, in *QueryEncryptedAttestationRequest, opts ...grpc.CallOption) (*QueryEncryptedAttestationResponse, error) {
	out := new(QueryEncryptedAttestationResponse)
	err := c.cc.Invoke(ctx, "/cert.attestation.v1.Query/EncryptedAttestation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Stats queries attestation statistics
func (c *queryClient) Stats(ctx context.Context, in *QueryStatsRequest, opts ...grpc.CallOption) (*QueryStatsResponse, error) {
	out := new(QueryStatsResponse)
	err := c.cc.Invoke(ctx, "/cert.attestation.v1.Query/Stats", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

package types

import (
	"cosmossdk.io/errors"
)

// Hardware module sentinel errors
var (
	ErrInvalidDevice = errors.Register(ModuleName, 2, "invalid device")
	ErrDeviceNotFound = errors.Register(ModuleName, 3, "device not found")
	ErrDeviceAlreadyExists = errors.Register(ModuleName, 4, "device already exists")
	ErrInvalidAttestation = errors.Register(ModuleName, 5, "invalid attestation")
	ErrAttestationFailed = errors.Register(ModuleName, 6, "attestation verification failed")
	ErrUnsupportedTEE = errors.Register(ModuleName, 7, "unsupported TEE type")
	ErrUnauthorized = errors.Register(ModuleName, 8, "unauthorized")
	ErrDeviceSuspended = errors.Register(ModuleName, 9, "device is suspended")
	ErrInvalidAddress = errors.Register(ModuleName, 10, "invalid address")
	ErrDeviceNotOwned = errors.Register(ModuleName, 11, "device not owned by sender")
	ErrLinkFailed = errors.Register(ModuleName, 12, "failed to link device to CertID")
	ErrChallengeMismatch = errors.Register(ModuleName, 13, "challenge nonce mismatch")
	ErrChallengeExpired = errors.Register(ModuleName, 14, "challenge has expired")
)

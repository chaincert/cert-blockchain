package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// CertID module sentinel errors
var (
	ErrProfileNotFound           = sdkerrors.Register(ModuleName, 1, "profile not found")
	ErrProfileAlreadyExists      = sdkerrors.Register(ModuleName, 2, "profile already exists")
	ErrUnauthorized              = sdkerrors.Register(ModuleName, 3, "unauthorized")
	ErrInvalidName               = sdkerrors.Register(ModuleName, 4, "invalid name: must be 100 characters or less")
	ErrInvalidBio                = sdkerrors.Register(ModuleName, 5, "invalid bio: must be 500 characters or less")
	ErrInvalidAttestationUID     = sdkerrors.Register(ModuleName, 6, "invalid attestation UID")
	ErrCredentialNotFound        = sdkerrors.Register(ModuleName, 7, "credential not found")
	ErrCredentialAlreadyExists   = sdkerrors.Register(ModuleName, 8, "credential already exists")
	ErrInvalidSocialVerification = sdkerrors.Register(ModuleName, 9, "invalid social verification")
	ErrSocialAlreadyVerified     = sdkerrors.Register(ModuleName, 10, "social account already verified")
	ErrInvalidPublicKey          = sdkerrors.Register(ModuleName, 11, "invalid public key")
	ErrVerificationFailed        = sdkerrors.Register(ModuleName, 12, "verification failed")
	ErrInvalidAvatarCID          = sdkerrors.Register(ModuleName, 13, "invalid avatar CID")
	ErrInvalidUsername           = sdkerrors.Register(ModuleName, 14, "invalid username")
	ErrInvalidBadgeName          = sdkerrors.Register(ModuleName, 15, "invalid badge name")
	ErrBadgeAlreadyAwarded       = sdkerrors.Register(ModuleName, 16, "badge already awarded")
	ErrBadgeNotFound             = sdkerrors.Register(ModuleName, 17, "badge not found")
	ErrInvalidTrustScore         = sdkerrors.Register(ModuleName, 18, "invalid trust score: must be 0-100")
	ErrOracleNotAuthorized       = sdkerrors.Register(ModuleName, 19, "oracle not authorized")
	ErrHandleAlreadyTaken        = sdkerrors.Register(ModuleName, 20, "handle already taken")
	ErrProfileNotActive          = sdkerrors.Register(ModuleName, 21, "profile is not active")
)

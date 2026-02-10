package types

import (
	"cosmossdk.io/errors"
)

// TrustScore module sentinel errors
var (
	ErrInvalidAddress  = errors.Register(ModuleName, 2, "invalid address")
	ErrScoreNotFound   = errors.Register(ModuleName, 3, "trust score not found")
	ErrInvalidFactors  = errors.Register(ModuleName, 4, "invalid scoring factors")
	ErrUnauthorized    = errors.Register(ModuleName, 5, "unauthorized")
)

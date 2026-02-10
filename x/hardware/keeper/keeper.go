package keeper

import (
	"encoding/json"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/hardware/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
// for the hardware module's state.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	
	// Authority is the module authority address (for governance actions)
	authority string

	// CertIDKeeper for linking devices to CertID profiles
	certidKeeper types.CertIDKeeperI
}

// NewKeeper creates a new Hardware Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
	certidKeeper types.CertIDKeeperI,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		authority:    authority,
		certidKeeper: certidKeeper,
	}
}

// SetCertIDKeeper sets the CertID keeper (called after app initialization)
func (k *Keeper) SetCertIDKeeper(ck types.CertIDKeeperI) {
	k.certidKeeper = ck
}

// GetAuthority returns the module's authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// RegisterDevice registers a new hardware device after TEE verification
func (k Keeper) RegisterDevice(ctx sdk.Context, msg *types.MsgRegisterDevice) (*types.Device, error) {
	// Generate device ID from public key and TEE type
	deviceID := types.GenerateDeviceID(msg.PublicKey, msg.TEEType)

	// Check if device already exists
	store := ctx.KVStore(k.storeKey)
	deviceKey := types.GetDeviceKey(deviceID)
	if store.Has(deviceKey) {
		return nil, types.ErrDeviceAlreadyExists
	}

	// Verify initial attestation
	// TODO: Implement TEE-specific verification (TrustZone, SecureEnclave)
	verified, err := k.VerifyAttestation(ctx, deviceID, msg.TEEType, msg.InitialAttestation, nil)
	if err != nil || !verified {
		return nil, types.ErrAttestationFailed.Wrap("initial attestation verification failed")
	}

	// Create device
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	device := types.NewDevice(deviceID, msg.Manufacturer, msg.TEEType, msg.PublicKey, creator)
	device.Model = msg.Model
	device.AttestationCount = 1

	// Store device (using JSON encoding for now, will migrate to protobuf)
	bz, err := json.Marshal(device)
	if err != nil {
		return nil, types.ErrInvalidDevice.Wrap("failed to marshal device")
	}
	store.Set(deviceKey, bz)

	// Create owner -> device index
	indexKey := types.GetOwnerDeviceIndexKey(msg.Creator, deviceID)
	store.Set(indexKey, []byte{0x01})

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeviceRegistered,
			sdk.NewAttribute(types.AttributeKeyDeviceID, deviceID),
			sdk.NewAttribute(types.AttributeKeyManufacturer, msg.Manufacturer),
			sdk.NewAttribute(types.AttributeKeyTEEType, string(msg.TEEType)),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
		),
	)

	k.Logger(ctx).Info("device registered",
		"device_id", deviceID,
		"manufacturer", msg.Manufacturer,
		"tee_type", msg.TEEType,
		"owner", msg.Creator,
	)

	return device, nil
}

// GetDevice retrieves a device by ID
func (k Keeper) GetDevice(ctx sdk.Context, deviceID string) (*types.Device, error) {
	store := ctx.KVStore(k.storeKey)
	deviceKey := types.GetDeviceKey(deviceID)

	bz := store.Get(deviceKey)
	if bz == nil {
		return nil, types.ErrDeviceNotFound
	}

	var device types.Device
	if err := json.Unmarshal(bz, &device); err != nil {
		return nil, types.ErrInvalidDevice.Wrap("failed to unmarshal device")
	}
	return &device, nil
}

// VerifyAttestation verifies a TEE attestation
// This is the core anti-Sybil mechanism per Whitepaper Section 3.1
func (k Keeper) VerifyAttestation(
	ctx sdk.Context,
	deviceID string,
	teeType types.TEEType,
	attestationData []byte,
	nonce []byte,
) (bool, error) {
	// TEE-specific verification logic
	switch teeType {
	case types.TEETypeTrustZone:
		return k.verifyTrustZoneAttestation(ctx, attestationData, nonce)
	case types.TEETypeSecureEnclave:
		return k.verifySecureEnclaveAttestation(ctx, attestationData, nonce)
	default:
		return false, types.ErrUnsupportedTEE
	}
}

// verifyTrustZoneAttestation verifies ARM TrustZone attestation
// Per Whitepaper: "ARM TrustZone for IoT devices and Android"
func (k Keeper) verifyTrustZoneAttestation(ctx sdk.Context, attestationData, nonce []byte) (bool, error) {
	// TODO: Implement ARM TrustZone attestation verification
	// This will involve:
	// 1. Parse attestation token format
	// 2. Verify signature against ARM root certificate
	// 3. Validate nonce if provided
	// 4. Check attestation freshness
	
	k.Logger(ctx).Debug("verifying TrustZone attestation", "data_len", len(attestationData))
	
	// Placeholder - return true for initial implementation
	// Real verification will be added in Phase 1 completion
	return len(attestationData) > 0, nil
}

// verifySecureEnclaveAttestation verifies Apple Secure Enclave attestation
// Per Whitepaper: "Apple Secure Enclave for iOS devices"
func (k Keeper) verifySecureEnclaveAttestation(ctx sdk.Context, attestationData, nonce []byte) (bool, error) {
	// TODO: Implement Apple Secure Enclave verification
	// This will involve:
	// 1. Parse Apple App Attest / DeviceCheck token
	// 2. Verify against Apple's attestation CA
	// 3. Validate nonce (challenge)
	// 4. Check risk metric if available
	
	k.Logger(ctx).Debug("verifying SecureEnclave attestation", "data_len", len(attestationData))
	
	// Placeholder - return true for initial implementation
	return len(attestationData) > 0, nil
}

// GetDevicesByOwner returns all devices owned by an address
func (k Keeper) GetDevicesByOwner(ctx sdk.Context, owner string) []types.Device {
	store := ctx.KVStore(k.storeKey)
	prefix := append(types.OwnerDeviceIndexPrefix, []byte(owner)...)
	
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var devices []types.Device
	for ; iterator.Valid(); iterator.Next() {
		// Extract device ID from key
		key := iterator.Key()
		// Key format: OwnerDeviceIndexPrefix | owner | "/" | deviceID
		parts := splitKeyAtSlash(key[len(types.OwnerDeviceIndexPrefix):])
		if len(parts) < 2 {
			continue
		}
		deviceID := string(parts[1])
		
		device, err := k.GetDevice(ctx, deviceID)
		if err == nil {
			devices = append(devices, *device)
		}
	}

	return devices
}

// splitKeyAtSlash splits a byte slice at the first slash
func splitKeyAtSlash(key []byte) [][]byte {
	for i, b := range key {
		if b == '/' {
			return [][]byte{key[:i], key[i+1:]}
		}
	}
	return [][]byte{key}
}

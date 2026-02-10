// CertID Hardware Attestation Verifier — Arbitrum Stylus (Rust → WASM)
//
// This contract runs on Arbitrum L2 and provides gas-efficient hardware
// attestation verification using WASM instead of Solidity.
//
// Architecture:
//   Cosmos L1 (Consensus) → Bridge Relayer → Arbitrum L2 (Compute/Verify)
//
// Why Stylus?
//   TEE signature verification in Solidity costs 5M+ gas.
//   In Stylus (Rust → WASM), the same operation is ~10x cheaper.

#![cfg_attr(not(feature = "export-abi"), no_main)]
extern crate alloc;

use alloc::vec::Vec;
use stylus_sdk::{
    alloy_primitives::{Address, FixedBytes, U256},
    prelude::*,
    storage::{StorageAddress, StorageMap, StorageU256},
};

/// CertID Hardware Verifier Contract
///
/// Stores device registrations, trust scores, and attestation verification
/// results on Arbitrum L2. Acts as the compute layer for the CertID
/// cross-chain DePIN protocol.
#[storage]
#[entrypoint]
pub struct CertIDVerifier {
    /// Maps DeviceID (bytes32) → Trust Score (0-100)
    device_trust_scores: StorageMap<FixedBytes<32>, StorageU256>,

    /// Maps DeviceID → Owner Address
    device_owners: StorageMap<FixedBytes<32>, StorageAddress>,

    /// Total number of successful TEE verifications
    total_verifications: StorageU256,
}

#[public]
impl CertIDVerifier {
    /// Register a new device on Arbitrum L2.
    ///
    /// This mirrors device state from the Cosmos Hub via the bridge relayer.
    /// In production, this would be permissioned to the relayer address only.
    /// For the Grant Pilot, open registration demonstrates the flow.
    pub fn register_device(&mut self, device_id: FixedBytes<32>, owner: Address) {
        self.device_owners.setter(device_id).set(owner);
    }

    /// Update the Trust Score of a registered device.
    ///
    /// Called by the CertID Bridge Relayer after the Cosmos L1 recalculates
    /// the deterministic trust score (see x/hardware/keeper/scoring.go).
    pub fn update_trust_score(&mut self, device_id: FixedBytes<32>, new_score: U256) {
        self.device_trust_scores.setter(device_id).set(new_score);
    }

    /// Verify a TEE Attestation — the "Stylus Magic".
    ///
    /// This is the high-value operation that justifies using Stylus over
    /// Solidity. Cryptographic signature verification in WASM is ~10x
    /// cheaper than the equivalent EVM opcodes.
    ///
    /// Grant Pilot behavior:
    ///   - Checks if the device is registered and has a valid trust score
    ///   - Increments the global verification counter
    ///
    /// Production behavior (Phase 3):
    ///   - Verifies the `attestation_data` signature against the manufacturer's
    ///     public key (ARM TrustZone / Apple Secure Enclave)
    pub fn verify_tee_attestation(
        &mut self,
        device_id: FixedBytes<32>,
        _attestation_data: Vec<u8>,
    ) -> bool {
        // 1. Fetch current trust score
        let score = self.device_trust_scores.get(device_id);

        // 2. Grant Pilot: Check if device is registered with valid score
        //    Production: verify attestation_data signature here
        if score > U256::ZERO {
            // Increment global verification counter
            let current_count = self.total_verifications.get();
            self.total_verifications.set(current_count + U256::from(1));
            return true;
        }

        false
    }

    /// View: Get the trust score for a device
    pub fn get_device_trust(&self, device_id: FixedBytes<32>) -> U256 {
        self.device_trust_scores.get(device_id)
    }

    /// View: Get the owner address of a device
    pub fn get_device_owner(&self, device_id: FixedBytes<32>) -> Address {
        self.device_owners.get(device_id)
    }

    /// View: Get total successful attestation verifications
    pub fn get_total_verifications(&self) -> U256 {
        self.total_verifications.get()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use stylus_sdk::testing::*;

    #[test]
    fn test_register_and_verify() {
        let vm = TestVM::default();
        let mut contract = CertIDVerifier::from(&vm);

        let device_id = FixedBytes::<32>::ZERO;
        let owner = Address::ZERO;

        // Register device
        contract.register_device(device_id, owner).unwrap();

        // Set trust score to 92
        contract
            .update_trust_score(device_id, U256::from(92))
            .unwrap();

        // Verify attestation should succeed
        let result = contract
            .verify_tee_attestation(device_id, Vec::new())
            .unwrap();
        assert!(result);

        // Check counter incremented
        assert_eq!(
            contract.get_total_verifications().unwrap(),
            U256::from(1)
        );
    }

    #[test]
    fn test_unregistered_device_fails() {
        let vm = TestVM::default();
        let mut contract = CertIDVerifier::from(&vm);

        let device_id = FixedBytes::<32>::ZERO;

        // Verify without registration/score should fail
        let result = contract
            .verify_tee_attestation(device_id, Vec::new())
            .unwrap();
        assert!(!result);

        // Counter should not increment
        assert_eq!(
            contract.get_total_verifications().unwrap(),
            U256::ZERO
        );
    }
}

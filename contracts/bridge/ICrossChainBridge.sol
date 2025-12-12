// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title ICrossChainBridge
 * @notice Interface for cross-chain bridge operations
 * @dev Per Whitepaper Section 13 (Roadmap Phase 2: Ecosystem)
 * Enables secure transfer of CERT tokens and attestation data across chains
 */
interface ICrossChainBridge {
    /// @notice Emitted when tokens are locked for cross-chain transfer
    event TokensLocked(
        address indexed sender,
        uint256 amount,
        uint256 targetChainId,
        bytes32 indexed transferId
    );

    /// @notice Emitted when tokens are released from bridge
    event TokensReleased(
        address indexed recipient,
        uint256 amount,
        uint256 sourceChainId,
        bytes32 indexed transferId
    );

    /// @notice Emitted when attestation data is bridged
    event AttestationBridged(
        bytes32 indexed attestationUID,
        uint256 targetChainId,
        bytes32 indexed bridgeId
    );

    /// @notice Emitted when a bridge request is validated
    event BridgeValidated(bytes32 indexed transferId, address[] validators);

    /**
     * @notice Lock tokens on source chain for cross-chain transfer
     * @param amount Amount of CERT tokens to lock
     * @param targetChainId Destination chain ID
     * @param recipient Recipient address on target chain
     * @return transferId Unique identifier for the transfer
     */
    function lockTokens(
        uint256 amount,
        uint256 targetChainId,
        address recipient
    ) external returns (bytes32 transferId);

    /**
     * @notice Release tokens on target chain after validation
     * @param transferId Transfer identifier from source chain
     * @param recipient Recipient address
     * @param amount Amount to release
     * @param signatures Validator signatures
     */
    function releaseTokens(
        bytes32 transferId,
        address recipient,
        uint256 amount,
        bytes[] calldata signatures
    ) external;

    /**
     * @notice Bridge attestation metadata to another chain
     * @param attestationUID Attestation UID to bridge
     * @param targetChainId Destination chain ID
     * @param attestationData Encoded attestation metadata
     * @return bridgeId Unique bridge operation ID
     */
    function bridgeAttestation(
        bytes32 attestationUID,
        uint256 targetChainId,
        bytes calldata attestationData
    ) external returns (bytes32 bridgeId);

    /**
     * @notice Receive bridged attestation from another chain
     * @param bridgeId Bridge operation ID
     * @param sourceChainId Source chain ID
     * @param attestationData Encoded attestation metadata
     * @param signatures Validator signatures
     */
    function receiveAttestation(
        bytes32 bridgeId,
        uint256 sourceChainId,
        bytes calldata attestationData,
        bytes[] calldata signatures
    ) external;

    /**
     * @notice Get the minimum required signatures for validation
     * @return Minimum signature threshold
     */
    function getSignatureThreshold() external view returns (uint256);

    /**
     * @notice Check if a transfer has been processed
     * @param transferId Transfer identifier
     * @return True if processed
     */
    function isTransferProcessed(bytes32 transferId) external view returns (bool);

    /**
     * @notice Get supported chain IDs
     * @return Array of supported chain IDs
     */
    function getSupportedChains() external view returns (uint256[] memory);
}


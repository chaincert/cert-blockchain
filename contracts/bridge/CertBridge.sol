// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "./ICrossChainBridge.sol";

/**
 * @title CertBridge
 * @notice Cross-chain bridge for CERT tokens and attestation data
 * @dev Per Whitepaper Section 13 (Roadmap Phase 2: Ecosystem)
 */
contract CertBridge is ICrossChainBridge, AccessControl, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;
    using ECDSA for bytes32;

    bytes32 public constant VALIDATOR_ROLE = keccak256("VALIDATOR_ROLE");
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    IERC20 public immutable certToken;
    uint256 public immutable chainId;
    uint256 public signatureThreshold;
    mapping(uint256 => bool) public supportedChains;
    mapping(bytes32 => bool) public processedTransfers;
    uint256 private _nonce;
    uint256[] private _supportedChainIds;

    uint256 public constant MAX_TRANSFER_AMOUNT = 1_000_000 * 10**18;
    uint256 public constant MIN_TRANSFER_AMOUNT = 1 * 10**18;

    constructor(address _certToken, uint256 _signatureThreshold, uint256[] memory _supportedChains) {
        require(_certToken != address(0), "Invalid token");
        require(_signatureThreshold > 0, "Invalid threshold");
        certToken = IERC20(_certToken);
        chainId = block.chainid;
        signatureThreshold = _signatureThreshold;
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(OPERATOR_ROLE, msg.sender);
        for (uint256 i = 0; i < _supportedChains.length; i++) {
            supportedChains[_supportedChains[i]] = true;
            _supportedChainIds.push(_supportedChains[i]);
        }
    }

    function lockTokens(uint256 amount, uint256 targetChainId, address recipient) 
        external override nonReentrant whenNotPaused returns (bytes32 transferId) 
    {
        require(supportedChains[targetChainId], "Unsupported chain");
        require(amount >= MIN_TRANSFER_AMOUNT && amount <= MAX_TRANSFER_AMOUNT, "Invalid amount");
        require(recipient != address(0), "Invalid recipient");
        transferId = keccak256(abi.encodePacked(chainId, targetChainId, msg.sender, recipient, amount, block.timestamp, _nonce++));
        certToken.safeTransferFrom(msg.sender, address(this), amount);
        emit TokensLocked(msg.sender, amount, targetChainId, transferId);
    }

    function releaseTokens(bytes32 transferId, address recipient, uint256 amount, bytes[] calldata signatures)
        external override nonReentrant whenNotPaused
    {
        require(!processedTransfers[transferId], "Already processed");
        require(signatures.length >= signatureThreshold, "Insufficient signatures");
        bytes32 messageHash = keccak256(abi.encodePacked(transferId, chainId, recipient, amount));
        bytes32 ethSignedHash = messageHash.toEthSignedMessageHash();
        address[] memory signers = new address[](signatures.length);
        for (uint256 i = 0; i < signatures.length; i++) {
            address signer = ethSignedHash.recover(signatures[i]);
            require(hasRole(VALIDATOR_ROLE, signer), "Invalid validator");
            for (uint256 j = 0; j < i; j++) { require(signers[j] != signer, "Duplicate signer"); }
            signers[i] = signer;
        }
        processedTransfers[transferId] = true;
        certToken.safeTransfer(recipient, amount);
        emit TokensReleased(recipient, amount, 0, transferId);
        emit BridgeValidated(transferId, signers);
    }

    function bridgeAttestation(bytes32 attestationUID, uint256 targetChainId, bytes calldata attestationData)
        external override nonReentrant whenNotPaused returns (bytes32 bridgeId)
    {
        require(supportedChains[targetChainId], "Unsupported chain");
        require(attestationData.length > 0, "Empty data");
        bridgeId = keccak256(abi.encodePacked(attestationUID, chainId, targetChainId, block.timestamp, _nonce++));
        emit AttestationBridged(attestationUID, targetChainId, bridgeId);
    }

    function receiveAttestation(bytes32 bridgeId, uint256 sourceChainId, bytes calldata attestationData, bytes[] calldata signatures)
        external override nonReentrant whenNotPaused
    {
        require(!processedTransfers[bridgeId], "Already processed");
        require(signatures.length >= signatureThreshold, "Insufficient signatures");
        bytes32 messageHash = keccak256(abi.encodePacked(bridgeId, sourceChainId, chainId, attestationData));
        bytes32 ethSignedHash = messageHash.toEthSignedMessageHash();
        for (uint256 i = 0; i < signatures.length; i++) {
            require(hasRole(VALIDATOR_ROLE, ethSignedHash.recover(signatures[i])), "Invalid validator");
        }
        processedTransfers[bridgeId] = true;
    }

    function getSignatureThreshold() external view override returns (uint256) { return signatureThreshold; }
    function isTransferProcessed(bytes32 transferId) external view override returns (bool) { return processedTransfers[transferId]; }
    function getSupportedChains() external view override returns (uint256[] memory) { return _supportedChainIds; }

    function setSignatureThreshold(uint256 _threshold) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(_threshold > 0, "Invalid threshold");
        signatureThreshold = _threshold;
    }

    function addSupportedChain(uint256 _chainId) external onlyRole(OPERATOR_ROLE) {
        require(!supportedChains[_chainId], "Already supported");
        supportedChains[_chainId] = true;
        _supportedChainIds.push(_chainId);
    }

    function pause() external onlyRole(OPERATOR_ROLE) { _pause(); }
    function unpause() external onlyRole(OPERATOR_ROLE) { _unpause(); }

    function emergencyWithdraw(address token, uint256 amount) external onlyRole(DEFAULT_ADMIN_ROLE) {
        IERC20(token).safeTransfer(msg.sender, amount);
    }
}


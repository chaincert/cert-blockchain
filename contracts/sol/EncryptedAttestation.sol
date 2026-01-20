// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./EAS.sol";

/**
 * @title IEncryptedAttestation
 * @notice Interface for Encrypted Attestation functionality
 * @dev Per CERT Whitepaper Section 3.2, 3.3, 3.4
 */
interface IEncryptedAttestation {
    struct EncryptedAttestationData {
        string ipfsCID;
        bytes32 encryptedDataHash;
        address[] recipients;
        bytes[] encryptedSymmetricKeys;
        bool revocable;
        uint64 expirationTime;
    }

    event EncryptedAttestationCreated(
        bytes32 indexed uid,
        address indexed attester,
        string ipfsCID,
        uint256 recipientCount
    );

    event RecipientAdded(
        bytes32 indexed uid,
        address indexed recipient
    );

    event RecipientRemoved(
        bytes32 indexed uid,
        address indexed recipient
    );

    function createEncryptedAttestation(
        bytes32 schemaUID,
        EncryptedAttestationData calldata data
    ) external returns (bytes32);

    function getEncryptedAttestation(bytes32 uid) external view returns (EncryptedAttestationData memory);
    function isAuthorizedRecipient(bytes32 uid, address recipient) external view returns (bool);
    function getEncryptedKey(bytes32 uid, address recipient) external view returns (bytes memory);
}

/**
 * @title EncryptedAttestation
 * @notice Implementation of the Encrypted Attestation System
 * @dev Core privacy feature per Whitepaper Section 3
 * 
 * This contract enables:
 * - On-chain anchoring of encrypted data stored on IPFS
 * - Access control for encrypted symmetric keys
 * - Multi-recipient support (up to 50 per Whitepaper Section 12)
 */
contract EncryptedAttestation is IEncryptedAttestation {
    IEAS public immutable eas;
    
    // Maximum recipients per attestation (Whitepaper Section 12)
    uint256 public constant MAX_RECIPIENTS = 50;
    
    // Encrypted attestation storage
    struct StoredEncryptedAttestation {
        bytes32 easUID;
        string ipfsCID;
        bytes32 encryptedDataHash;
        address attester;
        uint64 creationTime;
        uint64 expirationTime;
        bool revocable;
        bool revoked;
    }
    
    mapping(bytes32 => StoredEncryptedAttestation) private _attestations;
    mapping(bytes32 => mapping(address => bool)) private _authorizedRecipients;
    mapping(bytes32 => mapping(address => bytes)) private _encryptedKeys;
    mapping(bytes32 => address[]) private _recipientsList;
    
    // Index by IPFS CID
    mapping(string => bytes32) private _cidToUID;

    constructor(IEAS _eas) {
        eas = _eas;
    }

    /**
     * @notice Create an encrypted attestation
     * @dev Implements Step 4 of Whitepaper Section 3.2 - On-Chain Anchoring
     * @param schemaUID The schema to use for the attestation
     * @param data The encrypted attestation data
     * @return uid The unique identifier of the created attestation
     */
    function createEncryptedAttestation(
        bytes32 schemaUID,
        EncryptedAttestationData calldata data
    ) external override returns (bytes32) {
        require(data.recipients.length > 0, "At least one recipient required");
        require(data.recipients.length <= MAX_RECIPIENTS, "Too many recipients");
        require(data.recipients.length == data.encryptedSymmetricKeys.length, "Key count mismatch");
        require(bytes(data.ipfsCID).length >= 46, "Invalid IPFS CID");
        require(data.encryptedDataHash != bytes32(0), "Invalid data hash");

        // Generate unique ID
        bytes32 uid = keccak256(abi.encodePacked(
            schemaUID,
            msg.sender,
            block.timestamp,
            data.encryptedDataHash
        ));

        require(_attestations[uid].creationTime == 0, "Attestation exists");
        require(_cidToUID[data.ipfsCID] == bytes32(0), "CID already used");

        // Store attestation
        _attestations[uid] = StoredEncryptedAttestation({
            easUID: bytes32(0), // Will be set if EAS attestation is created
            ipfsCID: data.ipfsCID,
            encryptedDataHash: data.encryptedDataHash,
            attester: msg.sender,
            creationTime: uint64(block.timestamp),
            expirationTime: data.expirationTime,
            revocable: data.revocable,
            revoked: false
        });

        // Store CID index
        _cidToUID[data.ipfsCID] = uid;

        // Store recipients and their encrypted keys
        for (uint256 i = 0; i < data.recipients.length; i++) {
            address recipient = data.recipients[i];
            require(recipient != address(0), "Invalid recipient");
            
            _authorizedRecipients[uid][recipient] = true;
            _encryptedKeys[uid][recipient] = data.encryptedSymmetricKeys[i];
            _recipientsList[uid].push(recipient);

            emit RecipientAdded(uid, recipient);
        }

        emit EncryptedAttestationCreated(uid, msg.sender, data.ipfsCID, data.recipients.length);

        return uid;
    }

    /**
     * @notice Get encrypted attestation metadata
     * @param uid The attestation UID
     * @return data The encrypted attestation data (without keys)
     */
    function getEncryptedAttestation(
        bytes32 uid
    ) external view override returns (EncryptedAttestationData memory data) {
        StoredEncryptedAttestation storage stored = _attestations[uid];
        require(stored.creationTime != 0, "Attestation not found");

        data.ipfsCID = stored.ipfsCID;
        data.encryptedDataHash = stored.encryptedDataHash;
        data.revocable = stored.revocable;
        data.expirationTime = stored.expirationTime;
        data.recipients = _recipientsList[uid];
        // Note: encryptedSymmetricKeys not returned here for security
    }

    /**
     * @notice Check if an address is authorized to access an attestation
     * @dev Per Whitepaper Section 3.3 - Access Control
     * @param uid The attestation UID
     * @param recipient The address to check
     * @return authorized Whether the recipient is authorized
     */
    function isAuthorizedRecipient(
        bytes32 uid,
        address recipient
    ) external view override returns (bool) {
        StoredEncryptedAttestation storage stored = _attestations[uid];
        if (stored.creationTime == 0) return false;
        if (stored.revoked) return false;
        if (stored.expirationTime != 0 && stored.expirationTime < block.timestamp) return false;
        
        return _authorizedRecipients[uid][recipient] || stored.attester == recipient;
    }

    /**
     * @notice Get the encrypted symmetric key for a recipient
     * @dev Only authorized recipients can retrieve their key
     * @param uid The attestation UID
     * @param recipient The recipient address (must be msg.sender or authorized)
     * @return key The encrypted symmetric key
     */
    function getEncryptedKey(
        bytes32 uid,
        address recipient
    ) external view override returns (bytes memory) {
        require(
            msg.sender == recipient || 
            msg.sender == _attestations[uid].attester,
            "Not authorized"
        );
        require(_authorizedRecipients[uid][recipient], "Recipient not authorized");
        
        return _encryptedKeys[uid][recipient];
    }
}


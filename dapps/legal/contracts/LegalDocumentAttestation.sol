// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title LegalDocumentAttestation
 * @notice Confidential document sharing for legal agreements
 * @dev Per Whitepaper Section 11 - Use Case: Legal Documents
 * Demonstrates the EncryptedMultiRecipientAttestation schema
 */
contract LegalDocumentAttestation is AccessControl, ReentrancyGuard, Pausable {
    /// @notice Role for legal administrators
    bytes32 public constant LEGAL_ADMIN_ROLE = keccak256("LEGAL_ADMIN_ROLE");
    /// @notice Role for notaries
    bytes32 public constant NOTARY_ROLE = keccak256("NOTARY_ROLE");

    /// @notice Document types
    bytes32 public constant DOC_TYPE_CONTRACT = keccak256("CONTRACT");
    bytes32 public constant DOC_TYPE_NDA = keccak256("NDA");
    bytes32 public constant DOC_TYPE_MERGER = keccak256("MERGER");
    bytes32 public constant DOC_TYPE_COURT_FILING = keccak256("COURT_FILING");

    /// @notice Legal document structure
    struct LegalDocument {
        bytes32 attestationUID;
        bytes32 documentType;
        address creator;
        address[] parties;
        uint256 createdAt;
        uint256 effectiveDate;
        uint256 expirationDate;
        bool isRevoked;
        bytes32 ipfsCID;
        bytes32 documentHash;
        mapping(address => bool) signatures;
        uint256 signatureCount;
        uint256 requiredSignatures;
    }

    /// @notice Signature status for a party
    struct SignatureStatus {
        bool hasSigned;
        uint256 signedAt;
        bytes signature;
    }

    /// @notice Document UID to document
    mapping(bytes32 => LegalDocument) public documents;
    /// @notice Document signatures
    mapping(bytes32 => mapping(address => SignatureStatus)) public documentSignatures;
    /// @notice Party address to their document UIDs
    mapping(address => bytes32[]) public partyDocuments;

    /// @notice Maximum parties per document (per Whitepaper Section 12)
    uint256 public constant MAX_PARTIES = 50;

    event DocumentCreated(bytes32 indexed uid, bytes32 documentType, address indexed creator, address[] parties);
    event DocumentSigned(bytes32 indexed uid, address indexed signer, uint256 timestamp);
    event DocumentFullySigned(bytes32 indexed uid, uint256 timestamp);
    event DocumentRevoked(bytes32 indexed uid, address indexed revokedBy);
    event PartyAdded(bytes32 indexed uid, address indexed party);

    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(LEGAL_ADMIN_ROLE, msg.sender);
    }

    /**
     * @notice Create a new legal document attestation
     * @param attestationUID Reference to encrypted attestation
     * @param documentType Type of legal document
     * @param parties All parties involved in the document
     * @param effectiveDate When document becomes effective
     * @param expirationDate When document expires (0 for no expiration)
     * @param ipfsCID IPFS CID of encrypted document
     * @param documentHash Hash for integrity verification
     * @param requiredSignatures Number of signatures needed
     */
    function createDocument(
        bytes32 attestationUID,
        bytes32 documentType,
        address[] calldata parties,
        uint256 effectiveDate,
        uint256 expirationDate,
        bytes32 ipfsCID,
        bytes32 documentHash,
        uint256 requiredSignatures
    ) external nonReentrant whenNotPaused {
        require(parties.length > 0 && parties.length <= MAX_PARTIES, "Invalid party count");
        require(documents[attestationUID].createdAt == 0, "Document exists");
        require(requiredSignatures > 0 && requiredSignatures <= parties.length, "Invalid signature requirement");

        LegalDocument storage doc = documents[attestationUID];
        doc.attestationUID = attestationUID;
        doc.documentType = documentType;
        doc.creator = msg.sender;
        doc.parties = parties;
        doc.createdAt = block.timestamp;
        doc.effectiveDate = effectiveDate;
        doc.expirationDate = expirationDate;
        doc.ipfsCID = ipfsCID;
        doc.documentHash = documentHash;
        doc.requiredSignatures = requiredSignatures;

        // Track documents for each party
        for (uint256 i = 0; i < parties.length; i++) {
            partyDocuments[parties[i]].push(attestationUID);
        }

        emit DocumentCreated(attestationUID, documentType, msg.sender, parties);
    }

    /**
     * @notice Sign a legal document
     * @param attestationUID Document to sign
     * @param signature Cryptographic signature
     */
    function signDocument(bytes32 attestationUID, bytes calldata signature) external nonReentrant whenNotPaused {
        LegalDocument storage doc = documents[attestationUID];
        require(doc.createdAt > 0, "Document not found");
        require(!doc.isRevoked, "Document revoked");
        require(!doc.signatures[msg.sender], "Already signed");
        require(_isParty(attestationUID, msg.sender), "Not a party");

        doc.signatures[msg.sender] = true;
        doc.signatureCount++;

        documentSignatures[attestationUID][msg.sender] = SignatureStatus({
            hasSigned: true,
            signedAt: block.timestamp,
            signature: signature
        });

        emit DocumentSigned(attestationUID, msg.sender, block.timestamp);

        if (doc.signatureCount >= doc.requiredSignatures) {
            emit DocumentFullySigned(attestationUID, block.timestamp);
        }
    }

    /**
     * @notice Check if document has all required signatures
     */
    function isFullySigned(bytes32 attestationUID) external view returns (bool) {
        LegalDocument storage doc = documents[attestationUID];
        return doc.signatureCount >= doc.requiredSignatures;
    }

    /**
     * @notice Revoke a document (creator or admin only)
     */
    function revokeDocument(bytes32 attestationUID) external nonReentrant {
        LegalDocument storage doc = documents[attestationUID];
        require(doc.createdAt > 0, "Document not found");
        require(msg.sender == doc.creator || hasRole(LEGAL_ADMIN_ROLE, msg.sender), "Unauthorized");
        
        doc.isRevoked = true;
        emit DocumentRevoked(attestationUID, msg.sender);
    }

    function _isParty(bytes32 attestationUID, address addr) internal view returns (bool) {
        address[] storage parties = documents[attestationUID].parties;
        for (uint256 i = 0; i < parties.length; i++) {
            if (parties[i] == addr) return true;
        }
        return false;
    }

    function getDocumentParties(bytes32 attestationUID) external view returns (address[] memory) {
        return documents[attestationUID].parties;
    }

    function getPartyDocuments(address party) external view returns (bytes32[] memory) {
        return partyDocuments[party];
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) { _pause(); }
    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) { _unpause(); }
}


// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title HealthcareAttestation
 * @notice HIPAA-compliant medical record attestation system
 * @dev Per Whitepaper Section 11 - Use Case: Healthcare Records
 * Demonstrates the EncryptedFileAttestation schema for medical data
 */
contract HealthcareAttestation is AccessControl, ReentrancyGuard, Pausable {
    /// @notice Role for healthcare providers
    bytes32 public constant PROVIDER_ROLE = keccak256("PROVIDER_ROLE");
    /// @notice Role for auditors (compliance)
    bytes32 public constant AUDITOR_ROLE = keccak256("AUDITOR_ROLE");

    /// @notice Medical record attestation structure
    struct MedicalRecord {
        bytes32 attestationUID;        // Reference to encrypted attestation
        address patient;               // Patient wallet address
        address provider;              // Healthcare provider
        bytes32 recordType;            // Type of record (lab, imaging, etc.)
        uint256 timestamp;             // When record was created
        bool isRevoked;                // Revocation status
        bytes32 ipfsCID;               // IPFS CID of encrypted data
        bytes32 dataHash;              // Hash for integrity verification
    }

    /// @notice Patient consent for data sharing
    struct Consent {
        address patient;
        address authorizedParty;
        uint256 expiresAt;
        bool isActive;
        bytes32[] recordTypes;         // Types of records authorized
    }

    /// @notice Mapping of attestation UID to medical record
    mapping(bytes32 => MedicalRecord) public records;
    /// @notice Patient address to their record UIDs
    mapping(address => bytes32[]) public patientRecords;
    /// @notice Consent key to consent details
    mapping(bytes32 => Consent) public consents;

    /// @notice Events for audit trail
    event RecordCreated(bytes32 indexed uid, address indexed patient, address indexed provider, bytes32 recordType);
    event RecordRevoked(bytes32 indexed uid, address indexed revokedBy);
    event ConsentGranted(address indexed patient, address indexed authorizedParty, uint256 expiresAt);
    event ConsentRevoked(address indexed patient, address indexed authorizedParty);
    event RecordAccessed(bytes32 indexed uid, address indexed accessedBy, uint256 timestamp);

    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
    }

    /**
     * @notice Create a medical record attestation
     * @param attestationUID The encrypted attestation UID from the main EAS contract
     * @param patient Patient's wallet address
     * @param recordType Type of medical record
     * @param ipfsCID IPFS CID where encrypted data is stored
     * @param dataHash Hash of encrypted data for integrity
     */
    function createRecord(
        bytes32 attestationUID,
        address patient,
        bytes32 recordType,
        bytes32 ipfsCID,
        bytes32 dataHash
    ) external onlyRole(PROVIDER_ROLE) nonReentrant whenNotPaused {
        require(patient != address(0), "Invalid patient address");
        require(records[attestationUID].timestamp == 0, "Record already exists");

        records[attestationUID] = MedicalRecord({
            attestationUID: attestationUID,
            patient: patient,
            provider: msg.sender,
            recordType: recordType,
            timestamp: block.timestamp,
            isRevoked: false,
            ipfsCID: ipfsCID,
            dataHash: dataHash
        });

        patientRecords[patient].push(attestationUID);

        emit RecordCreated(attestationUID, patient, msg.sender, recordType);
    }

    /**
     * @notice Grant consent for data sharing
     * @param authorizedParty Address authorized to access records
     * @param duration How long consent is valid (seconds)
     * @param recordTypes Types of records to authorize
     */
    function grantConsent(
        address authorizedParty,
        uint256 duration,
        bytes32[] calldata recordTypes
    ) external nonReentrant {
        require(authorizedParty != address(0), "Invalid authorized party");
        require(duration > 0 && duration <= 365 days, "Invalid duration");

        bytes32 consentKey = keccak256(abi.encodePacked(msg.sender, authorizedParty));
        
        consents[consentKey] = Consent({
            patient: msg.sender,
            authorizedParty: authorizedParty,
            expiresAt: block.timestamp + duration,
            isActive: true,
            recordTypes: recordTypes
        });

        emit ConsentGranted(msg.sender, authorizedParty, block.timestamp + duration);
    }

    /**
     * @notice Revoke previously granted consent
     * @param authorizedParty Address to revoke consent from
     */
    function revokeConsent(address authorizedParty) external nonReentrant {
        bytes32 consentKey = keccak256(abi.encodePacked(msg.sender, authorizedParty));
        require(consents[consentKey].isActive, "No active consent");

        consents[consentKey].isActive = false;
        emit ConsentRevoked(msg.sender, authorizedParty);
    }

    /**
     * @notice Check if an address has valid consent to access patient records
     * @param patient Patient address
     * @param accessor Address attempting access
     * @param recordType Type of record being accessed
     */
    function hasValidConsent(address patient, address accessor, bytes32 recordType) public view returns (bool) {
        if (patient == accessor) return true;  // Patients can always access their own records
        
        bytes32 consentKey = keccak256(abi.encodePacked(patient, accessor));
        Consent storage consent = consents[consentKey];
        
        if (!consent.isActive || consent.expiresAt < block.timestamp) return false;
        
        for (uint i = 0; i < consent.recordTypes.length; i++) {
            if (consent.recordTypes[i] == recordType || consent.recordTypes[i] == bytes32(0)) {
                return true;
            }
        }
        return false;
    }

    /**
     * @notice Log record access for HIPAA compliance
     * @param attestationUID Record being accessed
     */
    function logAccess(bytes32 attestationUID) external {
        MedicalRecord storage record = records[attestationUID];
        require(record.timestamp > 0, "Record not found");
        require(hasValidConsent(record.patient, msg.sender, record.recordType), "No consent");

        emit RecordAccessed(attestationUID, msg.sender, block.timestamp);
    }

    function getPatientRecords(address patient) external view returns (bytes32[] memory) {
        return patientRecords[patient];
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) { _pause(); }
    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) { _unpause(); }
}


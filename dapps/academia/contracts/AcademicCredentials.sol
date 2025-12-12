// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title AcademicCredentials
 * @notice Academic credential attestation system
 * @dev Per Whitepaper Section 11 - Use Case: Academic Records
 * Demonstrates both public diplomas and private encrypted transcripts
 */
contract AcademicCredentials is AccessControl, ReentrancyGuard, Pausable {
    /// @notice Role for educational institutions
    bytes32 public constant INSTITUTION_ROLE = keccak256("INSTITUTION_ROLE");
    /// @notice Role for accreditation bodies
    bytes32 public constant ACCREDITOR_ROLE = keccak256("ACCREDITOR_ROLE");

    /// @notice Credential types
    bytes32 public constant DIPLOMA = keccak256("DIPLOMA");
    bytes32 public constant TRANSCRIPT = keccak256("TRANSCRIPT");
    bytes32 public constant CERTIFICATE = keccak256("CERTIFICATE");
    bytes32 public constant DEGREE = keccak256("DEGREE");

    /// @notice Public credential (diploma, certificate)
    struct PublicCredential {
        bytes32 attestationUID;
        address student;
        address institution;
        bytes32 credentialType;
        string credentialName;         // e.g., "Bachelor of Science"
        string fieldOfStudy;           // e.g., "Computer Science"
        uint256 conferredDate;
        uint256 createdAt;
        bool isRevoked;
        bytes32 institutionSignature;
    }

    /// @notice Private credential (transcript with grades)
    struct PrivateCredential {
        bytes32 attestationUID;
        address student;
        address institution;
        bytes32 credentialType;
        uint256 createdAt;
        bool isRevoked;
        bytes32 ipfsCID;               // IPFS CID for encrypted data
        bytes32 dataHash;              // Hash for integrity verification
    }

    /// @notice Access grant for employers to view transcripts
    struct AccessGrant {
        address employer;
        uint256 expiresAt;
        bool isActive;
    }

    /// @notice Public credentials by UID
    mapping(bytes32 => PublicCredential) public publicCredentials;
    /// @notice Private credentials by UID
    mapping(bytes32 => PrivateCredential) public privateCredentials;
    /// @notice Student to their credential UIDs
    mapping(address => bytes32[]) public studentPublicCredentials;
    mapping(address => bytes32[]) public studentPrivateCredentials;
    /// @notice Access grants: student => employer => grant
    mapping(address => mapping(address => AccessGrant)) public accessGrants;
    /// @notice Verified institutions
    mapping(address => bool) public verifiedInstitutions;

    event PublicCredentialCreated(bytes32 indexed uid, address indexed student, address indexed institution, string credentialName);
    event PrivateCredentialCreated(bytes32 indexed uid, address indexed student, address indexed institution);
    event CredentialRevoked(bytes32 indexed uid, address indexed revokedBy);
    event AccessGranted(address indexed student, address indexed employer, uint256 expiresAt);
    event AccessRevoked(address indexed student, address indexed employer);
    event InstitutionVerified(address indexed institution, address indexed verifiedBy);

    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
    }

    /**
     * @notice Create a public credential (diploma/certificate)
     * @param attestationUID EAS attestation UID
     * @param student Student wallet address
     * @param credentialType Type of credential
     * @param credentialName Name of the credential
     * @param fieldOfStudy Field of study
     * @param conferredDate Date credential was conferred
     * @param signature Institution signature
     */
    function createPublicCredential(
        bytes32 attestationUID,
        address student,
        bytes32 credentialType,
        string calldata credentialName,
        string calldata fieldOfStudy,
        uint256 conferredDate,
        bytes32 signature
    ) external onlyRole(INSTITUTION_ROLE) nonReentrant whenNotPaused {
        require(student != address(0), "Invalid student");
        require(publicCredentials[attestationUID].createdAt == 0, "Credential exists");

        publicCredentials[attestationUID] = PublicCredential({
            attestationUID: attestationUID,
            student: student,
            institution: msg.sender,
            credentialType: credentialType,
            credentialName: credentialName,
            fieldOfStudy: fieldOfStudy,
            conferredDate: conferredDate,
            createdAt: block.timestamp,
            isRevoked: false,
            institutionSignature: signature
        });

        studentPublicCredentials[student].push(attestationUID);
        emit PublicCredentialCreated(attestationUID, student, msg.sender, credentialName);
    }

    /**
     * @notice Create a private credential (encrypted transcript)
     * @param attestationUID Encrypted attestation UID
     * @param student Student wallet address
     * @param ipfsCID IPFS CID of encrypted transcript
     * @param dataHash Hash for integrity verification
     */
    function createPrivateCredential(
        bytes32 attestationUID,
        address student,
        bytes32 ipfsCID,
        bytes32 dataHash
    ) external onlyRole(INSTITUTION_ROLE) nonReentrant whenNotPaused {
        require(student != address(0), "Invalid student");
        require(privateCredentials[attestationUID].createdAt == 0, "Credential exists");

        privateCredentials[attestationUID] = PrivateCredential({
            attestationUID: attestationUID,
            student: student,
            institution: msg.sender,
            credentialType: TRANSCRIPT,
            createdAt: block.timestamp,
            isRevoked: false,
            ipfsCID: ipfsCID,
            dataHash: dataHash
        });

        studentPrivateCredentials[student].push(attestationUID);
        emit PrivateCredentialCreated(attestationUID, student, msg.sender);
    }

    /**
     * @notice Grant employer access to transcripts
     */
    function grantAccess(address employer, uint256 durationDays) external nonReentrant {
        require(employer != address(0), "Invalid employer");
        require(durationDays > 0 && durationDays <= 365, "Invalid duration");

        accessGrants[msg.sender][employer] = AccessGrant({
            employer: employer,
            expiresAt: block.timestamp + (durationDays * 1 days),
            isActive: true
        });

        emit AccessGranted(msg.sender, employer, block.timestamp + (durationDays * 1 days));
    }

    /**
     * @notice Revoke employer access
     */
    function revokeAccess(address employer) external nonReentrant {
        require(accessGrants[msg.sender][employer].isActive, "No active grant");
        accessGrants[msg.sender][employer].isActive = false;
        emit AccessRevoked(msg.sender, employer);
    }

    /**
     * @notice Check if employer has valid access
     */
    function hasAccess(address student, address employer) external view returns (bool) {
        AccessGrant storage grant = accessGrants[student][employer];
        return grant.isActive && grant.expiresAt > block.timestamp;
    }

    function verifyInstitution(address institution) external onlyRole(ACCREDITOR_ROLE) {
        verifiedInstitutions[institution] = true;
        _grantRole(INSTITUTION_ROLE, institution);
        emit InstitutionVerified(institution, msg.sender);
    }

    function getStudentPublicCredentials(address student) external view returns (bytes32[] memory) {
        return studentPublicCredentials[student];
    }

    function getStudentPrivateCredentials(address student) external view returns (bytes32[] memory) {
        return studentPrivateCredentials[student];
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) { _pause(); }
    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) { _unpause(); }
}


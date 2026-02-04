// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title CertID
 * @notice Decentralized Identity Registry with Soulbound Token (SBT) badges
 * @dev Per CERT Whitepaper - Identity layer for Cert Blockchain ecosystem
 */
contract CertID is Ownable, ReentrancyGuard {
    
    // Entity types for profile categorization
    enum EntityType { Individual, Institution, SystemAdmin, Bot }

    // Profile structure for identity data
    struct Profile {
        string handle;           // e.g., "university.cert"
        string metadataURI;      // IPFS link to logo/bio/social links
        EntityType entityType;   // Categorization of entity
        uint256 trustScore;      // Dynamic reputation score (0-100)
        bool isVerified;         // High-level verification status
        bool isActive;           // Profile active status
        uint256 createdAt;       // Profile creation timestamp
        uint256 updatedAt;       // Last update timestamp
    }

    // Mappings for profile and handle resolution
    mapping(address => Profile) private _profiles;
    mapping(string => address) private _handleToAddress;
    mapping(address => mapping(bytes32 => bool)) private _badges;
    mapping(address => bool) public authorizedOracles;

    // Standard badge identifiers
    bytes32 public constant BADGE_KYC_L1 = keccak256("KYC_L1");
    bytes32 public constant BADGE_KYC_L2 = keccak256("KYC_L2");
    bytes32 public constant BADGE_ACADEMIC = keccak256("ACADEMIC_ISSUER");
    bytes32 public constant BADGE_CREATOR = keccak256("VERIFIED_CREATOR");
    bytes32 public constant BADGE_GOV = keccak256("GOV_AGENCY");
    bytes32 public constant BADGE_LEGAL = keccak256("LEGAL_ENTITY");
    bytes32 public constant BADGE_ISO9001 = keccak256("ISO_9001_CERTIFIED");

    // Events
    event ProfileCreated(address indexed user, string handle);
    event ProfileUpdated(address indexed user, string handle);
    event BadgeAwarded(address indexed user, bytes32 indexed badgeId, string badgeName);
    event BadgeRevoked(address indexed user, bytes32 indexed badgeId);
    event TrustScoreUpdated(address indexed user, uint256 oldScore, uint256 newScore);
    event VerificationStatusChanged(address indexed user, bool isVerified);
    event OracleAuthorized(address indexed oracle);
    event OracleRevoked(address indexed oracle);

    // Modifiers
    modifier onlyAuthorized() {
        require(msg.sender == owner() || authorizedOracles[msg.sender], "CertID: Not authorized");
        _;
    }

    modifier profileExists(address user) {
        require(_profiles[user].isActive, "CertID: Profile does not exist");
        _;
    }

    constructor() Ownable(msg.sender) {}

    // ============ Profile Management ============

    /**
     * @notice Register a new CertID profile
     * @param handle Unique handle (e.g., "alice.cert")
     * @param metadataURI IPFS URI for extended metadata
     * @param entityType Type of entity (Individual, Institution, etc.)
     */
    function registerProfile(
        string memory handle,
        string memory metadataURI,
        EntityType entityType
    ) external nonReentrant {
        require(!_profiles[msg.sender].isActive, "CertID: Profile already exists");
        require(bytes(handle).length > 0, "CertID: Handle cannot be empty");
        require(_handleToAddress[handle] == address(0), "CertID: Handle already taken");

        _profiles[msg.sender] = Profile({
            handle: handle,
            metadataURI: metadataURI,
            entityType: entityType,
            trustScore: 0,
            isVerified: false,
            isActive: true,
            createdAt: block.timestamp,
            updatedAt: block.timestamp
        });

        _handleToAddress[handle] = msg.sender;
        emit ProfileCreated(msg.sender, handle);
    }

    /**
     * @notice Update profile metadata
     * @param metadataURI New IPFS URI for metadata
     */
    function updateMetadata(string memory metadataURI) external profileExists(msg.sender) {
        _profiles[msg.sender].metadataURI = metadataURI;
        _profiles[msg.sender].updatedAt = block.timestamp;
        emit ProfileUpdated(msg.sender, _profiles[msg.sender].handle);
    }

    // ============ Badge Management (Soulbound Tokens) ============

    /**
     * @notice Award a verification badge to a user (SBT - non-transferable)
     * @param user Address to receive the badge
     * @param badgeName Human-readable badge name
     */
    function awardBadge(address user, string memory badgeName) external onlyAuthorized profileExists(user) {
        bytes32 badgeId = keccak256(abi.encodePacked(badgeName));
        require(!_badges[user][badgeId], "CertID: Badge already awarded");
        
        _badges[user][badgeId] = true;
        emit BadgeAwarded(user, badgeId, badgeName);
    }

    /**
     * @notice Revoke a badge from a user
     * @param user Address to revoke badge from
     * @param badgeName Badge name to revoke
     */
    function revokeBadge(address user, string memory badgeName) external onlyAuthorized {
        bytes32 badgeId = keccak256(abi.encodePacked(badgeName));
        require(_badges[user][badgeId], "CertID: Badge not found");
        
        _badges[user][badgeId] = false;
        emit BadgeRevoked(user, badgeId);
    }

    // ============ Trust Score Management ============

    /**
     * @notice Update a user's trust score
     * @param user Address to update
     * @param score New trust score (0-100)
     */
    function updateTrustScore(address user, uint256 score) external onlyAuthorized profileExists(user) {
        require(score <= 100, "CertID: Score must be <= 100");
        uint256 oldScore = _profiles[user].trustScore;
        _profiles[user].trustScore = score;
        _profiles[user].updatedAt = block.timestamp;
        emit TrustScoreUpdated(user, oldScore, score);
    }

    /**
     * @notice Increment trust score (called by Chain Certify on successful attestations)
     * @param user Address to increment
     * @param amount Amount to increment
     */
    function incrementTrustScore(address user, uint256 amount) external onlyAuthorized profileExists(user) {
        uint256 oldScore = _profiles[user].trustScore;
        uint256 newScore = oldScore + amount;
        if (newScore > 100) newScore = 100;
        _profiles[user].trustScore = newScore;
        _profiles[user].updatedAt = block.timestamp;
        emit TrustScoreUpdated(user, oldScore, newScore);
    }

    // ============ Verification Status ============

    /**
     * @notice Set verification status for a profile
     * @param user Address to verify/unverify
     * @param verified Verification status
     */
    function setVerificationStatus(address user, bool verified) external onlyAuthorized profileExists(user) {
        _profiles[user].isVerified = verified;
        _profiles[user].updatedAt = block.timestamp;
        emit VerificationStatusChanged(user, verified);
    }

    // ============ Oracle Management ============

    function authorizeOracle(address oracle) external onlyOwner {
        authorizedOracles[oracle] = true;
        emit OracleAuthorized(oracle);
    }

    function revokeOracle(address oracle) external onlyOwner {
        authorizedOracles[oracle] = false;
        emit OracleRevoked(oracle);
    }

    // ============ View Functions ============

    function getProfile(address user) external view returns (
        string memory handle,
        string memory metadataURI,
        bool isVerified,
        uint256 trustScore,
        EntityType entityType,
        bool isActive
    ) {
        Profile storage p = _profiles[user];
        return (p.handle, p.metadataURI, p.isVerified, p.trustScore, p.entityType, p.isActive);
    }

    function hasBadge(address user, string memory badgeName) external view returns (bool) {
        return _badges[user][keccak256(abi.encodePacked(badgeName))];
    }

    function hasBadgeById(address user, bytes32 badgeId) external view returns (bool) {
        return _badges[user][badgeId];
    }

    function getHandle(address user) external view returns (string memory) {
        return _profiles[user].handle;
    }

    function resolveHandle(string memory handle) external view returns (address) {
        return _handleToAddress[handle];
    }

    function isProfileActive(address user) external view returns (bool) {
        return _profiles[user].isActive;
    }

    function getTrustScore(address user) external view returns (uint256) {
        return _profiles[user].trustScore;
    }
}


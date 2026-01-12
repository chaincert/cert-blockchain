// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

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

    constructor() Ownable() {}

    // ============ Profile Management ============

    /**
     * @notice Register a new CertID profile
     * @param _handle Unique handle (e.g., "alice.cert")
     * @param _metadataURI IPFS URI for extended metadata
     * @param _entityType Type of entity (Individual, Institution, etc.)
     */
    function registerProfile(
        string memory _handle,
        string memory _metadataURI,
        EntityType _entityType
    ) external nonReentrant {
        require(!_profiles[msg.sender].isActive, "CertID: Profile already exists");
        require(bytes(_handle).length > 0, "CertID: Handle cannot be empty");
        require(_handleToAddress[_handle] == address(0), "CertID: Handle already taken");

        _profiles[msg.sender] = Profile({
            handle: _handle,
            metadataURI: _metadataURI,
            entityType: _entityType,
            trustScore: 0,
            isVerified: false,
            isActive: true,
            createdAt: block.timestamp,
            updatedAt: block.timestamp
        });

        _handleToAddress[_handle] = msg.sender;
        emit ProfileCreated(msg.sender, _handle);
    }

    /**
     * @notice Update profile metadata
     * @param _metadataURI New IPFS URI for metadata
     */
    function updateMetadata(string memory _metadataURI) external profileExists(msg.sender) {
        _profiles[msg.sender].metadataURI = _metadataURI;
        _profiles[msg.sender].updatedAt = block.timestamp;
        emit ProfileUpdated(msg.sender, _profiles[msg.sender].handle);
    }

    // ============ Badge Management (Soulbound Tokens) ============

    /**
     * @notice Award a verification badge to a user (SBT - non-transferable)
     * @param _user Address to receive the badge
     * @param _badgeName Human-readable badge name
     */
    function awardBadge(address _user, string memory _badgeName) external onlyAuthorized profileExists(_user) {
        bytes32 badgeId = keccak256(abi.encodePacked(_badgeName));
        require(!_badges[_user][badgeId], "CertID: Badge already awarded");
        
        _badges[_user][badgeId] = true;
        emit BadgeAwarded(_user, badgeId, _badgeName);
    }

    /**
     * @notice Revoke a badge from a user
     * @param _user Address to revoke badge from
     * @param _badgeName Badge name to revoke
     */
    function revokeBadge(address _user, string memory _badgeName) external onlyAuthorized {
        bytes32 badgeId = keccak256(abi.encodePacked(_badgeName));
        require(_badges[_user][badgeId], "CertID: Badge not found");
        
        _badges[_user][badgeId] = false;
        emit BadgeRevoked(_user, badgeId);
    }

    // ============ Trust Score Management ============

    /**
     * @notice Update a user's trust score
     * @param _user Address to update
     * @param _score New trust score (0-100)
     */
    function updateTrustScore(address _user, uint256 _score) external onlyAuthorized profileExists(_user) {
        require(_score <= 100, "CertID: Score must be <= 100");
        uint256 oldScore = _profiles[_user].trustScore;
        _profiles[_user].trustScore = _score;
        _profiles[_user].updatedAt = block.timestamp;
        emit TrustScoreUpdated(_user, oldScore, _score);
    }

    /**
     * @notice Increment trust score (called by Chain Certify on successful attestations)
     * @param _user Address to increment
     * @param _amount Amount to increment
     */
    function incrementTrustScore(address _user, uint256 _amount) external onlyAuthorized profileExists(_user) {
        uint256 oldScore = _profiles[_user].trustScore;
        uint256 newScore = oldScore + _amount;
        if (newScore > 100) newScore = 100;
        _profiles[_user].trustScore = newScore;
        _profiles[_user].updatedAt = block.timestamp;
        emit TrustScoreUpdated(_user, oldScore, newScore);
    }

    // ============ Verification Status ============

    /**
     * @notice Set verification status for a profile
     * @param _user Address to verify/unverify
     * @param _verified Verification status
     */
    function setVerificationStatus(address _user, bool _verified) external onlyAuthorized profileExists(_user) {
        _profiles[_user].isVerified = _verified;
        _profiles[_user].updatedAt = block.timestamp;
        emit VerificationStatusChanged(_user, _verified);
    }

    // ============ Oracle Management ============

    function authorizeOracle(address _oracle) external onlyOwner {
        authorizedOracles[_oracle] = true;
        emit OracleAuthorized(_oracle);
    }

    function revokeOracle(address _oracle) external onlyOwner {
        authorizedOracles[_oracle] = false;
        emit OracleRevoked(_oracle);
    }

    // ============ View Functions ============

    function getProfile(address _user) external view returns (
        string memory handle,
        string memory metadataURI,
        bool isVerified,
        uint256 trustScore,
        EntityType entityType,
        bool isActive
    ) {
        Profile storage p = _profiles[_user];
        return (p.handle, p.metadataURI, p.isVerified, p.trustScore, p.entityType, p.isActive);
    }

    function hasBadge(address _user, string memory _badgeName) external view returns (bool) {
        return _badges[_user][keccak256(abi.encodePacked(_badgeName))];
    }

    function hasBadgeById(address _user, bytes32 _badgeId) external view returns (bool) {
        return _badges[_user][_badgeId];
    }

    function getHandle(address _user) external view returns (string memory) {
        return _profiles[_user].handle;
    }

    function resolveHandle(string memory _handle) external view returns (address) {
        return _handleToAddress[_handle];
    }

    function isProfileActive(address _user) external view returns (bool) {
        return _profiles[_user].isActive;
    }

    function getTrustScore(address _user) external view returns (uint256) {
        return _profiles[_user].trustScore;
    }
}


// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title IEAS
 * @notice Interface for the Ethereum Attestation Service
 * @dev Per CERT Whitepaper Section 2.2 and 3
 */
interface IEAS {
    struct AttestationRequest {
        bytes32 schema;
        AttestationRequestData data;
    }

    struct AttestationRequestData {
        address recipient;
        uint64 expirationTime;
        bool revocable;
        bytes32 refUID;
        bytes data;
        uint256 value;
    }

    struct Attestation {
        bytes32 uid;
        bytes32 schema;
        uint64 time;
        uint64 expirationTime;
        uint64 revocationTime;
        bytes32 refUID;
        address recipient;
        address attester;
        bool revocable;
        bytes data;
    }

    event Attested(
        address indexed recipient,
        address indexed attester,
        bytes32 uid,
        bytes32 indexed schemaUID
    );

    event Revoked(
        address indexed recipient,
        address indexed attester,
        bytes32 uid,
        bytes32 indexed schemaUID
    );

    function attest(AttestationRequest calldata request) external payable returns (bytes32);
    function revoke(bytes32 uid) external;
    function getAttestation(bytes32 uid) external view returns (Attestation memory);
    function isAttestationValid(bytes32 uid) external view returns (bool);
}

/**
 * @title ISchemaRegistry
 * @notice Interface for the Schema Registry
 */
interface ISchemaRegistry {
    struct SchemaRecord {
        bytes32 uid;
        address resolver;
        bool revocable;
        string schema;
    }

    event Registered(bytes32 indexed uid, address indexed registerer);

    function register(string calldata schema, address resolver, bool revocable) external returns (bytes32);
    function getSchema(bytes32 uid) external view returns (SchemaRecord memory);
}

/**
 * @title SchemaRegistry
 * @notice Registry for attestation schemas per EAS standard
 * @dev Deployed at genesis per Whitepaper Section 2.2
 */
contract SchemaRegistry is ISchemaRegistry {
    mapping(bytes32 => SchemaRecord) private _schemas;

    function register(
        string calldata schema,
        address resolver,
        bool revocable
    ) external override returns (bytes32) {
        bytes32 uid = _getUID(schema, resolver, revocable);
        
        require(_schemas[uid].uid == bytes32(0), "Schema already exists");
        
        _schemas[uid] = SchemaRecord({
            uid: uid,
            resolver: resolver,
            revocable: revocable,
            schema: schema
        });

        emit Registered(uid, msg.sender);
        return uid;
    }

    function getSchema(bytes32 uid) external view override returns (SchemaRecord memory) {
        return _schemas[uid];
    }

    function _getUID(
        string calldata schema,
        address resolver,
        bool revocable
    ) private pure returns (bytes32) {
        return keccak256(abi.encodePacked(schema, resolver, revocable));
    }
}

/**
 * @title EAS
 * @notice Ethereum Attestation Service implementation for CERT Blockchain
 * @dev Core protocol per Whitepaper Section 2.2, 3
 */
contract EAS is IEAS {
    ISchemaRegistry public immutable schemaRegistry;
    
    mapping(bytes32 => Attestation) private _attestations;
    uint64 private _attestationCount;

    constructor(ISchemaRegistry _schemaRegistry) {
        schemaRegistry = _schemaRegistry;
    }

    function attest(
        AttestationRequest calldata request
    ) external payable override returns (bytes32) {
        ISchemaRegistry.SchemaRecord memory schema = schemaRegistry.getSchema(request.schema);
        require(schema.uid != bytes32(0), "Schema not found");
        
        if (!schema.revocable) {
            require(!request.data.revocable, "Schema does not allow revocable attestations");
        }

        bytes32 uid = _getUID(
            request.schema,
            request.data.recipient,
            msg.sender,
            block.timestamp,
            request.data.data,
            _attestationCount
        );

        _attestations[uid] = Attestation({
            uid: uid,
            schema: request.schema,
            time: uint64(block.timestamp),
            expirationTime: request.data.expirationTime,
            revocationTime: 0,
            refUID: request.data.refUID,
            recipient: request.data.recipient,
            attester: msg.sender,
            revocable: request.data.revocable,
            data: request.data.data
        });

        _attestationCount++;

        emit Attested(request.data.recipient, msg.sender, uid, request.schema);
        return uid;
    }

    function revoke(bytes32 uid) external override {
        Attestation storage attestation = _attestations[uid];
        require(attestation.uid != bytes32(0), "Attestation not found");
        require(attestation.attester == msg.sender, "Only attester can revoke");
        require(attestation.revocable, "Attestation not revocable");
        require(attestation.revocationTime == 0, "Already revoked");

        attestation.revocationTime = uint64(block.timestamp);

        emit Revoked(attestation.recipient, msg.sender, uid, attestation.schema);
    }

    function getAttestation(bytes32 uid) external view override returns (Attestation memory) {
        return _attestations[uid];
    }

    function isAttestationValid(bytes32 uid) external view override returns (bool) {
        Attestation memory attestation = _attestations[uid];
        if (attestation.uid == bytes32(0)) return false;
        if (attestation.revocationTime != 0) return false;
        if (attestation.expirationTime != 0 && attestation.expirationTime < block.timestamp) return false;
        return true;
    }

    function _getUID(
        bytes32 schema,
        address recipient,
        address attester,
        uint256 time,
        bytes memory data,
        uint64 nonce
    ) private pure returns (bytes32) {
        return keccak256(abi.encodePacked(schema, recipient, attester, time, data, nonce));
    }
}


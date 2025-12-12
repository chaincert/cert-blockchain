// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/governance/Governor.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorCountingSimple.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotes.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotesQuorumFraction.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorTimelockControl.sol";

/**
 * @title CertGovernance
 * @notice On-chain governance for CERT network parameters
 * @dev Per Whitepaper Section 5.2 and 6.2 - Community Governance
 * Allows CERT token holders to propose and vote on network changes
 */
contract CertGovernance is
    Governor,
    GovernorSettings,
    GovernorCountingSimple,
    GovernorVotes,
    GovernorVotesQuorumFraction,
    GovernorTimelockControl
{
    /// @notice Proposal types for network parameter changes
    enum ProposalType {
        PARAMETER_CHANGE,      // Network parameters (gas limit, block time)
        VALIDATOR_CHANGE,      // Validator set changes
        TREASURY_SPENDING,     // Treasury fund allocation
        PROTOCOL_UPGRADE,      // Protocol version upgrades
        EMERGENCY_ACTION       // Emergency governance actions
    }

    /// @notice Network parameter proposal
    struct ParameterProposal {
        string parameterName;
        uint256 currentValue;
        uint256 proposedValue;
        string rationale;
    }

    /// @notice Mapping of proposal ID to parameter details
    mapping(uint256 => ParameterProposal) public parameterProposals;

    /// @notice Events for governance actions
    event ParameterProposalCreated(
        uint256 indexed proposalId,
        string parameterName,
        uint256 currentValue,
        uint256 proposedValue
    );
    event ParameterUpdated(string parameterName, uint256 oldValue, uint256 newValue);

    /**
     * @notice Initialize governance with CERT token
     * @param _token CERT token with voting power
     * @param _timelock Timelock controller for execution delay
     * @param _votingDelay Delay before voting starts (blocks)
     * @param _votingPeriod Duration of voting (blocks)
     * @param _proposalThreshold Minimum tokens to create proposal
     * @param _quorumPercentage Percentage of supply needed for quorum
     */
    constructor(
        IVotes _token,
        TimelockController _timelock,
        uint48 _votingDelay,
        uint32 _votingPeriod,
        uint256 _proposalThreshold,
        uint256 _quorumPercentage
    )
        Governor("CERT Governance")
        GovernorSettings(_votingDelay, _votingPeriod, _proposalThreshold)
        GovernorVotes(_token)
        GovernorVotesQuorumFraction(_quorumPercentage)
        GovernorTimelockControl(_timelock)
    {}

    /**
     * @notice Create a parameter change proposal
     * @param targets Contract addresses to call
     * @param values ETH values for calls
     * @param calldatas Function call data
     * @param description Proposal description
     * @param parameterName Name of parameter to change
     * @param currentValue Current parameter value
     * @param proposedValue Proposed new value
     * @param rationale Reason for the change
     */
    function proposeParameterChange(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description,
        string memory parameterName,
        uint256 currentValue,
        uint256 proposedValue,
        string memory rationale
    ) external returns (uint256) {
        uint256 proposalId = propose(targets, values, calldatas, description);
        
        parameterProposals[proposalId] = ParameterProposal({
            parameterName: parameterName,
            currentValue: currentValue,
            proposedValue: proposedValue,
            rationale: rationale
        });

        emit ParameterProposalCreated(proposalId, parameterName, currentValue, proposedValue);
        return proposalId;
    }

    /**
     * @notice Get parameter proposal details
     */
    function getParameterProposal(uint256 proposalId) external view returns (ParameterProposal memory) {
        return parameterProposals[proposalId];
    }

    // Required overrides for Governor extensions
    function votingDelay() public view override(Governor, GovernorSettings) returns (uint256) {
        return super.votingDelay();
    }

    function votingPeriod() public view override(Governor, GovernorSettings) returns (uint256) {
        return super.votingPeriod();
    }

    function quorum(uint256 blockNumber) public view override(Governor, GovernorVotesQuorumFraction) returns (uint256) {
        return super.quorum(blockNumber);
    }

    function state(uint256 proposalId) public view override(Governor, GovernorTimelockControl) returns (ProposalState) {
        return super.state(proposalId);
    }

    function proposalNeedsQueuing(uint256 proposalId) public view override(Governor, GovernorTimelockControl) returns (bool) {
        return super.proposalNeedsQueuing(proposalId);
    }

    function proposalThreshold() public view override(Governor, GovernorSettings) returns (uint256) {
        return super.proposalThreshold();
    }

    function _queueOperations(uint256 proposalId, address[] memory targets, uint256[] memory values, bytes[] memory calldatas, bytes32 descriptionHash)
        internal override(Governor, GovernorTimelockControl) returns (uint48)
    {
        return super._queueOperations(proposalId, targets, values, calldatas, descriptionHash);
    }

    function _executeOperations(uint256 proposalId, address[] memory targets, uint256[] memory values, bytes[] memory calldatas, bytes32 descriptionHash)
        internal override(Governor, GovernorTimelockControl)
    {
        super._executeOperations(proposalId, targets, values, calldatas, descriptionHash);
    }

    function _cancel(address[] memory targets, uint256[] memory values, bytes[] memory calldatas, bytes32 descriptionHash)
        internal override(Governor, GovernorTimelockControl) returns (uint256)
    {
        return super._cancel(targets, values, calldatas, descriptionHash);
    }

    function _executor() internal view override(Governor, GovernorTimelockControl) returns (address) {
        return super._executor();
    }
}


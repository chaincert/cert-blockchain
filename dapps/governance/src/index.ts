/**
 * Governance Dashboard - Network Parameter Management
 * Per Whitepaper Section 5.2 and 6.2 - On-Chain Governance
 * Allows CERT holders to view, propose, and vote on network changes
 */

import { ethers } from 'ethers';

// Governance contract ABI (essential functions)
const GOVERNANCE_ABI = [
  'function propose(address[],uint256[],bytes[],string) returns (uint256)',
  'function proposeParameterChange(address[],uint256[],bytes[],string,string,uint256,uint256,string) returns (uint256)',
  'function castVote(uint256,uint8) returns (uint256)',
  'function castVoteWithReason(uint256,uint8,string) returns (uint256)',
  'function queue(address[],uint256[],bytes[],bytes32) returns (uint256)',
  'function execute(address[],uint256[],bytes[],bytes32) returns (uint256)',
  'function state(uint256) view returns (uint8)',
  'function proposalVotes(uint256) view returns (uint256,uint256,uint256)',
  'function getParameterProposal(uint256) view returns (tuple(string,uint256,uint256,string))',
  'function proposalThreshold() view returns (uint256)',
  'function quorum(uint256) view returns (uint256)',
  'function votingDelay() view returns (uint256)',
  'function votingPeriod() view returns (uint256)',
];

// Proposal states
export enum ProposalState {
  Pending = 0,
  Active = 1,
  Canceled = 2,
  Defeated = 3,
  Succeeded = 4,
  Queued = 5,
  Expired = 6,
  Executed = 7,
}

// Vote types
export enum VoteType {
  Against = 0,
  For = 1,
  Abstain = 2,
}

// Network parameters that can be governed
export const GOVERNABLE_PARAMETERS = {
  MAX_GAS_PER_BLOCK: { name: 'maxGasPerBlock', current: 30_000_000, unit: 'gas' },
  BLOCK_TIME: { name: 'blockTime', current: 2, unit: 'seconds' },
  MAX_VALIDATORS: { name: 'maxValidators', current: 80, unit: 'count' },
  UNBONDING_PERIOD: { name: 'unbondingPeriod', current: 21, unit: 'days' },
  SLASHING_DOWNTIME: { name: 'slashingDowntime', current: 0.01, unit: 'percent' },
  SLASHING_DOUBLE_SIGN: { name: 'slashingDoubleSign', current: 5, unit: 'percent' },
  MIN_GAS_PRICE: { name: 'minGasPrice', current: 0.0001, unit: 'CERT' },
} as const;

export interface Proposal {
  id: string;
  proposer: string;
  targets: string[];
  values: bigint[];
  calldatas: string[];
  description: string;
  state: ProposalState;
  votes: {
    against: bigint;
    for: bigint;
    abstain: bigint;
  };
  parameterChange?: {
    parameterName: string;
    currentValue: bigint;
    proposedValue: bigint;
    rationale: string;
  };
}

/**
 * Governance Dashboard client
 */
export class GovernanceDashboard {
  private contract: ethers.Contract;

  constructor(
    governanceAddress: string,
    private signer: ethers.Signer
  ) {
    this.contract = new ethers.Contract(governanceAddress, GOVERNANCE_ABI, signer);
  }

  /**
   * Create a parameter change proposal
   */
  async proposeParameterChange(
    parameterName: keyof typeof GOVERNABLE_PARAMETERS,
    proposedValue: number,
    rationale: string,
    targetContract: string,
    calldata: string
  ): Promise<string> {
    const param = GOVERNABLE_PARAMETERS[parameterName];
    const description = `Change ${param.name} from ${param.current} to ${proposedValue} ${param.unit}`;

    const tx = await this.contract.proposeParameterChange(
      [targetContract],
      [0],
      [calldata],
      description,
      param.name,
      param.current,
      proposedValue,
      rationale
    );
    const receipt = await tx.wait();

    // Extract proposal ID from events
    const event = receipt.logs.find((log: ethers.Log) => 
      log.topics[0] === ethers.id('ProposalCreated(uint256,address,address[],uint256[],string[],bytes[],uint256,uint256,string)')
    );
    
    return event ? event.topics[1] : receipt.hash;
  }

  /**
   * Cast a vote on a proposal
   */
  async vote(proposalId: string, support: VoteType, reason?: string): Promise<string> {
    const tx = reason 
      ? await this.contract.castVoteWithReason(proposalId, support, reason)
      : await this.contract.castVote(proposalId, support);
    const receipt = await tx.wait();
    return receipt.hash;
  }

  /**
   * Queue a successful proposal for execution
   */
  async queueProposal(
    targets: string[],
    values: bigint[],
    calldatas: string[],
    descriptionHash: string
  ): Promise<string> {
    const tx = await this.contract.queue(targets, values, calldatas, descriptionHash);
    const receipt = await tx.wait();
    return receipt.hash;
  }

  /**
   * Execute a queued proposal
   */
  async executeProposal(
    targets: string[],
    values: bigint[],
    calldatas: string[],
    descriptionHash: string
  ): Promise<string> {
    const tx = await this.contract.execute(targets, values, calldatas, descriptionHash);
    const receipt = await tx.wait();
    return receipt.hash;
  }

  /**
   * Get proposal state
   */
  async getProposalState(proposalId: string): Promise<ProposalState> {
    return await this.contract.state(proposalId);
  }

  /**
   * Get proposal votes
   */
  async getProposalVotes(proposalId: string): Promise<{ against: bigint; for: bigint; abstain: bigint }> {
    const [against, forVotes, abstain] = await this.contract.proposalVotes(proposalId);
    return { against, for: forVotes, abstain };
  }

  /**
   * Get governance parameters
   */
  async getGovernanceParams(): Promise<{
    proposalThreshold: bigint;
    votingDelay: bigint;
    votingPeriod: bigint;
  }> {
    const [threshold, delay, period] = await Promise.all([
      this.contract.proposalThreshold(),
      this.contract.votingDelay(),
      this.contract.votingPeriod(),
    ]);
    return { proposalThreshold: threshold, votingDelay: delay, votingPeriod: period };
  }
}

export default GovernanceDashboard;


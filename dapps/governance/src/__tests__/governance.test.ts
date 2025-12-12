/**
 * Tests for Governance Dashboard SDK
 */

import {
  GovernanceDashboard,
  ProposalState,
  VoteType,
  GOVERNABLE_PARAMETERS,
  Proposal
} from '../index';

describe('Governance dApp Constants', () => {
  describe('GOVERNABLE_PARAMETERS', () => {
    it('should have all required parameters', () => {
      expect(GOVERNABLE_PARAMETERS.MAX_GAS_PER_BLOCK).toBeDefined();
      expect(GOVERNABLE_PARAMETERS.BLOCK_TIME).toBeDefined();
      expect(GOVERNABLE_PARAMETERS.MAX_VALIDATORS).toBeDefined();
      expect(GOVERNABLE_PARAMETERS.UNBONDING_PERIOD).toBeDefined();
      expect(GOVERNABLE_PARAMETERS.SLASHING_DOWNTIME).toBeDefined();
      expect(GOVERNABLE_PARAMETERS.SLASHING_DOUBLE_SIGN).toBeDefined();
      expect(GOVERNABLE_PARAMETERS.MIN_GAS_PRICE).toBeDefined();
    });

    it('should have correct current values per whitepaper', () => {
      expect(GOVERNABLE_PARAMETERS.MAX_GAS_PER_BLOCK.current).toBe(30_000_000);
      expect(GOVERNABLE_PARAMETERS.BLOCK_TIME.current).toBe(2);
      expect(GOVERNABLE_PARAMETERS.MAX_VALIDATORS.current).toBe(80);
      expect(GOVERNABLE_PARAMETERS.UNBONDING_PERIOD.current).toBe(21);
    });

    it('should have proper units for each parameter', () => {
      expect(GOVERNABLE_PARAMETERS.MAX_GAS_PER_BLOCK.unit).toBe('gas');
      expect(GOVERNABLE_PARAMETERS.BLOCK_TIME.unit).toBe('seconds');
      expect(GOVERNABLE_PARAMETERS.MAX_VALIDATORS.unit).toBe('count');
      expect(GOVERNABLE_PARAMETERS.UNBONDING_PERIOD.unit).toBe('days');
      expect(GOVERNABLE_PARAMETERS.SLASHING_DOWNTIME.unit).toBe('percent');
      expect(GOVERNABLE_PARAMETERS.SLASHING_DOUBLE_SIGN.unit).toBe('percent');
      expect(GOVERNABLE_PARAMETERS.MIN_GAS_PRICE.unit).toBe('CERT');
    });

    it('should have proper names for each parameter', () => {
      expect(GOVERNABLE_PARAMETERS.MAX_GAS_PER_BLOCK.name).toBe('maxGasPerBlock');
      expect(GOVERNABLE_PARAMETERS.BLOCK_TIME.name).toBe('blockTime');
      expect(GOVERNABLE_PARAMETERS.MAX_VALIDATORS.name).toBe('maxValidators');
      expect(GOVERNABLE_PARAMETERS.UNBONDING_PERIOD.name).toBe('unbondingPeriod');
    });
  });
});

describe('ProposalState enum', () => {
  it('should have all states in correct order', () => {
    expect(ProposalState.Pending).toBe(0);
    expect(ProposalState.Active).toBe(1);
    expect(ProposalState.Canceled).toBe(2);
    expect(ProposalState.Defeated).toBe(3);
    expect(ProposalState.Succeeded).toBe(4);
    expect(ProposalState.Queued).toBe(5);
    expect(ProposalState.Expired).toBe(6);
    expect(ProposalState.Executed).toBe(7);
  });

  it('should have 8 total states', () => {
    const stateCount = Object.keys(ProposalState).filter(k => isNaN(Number(k))).length;
    expect(stateCount).toBe(8);
  });
});

describe('VoteType enum', () => {
  it('should have all vote types', () => {
    expect(VoteType.Against).toBe(0);
    expect(VoteType.For).toBe(1);
    expect(VoteType.Abstain).toBe(2);
  });

  it('should have 3 total vote types', () => {
    const voteTypeCount = Object.keys(VoteType).filter(k => isNaN(Number(k))).length;
    expect(voteTypeCount).toBe(3);
  });
});

describe('Proposal interface', () => {
  it('should accept valid proposal data', () => {
    const proposal: Proposal = {
      id: '0x1234567890abcdef',
      proposer: '0x1234567890123456789012345678901234567890',
      targets: ['0x0987654321098765432109876543210987654321'],
      values: [0n],
      calldatas: ['0x'],
      description: 'Increase max gas per block to 35M',
      state: ProposalState.Active,
      votes: {
        against: 100n,
        for: 500n,
        abstain: 50n,
      },
    };

    expect(proposal.id).toBeDefined();
    expect(proposal.proposer).toBeDefined();
    expect(proposal.state).toBe(ProposalState.Active);
    expect(proposal.votes.for).toBe(500n);
  });

  it('should accept proposal with parameter change', () => {
    const proposal: Proposal = {
      id: '0xabcdef1234567890',
      proposer: '0x1234567890123456789012345678901234567890',
      targets: ['0x0987654321098765432109876543210987654321'],
      values: [0n],
      calldatas: ['0x'],
      description: 'Change block time',
      state: ProposalState.Pending,
      votes: {
        against: 0n,
        for: 0n,
        abstain: 0n,
      },
      parameterChange: {
        parameterName: 'blockTime',
        currentValue: 2n,
        proposedValue: 3n,
        rationale: 'Improve network stability',
      },
    };

    expect(proposal.parameterChange).toBeDefined();
    expect(proposal.parameterChange?.parameterName).toBe('blockTime');
    expect(proposal.parameterChange?.currentValue).toBe(2n);
    expect(proposal.parameterChange?.proposedValue).toBe(3n);
  });
});

describe('GovernanceDashboard class', () => {
  it('should export GovernanceDashboard class', () => {
    expect(GovernanceDashboard).toBeDefined();
    expect(typeof GovernanceDashboard).toBe('function');
  });
});


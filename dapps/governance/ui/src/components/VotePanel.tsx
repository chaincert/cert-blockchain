/**
 * Vote Panel Component
 * Per Whitepaper Section 5.2 and 6.2 - Governance Dashboard
 * Interface for casting votes on proposals
 */

import React, { useState, useEffect } from 'react';
import { ethers } from 'ethers';
import { GovernanceDashboard, VoteType, ProposalState, Proposal } from '../../../src/index';

interface VotePanelProps {
  proposal: Proposal;
  governanceAddress: string;
  signer: ethers.Signer;
  onVoteCast: () => void;
}

export const VotePanel: React.FC<VotePanelProps> = ({
  proposal,
  governanceAddress,
  signer,
  onVoteCast,
}) => {
  const [selectedVote, setSelectedVote] = useState<VoteType | null>(null);
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [votingPower, setVotingPower] = useState<bigint>(0n);

  const governance = new GovernanceDashboard(governanceAddress, signer);

  useEffect(() => {
    loadVotingPower();
  }, [signer]);

  const loadVotingPower = async () => {
    // In production, query token contract for voting power
    setVotingPower(BigInt(1000000)); // Placeholder
  };

  const handleVote = async () => {
    if (selectedVote === null) {
      setError('Please select a vote option');
      return;
    }

    setError(null);
    setLoading(true);

    try {
      await governance.vote(proposal.id, selectedVote, reason || undefined);
      onVoteCast();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cast vote');
    } finally {
      setLoading(false);
    }
  };

  const totalVotes = Number(proposal.votes.for) + Number(proposal.votes.against) + Number(proposal.votes.abstain);
  const forPercent = totalVotes > 0 ? (Number(proposal.votes.for) / totalVotes) * 100 : 0;
  const againstPercent = totalVotes > 0 ? (Number(proposal.votes.against) / totalVotes) * 100 : 0;

  const isVotingActive = proposal.state === ProposalState.Active;

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-xl font-bold mb-4">Cast Your Vote</h2>

      {/* Voting Power */}
      <div className="mb-4 p-3 bg-blue-50 rounded-lg">
        <span className="text-sm text-gray-600">Your Voting Power:</span>
        <span className="ml-2 font-bold">{votingPower.toString()} CERT</span>
      </div>

      {/* Vote Results */}
      <div className="mb-6">
        <h3 className="font-medium mb-2">Current Results</h3>
        
        {/* For */}
        <div className="mb-2">
          <div className="flex justify-between text-sm mb-1">
            <span className="text-green-600">For</span>
            <span>{forPercent.toFixed(1)}%</span>
          </div>
          <div className="h-3 bg-gray-200 rounded-full overflow-hidden">
            <div
              className="h-full bg-green-500 transition-all"
              style={{ width: `${forPercent}%` }}
            />
          </div>
          <span className="text-xs text-gray-500">{proposal.votes.for.toString()} CERT</span>
        </div>

        {/* Against */}
        <div className="mb-2">
          <div className="flex justify-between text-sm mb-1">
            <span className="text-red-600">Against</span>
            <span>{againstPercent.toFixed(1)}%</span>
          </div>
          <div className="h-3 bg-gray-200 rounded-full overflow-hidden">
            <div
              className="h-full bg-red-500 transition-all"
              style={{ width: `${againstPercent}%` }}
            />
          </div>
          <span className="text-xs text-gray-500">{proposal.votes.against.toString()} CERT</span>
        </div>

        {/* Abstain */}
        <div className="text-sm text-gray-500">
          Abstain: {proposal.votes.abstain.toString()} CERT
        </div>
      </div>

      {isVotingActive ? (
        <>
          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
              {error}
            </div>
          )}

          {/* Vote Options */}
          <div className="space-y-2 mb-4">
            {[
              { type: VoteType.For, label: 'For', color: 'border-green-500 bg-green-50' },
              { type: VoteType.Against, label: 'Against', color: 'border-red-500 bg-red-50' },
              { type: VoteType.Abstain, label: 'Abstain', color: 'border-gray-500 bg-gray-50' },
            ].map(({ type, label, color }) => (
              <button
                key={type}
                onClick={() => setSelectedVote(type)}
                className={`w-full p-3 rounded-lg border-2 text-left transition-all ${
                  selectedVote === type ? color : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <span className="font-medium">{label}</span>
              </button>
            ))}
          </div>

          {/* Reason (optional) */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Reason (optional)
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Share your reasoning..."
              rows={2}
              className="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
            />
          </div>

          {/* Submit */}
          <button
            onClick={handleVote}
            disabled={loading || selectedVote === null}
            className={`w-full py-2 px-4 rounded-md text-white font-medium ${
              loading || selectedVote === null
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-blue-500 hover:bg-blue-600'
            }`}
          >
            {loading ? 'Casting Vote...' : 'Cast Vote'}
          </button>
        </>
      ) : (
        <div className="text-center py-4 text-gray-500">
          Voting is not active for this proposal
        </div>
      )}
    </div>
  );
};

export default VotePanel;


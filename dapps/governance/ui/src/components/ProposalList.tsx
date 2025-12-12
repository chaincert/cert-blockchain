/**
 * Proposal List Component
 * Per Whitepaper Section 5.2 and 6.2 - Governance Dashboard
 * Displays active and past governance proposals
 */

import React, { useState, useEffect } from 'react';
import { ethers } from 'ethers';
import { GovernanceDashboard, ProposalState, Proposal } from '../../../src/index';

interface ProposalListProps {
  governanceAddress: string;
  signer: ethers.Signer;
  onSelectProposal: (proposal: Proposal) => void;
}

// Proposal state labels and colors
const STATE_CONFIG: Record<ProposalState, { label: string; color: string }> = {
  [ProposalState.Pending]: { label: 'Pending', color: 'bg-gray-500' },
  [ProposalState.Active]: { label: 'Active', color: 'bg-blue-500' },
  [ProposalState.Canceled]: { label: 'Canceled', color: 'bg-red-300' },
  [ProposalState.Defeated]: { label: 'Defeated', color: 'bg-red-500' },
  [ProposalState.Succeeded]: { label: 'Succeeded', color: 'bg-green-500' },
  [ProposalState.Queued]: { label: 'Queued', color: 'bg-yellow-500' },
  [ProposalState.Expired]: { label: 'Expired', color: 'bg-gray-400' },
  [ProposalState.Executed]: { label: 'Executed', color: 'bg-green-700' },
};

export const ProposalList: React.FC<ProposalListProps> = ({
  governanceAddress,
  signer,
  onSelectProposal,
}) => {
  const [proposals, setProposals] = useState<Proposal[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'active' | 'past'>('all');

  const governance = new GovernanceDashboard(governanceAddress, signer);

  useEffect(() => {
    loadProposals();
  }, [governanceAddress]);

  const loadProposals = async () => {
    setLoading(true);
    try {
      // In production, fetch from indexer or events
      // This is a placeholder for demonstration
      const mockProposals: Proposal[] = [];
      setProposals(mockProposals);
    } catch (error) {
      console.error('Failed to load proposals:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredProposals = proposals.filter((p) => {
    if (filter === 'active') {
      return p.state === ProposalState.Active || p.state === ProposalState.Pending;
    }
    if (filter === 'past') {
      return p.state >= ProposalState.Canceled;
    }
    return true;
  });

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Filter Tabs */}
      <div className="flex space-x-2 border-b border-gray-200 pb-2">
        {(['all', 'active', 'past'] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-4 py-2 rounded-t-lg font-medium ${
              filter === f
                ? 'bg-blue-500 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {f.charAt(0).toUpperCase() + f.slice(1)}
          </button>
        ))}
      </div>

      {/* Proposal Cards */}
      {filteredProposals.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No proposals found. Be the first to create one!
        </div>
      ) : (
        <div className="grid gap-4">
          {filteredProposals.map((proposal) => (
            <div
              key={proposal.id}
              onClick={() => onSelectProposal(proposal)}
              className="bg-white rounded-lg shadow p-4 cursor-pointer hover:shadow-md transition-shadow"
            >
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <h3 className="font-semibold text-lg">{proposal.description}</h3>
                  <p className="text-sm text-gray-500">ID: {proposal.id.slice(0, 10)}...</p>
                </div>
                <span
                  className={`px-3 py-1 rounded-full text-white text-sm ${
                    STATE_CONFIG[proposal.state].color
                  }`}
                >
                  {STATE_CONFIG[proposal.state].label}
                </span>
              </div>

              {/* Vote Progress */}
              {proposal.state === ProposalState.Active && (
                <div className="mt-4">
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-green-600">For: {proposal.votes.for.toString()}</span>
                    <span className="text-red-600">Against: {proposal.votes.against.toString()}</span>
                  </div>
                  <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-green-500"
                      style={{
                        width: `${
                          Number(proposal.votes.for) /
                          (Number(proposal.votes.for) + Number(proposal.votes.against) + 1) * 100
                        }%`,
                      }}
                    />
                  </div>
                </div>
              )}

              {/* Parameter Change Info */}
              {proposal.parameterChange && (
                <div className="mt-3 p-2 bg-blue-50 rounded text-sm">
                  <span className="font-medium">{proposal.parameterChange.parameterName}:</span>{' '}
                  {proposal.parameterChange.currentValue.toString()} →{' '}
                  <span className="text-blue-600">{proposal.parameterChange.proposedValue.toString()}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ProposalList;


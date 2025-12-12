/**
 * Create Proposal Component
 * Per Whitepaper Section 5.2 and 6.2 - Governance Dashboard
 * Form for creating new governance proposals
 */

import React, { useState } from 'react';
import { ethers } from 'ethers';
import { GovernanceDashboard, GOVERNABLE_PARAMETERS } from '../../../src/index';

interface CreateProposalProps {
  governanceAddress: string;
  signer: ethers.Signer;
  onProposalCreated: (proposalId: string) => void;
}

type ParameterKey = keyof typeof GOVERNABLE_PARAMETERS;

export const CreateProposal: React.FC<CreateProposalProps> = ({
  governanceAddress,
  signer,
  onProposalCreated,
}) => {
  const [selectedParam, setSelectedParam] = useState<ParameterKey>('MAX_GAS_PER_BLOCK');
  const [proposedValue, setProposedValue] = useState('');
  const [rationale, setRationale] = useState('');
  const [targetContract, setTargetContract] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const governance = new GovernanceDashboard(governanceAddress, signer);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      // Validate inputs
      if (!proposedValue || !rationale || !targetContract) {
        throw new Error('All fields are required');
      }

      const value = parseFloat(proposedValue);
      if (isNaN(value) || value <= 0) {
        throw new Error('Invalid proposed value');
      }

      // Build calldata for parameter change
      // This would be specific to the target contract's interface
      const calldata = ethers.AbiCoder.defaultAbiCoder().encode(
        ['uint256'],
        [BigInt(Math.floor(value))]
      );

      const proposalId = await governance.proposeParameterChange(
        selectedParam,
        value,
        rationale,
        targetContract,
        calldata
      );

      onProposalCreated(proposalId);
      
      // Reset form
      setProposedValue('');
      setRationale('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create proposal');
    } finally {
      setLoading(false);
    }
  };

  const currentParam = GOVERNABLE_PARAMETERS[selectedParam];

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-xl font-bold mb-4">Create Parameter Proposal</h2>
      
      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Parameter Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Parameter to Change
          </label>
          <select
            value={selectedParam}
            onChange={(e) => setSelectedParam(e.target.value as ParameterKey)}
            className="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
          >
            {Object.entries(GOVERNABLE_PARAMETERS).map(([key, param]) => (
              <option key={key} value={key}>
                {param.name} (Current: {param.current} {param.unit})
              </option>
            ))}
          </select>
        </div>

        {/* Current vs Proposed */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Current Value
            </label>
            <div className="p-2 bg-gray-100 rounded-md">
              {currentParam.current} {currentParam.unit}
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Proposed Value
            </label>
            <input
              type="number"
              value={proposedValue}
              onChange={(e) => setProposedValue(e.target.value)}
              placeholder={`New value in ${currentParam.unit}`}
              className="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              required
            />
          </div>
        </div>

        {/* Target Contract */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Target Contract Address
          </label>
          <input
            type="text"
            value={targetContract}
            onChange={(e) => setTargetContract(e.target.value)}
            placeholder="0x..."
            className="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
            required
          />
        </div>

        {/* Rationale */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Rationale
          </label>
          <textarea
            value={rationale}
            onChange={(e) => setRationale(e.target.value)}
            placeholder="Explain why this change is beneficial for the network..."
            rows={4}
            className="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
            required
          />
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          disabled={loading}
          className={`w-full py-2 px-4 rounded-md text-white font-medium ${
            loading
              ? 'bg-gray-400 cursor-not-allowed'
              : 'bg-blue-500 hover:bg-blue-600'
          }`}
        >
          {loading ? 'Creating Proposal...' : 'Create Proposal'}
        </button>
      </form>
    </div>
  );
};

export default CreateProposal;


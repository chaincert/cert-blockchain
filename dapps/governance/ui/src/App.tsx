/**
 * Governance Dashboard Main App
 * Per Whitepaper Section 5.2 and 6.2 - On-Chain Governance
 * Main entry point for the governance UI
 */

import React, { useState, useEffect } from 'react';
import { ethers } from 'ethers';
import { Proposal } from '../../src/index';
import { ProposalList } from './components/ProposalList';
import { CreateProposal } from './components/CreateProposal';
import { VotePanel } from './components/VotePanel';

// Default governance contract address (from genesis)
const GOVERNANCE_ADDRESS = process.env.REACT_APP_GOVERNANCE_ADDRESS || '0x0000000000000000000000000000000000000010';

type View = 'list' | 'create' | 'vote';

const App: React.FC = () => {
  const [signer, setSigner] = useState<ethers.Signer | null>(null);
  const [account, setAccount] = useState<string | null>(null);
  const [currentView, setCurrentView] = useState<View>('list');
  const [selectedProposal, setSelectedProposal] = useState<Proposal | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    checkWalletConnection();
  }, []);

  const checkWalletConnection = async () => {
    if (typeof window.ethereum !== 'undefined') {
      try {
        const provider = new ethers.BrowserProvider(window.ethereum);
        const accounts = await provider.listAccounts();
        if (accounts.length > 0) {
          const signer = await provider.getSigner();
          setSigner(signer);
          setAccount(accounts[0].address);
        }
      } catch (err) {
        console.error('Wallet check failed:', err);
      }
    }
  };

  const connectWallet = async () => {
    if (typeof window.ethereum === 'undefined') {
      setError('Please install MetaMask or another Web3 wallet');
      return;
    }

    try {
      const provider = new ethers.BrowserProvider(window.ethereum);
      await provider.send('eth_requestAccounts', []);
      const signer = await provider.getSigner();
      const address = await signer.getAddress();
      setSigner(signer);
      setAccount(address);
      setError(null);
    } catch (err) {
      setError('Failed to connect wallet');
    }
  };

  const handleSelectProposal = (proposal: Proposal) => {
    setSelectedProposal(proposal);
    setCurrentView('vote');
  };

  const handleProposalCreated = (proposalId: string) => {
    console.log('Proposal created:', proposalId);
    setCurrentView('list');
  };

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center space-x-4">
            <h1 className="text-2xl font-bold text-blue-600">CERT Governance</h1>
            <span className="text-sm text-gray-500">Network Parameter Management</span>
          </div>
          
          {account ? (
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">
                {account.slice(0, 6)}...{account.slice(-4)}
              </span>
              <div className="w-3 h-3 bg-green-500 rounded-full"></div>
            </div>
          ) : (
            <button
              onClick={connectWallet}
              className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
            >
              Connect Wallet
            </button>
          )}
        </div>
      </header>

      {/* Navigation */}
      <nav className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex space-x-4">
            {[
              { view: 'list' as View, label: 'Proposals' },
              { view: 'create' as View, label: 'Create Proposal' },
            ].map(({ view, label }) => (
              <button
                key={view}
                onClick={() => setCurrentView(view)}
                className={`px-4 py-3 border-b-2 font-medium ${
                  currentView === view
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                {label}
              </button>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        {error && (
          <div className="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
            {error}
          </div>
        )}

        {!signer ? (
          <div className="text-center py-12">
            <h2 className="text-xl font-semibold mb-2">Connect Your Wallet</h2>
            <p className="text-gray-600 mb-4">
              Connect your wallet to view and participate in governance
            </p>
            <button
              onClick={connectWallet}
              className="px-6 py-3 bg-blue-500 text-white rounded-md hover:bg-blue-600"
            >
              Connect Wallet
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2">
              {currentView === 'list' && (
                <ProposalList
                  governanceAddress={GOVERNANCE_ADDRESS}
                  signer={signer}
                  onSelectProposal={handleSelectProposal}
                />
              )}
              {currentView === 'create' && (
                <CreateProposal
                  governanceAddress={GOVERNANCE_ADDRESS}
                  signer={signer}
                  onProposalCreated={handleProposalCreated}
                />
              )}
              {currentView === 'vote' && selectedProposal && (
                <VotePanel
                  proposal={selectedProposal}
                  governanceAddress={GOVERNANCE_ADDRESS}
                  signer={signer}
                  onVoteCast={() => setCurrentView('list')}
                />
              )}
            </div>

            {/* Sidebar - Governance Stats */}
            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="font-bold mb-4">Governance Stats</h3>
              <div className="space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Proposal Threshold</span>
                  <span className="font-medium">100,000 CERT</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Voting Delay</span>
                  <span className="font-medium">1 day</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Voting Period</span>
                  <span className="font-medium">7 days</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Quorum</span>
                  <span className="font-medium">4%</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};

export default App;

// TypeScript declaration for window.ethereum
declare global {
  interface Window {
    ethereum?: ethers.Eip1193Provider;
  }
}


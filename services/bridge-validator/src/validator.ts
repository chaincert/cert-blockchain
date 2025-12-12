/**
 * Cross-Chain Bridge Validator Service
 * Per Whitepaper Section 13 (Roadmap Phase 2: Ecosystem)
 * Validates and signs cross-chain transfer requests
 */

import { ethers } from 'ethers';
import { EventEmitter } from 'events';

// Bridge contract ABI (essential functions)
const BRIDGE_ABI = [
  'event TokensLocked(address indexed sender, uint256 amount, uint256 indexed targetChainId, bytes32 transferId)',
  'event AttestationBridged(bytes32 indexed attestationUID, uint256 indexed targetChainId, bytes32 bridgeId)',
  'function releaseTokens(bytes32,address,uint256,bytes[]) external',
  'function receiveAttestation(bytes32,uint256,bytes,bytes[]) external',
  'function getSignatureThreshold() view returns (uint256)',
  'function isTransferProcessed(bytes32) view returns (bool)',
];

export interface BridgeConfig {
  sourceChainRpc: string;
  targetChainRpc: string;
  sourceBridgeAddress: string;
  targetBridgeAddress: string;
  validatorPrivateKey: string;
  pollIntervalMs: number;
}

export interface PendingTransfer {
  transferId: string;
  sender: string;
  recipient: string;
  amount: bigint;
  sourceChainId: number;
  targetChainId: number;
  timestamp: number;
  signatures: string[];
}

/**
 * Bridge Validator - monitors and validates cross-chain transfers
 */
export class BridgeValidator extends EventEmitter {
  private sourceProvider: ethers.JsonRpcProvider;
  private targetProvider: ethers.JsonRpcProvider;
  private sourceContract: ethers.Contract;
  private targetContract: ethers.Contract;
  private validatorWallet: ethers.Wallet;
  private pendingTransfers: Map<string, PendingTransfer> = new Map();
  private isRunning = false;

  constructor(private config: BridgeConfig) {
    super();
    this.sourceProvider = new ethers.JsonRpcProvider(config.sourceChainRpc);
    this.targetProvider = new ethers.JsonRpcProvider(config.targetChainRpc);
    this.validatorWallet = new ethers.Wallet(config.validatorPrivateKey);
    
    this.sourceContract = new ethers.Contract(
      config.sourceBridgeAddress,
      BRIDGE_ABI,
      this.sourceProvider
    );
    this.targetContract = new ethers.Contract(
      config.targetBridgeAddress,
      BRIDGE_ABI,
      this.validatorWallet.connect(this.targetProvider)
    );
  }

  /**
   * Start the validator service
   */
  async start(): Promise<void> {
    this.isRunning = true;
    console.log('[BridgeValidator] Starting validator service...');

    // Listen for TokensLocked events on source chain
    this.sourceContract.on('TokensLocked', async (sender, amount, targetChainId, transferId) => {
      console.log(`[BridgeValidator] Detected lock: ${transferId}`);
      await this.handleTokenLock(transferId, sender, amount, targetChainId);
    });

    // Listen for AttestationBridged events
    this.sourceContract.on('AttestationBridged', async (attestationUID, targetChainId, bridgeId) => {
      console.log(`[BridgeValidator] Detected attestation bridge: ${bridgeId}`);
      await this.handleAttestationBridge(bridgeId, attestationUID, targetChainId);
    });

    this.emit('started');
    console.log('[BridgeValidator] Validator service started');
  }

  /**
   * Stop the validator service
   */
  async stop(): Promise<void> {
    this.isRunning = false;
    this.sourceContract.removeAllListeners();
    this.emit('stopped');
    console.log('[BridgeValidator] Validator service stopped');
  }

  /**
   * Handle token lock event - sign and potentially relay
   */
  private async handleTokenLock(
    transferId: string,
    sender: string,
    amount: bigint,
    targetChainId: bigint
  ): Promise<void> {
    try {
      // Check if already processed on target chain
      const isProcessed = await this.targetContract.isTransferProcessed(transferId);
      if (isProcessed) {
        console.log(`[BridgeValidator] Transfer ${transferId} already processed`);
        return;
      }

      // Create signature for the transfer
      const messageHash = ethers.keccak256(
        ethers.solidityPacked(
          ['bytes32', 'uint256', 'address', 'uint256'],
          [transferId, targetChainId, sender, amount]
        )
      );
      const signature = await this.validatorWallet.signMessage(ethers.getBytes(messageHash));

      // Store pending transfer
      const transfer: PendingTransfer = {
        transferId,
        sender,
        recipient: sender, // Default recipient is sender
        amount,
        sourceChainId: Number((await this.sourceProvider.getNetwork()).chainId),
        targetChainId: Number(targetChainId),
        timestamp: Date.now(),
        signatures: [signature],
      };
      this.pendingTransfers.set(transferId, transfer);

      this.emit('transferSigned', { transferId, signature });
      console.log(`[BridgeValidator] Signed transfer ${transferId}`);

      // Try to relay if threshold met
      await this.tryRelayTransfer(transferId);
    } catch (error) {
      console.error(`[BridgeValidator] Error handling lock: ${error}`);
      this.emit('error', { transferId, error });
    }
  }

  /**
   * Handle attestation bridge event
   */
  private async handleAttestationBridge(
    bridgeId: string,
    attestationUID: string,
    targetChainId: bigint
  ): Promise<void> {
    // Implementation for attestation bridging
    console.log(`[BridgeValidator] Attestation ${attestationUID} bridged to chain ${targetChainId}`);
    this.emit('attestationBridged', { bridgeId, attestationUID, targetChainId });
  }

  /**
   * Attempt to relay transfer if signature threshold met
   */
  private async tryRelayTransfer(transferId: string): Promise<boolean> {
    const transfer = this.pendingTransfers.get(transferId);
    if (!transfer) return false;

    const threshold = await this.targetContract.getSignatureThreshold();
    if (transfer.signatures.length >= Number(threshold)) {
      // Relay to target chain
      const tx = await this.targetContract.releaseTokens(
        transferId,
        transfer.recipient,
        transfer.amount,
        transfer.signatures
      );
      await tx.wait();
      this.pendingTransfers.delete(transferId);
      this.emit('transferRelayed', { transferId, txHash: tx.hash });
      console.log(`[BridgeValidator] Relayed transfer ${transferId}`);
      return true;
    }
    return false;
  }

  /**
   * Add external signature to pending transfer
   */
  addSignature(transferId: string, signature: string): void {
    const transfer = this.pendingTransfers.get(transferId);
    if (transfer && !transfer.signatures.includes(signature)) {
      transfer.signatures.push(signature);
      this.tryRelayTransfer(transferId);
    }
  }

  /**
   * Get pending transfers
   */
  getPendingTransfers(): PendingTransfer[] {
    return Array.from(this.pendingTransfers.values());
  }
}

export default BridgeValidator;


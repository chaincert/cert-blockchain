/**
 * Bridge Validator Service Entry Point
 * Per Whitepaper Section 13 - Cross-Chain Bridge Infrastructure
 */

import { BridgeValidator, BridgeConfig } from './validator';
import express from 'express';

// Configuration from environment variables
const config: BridgeConfig = {
  sourceChainRpc: process.env.SOURCE_CHAIN_RPC || 'http://localhost:8545',
  targetChainRpc: process.env.TARGET_CHAIN_RPC || 'http://localhost:8546',
  sourceBridgeAddress: process.env.SOURCE_BRIDGE_ADDRESS || '',
  targetBridgeAddress: process.env.TARGET_BRIDGE_ADDRESS || '',
  validatorPrivateKey: process.env.VALIDATOR_PRIVATE_KEY || '',
  pollIntervalMs: parseInt(process.env.POLL_INTERVAL_MS || '5000'),
};

// Validate configuration
if (!config.sourceBridgeAddress || !config.targetBridgeAddress) {
  console.error('ERROR: Bridge addresses must be configured');
  process.exit(1);
}
if (!config.validatorPrivateKey) {
  console.error('ERROR: Validator private key must be configured');
  process.exit(1);
}

// Initialize validator
const validator = new BridgeValidator(config);

// Express API for validator status and signature submission
const app = express();
app.use(express.json());

/**
 * Health check endpoint
 */
app.get('/health', (_req, res) => {
  res.json({ status: 'healthy', timestamp: Date.now() });
});

/**
 * Get pending transfers
 */
app.get('/pending', (_req, res) => {
  const pending = validator.getPendingTransfers();
  res.json({ count: pending.length, transfers: pending });
});

/**
 * Submit signature for a pending transfer
 */
app.post('/signature', (req, res) => {
  const { transferId, signature } = req.body;
  
  if (!transferId || !signature) {
    return res.status(400).json({ error: 'Missing transferId or signature' });
  }
  
  validator.addSignature(transferId, signature);
  return res.json({ success: true, transferId });
});

/**
 * Get validator status
 */
app.get('/status', (_req, res) => {
  res.json({
    sourceChain: config.sourceChainRpc,
    targetChain: config.targetChainRpc,
    pendingCount: validator.getPendingTransfers().length,
    uptime: process.uptime(),
  });
});

// Event handlers
validator.on('started', () => console.log('[Service] Validator started'));
validator.on('stopped', () => console.log('[Service] Validator stopped'));
validator.on('transferSigned', (data) => console.log('[Service] Transfer signed:', data.transferId));
validator.on('transferRelayed', (data) => console.log('[Service] Transfer relayed:', data.transferId));
validator.on('error', (data) => console.error('[Service] Error:', data));

// Start service
const PORT = parseInt(process.env.PORT || '3001');

async function main() {
  try {
    await validator.start();
    
    app.listen(PORT, () => {
      console.log(`[Service] Bridge validator API listening on port ${PORT}`);
      console.log(`[Service] Source chain: ${config.sourceChainRpc}`);
      console.log(`[Service] Target chain: ${config.targetChainRpc}`);
    });
    
    // Graceful shutdown
    process.on('SIGINT', async () => {
      console.log('[Service] Shutting down...');
      await validator.stop();
      process.exit(0);
    });
  } catch (error) {
    console.error('[Service] Failed to start:', error);
    process.exit(1);
  }
}

main();

export { validator, app };


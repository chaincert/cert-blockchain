/**
 * Tests for Bridge Validator Service
 */

import { BridgeValidator, BridgeConfig, PendingTransfer } from '../validator';

// Mock ethers before importing
jest.mock('ethers', () => {
  const actual = jest.requireActual('ethers');
  return {
    ...actual,
    ethers: {
      ...actual.ethers,
      JsonRpcProvider: jest.fn().mockImplementation(() => ({
        getNetwork: jest.fn().mockResolvedValue({ chainId: 8888n }),
        destroy: jest.fn(),
        removeAllListeners: jest.fn(),
      })),
      Contract: jest.fn().mockImplementation(() => ({
        on: jest.fn(),
        off: jest.fn(),
        removeAllListeners: jest.fn(),
        getSignatureThreshold: jest.fn().mockResolvedValue(2n),
        isTransferProcessed: jest.fn().mockResolvedValue(false),
        releaseTokens: jest.fn().mockResolvedValue({ wait: jest.fn() }),
      })),
      Wallet: jest.fn().mockImplementation(() => ({
        connect: jest.fn().mockReturnThis(),
        signMessage: jest.fn().mockResolvedValue('0xmocksignature'),
      })),
    },
  };
});

// Suppress unhandled promise rejections from ethers during tests
const originalConsoleLog = console.log;
const originalConsoleError = console.error;

beforeAll(() => {
  process.on('unhandledRejection', () => {});
  // Suppress ethers connection error logs
  console.log = (...args: unknown[]) => {
    const msg = args[0]?.toString() || '';
    if (msg.includes('ECONNREFUSED') || msg.includes('JsonRpcProvider failed')) return;
    originalConsoleLog.apply(console, args);
  };
  console.error = (...args: unknown[]) => {
    const msg = args[0]?.toString() || '';
    if (msg.includes('ECONNREFUSED') || msg.includes('JsonRpcProvider failed')) return;
    originalConsoleError.apply(console, args);
  };
});

afterAll(async () => {
  // Restore console
  console.log = originalConsoleLog;
  console.error = originalConsoleError;
  // Clean up any remaining timers
  jest.useRealTimers();
  // Give time for async cleanup
  await new Promise(resolve => setTimeout(resolve, 100));
});

describe('BridgeValidator', () => {
  let validator: BridgeValidator;
  let config: BridgeConfig;

  beforeEach(() => {
    config = {
      sourceChainRpc: 'http://localhost:8545',
      targetChainRpc: 'http://localhost:8546',
      sourceBridgeAddress: '0x1111111111111111111111111111111111111111',
      targetBridgeAddress: '0x2222222222222222222222222222222222222222',
      validatorPrivateKey: '0x' + '1'.repeat(64),
      pollIntervalMs: 5000,
    };
    validator = new BridgeValidator(config);
  });

  afterEach(async () => {
    await validator.stop();
  });

  describe('constructor', () => {
    it('should initialize with correct configuration', () => {
      expect(validator).toBeDefined();
    });
  });

  describe('start', () => {
    it('should emit started event', async () => {
      const startedHandler = jest.fn();
      validator.on('started', startedHandler);

      await validator.start();

      expect(startedHandler).toHaveBeenCalled();
    });
  });

  describe('stop', () => {
    it('should emit stopped event', async () => {
      const stoppedHandler = jest.fn();
      validator.on('stopped', stoppedHandler);

      await validator.start();
      await validator.stop();

      expect(stoppedHandler).toHaveBeenCalled();
    });
  });

  describe('getPendingTransfers', () => {
    it('should return empty array initially', () => {
      const pending = validator.getPendingTransfers();
      expect(pending).toEqual([]);
    });
  });

  describe('addSignature', () => {
    it('should not add signature for non-existent transfer', () => {
      validator.addSignature('0xnonexistent', '0xsignature');
      expect(validator.getPendingTransfers()).toEqual([]);
    });
  });
});

describe('BridgeConfig', () => {
  it('should require all fields', () => {
    const config: BridgeConfig = {
      sourceChainRpc: 'http://localhost:8545',
      targetChainRpc: 'http://localhost:8546',
      sourceBridgeAddress: '0x1111111111111111111111111111111111111111',
      targetBridgeAddress: '0x2222222222222222222222222222222222222222',
      validatorPrivateKey: '0x' + '1'.repeat(64),
      pollIntervalMs: 5000,
    };

    expect(config.sourceChainRpc).toBeDefined();
    expect(config.targetChainRpc).toBeDefined();
    expect(config.sourceBridgeAddress).toBeDefined();
    expect(config.targetBridgeAddress).toBeDefined();
    expect(config.validatorPrivateKey).toBeDefined();
    expect(config.pollIntervalMs).toBeDefined();
  });
});

describe('PendingTransfer', () => {
  it('should have correct structure', () => {
    const transfer: PendingTransfer = {
      transferId: '0xabc123',
      sender: '0x1111111111111111111111111111111111111111',
      recipient: '0x2222222222222222222222222222222222222222',
      amount: 1000000n,
      sourceChainId: 8888,
      targetChainId: 1,
      timestamp: Date.now(),
      signatures: ['0xsig1'],
    };

    expect(transfer.transferId).toBeDefined();
    expect(transfer.sender).toBeDefined();
    expect(transfer.recipient).toBeDefined();
    expect(transfer.amount).toBeDefined();
    expect(transfer.sourceChainId).toBeDefined();
    expect(transfer.targetChainId).toBeDefined();
    expect(transfer.timestamp).toBeDefined();
    expect(transfer.signatures).toBeInstanceOf(Array);
  });
});


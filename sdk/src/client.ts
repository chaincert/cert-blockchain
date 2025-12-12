/**
 * Main client for CERT Blockchain SDK
 * Per Whitepaper Section 8 - SDK and API
 */

import { ethers } from 'ethers';
import type { ClientConfig, Schema, RegisterSchemaRequest } from './types';
import { EncryptedAttestation } from './attestation';
import { CertID } from './certid';
import { IPFS } from './ipfs';
import {
  CERT_RPC_URL,
  CERT_API_URL,
  CERT_IPFS_GATEWAY,
  CERT_EVM_CHAIN_ID,
} from './constants';

export class CertClient {
  private config: ClientConfig;
  private provider: ethers.JsonRpcProvider;
  private ipfs: IPFS;

  public attestation: EncryptedAttestation;
  public certid: CertID;

  constructor(config?: Partial<ClientConfig>) {
    this.config = {
      rpcUrl: config?.rpcUrl || CERT_RPC_URL,
      apiUrl: config?.apiUrl || CERT_API_URL,
      ipfsUrl: config?.ipfsUrl || 'http://localhost:5001',
      chainId: config?.chainId || CERT_EVM_CHAIN_ID.toString(),
    };

    this.provider = new ethers.JsonRpcProvider(this.config.rpcUrl);
    this.ipfs = new IPFS({
      url: this.config.ipfsUrl!,
      gateway: CERT_IPFS_GATEWAY,
    });

    this.attestation = new EncryptedAttestation(this.config.apiUrl, this.ipfs);
    this.certid = new CertID(this.config.apiUrl);
  }

  /**
   * Get the JSON-RPC provider
   */
  getProvider(): ethers.JsonRpcProvider {
    return this.provider;
  }

  /**
   * Get the IPFS client
   */
  getIPFS(): IPFS {
    return this.ipfs;
  }

  /**
   * Connect a signer (wallet) to the client
   * 
   * @param signer - Ethers signer
   * @returns Connected signer
   */
  connectSigner(signer: ethers.Signer): ethers.Signer {
    return signer.connect(this.provider);
  }

  /**
   * Create a signer from a private key
   * 
   * @param privateKey - Private key
   * @returns Wallet signer
   */
  createSigner(privateKey: string): ethers.Wallet {
    return new ethers.Wallet(privateKey, this.provider);
  }

  /**
   * Register a new schema
   * 
   * @param request - Schema registration request
   * @param signer - Signer for the transaction
   * @returns Registered schema
   */
  async registerSchema(
    request: RegisterSchemaRequest,
    signer: ethers.Signer
  ): Promise<Schema> {
    const creator = await signer.getAddress();
    const signature = await signer.signMessage(
      `Register schema: ${request.schema}`
    );

    const response = await fetch(`${this.config.apiUrl}/api/v1/schemas`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ...request,
        creator,
        signature,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to register schema');
    }

    return response.json();
  }

  /**
   * Get a schema by UID
   * 
   * @param uid - Schema UID
   * @returns Schema data
   */
  async getSchema(uid: string): Promise<Schema> {
    const response = await fetch(`${this.config.apiUrl}/api/v1/schemas/${uid}`);
    if (!response.ok) {
      throw new Error('Schema not found');
    }
    return response.json();
  }

  /**
   * Get the current block number
   */
  async getBlockNumber(): Promise<number> {
    return this.provider.getBlockNumber();
  }

  /**
   * Get the balance of an address
   * 
   * @param address - Address to check
   * @returns Balance in wei
   */
  async getBalance(address: string): Promise<bigint> {
    return this.provider.getBalance(address);
  }

  /**
   * Get the chain ID
   */
  async getChainId(): Promise<bigint> {
    const network = await this.provider.getNetwork();
    return network.chainId;
  }

  /**
   * Check if the client is connected to the correct network
   */
  async isCorrectNetwork(): Promise<boolean> {
    const chainId = await this.getChainId();
    return chainId === BigInt(CERT_EVM_CHAIN_ID);
  }

  /**
   * Wait for a transaction to be confirmed
   * 
   * @param txHash - Transaction hash
   * @param confirmations - Number of confirmations to wait for
   * @returns Transaction receipt
   */
  async waitForTransaction(
    txHash: string,
    confirmations: number = 1
  ): Promise<ethers.TransactionReceipt | null> {
    return this.provider.waitForTransaction(txHash, confirmations);
  }

  /**
   * Get health status of the API
   */
  async getHealth(): Promise<{ status: string; service: string }> {
    const response = await fetch(`${this.config.apiUrl}/health`);
    return response.json();
  }
}


require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.20",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
  networks: {
    // Local CERT blockchain (Ethermint JSON-RPC)
    cert: {
      url: process.env.CERT_RPC_URL || "http://localhost:8545",
      chainId: 11611,
      accounts: process.env.DEPLOYER_PRIVATE_KEY
        ? [process.env.DEPLOYER_PRIVATE_KEY]
        : [],
      gasPrice: "auto",
    },
    // Testnet configuration
    certTestnet: {
      url: process.env.CERT_TESTNET_RPC_URL || "http://localhost:8545",
      chainId: 11612,
      accounts: process.env.DEPLOYER_PRIVATE_KEY
        ? [process.env.DEPLOYER_PRIVATE_KEY]
        : [],
    },
    // Hardhat local for testing
    hardhat: {
      chainId: 31337,
    },
  },
  paths: {
    sources: "./sol",
    tests: "./test",
    cache: "./cache",
    artifacts: "./artifacts",
  },
};


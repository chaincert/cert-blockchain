const hre = require("hardhat");
const fs = require("fs");

async function main() {
  console.log("Deploying CertID contract to CERT blockchain...");
  console.log("Network:", hre.network.name);
  console.log("Chain ID:", hre.network.config.chainId);

  const [deployer] = await hre.ethers.getSigners();
  console.log("Deployer address:", deployer.address);

  const balance = await hre.ethers.provider.getBalance(deployer.address);
  console.log("Deployer balance:", hre.ethers.formatEther(balance), "CERT");

  // Deploy CertID
  console.log("\nDeploying CertID...");
  const CertID = await hre.ethers.getContractFactory("CertID");
  const certID = await CertID.deploy();
  await certID.waitForDeployment();

  const certIDAddress = await certID.getAddress();
  console.log("CertID deployed to:", certIDAddress);

  // Verify deployment
  console.log("\nVerifying deployment...");
  const owner = await certID.owner();
  console.log("Contract owner:", owner);

  // Log badge constants for reference
  console.log("\nStandard badge identifiers:");
  console.log("  BADGE_KYC_L1:", await certID.BADGE_KYC_L1());
  console.log("  BADGE_KYC_L2:", await certID.BADGE_KYC_L2());
  console.log("  BADGE_ACADEMIC:", await certID.BADGE_ACADEMIC());
  console.log("  BADGE_CREATOR:", await certID.BADGE_CREATOR());
  console.log("  BADGE_GOV:", await certID.BADGE_GOV());

  const deploymentInfo = {
    network: hre.network.name,
    chainId: hre.network.config.chainId,
    contracts: {
      CertID: certIDAddress,
    },
    deployer: deployer.address,
    timestamp: new Date().toISOString(),
    blockNumber: await hre.ethers.provider.getBlockNumber(),
  };

  const deploymentPath = `./deployments/${hre.network.name}.json`;
  fs.mkdirSync("./deployments", { recursive: true });
  fs.writeFileSync(deploymentPath, JSON.stringify(deploymentInfo, null, 2));
  console.log("\nDeployment info saved to:", deploymentPath);

  console.log("\n✅ CertID deployment complete!");
  return certIDAddress;
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });


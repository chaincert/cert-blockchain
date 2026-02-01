const hre = require("hardhat");

async function main() {
    console.log("Starting FORCE deployment of CertID...");

    const [deployer] = await hre.ethers.getSigners();
    console.log("Deployer:", deployer.address);

    // Force nonce 3 as requested by node
    const nonce = 3;
    console.log("Forcing Nonce:", nonce);

    const CertID = await hre.ethers.getContractFactory("CertID");
    const deployTx = await CertID.getDeployTransaction();

    // Get Fee Data
    const feeData = await hre.ethers.provider.getFeeData();
    console.log("Gas Price:", feeData.gasPrice?.toString());

    // Construct transaction
    const txRequest = {
        ...deployTx,
        gasLimit: 6000000,
        type: 2,
        maxFeePerGas: 500n, // Super high bump
        maxPriorityFeePerGas: 50n,
        nonce: nonce
    };

    // If fee data is null (some chains), fallback to legacy or gasPrice
    if (!txRequest.maxFeePerGas) {
        delete txRequest.type;
        delete txRequest.maxFeePerGas;
        delete txRequest.maxPriorityFeePerGas;
        txRequest.gasPrice = feeData.gasPrice; // fallback
    }

    console.log("Sending transaction...");
    // Sign and send
    const sentTx = await deployer.sendTransaction(txRequest);
    console.log("Transaction Hash:", sentTx.hash);

    console.log("Waiting for confirmation...");
    const receipt = await sentTx.wait();
    console.log("Deployed Contract Address:", receipt.contractAddress);

    // Log badge constants
    const certID = await hre.ethers.getContractAt("CertID", receipt.contractAddress);
    console.log("  BADGE_KYC_L1:", await certID.BADGE_KYC_L1());
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });

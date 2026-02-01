const hre = require("hardhat");

async function main() {
    console.log("Starting deployment of CertID at NEXT Nonce (4)...");

    const [deployer] = await hre.ethers.getSigners();
    console.log("Deployer:", deployer.address);

    const nonce = 4;
    console.log("Using Nonce:", nonce);

    const CertID = await hre.ethers.getContractFactory("CertID");
    const deployTx = await CertID.getDeployTransaction();

    // Construct transaction
    const txRequest = {
        ...deployTx,
        gasLimit: 8000000,
        type: 2,
        maxFeePerGas: 2000n, // Keep high gas just in case
        maxPriorityFeePerGas: 100n,
        nonce: nonce
    };

    console.log("Sending transaction...");
    console.log(`Gas Config: MaxFee=${txRequest.maxFeePerGas}, MaxPriority=${txRequest.maxPriorityFeePerGas}`);

    try {
        const sentTx = await deployer.sendTransaction(txRequest);
        console.log("Transaction Hash:", sentTx.hash);

        console.log("Waiting for confirmation...");
        const receipt = await sentTx.wait(1);
        console.log("Deployed Contract Address:", receipt.contractAddress);

        // Update verify script logic here manually or just log it
        const certID = await hre.ethers.getContractAt("CertID", receipt.contractAddress);
        console.log("Verification - BADGE_KYC_L1:", await certID.BADGE_KYC_L1());

    } catch (error) {
        console.error("Deployment failed:", error);
        process.exit(1);
    }
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });

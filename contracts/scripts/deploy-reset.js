const hre = require("hardhat");

async function main() {
    console.log("Retrying deployment of CertID at Nonce 0 with Lower Gas Limit...");

    const [deployer] = await hre.ethers.getSigners();
    console.log("Deployer:", deployer.address);

    const nonce = 0;
    
    const CertID = await hre.ethers.getContractFactory("CertID");
    const deployTx = await CertID.getDeployTransaction();

    // Bump to 60 Gwei to replace
    const fee = hre.ethers.parseUnits("60", "gwei");
    const priority = hre.ethers.parseUnits("15", "gwei");

    const txRequest = {
        ...deployTx,
        gasLimit: 4000000, // Reduced from 8M
        type: 2,
        maxFeePerGas: fee,
        maxPriorityFeePerGas: priority,
        nonce: nonce
    };

    console.log("Sending transaction...");
    console.log(`Gas Config: MaxFee=${hre.ethers.formatUnits(fee, "gwei")} Gwei, Limit=${txRequest.gasLimit}`);

    try {
        const sentTx = await deployer.sendTransaction(txRequest);
        console.log("Transaction Hash:", sentTx.hash);

        console.log("Waiting for confirmation...");
        const receipt = await sentTx.wait(1);
        console.log("Deployed Contract Address:", receipt.contractAddress);

        // Verify
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

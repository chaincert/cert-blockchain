const hre = require("hardhat");

async function main() {
    console.log("Attempting to CANCEL nonce 0 with standard Gas Price...");

    const [deployer] = await hre.ethers.getSigners();
    console.log("Deployer:", deployer.address);

    const nonce = 0;

    // 50 Gwei in Wei
    const safeGasPrice = hre.ethers.parseUnits("50", "gwei");

    console.log(`Sending 0 ETH to self with Nonce ${nonce} and Gas Price ${hre.ethers.formatUnits(safeGasPrice, "gwei")} Gwei`);

    const txRequest = {
        to: deployer.address,
        value: 0,
        nonce: nonce,
        gasLimit: 21000,
        maxFeePerGas: safeGasPrice,
        maxPriorityFeePerGas: hre.ethers.parseUnits("10", "gwei"),
        type: 2
    };

    try {
        const sentTx = await deployer.sendTransaction(txRequest);
        console.log("Cancellation Tx Sent. Hash:", sentTx.hash);
        console.log("Waiting for confirmation...");

        await sentTx.wait(1);
        console.log("✅ Nonce 0 Cancelled/Confirmed!");

    } catch (e) {
        console.error("Cancellation failed:", e);
    }
}

main().catch(console.error);

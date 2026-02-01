const hre = require("hardhat");

async function main() {
    const provider = new hre.ethers.JsonRpcProvider("http://localhost:8545");

    // Deployer from previous attempts/config
    const deployerAddress = "0xdB632Dda83F09d205EEEa9374CAa6EEbF0230375";

    console.log("--- Diagnostic Info ---");

    try {
        const network = await provider.getNetwork();
        console.log(`Connected to chain ID: ${network.chainId}`);
    } catch (e) {
        console.log("Error getting network:", e.message);
    }

    try {
        const blockNumber = await provider.getBlockNumber();
        console.log(`Latest Block: ${blockNumber}`);

        const feeData = await provider.getFeeData();
        console.log(`Gas Price: ${feeData.gasPrice}`);
        console.log(`Max Fee Per Gas: ${feeData.maxFeePerGas}`);
        console.log(`Max Priority Fee: ${feeData.maxPriorityFeePerGas}`);

    } catch (e) {
        console.log("Error getting block/fee data:", e.message);
    }

    try {
        const balance = await provider.getBalance(deployerAddress);
        console.log(`Deployer Balance: ${hre.ethers.formatEther(balance)} ETH`);
    } catch (e) {
        console.log("Error getting balance:", e.message);
    }

    try {
        const nonce = await provider.getTransactionCount(deployerAddress, "latest");
        console.log(`Deployer Nonce (latest): ${nonce}`);

        const noncePending = await provider.getTransactionCount(deployerAddress, "pending");
        console.log(`Deployer Nonce (pending): ${noncePending}`);
    } catch (e) {
        console.log("Error getting nonce:", e.message);
    }

    console.log("-----------------------");
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});

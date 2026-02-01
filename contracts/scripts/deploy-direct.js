const { ethers } = require('ethers');
const fs = require('fs');

async function main() {
    console.log("Direct EVM Deployment of CertID...\n");

    const provider = new ethers.JsonRpcProvider("http://localhost:8545");
    const network = await provider.getNetwork();
    console.log("Chain ID:", network.chainId.toString());

    const wallet = new ethers.Wallet(process.env.DEPLOYER_PRIVATE_KEY, provider);
    console.log("Deployer:", wallet.address);

    const balance = await provider.getBalance(wallet.address);
    console.log("Balance:", ethers.formatEther(balance), "CERT");

    // Get nonce and gas price
    const nonce = await provider.getTransactionCount(wallet.address);
    console.log("Nonce:", nonce);

    const feeData = await provider.getFeeData();
    // CERT network requires minimum 7 ucert gas price (7 gwei equivalent)
    const minGasPrice = 7000000000n; // 7 gwei in wei
    const gasPrice = feeData.gasPrice && feeData.gasPrice > minGasPrice ? feeData.gasPrice : minGasPrice;
    console.log("Gas Price:", ethers.formatUnits(gasPrice, 'gwei'), "gwei\n");

    // Read the compiled artifact
    const artifact = JSON.parse(fs.readFileSync('./artifacts/sol/CertID.sol/CertID.json', 'utf8'));

    // Create deployment transaction with higher gas limit
    const deployTx = {
        type: 0, // Legacy transaction
        nonce: nonce,
        gasPrice: gasPrice,
        gasLimit: 5000000n, // 5M gas for contract deployment (much higher)
        data: artifact.bytecode,
        chainId: network.chainId,
    };

    console.log("Signing and sending deployment transaction...");
    const signedTx = await wallet.signTransaction(deployTx);
    console.log("Signed TX length:", signedTx.length);

    try {
        const txResponse = await provider.broadcastTransaction(signedTx);
        console.log("Transaction hash:", txResponse.hash);

        console.log("Waiting for confirmation...");
        const receipt = await txResponse.wait(2); // Wait for 2 confirmations
        console.log("Contract address:", receipt.contractAddress);
        console.log("Gas used:", receipt.gasUsed.toString());
        console.log("Status:", receipt.status === 1 ? "SUCCESS" : "FAILED");

        // Verify code exists
        console.log("\nVerifying contract code...");
        const code = await provider.getCode(receipt.contractAddress);
        console.log("Code length:", code.length, "bytes");

        if (code === '0x' || code.length < 10) {
            console.error("❌ Contract code is empty! Deployment may have failed.");
        } else {
            console.log("✅ Contract code verified!");

            // Save deployment info
            const deploymentInfo = {
                network: "cert-production",
                chainId: network.chainId.toString(),
                contracts: { CertID: receipt.contractAddress },
                deployer: wallet.address,
                timestamp: new Date().toISOString(),
                txHash: txResponse.hash,
                gasUsed: receipt.gasUsed.toString(),
            };

            fs.mkdirSync("./deployments", { recursive: true });
            fs.writeFileSync("./deployments/production.json", JSON.stringify(deploymentInfo, null, 2));
            console.log("\n✅ CertID deployment complete!");
            console.log("Deployment info saved to: ./deployments/production.json");
        }
    } catch (e) {
        console.error("Deployment error:", e.message || e);
        if (e.error) console.error("RPC error:", JSON.stringify(e.error, null, 2));
    }
}

main().catch(console.error);

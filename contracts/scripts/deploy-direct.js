const { ethers } = require('ethers');
const fs = require('fs');

async function main() {
    console.log("Direct EVM Deployment of CertID with Legacy TX...\n");

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

    const gasPrice = await provider.getFeeData();
    console.log("Gas Price:", ethers.formatUnits(gasPrice.gasPrice || 0n, 'gwei'), "gwei\n");

    // Read the compiled artifact
    const artifact = JSON.parse(fs.readFileSync('./artifacts/sol/CertID.sol/CertID.json', 'utf8'));

    // Create deployment transaction as legacy
    const deployTx = {
        type: 0, // Legacy transaction
        nonce: nonce,
        gasPrice: gasPrice.gasPrice || 7000000000n, // 7 gwei minimum
        gasLimit: 3000000n, // 3M gas for contract deployment
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
        const receipt = await txResponse.wait();
        console.log("CertID deployed to:", receipt.contractAddress);

        // Save deployment info
        const deploymentInfo = {
            network: "cert-production",
            chainId: network.chainId.toString(),
            contracts: { CertID: receipt.contractAddress },
            deployer: wallet.address,
            timestamp: new Date().toISOString(),
            txHash: txResponse.hash,
        };

        fs.mkdirSync("./deployments", { recursive: true });
        fs.writeFileSync("./deployments/production.json", JSON.stringify(deploymentInfo, null, 2));
        console.log("\n✅ CertID deployment complete!");
    } catch (e) {
        console.error("Deployment error:", e.message || e);
        if (e.error) console.error("RPC error:", JSON.stringify(e.error, null, 2));
    }
}

main().catch(console.error);

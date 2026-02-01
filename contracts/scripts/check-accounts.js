const hre = require("hardhat");

async function main() {
    console.log("Checking balances of first 5 accounts from wallet mnemonic...");

    const mnemonic = "clarify luggage toddler behave squeeze report around reflect smart flight carry link";

    for (let i = 0; i < 5; i++) {
        const path = `m/44'/60'/0'/0/${i}`;
        // In ethers v6, this creates the wallet at the exact path
        const wallet = hre.ethers.HDNodeWallet.fromPhrase(mnemonic, null, path).connect(hre.ethers.provider);

        try {
            const balance = await hre.ethers.provider.getBalance(wallet.address);
            const nonce = await hre.ethers.provider.getTransactionCount(wallet.address);

            console.log(`Account [${i}] (${path}) Address: ${wallet.address}`);
            console.log(`  Balance: ${hre.ethers.formatEther(balance)} ETH`);
            console.log(`  Nonce:   ${nonce}`);
        } catch (e) {
            console.log(`Account [${i}] Error: ${e.message}`);
        }
    }
}

main().catch(console.error);

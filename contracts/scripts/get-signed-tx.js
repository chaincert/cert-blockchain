const hre = require("hardhat");

async function main() {
    const mnemonic = "clarify luggage toddler behave squeeze report around reflect smart flight carry link";
    const wallet = hre.ethers.Wallet.fromPhrase(mnemonic).connect(hre.ethers.provider);

    console.log("Wallet Address:", wallet.address);

    const nonce = 0;
    const gasPrice = hre.ethers.parseUnits("100", "gwei");

    const txRequest = {
        to: wallet.address,
        value: 0,
        nonce: nonce,
        gasLimit: 21000,
        maxFeePerGas: gasPrice,
        maxPriorityFeePerGas: hre.ethers.parseUnits("10", "gwei"),
        type: 2,
        chainId: 4283207343
    };

    const signedTx = await wallet.signTransaction(txRequest);

    console.log("SIGNED_TX_HEX:", signedTx);
}

main().catch(console.error);

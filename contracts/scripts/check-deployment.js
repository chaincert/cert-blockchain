const hre = require("hardhat");

async function main() {
    const [deployer] = await hre.ethers.getSigners();
    const address0 = hre.ethers.getCreateAddress({ from: deployer.address, nonce: 0 });
    const address1 = hre.ethers.getCreateAddress({ from: deployer.address, nonce: 1 });

    console.log("Checking address for nonce 0:", address0);
    const code0 = await hre.ethers.provider.getCode(address0);

    if (code0 !== "0x") {
        console.log("✅ Contract found at nonce 0!");
        console.log("CertID Address:", address0);
        try {
            const certID = await hre.ethers.getContractAt("CertID", address0);
            console.log("Verifying badge constant...");
            const kyc = await certID.BADGE_KYC_L1();
            console.log("BADGE_KYC_L1:", kyc);
        } catch (e) {
            console.log("Could not call contract:", e.message);
        }
    } else {
        console.log("No code at nonce 0.");
    }

    console.log("Checking address for nonce 1:", address1);
    const code1 = await hre.ethers.provider.getCode(address1);

    if (code1 !== "0x") {
        console.log("✅ Contract found at nonce 1!");
        console.log("CertID Address:", address1);
        try {
            const certID = await hre.ethers.getContractAt("CertID", address1);
            console.log("Verifying badge constant...");
            const kyc = await certID.BADGE_KYC_L1();
            console.log("BADGE_KYC_L1:", kyc);
        } catch (e) {
            console.log("Could not call contract:", e.message);
        }
    } else {
        console.log("No code at nonce 1.");
    }

    const address2 = hre.ethers.getCreateAddress({ from: deployer.address, nonce: 2 });
    console.log("Checking address for nonce 2:", address2);
    const code2 = await hre.ethers.provider.getCode(address2);

    if (code2 !== "0x") {
        console.log("✅ Contract found at nonce 2!");
        console.log("CertID Address:", address2);
        try {
            const certID = await hre.ethers.getContractAt("CertID", address2);
            console.log("Verifying badge constant...");
            const kyc = await certID.BADGE_KYC_L1();
            console.log("BADGE_KYC_L1:", kyc);
        } catch (e) {
            console.log("Could not call contract:", e.message);
        }
    } else {
        console.log("No code at nonce 2.");
    }

    const address3 = hre.ethers.getCreateAddress({ from: deployer.address, nonce: 3 });
    console.log("Checking address for nonce 3:", address3);
    const code3 = await hre.ethers.provider.getCode(address3);

    if (code3 !== "0x") {
        console.log("✅ Contract found at nonce 3!");
        console.log("CertID Address:", address3);
        try {
            const certID = await hre.ethers.getContractAt("CertID", address3);
            console.log("Verifying badge constant...");
            const kyc = await certID.BADGE_KYC_L1();
            console.log("BADGE_KYC_L1:", kyc);
        } catch (e) {
            console.log("Could not call contract:", e.message);
        }
    } else {
        console.log("No code at nonce 3.");
    }
}

main().catch(console.error);

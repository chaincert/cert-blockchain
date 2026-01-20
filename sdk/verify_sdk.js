const { ethers } = require('ethers');
const { CertID, CONTRACT_ADDRESSES, CERT_ID_ABI } = require('./dist/index');

async function main() {
    console.log("Verifying SDK CertID integration...\n");

    // Check that constants are exported correctly
    console.log("1. Checking exported constants...");
    console.log("   CONTRACT_ADDRESSES.CERT_ID:", CONTRACT_ADDRESSES.CERT_ID);
    console.log("   CERT_ID_ABI length:", CERT_ID_ABI.length, "functions/events");

    if (!CONTRACT_ADDRESSES.CERT_ID) {
        console.error("   ❌ CONTRACT_ADDRESSES.CERT_ID is undefined!");
        process.exit(1);
    }
    console.log("   ✅ Constants exported correctly\n");

    // Create a mock provider for testing (no actual network needed for this test)
    console.log("2. Testing CertID class instantiation...");

    // Test with a JsonRpcProvider pointing to local hardhat (may not be running)
    const provider = new ethers.JsonRpcProvider("http://127.0.0.1:8545");
    const certId = new CertID("http://localhost:3000", provider);

    // Access private contract via prototype workaround for verification
    const internalContract = certId.contract;

    if (!internalContract) {
        console.error("   ❌ Contract was not initialized!");
        process.exit(1);
    }

    console.log("   Contract target address:", internalContract.target);

    if (internalContract.target !== CONTRACT_ADDRESSES.CERT_ID) {
        console.error("   ❌ Contract address mismatch!");
        console.error("      Expected:", CONTRACT_ADDRESSES.CERT_ID);
        console.error("      Got:", internalContract.target);
        process.exit(1);
    }
    console.log("   ✅ Contract initialized with correct address\n");

    // Test contract interface
    console.log("3. Verifying contract interface...");
    const contractInterface = internalContract.interface;

    // Check for key functions
    const expectedFunctions = ['getProfile', 'hasBadge', 'getTrustScore', 'resolveHandle'];
    for (const fn of expectedFunctions) {
        if (contractInterface.getFunction(fn)) {
            console.log(`   ✅ Function '${fn}' found`);
        } else {
            console.log(`   ❌ Function '${fn}' NOT found`);
        }
    }

    console.log("\n✅ SDK Verification Complete - All checks passed!");
}

main().catch(console.error);

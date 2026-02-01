package main

import (
    "fmt"
    "os"
    
    "cosmossdk.io/x/tx/signing"
    
    "google.golang.org/protobuf/reflect/protoregistry"
    "google.golang.org/protobuf/reflect/protoreflect"
    protov2 "google.golang.org/protobuf/proto"
    
    evmapi "github.com/evmos/evmos/v20/api/ethermint/evm/v1"
)

func main() {
    // 1. Setup CustomGetSigners map as we do in app.go
    customGetSigners := map[protoreflect.FullName]signing.GetSignersFunc{
        "ethermint.evm.v1.MsgEthereumTx": func(msg protov2.Message) ([][]byte, error) {
            fmt.Println("Custom handler called successfully!")
            // Return dummy signers to avoid panics in downstream logic if we were running full verification
            return [][]byte{}, nil
        },
    }

    // 2. Create the Options (mock)
    opts := signing.Options{
        CustomGetSigners: customGetSigners,
    }

    // 3. Create a MsgEthereumTx (Pulsar)
    // We need to verify if we can create a valid msg that works with GetSigners
    // Manually pack (mocking) or use a helper if we can find one. 
    // Since we just want to test registration, we can try to use the message without Data first 
    // and handle the error gracefully in our mock handler.
    // OR we can try to use evmtypes.PackTxData if we import it correctly.
    
    // Let's just create a message and see if the handler is invoked. 
    // Our custom handler in THIS file calls evmapi.GetSigners which will error if Data is empty.
    // So we should modify our mock handler in THIS file to just print "Called" and verify registration.
    
    msg := &evmapi.MsgEthereumTx{} // Empty data
    
    // 4. Resolve the handler
    msgDescriptor := msg.ProtoReflect().Descriptor()
    fullName := msgDescriptor.FullName()
    fmt.Printf("Msg FullName: %s\n", fullName)
    
    handler, ok := opts.CustomGetSigners[fullName]
    if !ok {
        fmt.Println("ERROR: Handler not found for fullName!")
        os.Exit(1)
    }
    
    fmt.Println("Handler found. Invoking...")
    
    // 5. Invoke handler
    signers, err := handler(msg)
    if err != nil {
        fmt.Printf("Handler returned error (expected as msg is empty): %v\n", err)
    } else {
        fmt.Printf("Handler returned signers: %v\n", signers)
    }
    
    // 6. Test with SDK Context/Handler logic simulation?
    // The SDK's x/auth/ante/sigverify.go uses tx.GetSigners() which usually delegates to the Context/SignerHandler.
    // However, here we just want to verify our map key matches.
    
    // 7. Verify registry
    _, err = protoregistry.GlobalTypes.FindMessageByName(fullName)
    if err != nil {
         fmt.Printf("GlobalTypes lookup error: %v\n", err)
    } else {
         fmt.Println("GlobalTypes lookup success")
    }
}

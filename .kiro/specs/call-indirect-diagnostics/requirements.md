# Requirements Document

## Introduction

This feature enhances the WebAssembly simulator's diagnostic capabilities when a `call_indirect` instruction trap occurs. Currently, when a type mismatch happens during an indirect function call, the simulator only reports a generic error message. This feature will parse the WebAssembly function table at the trap site and provide detailed information about the expected versus actual function signatures, enabling developers to quickly identify and fix signature mismatches.

## Glossary

- **Simulator**: The Rust-based WebAssembly execution engine that runs Soroban smart contracts
- **call_indirect**: A WebAssembly instruction that invokes a function through a function table using a runtime index
- **Function Table**: A WebAssembly table containing function references that can be called indirectly
- **Type Signature**: The parameter and return types of a function (e.g., `(i32, i32) -> i64`)
- **Trap**: A WebAssembly runtime error that halts execution
- **Type Index**: An index into the WebAssembly module's type section that defines a function signature
- **Table Index**: A runtime value used to look up a function in the function table
- **WasmStackTrace**: The structured error representation used by the Simulator to report trap information
- **TrapKind**: An enumeration categorizing different types of WebAssembly traps

## Requirements

### Requirement 1

**User Story:** As a smart contract developer, I want to see the expected and actual function signatures when a call_indirect trap occurs, so that I can quickly identify which function signature is incorrect.

#### Acceptance Criteria

1. WHEN a call_indirect trap occurs due to type mismatch THEN the Simulator SHALL extract the expected type signature from the call_indirect instruction
2. WHEN a call_indirect trap occurs due to type mismatch THEN the Simulator SHALL extract the actual type signature from the function table entry
3. WHEN displaying the trap error THEN the Simulator SHALL format both signatures in human-readable form showing parameter types and return types
4. WHEN the function table entry is valid THEN the Simulator SHALL display the function name or index alongside the signature
5. WHEN formatting type signatures THEN the Simulator SHALL use standard WebAssembly type notation (i32, i64, f32, f64, v128, funcref, externref)

### Requirement 2

**User Story:** As a smart contract developer, I want to see the table index that caused the mismatch, so that I can trace back to the code that computed the incorrect index.

#### Acceptance Criteria

1. WHEN a call_indirect trap occurs THEN the Simulator SHALL extract the runtime table index from the trap context
2. WHEN displaying the trap error THEN the Simulator SHALL include the table index value in the diagnostic message
3. WHEN the table index is out of bounds THEN the Simulator SHALL report this as a distinct error from type mismatch
4. WHEN the table index points to an uninitialized element THEN the Simulator SHALL report this as a distinct error from type mismatch

### Requirement 3

**User Story:** As a smart contract developer, I want the enhanced diagnostics to integrate seamlessly with existing error reporting, so that I don't need to change my debugging workflow.

#### Acceptance Criteria

1. WHEN the Simulator generates enhanced call_indirect diagnostics THEN the Simulator SHALL include this information in the existing WasmStackTrace structure
2. WHEN the enhanced diagnostics are unavailable THEN the Simulator SHALL fall back to the current generic error message
3. WHEN displaying the error through the decode_error function THEN the Simulator SHALL include the enhanced signature information
4. WHEN the TrapKind is IndirectCallTypeMismatch THEN the Simulator SHALL store additional signature details in the trap structure

### Requirement 4

**User Story:** As a simulator maintainer, I want to parse WebAssembly module metadata to extract type information, so that I can provide accurate signature details.

#### Acceptance Criteria

1. WHEN the Simulator loads a WebAssembly module THEN the Simulator SHALL parse and cache the type section containing function signatures
2. WHEN the Simulator loads a WebAssembly module THEN the Simulator SHALL parse and cache the function table definitions
3. WHEN a call_indirect trap occurs THEN the Simulator SHALL look up the expected type index from the call_indirect instruction operand
4. WHEN a call_indirect trap occurs THEN the Simulator SHALL look up the actual function from the table at the runtime index
5. WHEN looking up type information THEN the Simulator SHALL handle missing or malformed metadata gracefully

### Requirement 5

**User Story:** As a smart contract developer, I want to see examples of correct function signatures, so that I can understand how to fix the type mismatch.

#### Acceptance Criteria

1. WHEN displaying a type mismatch error THEN the Simulator SHALL show the expected signature in the format `Expected: (param_types) -> (return_types)`
2. WHEN displaying a type mismatch error THEN the Simulator SHALL show the actual signature in the format `Actual: (param_types) -> (return_types)`
3. WHEN parameter or return types differ THEN the Simulator SHALL highlight which specific types are mismatched
4. WHEN the signatures have different parameter counts THEN the Simulator SHALL clearly indicate the count difference
5. WHEN the signatures have different return counts THEN the Simulator SHALL clearly indicate the count difference

### Requirement 6

**User Story:** As a simulator maintainer, I want comprehensive test coverage for the new diagnostic features, so that I can ensure reliability across different trap scenarios.

#### Acceptance Criteria

1. WHEN testing the feature THEN the Simulator SHALL include property-based tests that verify signature parsing across randomly generated type signatures
2. WHEN testing the feature THEN the Simulator SHALL include unit tests for table index boundary conditions (valid, out of bounds, uninitialized)
3. WHEN testing the feature THEN the Simulator SHALL include unit tests for signature formatting with various type combinations
4. WHEN testing the feature THEN the Simulator SHALL include integration tests using actual WebAssembly modules with call_indirect instructions
5. WHEN testing error handling THEN the Simulator SHALL verify graceful degradation when metadata is unavailable

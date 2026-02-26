# Implementation Plan

- [x] 1. Create WebAssembly type parsing infrastructure
  - Create `simulator/src/wasm_types.rs` module with ValueType enum and FunctionSignature struct
  - Implement signature formatting with human-readable output
  - Implement TypeSection parser using wasmparser crate
  - _Requirements: 1.5, 4.1_

- [ ]* 1.1 Write property test for signature parsing consistency
  - **Property 1: Type signature parsing round-trip consistency**
  - **Validates: Requirements 1.5**

- [ ]* 1.2 Write unit tests for signature formatting
  - Test formatting with all WebAssembly value types (i32, i64, f32, f64, v128, funcref, externref)
  - Test empty parameter and result lists
  - Test multi-value returns
  - _Requirements: 1.3, 1.5_

- [ ] 2. Implement function table inspection
  - Add FunctionTable and FunctionRef structs to wasm_types.rs
  - Implement table parsing to extract function references and type indices
  - Add table element lookup with bounds checking
  - _Requirements: 4.2, 4.4_

- [ ]* 2.1 Write property test for table index boundary detection
  - **Property 3: Table index boundary detection**
  - **Validates: Requirements 2.3**

- [ ]* 2.2 Write unit tests for table index edge cases
  - Test valid index within bounds
  - Test index at table length (out of bounds)
  - Test index beyond table length
  - Test uninitialized table elements
  - _Requirements: 2.3, 2.4_

- [ ] 3. Implement signature comparison logic
  - Add SignatureDiff struct to wasm_types.rs
  - Implement FunctionSignature::compare() method
  - Detect parameter count mismatches
  - Detect result count mismatches
  - Identify specific type mismatches with indices
  - _Requirements: 5.3, 5.4, 5.5_

- [ ]* 3.1 Write property test for comparison symmetry
  - **Property 2: Signature comparison is symmetric for equality**
  - **Validates: Requirements 5.3**

- [ ]* 3.2 Write unit tests for signature comparison
  - Test identical signatures (no differences)
  - Test different parameter counts
  - Test different result counts
  - Test same counts but different types
  - _Requirements: 5.3, 5.4, 5.5_

- [ ] 4. Enhance TrapKind enum with call_indirect details
  - Modify IndirectCallTypeMismatch variant in stack_trace.rs to include signature fields
  - Add expected_signature, actual_signature, table_index, type indices fields
  - Update classify_trap() to detect call_indirect traps
  - _Requirements: 3.4, 4.3_

- [ ] 5. Implement trap context extraction
  - Add extract_call_indirect_details() function to stack_trace.rs
  - Parse error messages to extract table index
  - Parse error messages to extract expected type index
  - Integrate with WasmStackTrace::from_host_error()
  - _Requirements: 2.1, 2.2, 4.3_

- [ ]* 5.1 Write unit tests for context extraction
  - Test extraction from various error message formats
  - Test handling of missing information
  - Test extraction of table indices
  - Test extraction of type indices
  - _Requirements: 2.1, 2.2_

- [ ] 6. Implement WebAssembly module caching
  - Add WasmModuleCache struct to wasm_types.rs
  - Implement from_wasm() to parse and cache type section and function table
  - Add hash-based cache key generation
  - Handle parsing errors gracefully
  - _Requirements: 4.1, 4.2, 4.5_

- [ ]* 6.1 Write property test for graceful degradation
  - **Property 5: Graceful degradation when metadata unavailable**
  - **Validates: Requirements 3.2, 4.5**

- [ ]* 6.2 Write unit tests for error handling
  - Test missing type section
  - Test missing function table
  - Test corrupted WebAssembly module
  - Test empty module
  - Verify fallback to generic messages
  - _Requirements: 3.2, 4.5_

- [ ] 7. Integrate enhanced diagnostics into error formatting
  - Update WasmStackTrace::display() to format call_indirect details
  - Implement format_signature_mismatch() helper function
  - Show expected vs actual signatures in readable format
  - Highlight specific type differences
  - Include table index in output
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.2, 5.1, 5.2_

- [ ]* 7.1 Write property test for signature formatting completeness
  - **Property 4: Signature formatting includes all type information**
  - **Validates: Requirements 1.3, 5.1, 5.2**

- [ ]* 7.2 Write unit tests for error message formatting
  - Test formatting with complete signature information
  - Test formatting with missing actual signature
  - Test formatting with missing expected signature
  - Test formatting with table index
  - _Requirements: 1.3, 5.1, 5.2_

- [ ] 8. Update decode_error function for backward compatibility
  - Modify decode_error() in stack_trace.rs to use enhanced diagnostics
  - Maintain fallback to generic message when details unavailable
  - Ensure existing error format is preserved when metadata missing
  - _Requirements: 3.2, 3.3_

- [ ]* 8.1 Write unit tests for backward compatibility
  - Test that existing error messages still work
  - Test fallback behavior
  - Test integration with decode_error()
  - _Requirements: 3.2, 3.3_

- [ ] 9. Add module bytes to simulation context
  - Modify SimulationRequest or runner to store WebAssembly module bytes
  - Pass module bytes through error handling chain
  - Make module bytes available when constructing WasmStackTrace
  - _Requirements: 4.1, 4.2_

- [ ] 10. Create integration tests with real WebAssembly modules
  - Create test contract with deliberate call_indirect type mismatch
  - Compile contract to WebAssembly
  - Add integration test that triggers the trap
  - Verify enhanced diagnostics appear in output
  - Test with multiple function tables
  - Test with complex signatures (many params/results)
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 3.1_

- [ ]* 10.1 Write property test for type index lookup consistency
  - **Property 6: Type index lookup consistency**
  - **Validates: Requirements 4.3**

- [ ]* 10.2 Write property test for uninitialized element detection
  - **Property 7: Uninitialized element detection**
  - **Validates: Requirements 2.4**

- [ ]* 10.3 Write property test for parameter count mismatch detection
  - **Property 8: Parameter count mismatch detection**
  - **Validates: Requirements 5.4**

- [ ]* 10.4 Write property test for return count mismatch detection
  - **Property 9: Return count mismatch detection**
  - **Validates: Requirements 5.5**

- [ ]* 10.5 Write property test for enhanced diagnostics integration
  - **Property 10: Enhanced diagnostics integration**
  - **Validates: Requirements 3.1, 3.4**

- [ ] 11. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

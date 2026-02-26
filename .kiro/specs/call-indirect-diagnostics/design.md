# Design Document: Enhanced call_indirect Trap Diagnostics

## Overview

This design enhances the WebAssembly simulator's trap diagnostics to provide detailed information when `call_indirect` instructions fail due to type mismatches. The current implementation only reports a generic "indirect call type mismatch" error. This enhancement will parse the WebAssembly module's type section and function table to extract and display the expected versus actual function signatures, along with the table index that caused the mismatch.

The implementation will integrate seamlessly with the existing `WasmStackTrace` infrastructure in `simulator/src/stack_trace.rs` and leverage the `wasmparser` crate already used in `simulator/src/vm.rs` for WebAssembly module parsing.

## Architecture

The enhancement consists of four main components:

1. **Type Signature Parser**: Parses WebAssembly type sections to extract function signatures
2. **Function Table Inspector**: Examines function tables to retrieve actual function types at runtime indices
3. **Trap Context Extractor**: Extracts call_indirect-specific information from trap error messages
4. **Enhanced Diagnostic Formatter**: Formats signature mismatches in human-readable form

### Component Interaction

```
Trap Occurs
    ↓
WasmStackTrace::from_host_error()
    ↓
classify_trap() → IndirectCallTypeMismatch
    ↓
extract_call_indirect_details()
    ├→ parse_wasm_module() → TypeSignatureParser
    ├→ extract_table_index()
    └→ lookup_function_signature() → FunctionTableInspector
    ↓
format_signature_mismatch()
    ↓
Enhanced Error Message
```

## Components and Interfaces

### 1. Type Signature Parser

**Module**: `simulator/src/wasm_types.rs` (new file)

**Purpose**: Parse WebAssembly type sections and represent function signatures

```rust
pub struct FunctionSignature {
    pub params: Vec<ValueType>,
    pub results: Vec<ValueType>,
}

pub enum ValueType {
    I32,
    I64,
    F32,
    F64,
    V128,
    FuncRef,
    ExternRef,
}

pub struct TypeSection {
    types: Vec<FunctionSignature>,
}

impl TypeSection {
    pub fn parse(wasm_bytes: &[u8]) -> Result<Self, String>;
    pub fn get_signature(&self, type_index: u32) -> Option<&FunctionSignature>;
}
```

### 2. Function Table Inspector

**Module**: `simulator/src/wasm_types.rs`

**Purpose**: Parse function tables and map table indices to type indices

```rust
pub struct FunctionTable {
    elements: Vec<Option<FunctionRef>>,
}

pub struct FunctionRef {
    pub func_index: u32,
    pub type_index: u32,
}

impl FunctionTable {
    pub fn parse(wasm_bytes: &[u8]) -> Result<Self, String>;
    pub fn get_function_at(&self, table_index: u32) -> Option<&FunctionRef>;
}
```

### 3. Enhanced TrapKind

**Module**: `simulator/src/stack_trace.rs` (modified)

**Purpose**: Store additional call_indirect-specific information

```rust
pub enum TrapKind {
    // ... existing variants ...
    IndirectCallTypeMismatch {
        expected_signature: Option<FunctionSignature>,
        actual_signature: Option<FunctionSignature>,
        table_index: Option<u32>,
        expected_type_index: Option<u32>,
        actual_type_index: Option<u32>,
    },
    // ... other variants ...
}
```

### 4. Call Indirect Details Extractor

**Module**: `simulator/src/stack_trace.rs` (modified)

**Purpose**: Extract call_indirect-specific information from trap context

```rust
struct CallIndirectDetails {
    table_index: Option<u32>,
    expected_type_index: Option<u32>,
    wasm_module: Option<Vec<u8>>,
}

fn extract_call_indirect_details(error_msg: &str, wasm_bytes: Option<&[u8]>) 
    -> CallIndirectDetails;
```

## Data Models

### FunctionSignature

Represents a WebAssembly function type with parameters and return values.

```rust
pub struct FunctionSignature {
    pub params: Vec<ValueType>,
    pub results: Vec<ValueType>,
}

impl FunctionSignature {
    pub fn format(&self) -> String {
        let params = self.params.iter()
            .map(|t| t.to_string())
            .collect::<Vec<_>>()
            .join(", ");
        let results = self.results.iter()
            .map(|t| t.to_string())
            .collect::<Vec<_>>()
            .join(", ");
        
        format!("({}) -> ({})", params, results)
    }
    
    pub fn compare(&self, other: &FunctionSignature) -> SignatureDiff {
        // Returns detailed comparison showing which types differ
    }
}
```

### SignatureDiff

Represents the differences between two function signatures.

```rust
pub struct SignatureDiff {
    pub param_count_match: bool,
    pub result_count_match: bool,
    pub param_mismatches: Vec<(usize, ValueType, ValueType)>, // (index, expected, actual)
    pub result_mismatches: Vec<(usize, ValueType, ValueType)>,
}
```

### WasmModuleCache

Caches parsed WebAssembly module metadata to avoid repeated parsing.

```rust
pub struct WasmModuleCache {
    type_section: Option<TypeSection>,
    function_table: Option<FunctionTable>,
    wasm_hash: Option<String>,
}

impl WasmModuleCache {
    pub fn from_wasm(wasm_bytes: &[u8]) -> Result<Self, String>;
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Type signature parsing round-trip consistency

*For any* valid WebAssembly type section, parsing it and then formatting each signature should produce a consistent representation that can be parsed again to yield equivalent type information.

**Validates: Requirements 1.5**

### Property 2: Signature comparison is symmetric for equality

*For any* two function signatures A and B, if A.compare(B) shows no mismatches, then B.compare(A) should also show no mismatches.

**Validates: Requirements 5.3**

### Property 3: Table index boundary detection

*For any* function table with N elements, attempting to access an index >= N should be detected and reported as out of bounds rather than type mismatch.

**Validates: Requirements 2.3**

### Property 4: Signature formatting includes all type information

*For any* function signature with parameters and results, the formatted string should contain representations of all parameter types and all result types in order.

**Validates: Requirements 1.3, 5.1, 5.2**

### Property 5: Graceful degradation when metadata unavailable

*For any* trap context where WebAssembly module metadata cannot be parsed, the system should fall back to the generic error message without crashing.

**Validates: Requirements 3.2, 4.5**

### Property 6: Type index lookup consistency

*For any* valid type index in a WebAssembly module's type section, looking up that index should return the same signature on repeated calls.

**Validates: Requirements 4.3**

### Property 7: Uninitialized element detection

*For any* function table element that is null/uninitialized, accessing it should be reported as an uninitialized element error rather than a type mismatch.

**Validates: Requirements 2.4**

### Property 8: Parameter count mismatch detection

*For any* two signatures with different parameter counts, the comparison should explicitly report the count difference.

**Validates: Requirements 5.4**

### Property 9: Return count mismatch detection

*For any* two signatures with different return counts, the comparison should explicitly report the count difference.

**Validates: Requirements 5.5**

### Property 10: Enhanced diagnostics integration

*For any* IndirectCallTypeMismatch trap with available metadata, the WasmStackTrace structure should contain the enhanced signature information.

**Validates: Requirements 3.1, 3.4**

## Error Handling

### Parsing Errors

When parsing WebAssembly modules fails:
- Log the parsing error at debug level
- Fall back to generic error message
- Do not crash or panic
- Include a note in the error that detailed diagnostics are unavailable

### Missing Metadata

When type sections or function tables are missing:
- Detect absence early in the parsing phase
- Return `None` for signature lookups
- Use generic error formatting
- Log at info level that enhanced diagnostics require complete metadata

### Invalid Indices

When type or table indices are out of bounds:
- Validate indices before lookup
- Return appropriate error variant (OutOfBounds vs TypeMismatch)
- Include the invalid index value in the error message
- Do not attempt to access out-of-bounds memory

### Malformed Signatures

When function signatures contain invalid type codes:
- Use wasmparser's built-in validation
- Map unknown types to a placeholder representation
- Include a warning in the diagnostic output
- Continue processing other valid signatures

## Testing Strategy

### Unit Tests

Unit tests will verify specific examples and edge cases:

1. **Signature Parsing Tests**
   - Parse signatures with various type combinations (i32, i64, multiple params/results)
   - Handle empty parameter lists
   - Handle empty result lists
   - Handle multi-value returns

2. **Signature Formatting Tests**
   - Format signatures with all WebAssembly value types
   - Verify parentheses and arrow placement
   - Test empty signatures `() -> ()`

3. **Signature Comparison Tests**
   - Compare identical signatures (should show no differences)
   - Compare signatures with different parameter counts
   - Compare signatures with different result counts
   - Compare signatures with same counts but different types

4. **Table Index Tests**
   - Valid index within bounds
   - Index exactly at table length (out of bounds)
   - Index far beyond table length
   - Negative index handling (if applicable)

5. **Uninitialized Element Tests**
   - Table with null elements
   - Table with mix of initialized and uninitialized elements

6. **Error Fallback Tests**
   - Missing type section
   - Missing function table
   - Corrupted WebAssembly module
   - Empty module

### Property-Based Tests

Property-based tests will verify universal properties across randomly generated inputs using the `proptest` or `quickcheck` crate:

1. **Property Test: Signature Parsing Consistency**
   - Generate random valid function signatures
   - Parse and format them
   - Verify consistent representation

2. **Property Test: Comparison Symmetry**
   - Generate pairs of random signatures
   - Verify A.compare(B) symmetry with B.compare(A) for equality cases

3. **Property Test: Boundary Detection**
   - Generate random table sizes
   - Test indices at boundaries (0, size-1, size, size+1)
   - Verify correct boundary classification

4. **Property Test: Type Preservation**
   - Generate random signatures
   - Verify all types are preserved through parse/format cycle

5. **Property Test: Graceful Degradation**
   - Generate random invalid/incomplete WebAssembly modules
   - Verify no panics or crashes
   - Verify fallback to generic messages

### Integration Tests

Integration tests will use actual WebAssembly modules:

1. **Real call_indirect Test**
   - Compile a Rust contract with call_indirect
   - Introduce a deliberate type mismatch
   - Verify enhanced diagnostics appear in output

2. **Multiple Table Test**
   - Module with multiple function tables
   - Verify correct table is inspected

3. **Complex Signature Test**
   - Functions with many parameters and multiple returns
   - Verify all types are correctly reported

4. **End-to-End Test**
   - Simulate a complete trap scenario
   - Verify WasmStackTrace contains enhanced information
   - Verify decode_error produces readable output

### Test Configuration

- Property-based tests should run a minimum of 100 iterations
- Each property-based test must include a comment explicitly referencing the correctness property from this design document
- Format: `// Feature: call-indirect-diagnostics, Property N: <property text>`
- Integration tests should use real Soroban contracts compiled to WebAssembly
- All tests should verify both success cases and error handling

## Implementation Notes

### WebAssembly Module Access

The simulator needs access to the original WebAssembly module bytes when a trap occurs. This may require:
- Storing the module bytes in the simulation context
- Passing module bytes through the error handling chain
- Caching parsed metadata for performance

### Performance Considerations

- Parse type sections and function tables once per module, cache results
- Avoid re-parsing on every trap
- Use lazy parsing: only parse when IndirectCallTypeMismatch occurs
- Consider memory overhead of caching full module bytes

### Compatibility

- Maintain backward compatibility with existing error messages
- Enhanced diagnostics are additive, not replacing
- Graceful degradation ensures no regression in error reporting
- Existing tests should continue to pass

### Future Enhancements

Potential future improvements not in scope for this feature:
- Source-level function name resolution
- Suggestion of correct signatures based on common patterns
- Interactive debugging mode to inspect function tables
- Integration with source maps for better function identification

// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

//! Integration tests asserting that a known crashing WASM resolves to exactly
//! "src/test.rs:42" across compiler versions (Rust stable and 1.78).
//!
//! The fixture WASM is synthesised in-memory using `gimli::write` so no binary
//! file needs to be committed and no external toolchain is required at test time.

use gimli::write::{
    Address, DebugLineStrOffsets, DebugStrOffsets, EndianVec, LineProgram, LineString, Sections,
};
use gimli::{EndianSlice, LineEncoding, LittleEndian, SectionId};
use object::{Object, ObjectSection};
use simulator::source_mapper::SourceMapper;

// Address in the fixture that maps to src/test.rs:42
const CRASH_ADDR: u64 = 0x1000;

// ─── Fixture builder ────────────────────────────────────────────────────────

/// Builds a minimal but valid WASM binary carrying two custom sections:
/// - `.debug_info` — a bare DWARF32/v4 compilation-unit header (satisfies the
///   symbol-detection check in `SourceMapper`).
/// - `.debug_line` — a DWARF32/v4 line program mapping `CRASH_ADDR` to
///   `src/test.rs` line 42.
fn build_wasm_fixture() -> Vec<u8> {
    let debug_line_bytes = build_debug_line_section();
    let debug_info_bytes = minimal_debug_info_header();

    let mut wasm = vec![0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
    wasm.extend(wasm_custom_section(b".debug_info", &debug_info_bytes));
    wasm.extend(wasm_custom_section(b".debug_line", &debug_line_bytes));
    wasm
}

/// Emits a DWARF32/v4 `.debug_line` section using `gimli::write` where
/// `CRASH_ADDR` → `src/test.rs` line 42.
fn build_debug_line_section() -> Vec<u8> {
    let encoding = gimli::Encoding {
        format: gimli::Format::Dwarf32,
        version: 4,
        address_size: 4,
    };

    // The compilation directory is empty; the comp_file carries the full relative path.
    let comp_dir = LineString::String(b"".to_vec());
    let comp_file = LineString::String(b"src/test.rs".to_vec());

    let mut program =
        LineProgram::new(encoding, LineEncoding::default(), comp_dir, comp_file, None);

    // Explicitly add "src/test.rs" to the file names table.
    // comp_file in LineProgram::new is not written to the file_names section
    // automatically; only files added via add_file appear in the header.
    let dir_id = program.default_directory();
    let file_id = program.add_file(LineString::String(b"src/test.rs".to_vec()), dir_id, None);

    // Emit one sequence: CRASH_ADDR → src/test.rs:42.
    // end_sequence takes an address_offset relative to the sequence start,
    // not an absolute address.
    program.begin_sequence(Some(Address::Constant(CRASH_ADDR)));
    program.row().file = file_id;
    program.row().line = 42;
    program.generate_row();
    program.end_sequence(4); // sequence spans 4 bytes past the first instruction

    // Write into a Sections container and extract the debug_line bytes.
    let mut sections = Sections::new(EndianVec::new(LittleEndian));
    program
        .write(
            &mut sections.debug_line,
            encoding,
            &DebugLineStrOffsets::none(),
            &DebugStrOffsets::none(),
        )
        .expect("gimli::write failed to serialize line program");

    let mut debug_line_bytes = Vec::new();
    sections
        .for_each(|id, writer| {
            if id == SectionId::DebugLine {
                debug_line_bytes = writer.slice().to_vec();
            }
            Ok::<_, gimli::write::Error>(())
        })
        .expect("sections.for_each failed");

    debug_line_bytes
}

/// Returns an 11-byte DWARF32/v4 compilation-unit header with no DIEs.
/// This is the minimum content required for `object` to detect a `.debug_info`
/// section and for `SourceMapper::has_debug_symbols` to return `true`.
fn minimal_debug_info_header() -> Vec<u8> {
    vec![
        0x07, 0x00, 0x00, 0x00, // unit_length = 7 (bytes after this field)
        0x04, 0x00, // DWARF version = 4
        0x00, 0x00, 0x00, 0x00, // debug_abbrev_offset = 0
        0x04, // address_size = 4
    ]
}

/// Encodes `content` as a WASM custom section with the given `name`.
/// Format: [section_id=0x00] [section_size: uleb128] [name_len: uleb128] [name] [content]
fn wasm_custom_section(name: &[u8], content: &[u8]) -> Vec<u8> {
    let mut body = Vec::new();
    write_leb128(&mut body, name.len() as u64);
    body.extend_from_slice(name);
    body.extend_from_slice(content);

    let mut section = vec![0x00]; // custom section ID
    write_leb128(&mut section, body.len() as u64);
    section.extend(body);
    section
}

/// Unsigned LEB128 encoder (stdlib only).
fn write_leb128(buf: &mut Vec<u8>, mut val: u64) {
    loop {
        let byte = (val & 0x7f) as u8;
        val >>= 7;
        if val == 0 {
            buf.push(byte);
            break;
        }
        buf.push(byte | 0x80);
    }
}

// ─── Tests ──────────────────────────────────────────────────────────────────

/// Primary assertion: the full SourceMapper pipeline resolves the known crash
/// address to exactly "src/test.rs" line 42.
#[test]
fn source_map_known_crash_wasm_yields_src_test_rs_42() {
    let wasm = build_wasm_fixture();
    let mapper = SourceMapper::new(wasm);

    assert!(
        mapper.has_debug_symbols(),
        "fixture WASM must be detected as carrying debug symbols"
    );

    let loc = mapper
        .map_wasm_offset_to_source(CRASH_ADDR)
        .expect("must resolve a source location for the known crash offset");

    assert_eq!(loc.file, "src/test.rs", "wrong source file");
    assert_eq!(loc.line, 42, "wrong line number");
}

/// Cross-validation: parse the fixture's DWARF content directly with `gimli`
/// to confirm that what `gimli::write` emits matches the assertion above.
/// This catches any drift between what the fixture builder emits and what the
/// inline DWARF parser in `source_mapper.rs` reads.
#[test]
fn wasm_fixture_dwarf_content_is_canonical() {
    let wasm = build_wasm_fixture();

    let obj = object::File::parse(wasm.as_slice()).expect("fixture must be a valid WASM binary");
    let section = obj
        .section_by_name(".debug_line")
        .expect("fixture must have a .debug_line section");
    let data = section.data().expect(".debug_line data must be readable");

    let debug_line = gimli::DebugLine::new(data, LittleEndian);
    let program = debug_line
        .program(
            gimli::DebugLineOffset(0),
            4u8,
            None::<EndianSlice<'_, LittleEndian>>,
            None::<EndianSlice<'_, LittleEndian>>,
        )
        .expect("must parse line program at offset 0");

    let mut rows = program.rows();
    let mut found = false;

    while let Some((header, row)) = rows.next_row().expect("row iteration must not fail") {
        if row.address() == CRASH_ADDR {
            let file_entry = header
                .file(row.file_index())
                .expect("row must reference a valid file entry");
            let raw_name = match file_entry.path_name() {
                gimli::AttributeValue::String(s) => s.slice(),
                other => panic!("unexpected path_name form: {:?}", other),
            };
            let file_name = std::str::from_utf8(raw_name).expect("file name must be valid UTF-8");

            assert!(
                file_name.contains("test.rs"),
                "expected file name to contain 'test.rs', got '{}'",
                file_name
            );
            assert_eq!(
                row.line().map(|l| l.get()),
                Some(42u64),
                "expected line 42 at address {:#x}",
                CRASH_ADDR
            );

            found = true;
            break;
        }
    }

    assert!(
        found,
        "address {:#x} not found in the .debug_line table",
        CRASH_ADDR
    );
}

/// Control: a WASM without debug sections must yield `None` from both
/// `has_debug_symbols` and `map_wasm_offset_to_source`.
#[test]
fn source_map_wasm_without_symbols_yields_none() {
    // Bare WASM header, no sections
    let wasm = vec![0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
    let mapper = SourceMapper::new(wasm);

    assert!(
        !mapper.has_debug_symbols(),
        "WASM without debug sections must not be detected as having symbols"
    );
    assert!(
        mapper.map_wasm_offset_to_source(CRASH_ADDR).is_none(),
        "lookup on a symbol-free WASM must return None"
    );
}

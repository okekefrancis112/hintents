// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

#![allow(
    dead_code,
    clippy::missing_errors_doc,
    clippy::must_use_candidate,
    clippy::doc_markdown,
    clippy::uninlined_format_args,
    clippy::module_name_repetitions,
    clippy::needless_pass_by_value,
    clippy::unnecessary_wraps,
    clippy::option_if_let_else,
    clippy::redundant_clone,
    clippy::redundant_closure_for_method_calls,
    clippy::cast_possible_truncation,
    clippy::cast_lossless,
    clippy::too_many_lines,
    clippy::unused_self,
    clippy::const_is_empty,
    clippy::unnecessary_semicolon,
    clippy::cast_sign_loss,
    clippy::cast_possible_wrap,
    clippy::missing_const_for_fn,
    clippy::map_unwrap_or
)]

use crate::source_map_cache::SourceMapCache;
use object::{Object, ObjectSection};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

pub struct SourceMapper {
    debug_line_data: Option<Vec<u8>>,
    has_symbols: bool,
    wasm_hash: String,
    cached_mappings: Option<HashMap<u64, SourceLocation>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SourceLocation {
    pub file: String,
    pub line: u32,
    pub column: u32,
    pub column_end: Option<u32>,
}

impl SourceMapper {
    /// Creates a new SourceMapper with caching enabled
    pub fn new(wasm_bytes: Vec<u8>) -> Self {
        let has_symbols = Self::check_debug_symbols(&wasm_bytes);
        let wasm_hash = SourceMapCache::compute_wasm_hash(&wasm_bytes);
        let debug_line_data = has_symbols
            .then(|| Self::extract_debug_line(&wasm_bytes))
            .flatten();

        // Try to load from cache first
        let cached_mappings = if let Ok(cache) = SourceMapCache::new() {
            if let Some(entry) = cache.get(&wasm_hash) {
                if entry.has_symbols == has_symbols {
                    Some(entry.mappings)
                } else {
                    None
                }
            } else {
                None
            }
        } else {
            None
        };

        Self {
            debug_line_data,
            has_symbols,
            wasm_hash,
            cached_mappings,
        }
    }

    /// Creates a new SourceMapper without caching (for testing)
    pub fn new_without_cache(wasm_bytes: Vec<u8>) -> Self {
        let has_symbols = Self::check_debug_symbols(&wasm_bytes);
        let wasm_hash = SourceMapCache::compute_wasm_hash(&wasm_bytes);
        let debug_line_data = has_symbols
            .then(|| Self::extract_debug_line(&wasm_bytes))
            .flatten();
        Self {
            debug_line_data,
            has_symbols,
            wasm_hash,
            cached_mappings: None,
        }
    }

    /// Creates a new SourceMapper with a custom cache directory (for testing)
    pub fn new_with_cache(wasm_bytes: Vec<u8>, cache_dir: std::path::PathBuf) -> Self {
        let has_symbols = Self::check_debug_symbols(&wasm_bytes);
        let wasm_hash = SourceMapCache::compute_wasm_hash(&wasm_bytes);
        let debug_line_data = has_symbols
            .then(|| Self::extract_debug_line(&wasm_bytes))
            .flatten();

        // Try to load from cache first
        let cached_mappings = if let Ok(cache) = SourceMapCache::with_cache_dir(cache_dir) {
            if let Some(entry) = cache.get(&wasm_hash) {
                if entry.has_symbols == has_symbols {
                    Some(entry.mappings)
                } else {
                    None
                }
            } else {
                None
            }
        } else {
            None
        };

        Self {
            debug_line_data,
            has_symbols,
            wasm_hash,
            cached_mappings,
        }
    }

    fn check_debug_symbols(wasm_bytes: &[u8]) -> bool {
        if let Ok(obj_file) = object::File::parse(wasm_bytes) {
            obj_file.section_by_name(".debug_info").is_some()
                && obj_file.section_by_name(".debug_line").is_some()
        } else {
            false
        }
    }

    fn extract_debug_line(wasm_bytes: &[u8]) -> Option<Vec<u8>> {
        let obj = object::File::parse(wasm_bytes).ok()?;
        let section = obj.section_by_name(".debug_line")?;
        section.data().ok().map(|d| d.to_vec())
    }

    pub fn map_wasm_offset_to_source(&self, wasm_offset: u64) -> Option<SourceLocation> {
        // Check cached mappings first
        if let Some(ref cached) = self.cached_mappings {
            if let Some(loc) = cached.get(&wasm_offset) {
                return Some(loc.clone());
            }
        }

        // Fall back to real DWARF .debug_line parsing
        // TODO: iterate all CUs from .debug_info offsets for multi-CU WASM
        parse_debug_line(self.debug_line_data.as_deref()?, wasm_offset)
    }

    pub fn has_debug_symbols(&self) -> bool {
        self.has_symbols
    }

    /// Returns the WASM hash used for caching
    pub fn get_wasm_hash(&self) -> &str {
        &self.wasm_hash
    }
}

// Parses a DWARF32 v2-v5 .debug_line section (little-endian) and returns the
// SourceLocation for `target_addr`, or None if not found or on any parse error.
// Only the opcode subset emitted by gimli::write for a simple line program is
// required; unsupported opcodes are skipped by consuming their operand bytes.
fn parse_debug_line(data: &[u8], target_addr: u64) -> Option<SourceLocation> {
    let mut pos = 0usize;

    // --- Unit header ---
    // unit_length (DWARF32: 4 bytes; skip 64-bit DWARF which begins with 0xffffffff)
    let unit_length = read_u32_le(data, pos)? as usize;
    if unit_length == 0xffff_ffff {
        return None; // 64-bit DWARF not supported
    }
    pos += 4;

    let unit_end = pos + unit_length;
    if unit_end > data.len() {
        return None;
    }

    let version = read_u16_le(data, pos)?;
    pos += 2;
    if !(2..=5).contains(&version) {
        return None;
    }

    // header_length (4 bytes for DWARF32)
    let header_length = read_u32_le(data, pos)? as usize;
    pos += 4;
    let program_start = pos + header_length;

    let minimum_instruction_length = read_u8(data, pos)?;
    pos += 1;

    // maximum_ops_per_instruction introduced in DWARF v4
    let maximum_ops_per_instruction = if version >= 4 {
        let v = read_u8(data, pos)?;
        pos += 1;
        v
    } else {
        1
    };
    if minimum_instruction_length == 0 || maximum_ops_per_instruction == 0 {
        return None;
    }

    let _default_is_stmt = read_u8(data, pos)?;
    pos += 1;

    let line_base = read_i8(data, pos)?;
    pos += 1;

    let line_range = read_u8(data, pos)?;
    pos += 1;
    if line_range == 0 {
        return None;
    }

    let opcode_base = read_u8(data, pos)?;
    pos += 1;

    // standard_opcode_lengths: one byte per standard opcode (opcode_base - 1 entries)
    let std_opcodes_count = opcode_base.saturating_sub(1) as usize;
    if pos + std_opcodes_count > data.len() {
        return None;
    }
    let standard_opcode_lengths: Vec<u8> = data[pos..pos + std_opcodes_count].to_vec();
    pos += std_opcodes_count;

    // include_directories: null-terminated strings, list terminated by empty string
    let mut include_dirs: Vec<String> = vec![String::new()]; // index 0 = compilation directory
    loop {
        if pos >= data.len() {
            return None;
        }
        if data[pos] == 0 {
            pos += 1; // terminator
            break;
        }
        let (s, n) = read_cstr(data, pos)?;
        include_dirs.push(s);
        pos += n;
    }

    // file_names: (name, dir_index, last_modified, file_length) per entry; list terminated by 0x00
    let mut file_names: Vec<(String, usize)> = vec![(String::new(), 0)]; // index 0 unused per spec
    loop {
        if pos >= data.len() {
            return None;
        }
        if data[pos] == 0 {
            // Terminator byte; pos is overridden to program_start below so no need to advance.
            break;
        }
        let (name, n) = read_cstr(data, pos)?;
        pos += n;
        let (dir_idx, n) = read_uleb128(data, pos)?;
        pos += n;
        let (_, n) = read_uleb128(data, pos)?; // last_modified
        pos += n;
        let (_, n) = read_uleb128(data, pos)?; // file_length
        pos += n;
        file_names.push((name, dir_idx as usize));
    }

    // Advance to the line number program
    pos = program_start;

    // --- State machine registers ---
    let mut address: u64 = 0;
    let mut file_idx: usize = 1;
    let mut line: i64 = 1;
    let mut column: u64 = 0;

    while pos < unit_end {
        let opcode = read_u8(data, pos)?;
        pos += 1;

        if opcode == 0 {
            // Extended opcode
            let (ext_len, n) = read_uleb128(data, pos)?;
            pos += n;
            let ext_end = pos + ext_len as usize;
            if ext_end > data.len() {
                return None;
            }
            let ext_opcode = read_u8(data, pos)?;
            pos += 1;

            match ext_opcode {
                1 => {
                    // DW_LNE_end_sequence -- reset state, do not emit
                    address = 0;
                    file_idx = 1;
                    line = 1;
                    column = 0;
                    pos = ext_end;
                }
                2 => {
                    // DW_LNE_set_address (4-byte address for 32-bit WASM)
                    if pos + 4 > ext_end {
                        return None;
                    }
                    address = read_u32_le(data, pos)? as u64;
                    pos = ext_end;
                }
                _ => {
                    pos = ext_end; // skip unknown extended opcodes
                }
            }
        } else if opcode < opcode_base {
            // Standard opcode
            match opcode {
                1 => {
                    // DW_LNS_copy -- emit a row
                    if address == target_addr {
                        return build_location(&file_names, &include_dirs, file_idx, line, column);
                    }
                }
                2 => {
                    // DW_LNS_advance_pc
                    let (op_advance, n) = read_uleb128(data, pos)?;
                    pos += n;
                    address += (op_advance * minimum_instruction_length as u64)
                        / maximum_ops_per_instruction as u64;
                }
                3 => {
                    // DW_LNS_advance_line
                    let (delta, n) = read_sleb128(data, pos)?;
                    pos += n;
                    line = line.wrapping_add(delta);
                }
                4 => {
                    // DW_LNS_set_file
                    let (f, n) = read_uleb128(data, pos)?;
                    pos += n;
                    file_idx = f as usize;
                }
                5 => {
                    // DW_LNS_set_column
                    let (c, n) = read_uleb128(data, pos)?;
                    pos += n;
                    column = c;
                }
                6 => {
                    // DW_LNS_negate_stmt -- no operands, ignore for our purposes
                }
                _ => {
                    // Skip other standard opcodes by consuming their operands
                    let num_operands = standard_opcode_lengths
                        .get(opcode as usize - 1)
                        .copied()
                        .unwrap_or(0);
                    for _ in 0..num_operands {
                        let (_, n) = read_uleb128(data, pos)?;
                        pos += n;
                    }
                }
            }
        } else {
            // Special opcode -- encodes an address+line delta and emits a row
            let adjusted = opcode - opcode_base;
            let op_advance = adjusted / line_range;
            let line_delta = line_base as i64 + (adjusted % line_range) as i64;

            address += (op_advance as u64 * minimum_instruction_length as u64)
                / maximum_ops_per_instruction as u64;
            line = line.wrapping_add(line_delta);

            if address == target_addr {
                return build_location(&file_names, &include_dirs, file_idx, line, column);
            }
        }
    }

    None
}

fn build_location(
    file_names: &[(String, usize)],
    include_dirs: &[String],
    file_idx: usize,
    line: i64,
    column: u64,
) -> Option<SourceLocation> {
    let (file_name, dir_idx) = file_names.get(file_idx)?;
    let dir = include_dirs.get(*dir_idx).map(String::as_str).unwrap_or("");
    let full_path = if dir.is_empty() {
        file_name.clone()
    } else {
        format!("{}/{}", dir, file_name)
    };
    Some(SourceLocation {
        file: full_path,
        line: line.max(0) as u32,
        column: if column > 0 { column as u32 } else { 0 },
        column_end: None,
    })
}

// --- Byte-level helpers (no external dependencies) ---

fn read_u8(data: &[u8], pos: usize) -> Option<u8> {
    data.get(pos).copied()
}

fn read_i8(data: &[u8], pos: usize) -> Option<i8> {
    data.get(pos).map(|&b| b as i8)
}

fn read_u16_le(data: &[u8], pos: usize) -> Option<u16> {
    let bytes: [u8; 2] = data.get(pos..pos + 2)?.try_into().ok()?;
    Some(u16::from_le_bytes(bytes))
}

fn read_u32_le(data: &[u8], pos: usize) -> Option<u32> {
    let bytes: [u8; 4] = data.get(pos..pos + 4)?.try_into().ok()?;
    Some(u32::from_le_bytes(bytes))
}

fn read_cstr(data: &[u8], pos: usize) -> Option<(String, usize)> {
    let end = data[pos..].iter().position(|&b| b == 0)?;
    let s = std::str::from_utf8(&data[pos..pos + end]).ok()?.to_string();
    Some((s, end + 1)) // +1 for the null terminator
}

fn read_uleb128(data: &[u8], pos: usize) -> Option<(u64, usize)> {
    let mut result: u64 = 0;
    let mut shift = 0u32;
    let mut consumed = 0usize;
    loop {
        let byte = *data.get(pos + consumed)?;
        consumed += 1;
        result |= ((byte & 0x7f) as u64) << shift;
        shift += 7;
        if byte & 0x80 == 0 {
            break;
        }
        if shift >= 64 {
            return None; // overflow guard
        }
    }
    Some((result, consumed))
}

fn read_sleb128(data: &[u8], pos: usize) -> Option<(i64, usize)> {
    let mut result: i64 = 0;
    let mut shift = 0u32;
    let mut consumed = 0usize;
    let mut byte;
    loop {
        byte = *data.get(pos + consumed)?;
        consumed += 1;
        result |= ((byte & 0x7f) as i64) << shift;
        shift += 7;
        if byte & 0x80 == 0 {
            break;
        }
        if shift >= 64 {
            return None; // overflow guard
        }
    }
    // Sign-extend if the sign bit of the last group is set
    if shift < 64 && (byte & 0x40) != 0 {
        result |= !0i64 << shift;
    }
    Some((result, consumed))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::source_map_cache::SourceMapCacheEntry;
    use tempfile::TempDir;

    #[test]
    fn test_source_mapper_without_symbols() {
        let wasm_bytes = vec![0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
        let mapper = SourceMapper::new_without_cache(wasm_bytes);

        assert!(!mapper.has_debug_symbols());
        assert!(mapper.map_wasm_offset_to_source(0x1234).is_none());
    }

    #[test]
    fn test_source_mapper_with_mock_symbols() {
        // Minimal WASM header only -- no debug sections
        let wasm_bytes = vec![0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
        let mapper = SourceMapper::new_without_cache(wasm_bytes);

        assert!(!mapper.has_debug_symbols());
    }

    #[test]
    fn test_source_location_serialization() {
        let location = SourceLocation {
            file: "test.rs".to_string(),
            line: 42,
            column: 10,
            column_end: Some(15),
        };

        let json = serde_json::to_string(&location).unwrap();
        assert!(json.contains("test.rs"));
        assert!(json.contains("42"));
    }

    #[test]
    fn test_uleb128_decode() {
        // 300 = 0xAC 0x02 in unsigned LEB128
        let data = [0xAC, 0x02];
        let (val, n) = read_uleb128(&data, 0).unwrap();
        assert_eq!(val, 300);
        assert_eq!(n, 2);
    }

    #[test]
    fn test_sleb128_decode_negative() {
        // -1 = 0x7F in signed LEB128
        let data = [0x7F];
        let (val, n) = read_sleb128(&data, 0).unwrap();
        assert_eq!(val, -1);
        assert_eq!(n, 1);
    }

    #[test]
    fn test_source_mapper_with_cache() {
        let temp_dir = TempDir::new().unwrap();
        let wasm_bytes = vec![0x00, 0x61, 0x73, 0x6d];
        let wasm_hash = SourceMapCache::compute_wasm_hash(&wasm_bytes);

        // First create - this will NOT populate cache because has_symbols is false
        {
            let mapper =
                SourceMapper::new_with_cache(wasm_bytes.clone(), temp_dir.path().to_path_buf());
            assert!(!mapper.has_debug_symbols());

            let result = mapper.map_wasm_offset_to_source(0x1234);
            assert!(result.is_none());
        }

        // Verify cache was NOT created (since no debug symbols)
        let cache = SourceMapCache::with_cache_dir(temp_dir.path().to_path_buf()).unwrap();
        let entries = cache.list_cached().unwrap();
        assert_eq!(entries.len(), 0);

        // Test that we can create cache entries directly
        let mut mappings = std::collections::HashMap::new();
        mappings.insert(
            0x1234,
            SourceLocation {
                file: "test.rs".to_string(),
                line: 42,
                column: 10,
                column_end: None,
            },
        );

        let entry = SourceMapCacheEntry {
            wasm_hash: wasm_hash.clone(),
            has_symbols: true,
            mappings,
            created_at: 1_234_567_890,
        };

        cache.store(entry).unwrap();

        // Verify cache was created
        let entries = cache.list_cached().unwrap();
        assert_eq!(entries.len(), 1);
        assert_eq!(entries[0].wasm_hash, wasm_hash);
    }

    #[test]
    fn test_wasm_hash() {
        let wasm_bytes = vec![0x00, 0x61, 0x73, 0x6d];
        let hash = SourceMapCache::compute_wasm_hash(&wasm_bytes);
        assert_eq!(hash.len(), 64);
    }
}

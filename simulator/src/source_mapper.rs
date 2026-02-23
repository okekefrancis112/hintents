// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

use crate::git_detector::GitRepository;
use object::Object;
use serde::Serialize;
use std::path::Path;

pub struct SourceMapper {
    has_symbols: bool,
    git_repo: Option<GitRepository>,
}

#[derive(Debug, Clone, Serialize)]
pub struct SourceLocation {
    pub file: String,
    pub line: u32,
    pub column: Option<u32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub github_link: Option<String>,
}

impl SourceMapper {
    pub fn new(wasm_bytes: Vec<u8>) -> Self {
        let has_symbols = Self::check_debug_symbols(&wasm_bytes);
        let git_repo = Self::detect_git_repository();
        Self { has_symbols, git_repo }
    }

    pub fn new_with_path(wasm_bytes: Vec<u8>, source_path: Option<&Path>) -> Self {
        let has_symbols = Self::check_debug_symbols(&wasm_bytes);
        let git_repo = source_path
            .and_then(|p| GitRepository::detect(p))
            .or_else(|| Self::detect_git_repository());
        Self { has_symbols, git_repo }
    }

    fn detect_git_repository() -> Option<GitRepository> {
        let current_dir = std::env::current_dir().ok()?;
        GitRepository::detect(&current_dir)
    }

    fn check_debug_symbols(wasm_bytes: &[u8]) -> bool {
        // Check if WASM contains debug sections
        if let Ok(obj_file) = object::File::parse(wasm_bytes) {
            obj_file.section_by_name(".debug_info").is_some()
                && obj_file.section_by_name(".debug_line").is_some()
        } else {
            false
        }
    }

    pub fn map_wasm_offset_to_source(&self, _wasm_offset: u64) -> Option<SourceLocation> {
        if !self.has_symbols {
            return None;
        }

        // For demonstration purposes, simulate mapping
        // In a real implementation, this would use addr2line or similar
        let file = "token.rs".to_string();
        let line = 45;
        let column = Some(12);

        let github_link = self.git_repo
            .as_ref()
            .and_then(|repo| repo.generate_file_link(&file, line));

        Some(SourceLocation {
            file,
            line,
            column,
            github_link,
        })
    }

    pub fn create_source_location(&self, file: String, line: u32, column: Option<u32>) -> SourceLocation {
        let github_link = self.git_repo
            .as_ref()
            .and_then(|repo| repo.generate_file_link(&file, line));

        SourceLocation {
            file,
            line,
            column,
            github_link,
        }
    }

    pub fn has_debug_symbols(&self) -> bool {
        self.has_symbols
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_source_mapper_without_symbols() {
        let wasm_bytes = vec![0x00, 0x61, 0x73, 0x6d]; // Basic WASM header
        let mapper = SourceMapper::new(wasm_bytes);

        assert!(!mapper.has_debug_symbols());
        assert!(mapper.map_wasm_offset_to_source(0x1234).is_none());
    }

    #[test]
    fn test_source_mapper_with_mock_symbols() {
        // This would be a WASM file with debug symbols in a real test
        let wasm_bytes = vec![0x00, 0x61, 0x73, 0x6d];
        let mapper = SourceMapper::new(wasm_bytes);

        // For now, this will return false since we don't have real debug symbols
        // In a real implementation with proper WASM + debug symbols, this would be true
        assert!(!mapper.has_debug_symbols());
    }

    #[test]
    fn test_source_location_serialization() {
        let location = SourceLocation {
            file: "test.rs".to_string(),
            line: 42,
            column: Some(10),
            github_link: None,
        };

        let json = serde_json::to_string(&location).unwrap();
        assert!(json.contains("test.rs"));
        assert!(json.contains("42"));
    }

    #[test]
    fn test_source_location_with_github_link() {
        let location = SourceLocation {
            file: "test.rs".to_string(),
            line: 42,
            column: Some(10),
            github_link: Some("https://github.com/user/repo/blob/abc123/test.rs#L42".to_string()),
        };

        let json = serde_json::to_string(&location).unwrap();
        assert!(json.contains("test.rs"));
        assert!(json.contains("42"));
        assert!(json.contains("github.com"));
    }
}

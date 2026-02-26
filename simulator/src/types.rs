// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

#![allow(dead_code)]

use crate::gas_optimizer::OptimizationReport;
use crate::stack_trace::WasmStackTrace;
use crate::source_mapper::SourceLocation;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

//
// ───────────────────────────── REQUEST ─────────────────────────────
//

#[derive(Debug, Deserialize)]
pub struct SimulationRequest {
    pub envelope_xdr: String,
    pub result_meta_xdr: String,

    pub ledger_entries: Option<HashMap<String, String>>,
    pub contract_wasm: Option<String>,

    // Local wasm loading support
    pub wasm_path: Option<String>,

    pub enable_optimization_advisor: bool,
    pub profile: Option<bool>,

    /// RFC 3339 timestamp supplied by caller (reserved for future use)
    pub timestamp: String,

    // Mocking options
    pub mock_base_fee: Option<u32>,
    pub mock_gas_price: Option<u64>,

    // Optional simulator restore preamble
    #[serde(default)]
    pub restore_preamble: Option<serde_json::Value>,
}

//
// ───────────────────── RESOURCE CALIBRATION ─────────────────────
//

#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct ResourceCalibration {
    pub sha256_fixed: u64,
    pub sha256_per_byte: u64,
    pub keccak256_fixed: u64,
    pub keccak256_per_byte: u64,
    pub ed25519_fixed: u64,
}

//
// ───────────────────────────── RESPONSE ─────────────────────────────
//

#[derive(Debug, Serialize)]
pub struct SimulationResponse {
    pub status: String,
    pub error: Option<String>,

    pub events: Vec<String>,
    pub diagnostic_events: Vec<DiagnosticEvent>,
    pub categorized_events: Vec<CategorizedEvent>,
    pub logs: Vec<String>,

    pub flamegraph: Option<String>,
    pub optimization_report: Option<OptimizationReport>,
    pub budget_usage: Option<BudgetUsage>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub source_location: Option<String>,

    // Debugging additions
    #[serde(skip_serializing_if = "Option::is_none")]
    pub stack_trace: Option<WasmStackTrace>,

    pub wasm_offset: Option<u64>,
}

//
// ───────────────────────────── EVENTS ─────────────────────────────
//

#[derive(Debug, Serialize)]
pub struct DiagnosticEvent {
    pub event_type: String,
    pub contract_id: Option<String>,
    pub topics: Vec<String>,
    pub data: String,
    pub in_successful_contract_call: bool,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub wasm_instruction: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct CategorizedEvent {
    pub category: String,
    pub event: DiagnosticEvent,
}

//
// ───────────────────────────── BUDGET ─────────────────────────────
//

#[derive(Debug, Serialize)]
pub struct BudgetUsage {
    pub cpu_instructions: u64,
    pub memory_bytes: u64,
    pub operations_count: usize,
    pub cpu_limit: u64,
    pub memory_limit: u64,
    pub cpu_usage_percent: f64,
    pub memory_usage_percent: f64,
}

//
// ───────────────────────────── ERRORS ─────────────────────────────
//

#[derive(Debug, Serialize)]
pub struct StructuredError {
    pub error_type: String,
    pub message: String,
    pub details: Option<String>,
}

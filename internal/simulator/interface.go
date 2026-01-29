// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package simulator

type Runner interface {
	Run(req *SimulationRequest) (*SimulationResponse, error)
}

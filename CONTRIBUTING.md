# Contributing to Erst

Thank you for your interest in contributing to Erst! We welcome contributions from the community to help make Stellar debugging better for everyone.

## Getting Started

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** locally:
    ```bash
    git clone https://github.com/your-username/hintents.git
    cd hintents
    ```
3.  **Create a branch** for your feature or bug fix:
    ```bash
    git checkout -b feature/my-new-feature
    ```

## Development Workflow

Erst consists of two parts:
1.  **Go CLI (`cmd/erst`)**: The user-facing tool.
2.  **Rust Simulator (`simulator/`)**: The core logic that replays transactions using `soroban-env-host`.

### Prerequisites
-   Go 1.21+
-   Rust (Standard Stable Toolchain)

### Building
You can build both components using the provided Dockerfile or manually:
```bash
# Build Rust Simulator
cd simulator
cargo build --release
cd ..

# Build Go CLI
go build -o erst cmd/erst/main.go
```

## Submitting a Pull Request

1.  Ensure all tests pass.
2.  Update documentation if you change functionality.
3.  Submit your PR to the `main` branch.

## License

By contributing, you agree that your contributions will be licensed under the Apache License, Version 2.0.

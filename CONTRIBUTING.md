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

## Feature Requests & Voting

We believe the community should drive the project's priorities. Here's how you can participate:

### Requesting a New Feature

1.  **Check existing requests**: Before creating a new feature request, browse [GitHub Discussions](https://github.com/dotandev/hintents/discussions) (under the "Feature Requests" category) to see if someone has already suggested it.
2.  **Create a discussion**: If your idea is new, start a discussion in the **Feature Requests** category with:
    -   A clear, descriptive title
    -   The problem or use case you're trying to solve
    -   Your proposed solution or approach
    -   Any relevant examples or context
3.  **Engage with the community**: Respond to questions and feedback to help refine the idea.

> **Note**: If GitHub Discussions are not yet enabled, please open a regular [GitHub Issue](https://github.com/dotandev/hintents/issues) with the `feature-request` label instead.

### Voting on Features

-   **Use reactions**: Vote for features you'd like to see by adding a 👍 (thumbs up) reaction to the original discussion post.
-   **Avoid "+1" comments**: Please use reactions instead of comments to keep discussions focused.
-   **Priority ranking**: Features with the most 👍 reactions will be prioritized in our development roadmap.

### From Discussion to Implementation

1.  Popular feature requests will be reviewed by maintainers and converted into GitHub Issues when approved.
2.  Approved issues will be labeled with `feature` and `community-requested`.
3.  You're welcome to implement features yourself! Comment on the issue to let others know you're working on it.

### Tips for Great Feature Requests

-   **Be specific**: Vague requests are hard to implement.
-   **Explain the "why"**: Help us understand the problem you're solving.
-   **Consider scope**: Smaller, focused features are easier to review and merge.
-   **Think about compatibility**: How does this fit with Stellar/Soroban's ecosystem?

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

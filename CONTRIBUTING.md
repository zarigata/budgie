# Contributing to Budgie

First off, thank you for considering contributing to Budgie! It's people like you that make Budgie such a great tool.

We welcome any form of contribution, from reporting bugs and suggesting enhancements to submitting pull requests.

## Code of Conduct

This project and everyone participating in it is governed by the [Budgie Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [email@example.com](mailto:email@example.com).

## How Can I Contribute?

### Reporting Bugs

This is one of the most helpful ways you can contribute. Before creating a bug report, please check existing issues to see if someone has already reported it.

When you create a bug report, please include as many details as possible:

- A clear and descriptive title.
- A step-by-step description of how to reproduce the bug.
- The expected behavior and what actually happened.
- Your system information (OS, Budgie version, etc.).

### Suggesting Enhancements

If you have an idea for a new feature or an improvement to an existing one, we'd love to hear it. Please create an issue with a clear description of your suggestion:

- A clear and descriptive title.
- A detailed description of the proposed enhancement.
- The use case or problem that this enhancement would solve.
- Any alternative solutions or features you've considered.

### Submitting Pull Requests

If you'd like to contribute code to Budgie, you can do so by submitting a pull request.

1.  **Fork the repository** and create your branch from `main`.
2.  **Make your changes** in a new git branch.
3.  **Create a pull request** with a clear description of your changes.
4.  **Ensure all tests pass** before submitting.

#### Code Style

- We follow the standard Go formatting (`gofmt`).
- Write clear and concise comments for complex logic.
- Keep functions small and focused on a single task.

#### Commit Messages

- Use a descriptive commit message that explains the "what" and "why" of your changes.
- Reference any related issues in your commit message (e.g., `Fixes #123`).

## Development Setup

To get started with Budgie development, you'll need:

- Go (version 1.21 or later)
- Git

1.  Clone the repository:
    ```bash
    git clone https://github.com/zarigata/budgie.git
    cd budgie
    ```
2.  Install the dependencies:
    ```bash
    go mod tidy
    ```
3.  Run the tests:
    ```bash
    make test
    ```
4.  Build the project:
    ```bash
    make build
    ```

## License

By contributing to Budgie, you agree that your contributions will be licensed under its MIT License.
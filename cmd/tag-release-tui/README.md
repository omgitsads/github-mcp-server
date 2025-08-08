# Tag Release TUI

A terminal user interface (TUI) for creating GitHub releases using Charmbracelet's Bubbletea library.

## Building

```bash
# Build the application into bin/ directory (ignored by git)
go build -o bin/tag-release-tui ./cmd/tag-release-tui
```

## Usage

```bash
# Basic usage with default remote (origin)
./bin/tag-release-tui v1.2.3

# Specify a different remote (e.g., your fork)
./bin/tag-release-tui v1.2.3 --remote omgitsads

# Run in test mode (validation only, no actual changes)
./bin/tag-release-tui v1.2.3 --test

# Test with a specific remote
./bin/tag-release-tui v1.2.3 --remote omgitsads --test
```

## Features

- **Interactive Validation**: Shows real-time validation of release requirements
- **Test Mode**: Run with `--test` flag to validate without making any actual changes
- **Remote Selection**: Specify git remote with `--remote` flag (default: origin)
- **Flexible Branch Support**: Can be configured to work with any branch (currently set to `tag-release-charmbracelet`)
- **Confirmation Screen**: Displays a summary and asks for confirmation before proceeding
- **Live Execution**: Shows progress as the release is being created
- **Post-Release Instructions**: Provides next steps after successful release creation
- **Safe Testing**: Perfect for testing against your fork without affecting upstream

## Flow

1. **Validation Phase**: 
   - Checks tag format (semantic versioning)
   - Verifies you're on the main branch
   - Fetches latest changes
   - Checks working directory is clean
   - Validates branch is up-to-date
   - Ensures tag doesn't already exist

2. **Confirmation Phase**:
   - Shows release summary
   - Lists actions that will be performed
   - Prompts for confirmation (y/n)

3. **Execution Phase**:
   - Creates the release tag
   - Pushes tag to origin
   - Updates latest-release tag
   - Pushes latest-release tag

4. **Completion Phase**:
   - Shows success message
   - Provides post-release instructions
   - Shows relevant links

## Keyboard Controls

- `y` - Confirm release creation (or test simulation)
- `n` - Cancel release creation
- `q` or `Ctrl+C` - Quit at any time
- `Enter` - Exit after completion or error

## Test Mode

Use the `--test` flag to run the application in test mode:

```bash
./bin/tag-release-tui v1.2.3 --test
```

In test mode, the application will:
- Perform all validation checks
- Show the confirmation screen with simulated actions
- Complete without making any actual git operations
- Display a test results summary

This is perfect for:
- Testing the application functionality
- Validating release requirements without risk
- Training or demonstration purposes
- CI/CD validation workflows

## Remote Selection

Use the `--remote` flag to specify which git remote to use:

```bash
# Test against your fork instead of upstream
./bin/tag-release-tui v1.2.3 --remote omgitsads --test

# Actually release to your fork (be careful!)
./bin/tag-release-tui v1.2.3 --remote omgitsads
```

This is especially useful for:
- Testing against your personal fork
- Avoiding accidental releases to upstream repositories
- Working in organizations with multiple remotes

## Error Handling

The application will show detailed error messages if any validation step fails, such as:
- Invalid tag format
- Not on main branch
- Uncommitted changes
- Out-of-date branch
- Tag already exists

## Comparison with Original Script

This TUI version provides the same functionality as the original `script/tag-release` but with:
- Better visual feedback
- Interactive confirmation
- Real-time progress updates
- Improved error presentation
- Built-in test mode (no need for `--dry-run` flag)
- Support for different branches during development
- Enhanced user experience with modern terminal UI

## Configuration

Currently configured for:
- **Allowed Branch**: `tag-release-charmbracelet` (for development/testing)
- **Default Remote**: `origin` (can be overridden with `--remote` flag)
- **Target Branch**: Can be modified in the source code for production use

To use with the main branch in production, change the `allowedBranch` parameter in the `performValidation` call.

## Safety Features

- **Build Location**: Application builds to `bin/` directory (ignored by git)
- **Remote Selection**: Safely test against your fork instead of upstream
- **Test Mode**: Comprehensive validation without making changes
- **Clear Confirmation**: Shows exactly what will happen before proceeding
- **Error Prevention**: Validates all requirements before starting

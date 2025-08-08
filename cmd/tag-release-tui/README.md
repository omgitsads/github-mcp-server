# Tag Release TUI

A terminal user interface (TUI) for creating GitHub releases using Charmbracelet's Bubbletea library.

## Usage

```bash
# Build the application
go build -o tag-release-tui ./cmd/tag-release-tui

# Run with a version tag
./tag-release-tui v1.2.3

# Run in test mode (validation only, no actual changes)
./tag-release-tui v1.2.3 --test

# Or use the convenience script
./run-tag-release-tui.sh v1.2.3 --test
```

## Features

- **Interactive Validation**: Shows real-time validation of release requirements
- **Test Mode**: Run with `--test` flag to validate without making any actual changes
- **Flexible Branch Support**: Can be configured to work with any branch (currently set to `tag-release-charmbracelet`)
- **Confirmation Screen**: Displays a summary and asks for confirmation before proceeding
- **Live Execution**: Shows progress as the release is being created
- **Post-Release Instructions**: Provides next steps after successful release creation

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
./tag-release-tui v1.2.3 --test
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
- **Target Branch**: Can be modified in the source code for production use

To use with the main branch in production, change the `allowedBranch` parameter in the `performValidation` call.

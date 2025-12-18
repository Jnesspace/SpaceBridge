# SpaceBriddge

migration toolkit for Spacelift. Migrate stacks, contexts, policies, and OpenTofu state between Spacelift accounts.

## Features

- **Full Resource Discovery** - Discovers spaces, stacks, contexts, policies, and their relationships
- **OpenTofu Code Generation** - Generates Spacelift provider OpenTofu code from discovered resources
- **Safe State Migration** - Streams OpenTofu state directly between accounts (no local disk storage)
- **Space Filtering** - Migrate specific spaces and their children
- **Disabled Stack Creation** - Creates stacks disabled to prevent runs before state is migrated

## Installation

```bash
# Clone the repository
git clone https://github.com/jnesspace/spacebridge.git
cd spacebridge

# Build
go build -o bin/spacebridge ./cmd/spacebridge

# Or install directly
go install ./cmd/spacebridge
```

## Configuration

SpaceBridge uses environment variables for authentication. Create a `.env` file or export these variables:

```bash
# Source Account (required for discovery and state download)
SOURCE_SPACELIFT_URL=https://your-source.app.spacelift.io
SOURCE_SPACELIFT_KEY_ID=your-api-key-id
SOURCE_SPACELIFT_SECRET_KEY=your-api-secret-key

# Destination Account (required for state upload and enabling stacks)
DESTINATION_SPACELIFT_URL=https://your-destination.app.spacelift.io
DESTINATION_SPACELIFT_KEY_ID=your-api-key-id
DESTINATION_SPACELIFT_SECRET_KEY=your-api-secret-key
```

## Migration Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                    SPACEBRIDGE MIGRATION WORKFLOW               │
└─────────────────────────────────────────────────────────────────┘

1. DISCOVER & GENERATE
   spacebridge generate -o ./terraform/ --disabled

2. ENABLE EXTERNAL STATE ACCESS (source)
   spacebridge state enable-access

3. VERIFY READINESS
   spacebridge state plan

4. APPLY TERRAFORM (destination)
   cd ./terraform/ && tofu init && tofu apply

5. MIGRATE STATE
   spacebridge state migrate

6. ENABLE STACKS (destination)
   spacebridge stacks enable
```

### Step-by-Step Guide

#### 1. Generate Terraform Code

```bash
# Generate Terraform for ALL resources (stacks created disabled)
spacebridge generate -o ./terraform/ --disabled

# Generate for a specific space only
spacebridge generate -o ./terraform/ --disabled -s your-space-id
```

This creates:
- `main.tf` - All Spacelift resources
- `variables.tf` - Variable declarations for secrets
- `secrets.auto.tfvars.template` - Template for secret values
- `provider.tf` - Spacelift provider configuration

#### 2. Enable External State Access

Before state can be downloaded, external state access must be enabled on source stacks (applies to OpenTofu/Terraform stacks):

```bash
# Enable on all stacks
spacebridge state enable-access

# Enable for specific space only
spacebridge state enable-access -s your-space-id
```

#### 3. Verify Migration Readiness

```bash
spacebridge state plan
```

This shows:
- **Ready** - Stacks with managed state + external access enabled
- **Blocked** - Stacks needing external access enabled
- **Skipped** - Self-managed state (uses external backend like S3)
- **N/A** - Non-OpenTofu stacks (Ansible, Kubernetes, etc.)

#### 4. Apply OpenTofu in Destination

```bash
cd ./terraform/

# Fill in secret values
cp secrets.auto.tfvars.template secrets.auto.tfvars
# Edit secrets.auto.tfvars with actual values

# Apply
tofu init
tofu plan
tofu apply
```

Stacks are created **disabled** (`is_disabled = true`) so they won't run before state is migrated.

#### 5. Migrate State

```bash
# Preview what will be migrated
spacebridge state migrate --dry-run

# Migrate all states
spacebridge state migrate

# Migrate specific space only
spacebridge state migrate -s your-space-id
```

State is streamed directly from source to destination - never written to disk.

#### 6. Enable Stacks

```bash
# Preview
spacebridge stacks enable --dry-run

# Enable all disabled stacks
spacebridge stacks enable

# Enable specific space only
spacebridge stacks enable -s your-space-id
```

#### 7. Verify

Trigger a plan on key stacks - they should show "No changes" if state was migrated correctly.

## Commands Reference

### Discovery Commands

```bash
# Discover all resources
spacebridge discover all

# Discover specific resource types
spacebridge discover spaces
spacebridge discover stacks
spacebridge discover contexts
spacebridge discover policies

# Export to manifest file
spacebridge export -o manifest.json
```

### Generate Command

```bash
spacebridge generate [flags]

Flags:
  -o, --output string     Output directory (default "./generated")
  -m, --manifest string   Input manifest file (optional)
  -d, --disabled          Create stacks as disabled for safe migration
  -s, --space string      Only include resources from this space
```

### State Commands

```bash
# Show migration plan
spacebridge state plan [-s space-id]

# Enable external state access on source stacks
spacebridge state enable-access [-s space-id]

# Migrate state from source to destination
spacebridge state migrate [--dry-run] [-s space-id]
```

### Stacks Commands

```bash
# Enable disabled stacks in destination
spacebridge stacks enable [--dry-run] [-s space-id]
```

### Global Flags

```bash
-v, --verbose   Enable verbose output (shows auth details, API calls)
```

## Space Filtering

All commands support `-s, --space` to filter by space ID:

```bash
# List available spaces
spacebridge discover spaces

# Migrate only a specific space
spacebridge generate -o ./terraform/ --disabled -s demo-01K31FVPGCF3656DERFW7YZ0D4
spacebridge state enable-access -s demo-01K31FVPGCF3656DERFW7YZ0D4
spacebridge state migrate -s demo-01K31FVPGCF3656DERFW7YZ0D4
spacebridge stacks enable -s demo-01K31FVPGCF3656DERFW7YZ0D4
```

**Note:** The `generate` command includes child spaces automatically when filtering.

## State Migration Details

### Spacelift-Managed State (`manage_state = true`)

SpaceBridge handles this automatically:
1. Gets pre-signed download URL from source
2. Gets pre-signed upload URL from destination
3. Streams state directly (no disk storage)
4. Triggers state import on destination

### Self-Managed State (`manage_state = false`)

For stacks using external backends (S3, GCS, etc.):
- No migration needed if using the same backend
- The destination stack will read from the existing backend
- Ensure destination account has access to the backend

### Non-OpenTofu Stacks

Ansible, Kubernetes, Pulumi, and CloudFormation stacks don't have OpenTofu state - they're skipped automatically.

## Handling Secrets

Secrets (write-only config values) cannot be read from the API. SpaceBridge:
1. Detects secrets during discovery
2. Creates variable declarations in `variables.tf`
3. Creates a template file `secrets.auto.tfvars.template`

You must manually fill in secret values before applying.

## Example: Full Migration

```bash
# 1. Set up environment
export SOURCE_SPACELIFT_URL=https://old-account.app.spacelift.io
export SOURCE_SPACELIFT_KEY_ID=...
export SOURCE_SPACELIFT_SECRET_KEY=...
export DESTINATION_SPACELIFT_URL=https://new-account.app.spacelift.io
export DESTINATION_SPACELIFT_KEY_ID=...
export DESTINATION_SPACELIFT_SECRET_KEY=...

# 2. Generate Terraform (stacks disabled)
spacebridge generate -o ./migration/ --disabled

# 3. Enable external state access on source
spacebridge state enable-access

# 4. Verify all stacks ready
spacebridge state plan

# 5. Apply OpenTofu to create resources in destination
cd ./migration/
cp secrets.auto.tfvars.template secrets.auto.tfvars
# Edit secrets.auto.tfvars
tofu init && tofu apply

# 6. Migrate state
cd ..
spacebridge state migrate

# 7. Enable stacks
spacebridge stacks enable

# 8. Verify - trigger plans on stacks, should show no changes
```

## Troubleshooting

### "External state access disabled"

Run `spacebridge state enable-access` to enable it via API, or manually enable in Spacelift UI:
Stack Settings > Backend > Enable "External State Access"

### "Stack not found in destination"

Ensure you've run `tofu apply` to create the destination stacks before migrating state.

### "Destination configuration error"

Set the `DESTINATION_*` environment variables. These are required for `state migrate` and `stacks enable`.

### Secrets not migrating

Secrets (write-only values) cannot be read from the API. Fill them in manually in `secrets.auto.tfvars`.

 

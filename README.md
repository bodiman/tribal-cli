# Tribal CLI

Tribal CLI provides commands for managing graph-based development workflows.

## Installation

### Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap bodiman/tribal

# Install tribal
brew install tribal
```

### Manual Installation

Download the latest release from [GitHub Releases](https://github.com/bodiman/tribal-cli/releases) and add it to your PATH.

## Usage

### Initialize a tribal repository

```bash
tribal init
```

This will add the TRIBAL.md file and update Claude Code's CLAUDE.md file.

### Create / retrieve a graph

```bash
tribal checkout -g"<graph title>"
```

### Search for graphs by title / description semantic similarity

```bash
tribal search --context "<context description>"
```

### Stage graph

```bash
tribal add -A
```

### Commit graph

```bash
tribal commit -m"<message>"
```

This will return a diff of the graphs, which should be reviewed before pushing.

### Push a graph

```bash
tribal push
```

## Commands

- `tribal init` - Initialize a tribal repository
- `tribal checkout -g"<title>"` - Create or retrieve a graph by title
- `tribal search --context "<description>"` - Search for graphs semantically
- `tribal add -A` - Stage all graph changes
- `tribal commit -m"<message>"` - Commit staged changes with a message
- `tribal push` - Push committed changes

## Development

### Building from source

```bash
git clone https://github.com/bodiman/tribal-cli.git
cd tribal-cli
go build -o tribal .
```

### Creating a release

This project uses [GoReleaser](https://goreleaser.com/) for releases and automatically updates the Homebrew tap:

```bash
git tag v1.0.4
git push origin v1.0.4
# GitHub Actions will automatically create the release and update the Homebrew formula
```

## License

MIT
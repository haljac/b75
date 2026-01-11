# Blind 75 Go Trainer (b75)

A TUI-based trainer for the Blind 75 LeetCode problems in Go.

## Features

- 🖥️ **TUI Interface**: Browse and select problems.
- 🛠️ **Integrated Workflow**: Opens your `$EDITOR` (vim, nvim, code, etc.) automatically.
- ⚡ **Instant Feedback**: Run tests with a single keystroke (`t`).
- 🤖 **AI Tutor**: Stuck? Ask Gemini (`?`) for a Socratic hint without spoiling the answer.
- 📂 **Isolated Workspaces**: Problems are created in `~/.local/share/b75/problems/` (XDG compliant).

## Installation

```bash
git clone https://github.com/haljac/b75.git
cd b75
go build -o b75 cmd/b75/main.go
mv b75 /usr/local/bin/ # Optional
```

## Usage

1. Set your Gemini API Key (optional, for tutor features):
   ```bash
   export GEMINI_API_KEY="your_api_key"
   ```

2. Run the tool:
   ```bash
   ./b75
   ```

3. Controls:
   - `Arrow Keys`: Navigate list
   - `Enter` / `e`: Open problem in `$EDITOR`
   - `t`: Run tests
   - `?`: Ask AI Tutor for help
   - `q`: Quit

## Problems
The tool comes pre-loaded with all **Blind 75** problems, fully offline.
Internal structure is defined in `internal/workspace/assets/problems`.

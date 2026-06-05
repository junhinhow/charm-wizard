# charm-wizard

A multi-step wizard framework for terminal UIs built with [Lipgloss](https://charm.land/lipgloss).

> [Leia em Portugues](README.pt-br.md)

## Features

- Sequential multi-step execution with progress tracking
- Back/forward navigation between steps
- Shared state between steps via `map[string]any`
- Visual progress indicator with colored step dots and progress bar
- Fluent API for easy wizard construction
- Lipgloss-styled output (green = completed, yellow = current, gray = remaining)

## Install

```bash
go get github.com/junhinhow/charm-wizard
```

## Usage

Here's a deploy wizard example:

```go
package main

import (
	"fmt"
	"log"

	wizard "github.com/junhinhow/charm-wizard"
)

func main() {
	w := wizard.New().
		WithTitle("Deploy Wizard").
		AddStep("validate", "Validate configuration", func(data map[string]any) error {
			fmt.Println("  Checking config files...")
			data["config_valid"] = true
			return nil
		}).
		AddStep("build", "Build application", func(data map[string]any) error {
			fmt.Println("  Compiling project...")
			data["build_artifact"] = "app-v1.2.0.tar.gz"
			return nil
		}).
		AddStep("test", "Run test suite", func(data map[string]any) error {
			artifact := data["build_artifact"].(string)
			fmt.Printf("  Testing %s...\n", artifact)
			data["tests_passed"] = true
			return nil
		}).
		AddStep("deploy", "Deploy to production", func(data map[string]any) error {
			artifact := data["build_artifact"].(string)
			fmt.Printf("  Deploying %s to production...\n", artifact)
			return nil
		})

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## API

### Wizard

| Method | Description |
|--------|-------------|
| `New()` | Create a new wizard |
| `WithTitle(title)` | Set the wizard title |
| `AddStep(name, desc, fn)` | Add a step with a run function |
| `SetData(key, value)` | Pre-set shared data |
| `Data()` | Get the shared data map |
| `Run()` | Execute all steps sequentially |
| `RunStep()` | Execute only the current step |
| `Back()` | Navigate to previous step |
| `Forward()` | Navigate to next step |
| `GoTo(index)` | Jump to a specific step |
| `CurrentStep()` | Get current step index |
| `TotalSteps()` | Get total number of steps |

### Progress

| Function | Description |
|----------|-------------|
| `RenderProgress(steps, index)` | Render the progress indicator |

## License

MIT

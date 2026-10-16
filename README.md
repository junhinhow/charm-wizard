# charm-wizard

A multi-step wizard framework for terminal UIs built with [Lipgloss](https://charm.land/lipgloss).

> [Leia em Português (PT-BR)](README.pt-br.md)

## Features

- Sequential multi-step execution with progress tracking
- Back/forward navigation between steps
- Shared state between steps via `map[string]any`
- Visual progress indicator with colored step dots and progress bar
- Typed getters for the shared data (`GetString`, `GetBool`, `GetInt`)
- Fluent API for easy wizard construction
- Lipgloss-styled output (green = completed, yellow = current, gray = remaining)

The wizard is safe by default: `RenderProgress` never panics with an invalid
index, `Forward()` refuses to move past a step that failed, and `Back()` undoes
the side effects of the step being abandoned (the shared data returns to the
state it had before that step ran).

`Run()` starts from the **current** step, so a wizard that already navigated to
the third step does not re-run the first two. The progress indicator is redrawn
in place on a terminal and written to the configured writer otherwise.

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
			artifact, ok := w.GetString("build_artifact")
			if !ok {
				return fmt.Errorf("build_artifact ausente")
			}
			fmt.Printf("  Testing %s...\n", artifact)
			data["tests_passed"] = true
			return nil
		}).
		AddStep("deploy", "Deploy to production", func(data map[string]any) error {
			artifact, ok := w.GetString("build_artifact")
			if !ok {
				return fmt.Errorf("build_artifact ausente")
			}
			fmt.Printf("  Deploying %s to production...\n", artifact)
			return nil
		})

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
```

The typed getters return `(value, ok)`, so a missing or mistyped key never
causes a panic — unlike a direct type assertion on the shared map.

## API

### Wizard

| Method | Description |
|--------|-------------|
| `New()` | Create a new wizard |
| `WithTitle(title)` | Set the wizard title |
| `WithWriter(w)` | Set the output writer (defaults to `os.Stdout`) |
| `AddStep(name, desc, fn)` | Add a step with a run function |
| `SetData(key, value)` | Pre-set shared data |
| `Data()` | Get the shared data map |
| `GetString(key)` | Get a shared value as `string` (`value, ok`) |
| `GetBool(key)` | Get a shared value as `bool` (`value, ok`) |
| `GetInt(key)` | Get a shared value as `int` (`value, ok`) |
| `Run()` | Execute the steps sequentially, from the current step |
| `RunStep()` | Execute only the current step and render the progress |
| `Back()` | Navigate to previous step (undoes its side effects) |
| `Forward()` | Navigate to next step (refused if the current step failed) |
| `GoTo(index)` | Jump to a specific step |
| `CurrentStep()` | Get current step index |
| `CurrentStepName()` | Get the name of the current step |
| `TotalSteps()` | Get total number of steps |
| `Steps()` | Get a copy of the registered steps |
| `IsFirst()` | Whether the current step is the first one |
| `IsLast()` | Whether the current step is the last one |
| `Progress()` | Progress as a fraction between 0 and 1 |

### Progress

| Function | Description |
|----------|-------------|
| `RenderProgress(steps, index)` | Render the progress indicator (empty string for an invalid index) |

### Errors

| Error | Description |
|-------|-------------|
| `ErrNoSteps` | The wizard has no registered steps |
| `ErrInvalidStepIndex` | The step index is out of range |

## License

MIT

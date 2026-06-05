// Pacote wizard fornece um framework de assistente multi-etapas para interfaces de terminal.
// Permite criar wizards com navegacao para frente/tras, estado compartilhado entre etapas
// e indicador visual de progresso estilizado com Lipgloss.
package wizard

import (
	"fmt"
)

// Step representa uma etapa individual do wizard.
type Step struct {
	// Nome curto da etapa (ex: "config", "deploy")
	Name string
	// Descricao exibida ao usuario durante a execucao
	Description string
	// RunFunc executa a logica da etapa. Recebe o mapa de dados compartilhado
	// e retorna erro caso a etapa falhe.
	RunFunc func(data map[string]any) error
}

// Wizard orquestra a execucao sequencial de etapas com estado compartilhado.
type Wizard struct {
	steps       []Step
	currentStep int
	data        map[string]any
	title       string
}

// New cria um novo Wizard vazio com dados inicializados.
func New() *Wizard {
	return &Wizard{
		steps:       make([]Step, 0),
		currentStep: 0,
		data:        make(map[string]any),
		title:       "Wizard",
	}
}

// WithTitle define o titulo do wizard (API fluente).
func (w *Wizard) WithTitle(title string) *Wizard {
	w.title = title
	return w
}

// AddStep adiciona uma nova etapa ao wizard (API fluente).
func (w *Wizard) AddStep(name, description string, runFunc func(data map[string]any) error) *Wizard {
	w.steps = append(w.steps, Step{
		Name:        name,
		Description: description,
		RunFunc:     runFunc,
	})
	return w
}

// Data retorna o mapa de dados compartilhado entre etapas.
func (w *Wizard) Data() map[string]any {
	return w.data
}

// SetData define um valor no mapa de dados compartilhado.
func (w *Wizard) SetData(key string, value any) *Wizard {
	w.data[key] = value
	return w
}

// CurrentStep retorna o indice da etapa atual (base zero).
func (w *Wizard) CurrentStep() int {
	return w.currentStep
}

// TotalSteps retorna o numero total de etapas.
func (w *Wizard) TotalSteps() int {
	return len(w.steps)
}

// Steps retorna uma copia das etapas registradas.
func (w *Wizard) Steps() []Step {
	copia := make([]Step, len(w.steps))
	copy(copia, w.steps)
	return copia
}

// Back retrocede uma etapa. Retorna false se ja estiver na primeira.
func (w *Wizard) Back() bool {
	if w.currentStep <= 0 {
		return false
	}
	w.currentStep--
	return true
}

// Forward avanca uma etapa. Retorna false se ja estiver na ultima.
func (w *Wizard) Forward() bool {
	if w.currentStep >= len(w.steps)-1 {
		return false
	}
	w.currentStep++
	return true
}

// GoTo navega para uma etapa especifica pelo indice. Retorna erro se fora dos limites.
func (w *Wizard) GoTo(index int) error {
	if index < 0 || index >= len(w.steps) {
		return fmt.Errorf("indice de etapa invalido: %d (total: %d)", index, len(w.steps))
	}
	w.currentStep = index
	return nil
}

// Run executa todas as etapas sequencialmente, exibindo progresso no terminal.
// Retorna erro caso alguma etapa falhe.
func (w *Wizard) Run() error {
	if len(w.steps) == 0 {
		return fmt.Errorf("wizard sem etapas registradas")
	}

	fmt.Println(renderTitle(w.title))
	fmt.Println()

	for i := range w.steps {
		w.currentStep = i
		etapa := w.steps[i]

		// Renderiza indicador de progresso
		fmt.Println(RenderProgress(w.steps, i))
		fmt.Println()

		if etapa.RunFunc != nil {
			if err := etapa.RunFunc(w.data); err != nil {
				return fmt.Errorf("falha na etapa %q: %w", etapa.Name, err)
			}
		}

		fmt.Println()
	}

	fmt.Println(renderComplete(w.title))
	return nil
}

// RunStep executa apenas a etapa atual sem avancar automaticamente.
// Util para wizards interativos com navegacao manual.
func (w *Wizard) RunStep() error {
	if len(w.steps) == 0 {
		return fmt.Errorf("wizard sem etapas registradas")
	}
	if w.currentStep < 0 || w.currentStep >= len(w.steps) {
		return fmt.Errorf("indice de etapa invalido: %d", w.currentStep)
	}

	etapa := w.steps[w.currentStep]
	if etapa.RunFunc != nil {
		if err := etapa.RunFunc(w.data); err != nil {
			return fmt.Errorf("falha na etapa %q: %w", etapa.Name, err)
		}
	}
	return nil
}

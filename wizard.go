// Package wizard fornece um framework de assistente multi-etapas para interfaces de terminal.
// Permite criar wizards com navegação para frente/trás, estado compartilhado entre etapas
// e indicador visual de progresso estilizado com Lipgloss.
package wizard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

// Erros retornados pelas operações do wizard.
var (
	// ErrNoSteps indica que o wizard não possui etapas registradas.
	ErrNoSteps = errors.New("wizard sem etapas registradas")
	// ErrInvalidStepIndex indica um índice de etapa fora dos limites válidos.
	ErrInvalidStepIndex = errors.New("índice de etapa inválido")
)

// statusEtapa representa o estado de uma etapa dentro do wizard.
type statusEtapa int

const (
	// etapaPendente indica uma etapa que ainda não foi executada com sucesso.
	etapaPendente statusEtapa = iota
	// etapaConcluida indica uma etapa executada com sucesso.
	etapaConcluida
	// etapaFalhou indica uma etapa que retornou erro.
	etapaFalhou
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
	out         io.Writer
	status      []statusEtapa
	// beforeState guarda o estado dos dados antes da execução de cada etapa,
	// permitindo desfazer os efeitos colaterais ao voltar.
	beforeState map[int]map[string]any
}

// New cria um novo Wizard vazio com dados inicializados.
func New() *Wizard {
	return &Wizard{
		steps:       make([]Step, 0),
		currentStep: 0,
		data:        make(map[string]any),
		title:       "Wizard",
		out:         os.Stdout,
		status:      make([]statusEtapa, 0),
		beforeState: make(map[int]map[string]any),
	}
}

// WithTitle define o titulo do wizard (API fluente).
func (w *Wizard) WithTitle(title string) *Wizard {
	w.title = title
	return w
}

// WithWriter define o destino da saida do wizard (API fluente).
// Por padrao a saida e o os.Stdout. Writers que nao sejam terminais recebem
// os blocos de progresso em sequencia, sem sequencias de controle do cursor.
func (w *Wizard) WithWriter(out io.Writer) *Wizard {
	if out != nil {
		w.out = out
	}
	return w
}

// AddStep adiciona uma nova etapa ao wizard (API fluente).
func (w *Wizard) AddStep(name, description string, runFunc func(data map[string]any) error) *Wizard {
	w.steps = append(w.steps, Step{
		Name:        name,
		Description: description,
		RunFunc:     runFunc,
	})
	w.status = append(w.status, etapaPendente)
	return w
}

// Data retorna o mapa de dados compartilhado entre etapas.
func (w *Wizard) Data() map[string]any {
	return w.data
}

// SetData define um valor no mapa de dados compartilhado.
func (w *Wizard) SetData(key string, value any) *Wizard {
	if w.data == nil {
		w.data = make(map[string]any)
	}
	w.data[key] = value
	return w
}

// GetString devolve o valor da chave como string.
// O segundo retorno indica se a chave existe e possui o tipo string.
func (w *Wizard) GetString(key string) (string, bool) {
	valor, ok := w.data[key]
	if !ok {
		return "", false
	}
	texto, ok := valor.(string)
	return texto, ok
}

// GetBool devolve o valor da chave como bool.
// O segundo retorno indica se a chave existe e possui o tipo bool.
func (w *Wizard) GetBool(key string) (bool, bool) {
	valor, ok := w.data[key]
	if !ok {
		return false, false
	}
	booleano, ok := valor.(bool)
	return booleano, ok
}

// GetInt devolve o valor da chave como int.
// O segundo retorno indica se a chave existe e possui o tipo int.
func (w *Wizard) GetInt(key string) (int, bool) {
	valor, ok := w.data[key]
	if !ok {
		return 0, false
	}
	inteiro, ok := valor.(int)
	return inteiro, ok
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

// CurrentStepName retorna o nome da etapa atual.
// Devolve uma string vazia quando nao ha etapa atual.
func (w *Wizard) CurrentStepName() string {
	if len(w.steps) == 0 || w.currentStep < 0 || w.currentStep >= len(w.steps) {
		return ""
	}
	return w.steps[w.currentStep].Name
}

// IsFirst informa se a etapa atual e a primeira do wizard.
func (w *Wizard) IsFirst() bool {
	return w.currentStep <= 0
}

// IsLast informa se a etapa atual e a ultima do wizard.
func (w *Wizard) IsLast() bool {
	return len(w.steps) == 0 || w.currentStep >= len(w.steps)-1
}

// Progress devolve o progresso do wizard como uma fracao entre 0 e 1,
// considerando apenas as etapas concluidas com sucesso.
func (w *Wizard) Progress() float64 {
	if len(w.steps) == 0 {
		return 0
	}
	concluidas := 0
	for _, estado := range w.status {
		if estado == etapaConcluida {
			concluidas++
		}
	}
	return float64(concluidas) / float64(len(w.steps))
}

// Back retrocede uma etapa. Retorna false se ja estiver na primeira.
//
// Ao retroceder, os efeitos colaterais da etapa abandonada sao desfeitos: os
// dados voltam ao estado anterior a execucao daquela etapa.
func (w *Wizard) Back() bool {
	if w.currentStep <= 0 {
		return false
	}

	// Desfaz os efeitos colaterais da etapa que esta sendo abandonada.
	if snapshot, ok := w.beforeState[w.currentStep]; ok {
		w.restoreData(snapshot)
		w.status[w.currentStep] = etapaPendente
	}

	w.currentStep--
	return true
}

// Forward avanca uma etapa. Retorna false se ja estiver na ultima ou se a
// etapa atual tiver falhado: nao e possivel avancar com uma etapa quebrada.
func (w *Wizard) Forward() bool {
	if len(w.steps) == 0 || w.currentStep >= len(w.steps)-1 {
		return false
	}
	if w.currentStep < len(w.status) && w.status[w.currentStep] == etapaFalhou {
		return false
	}
	w.currentStep++
	return true
}

// GoTo navega para uma etapa especifica pelo indice. Retorna erro se fora dos limites.
func (w *Wizard) GoTo(index int) error {
	if index < 0 || index >= len(w.steps) {
		return fmt.Errorf("%w: %d (total: %d)", ErrInvalidStepIndex, index, len(w.steps))
	}
	w.currentStep = index
	return nil
}

// Run executa as etapas sequencialmente a partir da etapa atual, exibindo o
// progresso na saida do wizard. Retorna erro caso alguma etapa falhe.
//
// A etapa inicial e a etapa corrente: um wizard que ja navegou ate a terceira
// etapa executa apenas dela em diante.
func (w *Wizard) Run() error {
	if len(w.steps) == 0 {
		return ErrNoSteps
	}

	w.write(renderTitle(w.title) + "\n\n")

	// O redesenho no lugar so faz sentido em um terminal interativo.
	noLugar := isTerminal(w.out)

	linhas := 0
	for i := w.currentStep; i < len(w.steps); i++ {
		w.currentStep = i

		// Apaga o bloco de progresso anterior para redesenhar no lugar.
		if noLugar && linhas > 0 {
			w.write(eraseSequence(linhas))
			linhas = 0
		}

		erro := w.runStepAt(i)

		// O progresso e desenhado mesmo quando a etapa falha, para que o
		// usuario saiba em qual etapa o wizard parou.
		bloco := RenderProgress(w.steps, i)
		w.write(bloco + "\n")
		linhas = countLines(bloco)

		if erro != nil {
			return erro
		}
	}

	// Remove o bloco final antes da mensagem de conclusao.
	if noLugar && linhas > 0 {
		w.write(eraseSequence(linhas))
	}
	w.write(renderComplete(w.title) + "\n")

	return nil
}

// RunStep executa apenas a etapa atual sem avancar automaticamente.
// Util para wizards interativos com navegacao manual.
//
// Ao final, o indicador de progresso e desenhado na saida do wizard, de modo
// que o usuario nao fica sem retorno visual.
func (w *Wizard) RunStep() error {
	if len(w.steps) == 0 {
		return ErrNoSteps
	}
	if w.currentStep < 0 || w.currentStep >= len(w.steps) {
		return fmt.Errorf("%w: %d (total: %d)", ErrInvalidStepIndex, w.currentStep, len(w.steps))
	}

	err := w.runStepAt(w.currentStep)

	// Mostra o progresso mesmo quando a etapa falha, para que o usuario
	// saiba em qual etapa o wizard parou.
	w.write(RenderProgress(w.steps, w.currentStep) + "\n")

	return err
}

// runStepAt executa a etapa de indice informado, atualizando o estado do wizard.
func (w *Wizard) runStepAt(index int) error {
	w.ensureStatus()

	// Guarda o estado dos dados antes da execucao para permitir desfazer
	// os efeitos colaterais quando o usuario voltar uma etapa.
	w.beforeState[index] = cloneData(w.data)

	etapa := w.steps[index]
	if etapa.RunFunc != nil {
		if err := etapa.RunFunc(w.data); err != nil {
			w.status[index] = etapaFalhou
			return fmt.Errorf("falha na etapa %q: %w", etapa.Name, err)
		}
	}

	w.status[index] = etapaConcluida
	return nil
}

// ensureStatus garante que o estado das etapas acompanhe a lista de etapas.
func (w *Wizard) ensureStatus() {
	for len(w.status) < len(w.steps) {
		w.status = append(w.status, etapaPendente)
	}
}

// restoreData substitui o conteudo dos dados compartilhados pelo conteudo do
// snapshot informado, preservando o mapa original.
func (w *Wizard) restoreData(snapshot map[string]any) {
	if w.data == nil {
		w.data = make(map[string]any)
	}
	for chave := range w.data {
		delete(w.data, chave)
	}
	for chave, valor := range snapshot {
		w.data[chave] = valor
	}
}

// cloneData devolve uma copia do mapa de dados compartilhados.
func cloneData(data map[string]any) map[string]any {
	copia := make(map[string]any, len(data))
	for chave, valor := range data {
		copia[chave] = valor
	}
	return copia
}

// write escreve na saida configurada do wizard.
func (w *Wizard) write(texto string) {
	if w.out == nil {
		w.out = os.Stdout
	}
	_, _ = w.out.Write([]byte(texto))
}

// isTerminal informa se o writer configurado e um terminal interativo.
func isTerminal(out io.Writer) bool {
	arquivo, ok := out.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(arquivo.Fd())
}

// eraseSequence devolve a sequencia ANSI que apaga as ultimas n linhas
// escritas, permitindo redesenhar o progresso no mesmo lugar.
func eraseSequence(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dA\x1b[J", n)
}

// countLines conta as linhas de um texto.
func countLines(texto string) int {
	if texto == "" {
		return 0
	}
	return strings.Count(texto, "\n") + 1
}

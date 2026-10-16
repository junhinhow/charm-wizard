package wizard

import (
	"bytes"
	"errors"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestRenderProgressIndiceInvalidoSemPanico garante que a função exportada
// RenderProgress não entra em pânico com um índice fora do range.
func TestRenderProgressIndiceInvalidoSemPanico(t *testing.T) {
	etapas := []Step{
		{Name: "primeira", Description: "Primeira etapa"},
		{Name: "segunda", Description: "Segunda etapa"},
	}

	indices := []int{5, 2, -1, -100, 1000}
	for _, indice := range indices {
		saida := RenderProgress(etapas, indice)
		if saida != "" {
			t.Errorf("índice %d: esperava string vazia, obtive %q", indice, saida)
		}
	}

	// A lista vazia também precisa ser tratada.
	if saida := RenderProgress(nil, 0); saida != "" {
		t.Errorf("esperava string vazia para lista vazia, obtive %q", saida)
	}
	if saida := RenderProgress([]Step{}, 0); saida != "" {
		t.Errorf("esperava string vazia para lista vazia, obtive %q", saida)
	}
}

// TestRenderProgressConteudo garante que o progresso traz as informações da
// etapa, os marcadores e a barra.
func TestRenderProgressConteudo(t *testing.T) {
	etapas := []Step{
		{Name: "validar", Description: "Validar configuracao"},
		{Name: "build", Description: "Compilar aplicacao"},
		{Name: "testar", Description: "Rodar suite de testes"},
	}

	saida := RenderProgress(etapas, 1)
	if saida == "" {
		t.Fatal("esperava saída não vazia")
	}
	if !strings.Contains(saida, "Etapa 2 de 3") {
		t.Errorf("esperava a informação da etapa: %q", saida)
	}
	if !strings.Contains(saida, "validar") || !strings.Contains(saida, "build") {
		t.Errorf("esperava os nomes das etapas: %q", saida)
	}
	if !strings.Contains(saida, "[>] build") {
		t.Errorf("esperava a etapa atual marcada: %q", saida)
	}
	if !strings.Contains(saida, "[*] validar") {
		t.Errorf("esperava a etapa concluída marcada: %q", saida)
	}
	if !strings.Contains(saida, "[ ] testar") {
		t.Errorf("esperava a etapa pendente marcada: %q", saida)
	}
	if !strings.Contains(saida, "66%") {
		t.Errorf("esperava o percentual de progresso: %q", saida)
	}
}

// TestRenderBarPercentualLimitado garante que o percentual nunca passa de
// 100%, mesmo com um índice fora do range.
func TestRenderBarPercentualLimitado(t *testing.T) {
	for current := -5; current <= 20; current++ {
		saida := renderBar(4, current)

		idx := strings.LastIndex(saida, " ")
		if idx < 0 {
			t.Fatalf("esperava o percentual na saída: %q", saida)
		}
		percentual, err := strconv.Atoi(strings.TrimSuffix(saida[idx+1:], "%"))
		if err != nil {
			t.Fatalf("percentual inválido em %q: %v", saida, err)
		}
		if percentual < 0 || percentual > 100 {
			t.Fatalf("percentual %d fora do intervalo para current=%d: %q", percentual, current, saida)
		}
	}

	// Com current >= total o percentual deve ser 100%, não 300%.
	if !strings.HasSuffix(renderBar(2, 5), "100%") {
		t.Fatalf("esperava 100%%, obtive %q", renderBar(2, 5))
	}

	// A barra tem sempre 30 colunas.
	if largura := lipgloss.Width(renderBar(4, 1)); largura != 30+1+len("25%") {
		t.Fatalf("largura inesperada da barra: %d", largura)
	}

	// Total inválido não pode provocar divisão por zero.
	if saida := renderBar(0, 0); saida == "" {
		t.Fatal("esperava uma barra vazia para total inválido")
	}
}

// TestTransicoesBackForward garante as transições de navegação.
func TestTransicoesBackForward(t *testing.T) {
	w := New()
	w.AddStep("a", "Etapa A", nil).
		AddStep("b", "Etapa B", nil).
		AddStep("c", "Etapa C", nil)

	if !w.IsFirst() {
		t.Fatal("a primeira etapa deveria ser a atual")
	}
	if w.IsLast() {
		t.Fatal("a primeira etapa não deveria ser a última")
	}
	if w.CurrentStepName() != "a" {
		t.Fatalf("esperava o nome da etapa atual, obtive %q", w.CurrentStepName())
	}

	if w.Back() {
		t.Fatal("Back na primeira etapa deveria retornar false")
	}
	if w.CurrentStep() != 0 {
		t.Fatalf("a etapa atual mudou para %d", w.CurrentStep())
	}

	if !w.Forward() {
		t.Fatal("Forward deveria retornar true")
	}
	if w.CurrentStep() != 1 {
		t.Fatalf("esperava a etapa 1, obtive %d", w.CurrentStep())
	}
	if w.IsFirst() {
		t.Fatal("não deveria mais ser a primeira etapa")
	}

	w.Forward()
	if w.CurrentStep() != 2 {
		t.Fatalf("esperava a etapa 2, obtive %d", w.CurrentStep())
	}
	if !w.IsLast() {
		t.Fatal("a terceira etapa deveria ser a última")
	}
	if w.Forward() {
		t.Fatal("Forward na última etapa deveria retornar false")
	}
	if w.CurrentStep() != 2 {
		t.Fatalf("a etapa atual mudou para %d", w.CurrentStep())
	}

	if !w.Back() {
		t.Fatal("Back deveria retornar true")
	}
	if w.CurrentStep() != 1 {
		t.Fatalf("esperava a etapa 1, obtive %d", w.CurrentStep())
	}
}

// TestForwardBloqueadoAposFalha garante que não é possível avançar quando a
// etapa atual falhou.
func TestForwardBloqueadoAposFalha(t *testing.T) {
	w := New()
	w.AddStep("a", "Etapa A", func(data map[string]any) error {
		return errors.New("falhou de propósito")
	})
	w.AddStep("b", "Etapa B", nil)

	if err := w.RunStep(); err == nil {
		t.Fatal("esperava erro da etapa")
	}

	if w.Forward() {
		t.Fatal("Forward não deveria avançar com uma etapa que falhou")
	}
	if w.CurrentStep() != 0 {
		t.Fatalf("a etapa atual mudou para %d", w.CurrentStep())
	}
}

// TestBackDesfazEfeitosColaterais garante que voltar uma etapa descarta os
// dados escritos pela etapa abandonada.
func TestBackDesfazEfeitosColaterais(t *testing.T) {
	w := New()
	w.AddStep("a", "Etapa A", func(data map[string]any) error {
		data["a"] = 1
		return nil
	})
	w.AddStep("b", "Etapa B", func(data map[string]any) error {
		data["b"] = 2
		return nil
	})

	if err := w.RunStep(); err != nil {
		t.Fatal(err)
	}
	if !w.Forward() {
		t.Fatal("Forward deveria retornar true")
	}
	if err := w.RunStep(); err != nil {
		t.Fatal(err)
	}
	if _, ok := w.GetInt("b"); !ok {
		t.Fatal("esperava a chave \"b\" após a execução da etapa")
	}

	if !w.Back() {
		t.Fatal("Back deveria retornar true")
	}
	if _, ok := w.GetInt("b"); ok {
		t.Fatal("Back deveria desfazer os efeitos colaterais da etapa abandonada")
	}
	if valor, ok := w.GetInt("a"); !ok || valor != 1 {
		t.Fatal("Back deveria preservar os dados da etapa anterior")
	}
	if w.CurrentStep() != 0 {
		t.Fatalf("esperava a etapa 0, obtive %d", w.CurrentStep())
	}
}

// TestGettersTipados garante os getters com valor e ok.
func TestGettersTipados(t *testing.T) {
	w := New().
		SetData("nome", "valor").
		SetData("ativo", true).
		SetData("contagem", 3).
		SetData("texto", "42")

	if valor, ok := w.GetString("nome"); !ok || valor != "valor" {
		t.Fatalf("GetString: esperava (\"valor\", true), obtive (%q, %v)", valor, ok)
	}
	if _, ok := w.GetString("ativo"); ok {
		t.Fatal("GetString deveria falhar para um valor de outro tipo")
	}
	if _, ok := w.GetString("inexistente"); ok {
		t.Fatal("GetString deveria falhar para uma chave inexistente")
	}

	if valor, ok := w.GetBool("ativo"); !ok || !valor {
		t.Fatalf("GetBool: esperava (true, true), obtive (%v, %v)", valor, ok)
	}
	if _, ok := w.GetBool("nome"); ok {
		t.Fatal("GetBool deveria falhar para um valor de outro tipo")
	}
	if _, ok := w.GetBool("inexistente"); ok {
		t.Fatal("GetBool deveria falhar para uma chave inexistente")
	}

	if valor, ok := w.GetInt("contagem"); !ok || valor != 3 {
		t.Fatalf("GetInt: esperava (3, true), obtive (%d, %v)", valor, ok)
	}
	if _, ok := w.GetInt("texto"); ok {
		t.Fatal("GetInt deveria falhar para um valor de outro tipo")
	}
	if _, ok := w.GetInt("inexistente"); ok {
		t.Fatal("GetInt deveria falhar para uma chave inexistente")
	}
}

// TestRunRespeitaNavegacao garante que Run começa na etapa atual, e não
// sempre na etapa zero.
func TestRunRespeitaNavegacao(t *testing.T) {
	var executadas []string
	registrar := func(nome string) func(map[string]any) error {
		return func(data map[string]any) error {
			executadas = append(executadas, nome)
			return nil
		}
	}

	w := New()
	w.AddStep("a", "Etapa A", registrar("a")).
		AddStep("b", "Etapa B", registrar("b")).
		AddStep("c", "Etapa C", registrar("c"))

	if err := w.GoTo(1); err != nil {
		t.Fatal(err)
	}
	if err := w.Run(); err != nil {
		t.Fatal(err)
	}

	if len(executadas) != 2 || executadas[0] != "b" || executadas[1] != "c" {
		t.Fatalf("esperava a execução das etapas b e c, obtive %v", executadas)
	}
	if w.CurrentStep() != 2 {
		t.Fatalf("esperava a etapa 2 ao final, obtive %d", w.CurrentStep())
	}
}

// TestRunExecutaEtapasEEscreveProgresso garante a execução completa e a
// escrita do progresso na saída do wizard.
func TestRunExecutaEtapasEEscreveProgresso(t *testing.T) {
	var buf bytes.Buffer

	w := New().WithTitle("Meu Wizard").WithWriter(&buf)
	w.AddStep("a", "Primeira etapa", func(data map[string]any) error {
		data["a"] = true
		return nil
	})
	w.AddStep("b", "Segunda etapa", func(data map[string]any) error {
		data["b"] = true
		return nil
	})

	if err := w.Run(); err != nil {
		t.Fatal(err)
	}

	saida := buf.String()
	if saida == "" {
		t.Fatal("Run não escreveu nada na saída")
	}
	for _, esperado := range []string{"Meu Wizard", "Etapa 1 de 2", "Etapa 2 de 2", "concluído com sucesso"} {
		if !strings.Contains(saida, esperado) {
			t.Errorf("esperava %q na saída: %q", esperado, saida)
		}
	}

	if valor, ok := w.GetBool("a"); !ok || !valor {
		t.Error("esperava a etapa \"a\" executada")
	}
	if valor, ok := w.GetBool("b"); !ok || !valor {
		t.Error("esperava a etapa \"b\" executada")
	}
	if w.Progress() != 1 {
		t.Fatalf("esperava progresso 1, obtive %v", w.Progress())
	}
	if !w.IsLast() {
		t.Error("após o Run a última etapa deveria ser a atual")
	}
}

// TestRunParaNaEtapaComErro garante que o Run interrompe na etapa que falhou.
func TestRunParaNaEtapaComErro(t *testing.T) {
	var buf bytes.Buffer
	erroOriginal := errors.New("falhou de propósito")
	executouSegunda := false

	w := New().WithWriter(&buf)
	w.AddStep("a", "Primeira etapa", func(data map[string]any) error {
		return erroOriginal
	})
	w.AddStep("b", "Segunda etapa", func(data map[string]any) error {
		executouSegunda = true
		return nil
	})

	err := w.Run()
	if err == nil {
		t.Fatal("esperava erro do Run")
	}
	if !errors.Is(err, erroOriginal) {
		t.Fatalf("esperava o erro original envolvido, obtive %v", err)
	}
	if executouSegunda {
		t.Fatal("a etapa seguinte não deveria ter sido executada")
	}
	if w.CurrentStep() != 0 {
		t.Fatalf("esperava parar na etapa 0, obtive %d", w.CurrentStep())
	}
}

// TestRunMostraProgressoDaEtapaQueFalhou garante que o usuário vê em qual
// etapa o wizard parou.
func TestRunMostraProgressoDaEtapaQueFalhou(t *testing.T) {
	var buf bytes.Buffer

	w := New().WithWriter(&buf)
	w.AddStep("a", "Primeira etapa", func(data map[string]any) error {
		return errors.New("boom")
	})
	w.AddStep("b", "Segunda etapa", nil)

	if err := w.Run(); err == nil {
		t.Fatal("esperava erro do Run")
	}
	if !strings.Contains(buf.String(), "Etapa 1 de 2") {
		t.Fatalf("esperava o progresso da etapa que falhou: %q", buf.String())
	}
}

// TestRunStepRenderizaProgresso garante que RunStep não é silencioso.
func TestRunStepRenderizaProgresso(t *testing.T) {
	var buf bytes.Buffer

	w := New().WithWriter(&buf)
	w.AddStep("a", "Executar agora", func(data map[string]any) error {
		data["executado"] = true
		return nil
	})

	if err := w.RunStep(); err != nil {
		t.Fatal(err)
	}

	saida := buf.String()
	if saida == "" {
		t.Fatal("RunStep não renderizou nada")
	}
	if !strings.Contains(saida, "Etapa 1 de 1") {
		t.Fatalf("esperava o indicador de progresso na saída: %q", saida)
	}
	if valor, ok := w.GetBool("executado"); !ok || !valor {
		t.Error("esperava a etapa executada")
	}
}

// TestRunStepComErroRenderizaProgresso garante que o progresso é desenhado
// mesmo quando a etapa falha.
func TestRunStepComErroRenderizaProgresso(t *testing.T) {
	var buf bytes.Buffer

	w := New().WithWriter(&buf)
	w.AddStep("a", "Executar agora", func(data map[string]any) error {
		return errors.New("boom")
	})

	if err := w.RunStep(); err == nil {
		t.Fatal("esperava erro da etapa")
	}
	if !strings.Contains(buf.String(), "Etapa 1 de 1") {
		t.Fatalf("esperava o indicador de progresso na saída: %q", buf.String())
	}
}

// TestIsFirstIsLastCurrentStepNameProgress garante a API de consulta.
func TestIsFirstIsLastCurrentStepNameProgress(t *testing.T) {
	w := New()
	if w.CurrentStepName() != "" {
		t.Fatalf("esperava nome vazio sem etapas, obtive %q", w.CurrentStepName())
	}
	if w.Progress() != 0 {
		t.Fatalf("esperava progresso 0, obtive %v", w.Progress())
	}
	if !w.IsFirst() {
		t.Error("sem etapas a posição atual é a primeira")
	}
	if !w.IsLast() {
		t.Error("sem etapas a posição atual é a última")
	}

	w.AddStep("a", "Etapa A", func(data map[string]any) error {
		data["a"] = true
		return nil
	})
	w.AddStep("b", "Etapa B", nil)

	if w.Progress() != 0 {
		t.Fatalf("esperava progresso 0 antes de executar, obtive %v", w.Progress())
	}
	if err := w.RunStep(); err != nil {
		t.Fatal(err)
	}
	if w.Progress() != 0.5 {
		t.Fatalf("esperava progresso 0.5, obtive %v", w.Progress())
	}
	if w.CurrentStepName() != "a" {
		t.Fatalf("esperava o nome \"a\", obtive %q", w.CurrentStepName())
	}
}

// TestGoToInvalido garante a validação do índice.
func TestGoToInvalido(t *testing.T) {
	w := New()
	w.AddStep("a", "Etapa A", nil)

	if err := w.GoTo(1); !errors.Is(err, ErrInvalidStepIndex) {
		t.Fatalf("esperava ErrInvalidStepIndex, obtive %v", err)
	}
	if err := w.GoTo(-1); !errors.Is(err, ErrInvalidStepIndex) {
		t.Fatalf("esperava ErrInvalidStepIndex, obtive %v", err)
	}
	if err := w.GoTo(0); err != nil {
		t.Fatalf("esperava sucesso no índice 0, obtive %v", err)
	}
}

// TestWizardSemEtapas garante o erro para wizards sem etapas.
func TestWizardSemEtapas(t *testing.T) {
	w := New()
	if err := w.Run(); !errors.Is(err, ErrNoSteps) {
		t.Fatalf("esperava ErrNoSteps, obtive %v", err)
	}
	if err := w.RunStep(); !errors.Is(err, ErrNoSteps) {
		t.Fatalf("esperava ErrNoSteps, obtive %v", err)
	}
}

// TestComentarioDoPacoteSegueConvencaoGo garante que o comentário do pacote
// começa com "Package wizard", como exige a convenção do Go.
func TestComentarioDoPacoteSegueConvencaoGo(t *testing.T) {
	arquivo, err := parser.ParseFile(token.NewFileSet(), "wizard.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if arquivo.Doc == nil {
		t.Fatal("wizard.go não tem comentário de pacote")
	}

	texto := arquivo.Doc.Text()
	if !strings.HasPrefix(texto, "Package wizard") {
		t.Fatalf("o comentário do pacote deve começar com \"Package wizard\": %q", texto)
	}
}

// TestEraseSequence garante a sequência de apagamento do bloco de progresso.
func TestEraseSequence(t *testing.T) {
	if eraseSequence(0) != "" {
		t.Fatal("esperava sequência vazia para zero linhas")
	}
	if eraseSequence(-1) != "" {
		t.Fatal("esperava sequência vazia para número negativo de linhas")
	}
	if got := eraseSequence(3); got != "\x1b[3A\x1b[J" {
		t.Fatalf("sequência inesperada: %q", got)
	}
}

// TestCountLines garante a contagem de linhas do bloco de progresso.
func TestCountLines(t *testing.T) {
	if countLines("") != 0 {
		t.Fatal("esperava zero linhas para texto vazio")
	}
	if countLines("uma") != 1 {
		t.Fatal("esperava uma linha")
	}
	if countLines("uma\nduas\ntres") != 3 {
		t.Fatal("esperava três linhas")
	}
}

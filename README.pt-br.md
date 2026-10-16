# charm-wizard

Framework de wizard multi-etapas para interfaces de terminal, construído com [Lipgloss](https://charm.land/lipgloss).

> [Read in English](README.md)

## Funcionalidades

- Execução sequencial de múltiplas etapas com acompanhamento de progresso
- Navegação para frente e para trás entre etapas
- Estado compartilhado entre etapas via `map[string]any`
- Indicador visual de progresso com pontos coloridos e barra de progresso
- Getters tipados para os dados compartilhados (`GetString`, `GetBool`, `GetInt`)
- API fluente para construção fácil do wizard
- Saída estilizada com Lipgloss (verde = concluída, amarelo = atual, cinza = pendente)

O wizard é seguro por padrão: `RenderProgress` nunca entra em pânico com um
índice inválido, `Forward()` recusa avançar quando a etapa atual falhou e
`Back()` desfaz os efeitos colaterais da etapa abandonada (os dados
compartilhados voltam ao estado anterior à execução daquela etapa).

O `Run()` começa pela etapa **atual**, de modo que um wizard que já navegou
até a terceira etapa não reexecuta as duas primeiras. O indicador de progresso
é redesenhado no lugar quando a saída é um terminal e é escrito no writer
configurado nos demais casos.

## Instalação

```bash
go get github.com/junhinhow/charm-wizard
```

## Uso

Exemplo de wizard de deploy:

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
		AddStep("validar", "Validar configuração", func(data map[string]any) error {
			fmt.Println("  Verificando arquivos de configuração...")
			data["config_valida"] = true
			return nil
		}).
		AddStep("build", "Compilar aplicação", func(data map[string]any) error {
			fmt.Println("  Compilando projeto...")
			data["artefato"] = "app-v1.2.0.tar.gz"
			return nil
		}).
		AddStep("testes", "Executar suíte de testes", func(data map[string]any) error {
			artefato, ok := w.GetString("artefato")
			if !ok {
				return fmt.Errorf("artefato ausente")
			}
			fmt.Printf("  Testando %s...\n", artefato)
			data["testes_ok"] = true
			return nil
		}).
		AddStep("deploy", "Publicar em produção", func(data map[string]any) error {
			artefato, ok := w.GetString("artefato")
			if !ok {
				return fmt.Errorf("artefato ausente")
			}
			fmt.Printf("  Publicando %s em produção...\n", artefato)
			return nil
		})

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Os getters tipados devolvem `(valor, ok)`, portanto uma chave ausente ou com
tipo inesperado nunca provoca pânico — ao contrário de uma asserção de tipo
direta sobre o mapa compartilhado.

## API

### Wizard

| Método | Descrição |
|--------|-----------|
| `New()` | Cria um novo wizard |
| `WithTitle(título)` | Define o título do wizard |
| `WithWriter(w)` | Define o writer de saída (por padrão `os.Stdout`) |
| `AddStep(nome, desc, fn)` | Adiciona uma etapa com função de execução |
| `SetData(chave, valor)` | Pré-define dados compartilhados |
| `Data()` | Retorna o mapa de dados compartilhados |
| `GetString(chave)` | Obtém um valor compartilhado como `string` (`valor, ok`) |
| `GetBool(chave)` | Obtém um valor compartilhado como `bool` (`valor, ok`) |
| `GetInt(chave)` | Obtém um valor compartilhado como `int` (`valor, ok`) |
| `Run()` | Executa as etapas sequencialmente, a partir da etapa atual |
| `RunStep()` | Executa apenas a etapa atual e renderiza o progresso |
| `Back()` | Navega para a etapa anterior (desfaz seus efeitos colaterais) |
| `Forward()` | Navega para a próxima etapa (recusado se a atual falhou) |
| `GoTo(índice)` | Pula para uma etapa específica |
| `CurrentStep()` | Retorna o índice da etapa atual |
| `CurrentStepName()` | Retorna o nome da etapa atual |
| `TotalSteps()` | Retorna o número total de etapas |
| `Steps()` | Retorna uma cópia das etapas registradas |
| `IsFirst()` | Indica se a etapa atual é a primeira |
| `IsLast()` | Indica se a etapa atual é a última |
| `Progress()` | Progresso como fração entre 0 e 1 |

### Progresso

| Função | Descrição |
|--------|-----------|
| `RenderProgress(steps, índice)` | Renderiza o indicador de progresso (string vazia para índice inválido) |

### Erros

| Erro | Descrição |
|------|-----------|
| `ErrNoSteps` | O wizard não possui etapas registradas |
| `ErrInvalidStepIndex` | O índice da etapa está fora dos limites |

## Licença

MIT

# charm-wizard

Framework de wizard multi-etapas para interfaces de terminal, construido com [Lipgloss](https://charm.land/lipgloss).

> [Read in English](README.md)

## Funcionalidades

- Execucao sequencial de multiplas etapas com acompanhamento de progresso
- Navegacao para frente e para tras entre etapas
- Estado compartilhado entre etapas via `map[string]any`
- Indicador visual de progresso com pontos coloridos e barra de progresso
- API fluente para construcao facil do wizard
- Saida estilizada com Lipgloss (verde = concluida, amarelo = atual, cinza = pendente)

## Instalacao

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
		AddStep("validar", "Validar configuracao", func(data map[string]any) error {
			fmt.Println("  Verificando arquivos de configuracao...")
			data["config_valida"] = true
			return nil
		}).
		AddStep("build", "Compilar aplicacao", func(data map[string]any) error {
			fmt.Println("  Compilando projeto...")
			data["artefato"] = "app-v1.2.0.tar.gz"
			return nil
		}).
		AddStep("testes", "Executar suite de testes", func(data map[string]any) error {
			artefato := data["artefato"].(string)
			fmt.Printf("  Testando %s...\n", artefato)
			data["testes_ok"] = true
			return nil
		}).
		AddStep("deploy", "Publicar em producao", func(data map[string]any) error {
			artefato := data["artefato"].(string)
			fmt.Printf("  Publicando %s em producao...\n", artefato)
			return nil
		})

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## API

### Wizard

| Metodo | Descricao |
|--------|-----------|
| `New()` | Cria um novo wizard |
| `WithTitle(titulo)` | Define o titulo do wizard |
| `AddStep(nome, desc, fn)` | Adiciona uma etapa com funcao de execucao |
| `SetData(chave, valor)` | Pre-define dados compartilhados |
| `Data()` | Retorna o mapa de dados compartilhados |
| `Run()` | Executa todas as etapas sequencialmente |
| `RunStep()` | Executa apenas a etapa atual |
| `Back()` | Navega para a etapa anterior |
| `Forward()` | Navega para a proxima etapa |
| `GoTo(indice)` | Pula para uma etapa especifica |
| `CurrentStep()` | Retorna o indice da etapa atual |
| `TotalSteps()` | Retorna o numero total de etapas |

### Progresso

| Funcao | Descricao |
|--------|-----------|
| `RenderProgress(steps, indice)` | Renderiza o indicador de progresso |

## Licenca

MIT

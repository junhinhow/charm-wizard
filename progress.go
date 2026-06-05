package wizard

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// Estilos Lipgloss para o indicador de progresso
var (
	// Estilo para etapas concluidas (verde)
	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	// Estilo para a etapa atual (amarelo)
	currentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true)

	// Estilo para etapas pendentes (cinza)
	remainingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	// Estilo para o titulo do wizard
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED"))

	// Estilo para mensagem de conclusao
	completeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	// Estilo para a barra de progresso preenchida
	barFilledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	// Estilo para a barra de progresso vazia
	barEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))

	// Estilo para informacao da etapa
	stepInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))
)

// RenderProgress renderiza o indicador visual de progresso para as etapas.
// Mostra pontos para cada etapa (concluida/atual/pendente) e uma barra de progresso.
func RenderProgress(steps []Step, currentIndex int) string {
	if len(steps) == 0 {
		return ""
	}

	var sb strings.Builder

	// Linha de informacao: "Etapa 2 de 4: Configuracao"
	etapaAtual := steps[currentIndex]
	info := fmt.Sprintf("Step %d of %d: %s", currentIndex+1, len(steps), etapaAtual.Description)
	sb.WriteString(stepInfoStyle.Render(info))
	sb.WriteString("\n")

	// Pontos indicadores de etapa
	sb.WriteString(renderDots(steps, currentIndex))
	sb.WriteString("\n")

	// Barra de progresso
	sb.WriteString(renderBar(len(steps), currentIndex))

	return sb.String()
}

// renderDots renderiza os pontos indicadores de etapas.
// Concluida: verde [*], Atual: amarelo [>], Pendente: cinza [ ]
func renderDots(steps []Step, currentIndex int) string {
	var partes []string

	for i, etapa := range steps {
		var ponto string
		nome := etapa.Name

		switch {
		case i < currentIndex:
			// Etapa concluida
			ponto = completedStyle.Render(fmt.Sprintf("[*] %s", nome))
		case i == currentIndex:
			// Etapa atual
			ponto = currentStyle.Render(fmt.Sprintf("[>] %s", nome))
		default:
			// Etapa pendente
			ponto = remainingStyle.Render(fmt.Sprintf("[ ] %s", nome))
		}

		partes = append(partes, ponto)
	}

	separador := remainingStyle.Render(" --- ")
	return strings.Join(partes, separador)
}

// renderBar renderiza uma barra de progresso visual.
func renderBar(total, current int) string {
	largura := 30
	preenchido := 0
	if total > 0 {
		preenchido = ((current + 1) * largura) / total
	}
	if preenchido > largura {
		preenchido = largura
	}

	barraCheia := barFilledStyle.Render(strings.Repeat("█", preenchido))
	barraVazia := barEmptyStyle.Render(strings.Repeat("░", largura-preenchido))
	percentual := ((current + 1) * 100) / total

	return fmt.Sprintf("%s%s %d%%", barraCheia, barraVazia, percentual)
}

// renderTitle renderiza o titulo estilizado do wizard.
func renderTitle(title string) string {
	return titleStyle.Render(title)
}

// renderComplete renderiza a mensagem de conclusao.
func renderComplete(title string) string {
	return completeStyle.Render(fmt.Sprintf("✓ %s completed successfully!", title))
}

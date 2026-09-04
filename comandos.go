package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	reSeparador = regexp.MustCompile(`(?i)\s*(?:&|\+|,|/|\se\s|\scom\s|\sx\s)\s*`)
	rePlacar    = regexp.MustCompile(`(?i)^(?:j?(\d+)\s+)?(\d+)\s*[x×:-]\s*(\d+)$`)
)

func separaNomes(s string) []string {
	partes := reSeparador.Split(strings.TrimSpace(s), -1)
	var out []string
	for _, p := range partes {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 1 {
		campos := strings.Fields(out[0])
		if len(campos) == 2 {
			return campos
		}
	}
	return out
}

func (b *Bot) Executar(t *Torneio, texto string) (string, bool) {
	texto = strings.TrimSpace(texto)
	if !strings.HasPrefix(texto, "!") {
		return "", false
	}
	campos := strings.Fields(texto[1:])
	if len(campos) == 0 {
		return "", false
	}
	cmd := strings.ToLower(campos[0])
	resto := strings.TrimSpace(strings.TrimPrefix(texto[1:], campos[0]))

	switch cmd {
	case "ajuda", "help", "comandos", "menu":
		return RenderAjuda(), false

	case "dupla", "inscrever", "entrar":
		nomes := separaNomes(resto)
		if len(nomes) != 2 {
			return "Manda assim: `!dupla João & Maria`", false
		}
		d, err := t.AddDupla(nomes[0], nomes[1])
		if err != nil {
			return "⚠️ " + err.Error(), false
		}
		return fmt.Sprintf("✅ Dupla %d inscrita: *%s*\nTotal: %d duplas.", len(t.Duplas), d.Nome(), len(t.Duplas)), true

	case "duplas", "lista", "inscritos":
		return t.RenderDuplas(), false

	case "remover", "tirar":
		if resto == "" {
			return "Manda `!remover João`", false
		}
		d, err := t.RemoveDupla(resto)
		if err != nil {
			return "⚠️ " + err.Error(), false
		}
		return fmt.Sprintf("🗑️ Removi a dupla *%s*.", d.Nome()), true

	case "ate", "até", "alvo":
		f := strings.ToLower(strings.TrimSpace(resto))
		if f == "" {
			if t.Alvo == 0 {
				return "Ainda não sei até quantos pontos vão os jogos — eu anoto sozinho no primeiro `!placar`.\nOu define agora: `!ate 17`.", false
			}
			return fmt.Sprintf("Os jogos vão até *%d* pontos. Pra mudar: `!ate 21`. Pra desligar a conferência: `!ate livre`.", t.Alvo), false
		}
		if f == "livre" || f == "0" {
			t.Alvo = 0
			return "🔓 Conferência desligada. Agora eu aceito qualquer placar.", true
		}
		n, err := strconv.Atoi(f)
		if err != nil {
			return "Manda `!ate 17` (ou `!ate livre` pra não conferir).", false
		}
		if err := t.DefinirAlvo(n); err != nil {
			return "⚠️ " + err.Error(), false
		}
		return fmt.Sprintf("🎯 Anotado: os jogos vão até *%d* pontos. Vou conferir os placares por isso.", n), true

	case "sortear", "sorteio", "gerar", "comecar", "começar":
		if t.Sorteado {
			return "A tabela já está gerada. Se quiser refazer tudo: `!zerar CONFIRMA`.", false
		}
		if err := t.Sortear(); err != nil {
			return "⚠️ " + err.Error(), false
		}
		var b2 strings.Builder
		b2.WriteString("🎲 *TABELA GERADA!*\n\n")
		fmt.Fprintf(&b2, "Pontos corridos, turno e returno: %d duplas, %d jogos em %d rodadas.\n", len(t.Duplas), len(t.Jogos), t.Rodadas())
		fmt.Fprintf(&b2, "Cada dupla enfrenta cada uma das outras *2 vezes*. Vitória vale %d pontos; campeã é quem somar mais no fim.\n\n%s", pontosPorVitoria, t.RenderJogos())
		return b2.String(), true

	case "campeonato", "chaves", "chave", "copa", "rodadas":
		return t.RenderCampeonato(), false

	case "tabela", "classificacao", "classificação":
		return t.RenderTabela(), false

	case "jogos", "proximo", "próximo", "agenda":
		return t.RenderJogos(), false

	case "placar", "resultado", "jogo":
		if !t.Sorteado {
			return "Gera a tabela primeiro: `!sortear`.", false
		}
		m := rePlacar.FindStringSubmatch(strings.TrimSpace(resto))
		if m == nil || m[1] == "" {
			return "Manda o *ID do jogo* junto: `!placar 7 21x18` (jogo J7).\n\n" + t.RenderJogos(), false
		}
		id, _ := strconv.Atoi(m[1])
		a, _ := strconv.Atoi(m[2])
		bb, _ := strconv.Atoi(m[3])
		alvoAntes := t.Alvo
		j, corrigido, err := t.RegistrarPlacar(id, a, bb)
		if err != nil {
			return "⚠️ " + err.Error(), false
		}
		t.Atualizar()

		titulo := "registrado"
		if corrigido {
			titulo = "corrigido"
		}
		var out strings.Builder
		fmt.Fprintf(&out, "📝 *J%d %s*\n%s %d x %d %s\n🏅 %s leva.",
			j.ID, titulo, t.NomeDupla(j.A), j.PlacarA, j.PlacarB, t.NomeDupla(j.B), t.NomeDupla(j.Vencedor()))
		if alvoAntes == 0 && t.Alvo != 0 {
			fmt.Fprintf(&out, "\n\n🎯 Anotei que os jogos vão até *%d* pontos — vou conferir os próximos por isso. Se não for, manda `!ate <n>` ou `!ate livre`.", t.Alvo)
		}
		out.WriteString("\n\nManda a foto da dupla comemorando que eu grudo nesse jogo. 📸")
		out.WriteString("\n\n")
		out.WriteString(t.RenderTabela())
		if t.Campea == 0 {
			if p := t.ProximoJogo(); p != nil {
				fmt.Fprintf(&out, "\n\n▶ Próximo: %s", t.LinhaJogo(p))
			}
		}
		return out.String(), true

	case "desfazer", "corrigir":
		j, err := t.Desfazer()
		if err != nil {
			return "⚠️ " + err.Error(), false
		}
		t.Atualizar()
		return fmt.Sprintf("↩️ Desfiz o resultado do J%d. Manda de novo: `!placar %d 21x18`", j.ID, j.ID), true

	case "zerar", "reset":
		if strings.ToUpper(strings.TrimSpace(resto)) != "CONFIRMA" {
			return "Isso apaga duplas, tabela e resultados. Se for isso mesmo: `!zerar CONFIRMA`", false
		}
		*t = *NovoTorneio(t.Chat)
		return "🧹 Zerei tudo. Comece cadastrando: `!dupla João & Maria`", true
	}

	return "", false
}

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const separadorLegenda = "\x1e"

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
		return strings.Fields(out[0])
	}
	return out
}

const prefixo = "!cup"

// ehComando diz se a mensagem é pra copa: todo comando começa com !cup.
func ehComando(texto string) bool {
	campos := strings.Fields(texto)
	return len(campos) > 0 && strings.ToLower(campos[0]) == prefixo
}

func (b *Bot) Executar(t *Torneio, texto string) (string, bool) {
	texto = strings.TrimSpace(texto)
	if !ehComando(texto) {
		return "", false
	}
	texto = strings.TrimSpace(texto[len(prefixo):])
	campos := strings.Fields(texto)
	if len(campos) == 0 {
		return RenderAjuda(), false
	}
	cmd := strings.ToLower(campos[0])
	resto := strings.TrimSpace(strings.TrimPrefix(texto, campos[0]))

	switch cmd {
	case "ajuda", "help", "comandos", "menu":
		return RenderAjuda(), false

	case "time", "inscrever", "entrar":
		nomes := separaNomes(resto)
		if len(nomes) == 0 {
			return "Manda assim: `!cup time João & Maria & Pedro`", false
		}
		e, err := t.AddTime(nomes)
		if err != nil {
			return "⚠️ " + err.Error(), false
		}
		return fmt.Sprintf("✅ Time %d inscrito: *%s*\nTotal: %d times.", len(t.Times), e.Nome(), len(t.Times)), true

	case "times", "lista", "inscritos":
		return t.RenderTimes(), false

	case "remover", "tirar":
		if resto == "" {
			return "Manda `!cup remover João`", false
		}
		e, err := t.RemoveTime(resto)
		if err != nil {
			return "⚠️ " + err.Error(), false
		}
		return fmt.Sprintf("🗑️ Removi o time *%s*.", e.Nome()), true

	case "ate", "até", "alvo":
		f := strings.ToLower(strings.TrimSpace(resto))
		if f == "" {
			if t.Alvo == 0 {
				return "Ainda não sei até quantos pontos vão os jogos — eu anoto sozinho no primeiro `!cup placar`.\nOu define agora: `!cup ate 17`.", false
			}
			return fmt.Sprintf("Os jogos vão até *%d* pontos. Pra mudar: `!cup ate 21`. Pra desligar a conferência: `!cup ate livre`.", t.Alvo), false
		}
		if f == "livre" || f == "0" {
			t.Alvo = 0
			return "🔓 Conferência desligada. Agora eu aceito qualquer placar.", true
		}
		n, err := strconv.Atoi(f)
		if err != nil {
			return "Manda `!cup ate 17` (ou `!cup ate livre` pra não conferir).", false
		}
		if err := t.DefinirAlvo(n); err != nil {
			return "⚠️ " + err.Error(), false
		}
		return fmt.Sprintf("🎯 Anotado: os jogos vão até *%d* pontos. Vou conferir os placares por isso.", n), true

	case "sortear", "sorteio", "gerar", "comecar", "começar":
		if t.Sorteado {
			return "A tabela já está gerada. Se quiser refazer tudo: `!cup zerar CONFIRMA`.", false
		}
		if err := t.Sortear(); err != nil {
			return "⚠️ " + err.Error(), false
		}
		var b2 strings.Builder
		b2.WriteString("🎲 *TABELA GERADA!*\n\n")
		fmt.Fprintf(&b2, "Pontos corridos, turno e returno: %d times, %d jogos em %d rodadas.\n", len(t.Times), len(t.Jogos), t.Rodadas())
		fmt.Fprintf(&b2, "Cada time enfrenta cada um dos outros *2 vezes*. Vitória vale %d pontos; campeão é quem somar mais no fim.", pontosPorVitoria)
		b2.WriteString(separadorLegenda)
		b2.WriteString(t.RenderJogos())
		return b2.String(), true

	case "campeonato", "chaves", "chave", "copa", "rodadas":
		return t.RenderCampeonato(), false

	case "tabela", "classificacao", "classificação":
		return t.RenderTabela(), false

	case "jogos", "proximo", "próximo", "agenda":
		return t.RenderJogos(), false

	case "placar", "resultado", "jogo":
		if !t.Sorteado {
			return "Gera a tabela primeiro: `!cup sortear`.", false
		}
		m := rePlacar.FindStringSubmatch(strings.TrimSpace(resto))
		if m == nil || m[1] == "" {
			return "Manda o *ID do jogo* junto: `!cup placar 7 21x18` (jogo J7).\n\n" + t.RenderJogos(), false
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
			j.ID, titulo, t.NomeTime(j.A), j.PlacarA, j.PlacarB, t.NomeTime(j.B), t.NomeTime(j.Vencedor()))
		if alvoAntes == 0 && t.Alvo != 0 {
			fmt.Fprintf(&out, "\n\n🎯 Anotei que os jogos vão até *%d* pontos — vou conferir os próximos por isso. Se não for, manda `!cup ate <n>` ou `!cup ate livre`.", t.Alvo)
		}
		out.WriteString("\n\nManda a foto do time comemorando que eu grudo nesse jogo. 📸")
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
		return fmt.Sprintf("↩️ Desfiz o resultado do J%d. Manda de novo: `!cup placar %d 21x18`", j.ID, j.ID), true

	case "zerar", "reset":
		if strings.ToUpper(strings.TrimSpace(resto)) != "CONFIRMA" {
			return "Isso apaga times, tabela e resultados. Se for isso mesmo: `!cup zerar CONFIRMA`", false
		}
		*t = *NovoTorneio(t.Chat)
		return "🧹 Zerei tudo. Comece cadastrando: `!cup time João & Maria & Pedro`", true
	}

	return "", false
}

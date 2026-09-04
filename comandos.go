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

	case "formato":
		f := strings.ToLower(strings.TrimSpace(resto))
		if f != "sets" && f != "pontos" {
			return "Use `!formato sets` (melhor de 3) ou `!formato pontos` (set único, ex: 21x18).\nAtual: *" + t.Formato + "*", false
		}
		t.Formato = f
		if f == "sets" {
			return "📐 Formato: *melhor de 3 sets*. Placares válidos: 2x0, 2x1, 1x2, 0x2.", true
		}
		return "📐 Formato: *set único*. Manda o placar em pontos, ex: `!placar 21x18`.", true

	case "sortear", "sorteio", "sortear!":
		n := 0
		if resto != "" {
			if v, err := strconv.Atoi(strings.Fields(resto)[0]); err == nil {
				n = v
			}
		}
		if t.Sorteado {
			return "Já sorteei. Se quiser refazer tudo: `!zerar CONFIRMA`.", false
		}
		if err := t.Sortear(n); err != nil {
			return "⚠️ " + err.Error(), false
		}
		var b2 strings.Builder
		b2.WriteString("🎲 *SORTEIO FEITO!*\n")
		for g := range t.Grupos {
			fmt.Fprintf(&b2, "\n*Grupo %s*\n", letraGrupo(g))
			for _, id := range t.Grupos[g] {
				fmt.Fprintf(&b2, "• %s\n", t.NomeDupla(id))
			}
		}
		fmt.Fprintf(&b2, "\n%d jogos na fase de grupos. As 2 melhores de cada grupo vão pro mata-mata.\n\n%s", len(t.JogosDaFaseDeGrupos()), t.RenderJogos())
		return b2.String(), true

	case "chaves", "chave", "bracket", "copa":
		return t.RenderChaves(), false

	case "tabela", "classificacao", "classificação":
		return t.RenderTabela(), false

	case "jogos", "proximo", "próximo", "agenda":
		return t.RenderJogos(), false

	case "placar", "resultado", "jogo":
		if !t.Sorteado {
			return "Sorteia primeiro: `!sortear`.", false
		}
		m := rePlacar.FindStringSubmatch(strings.TrimSpace(resto))
		if m == nil {
			return "Manda `!placar 2x1` (jogo da vez) ou `!placar 7 2x1` (jogo J7).", false
		}
		id := 0
		if m[1] != "" {
			id, _ = strconv.Atoi(m[1])
		} else {
			p := t.ProximoJogo()
			if p == nil {
				return "Não tem jogo pendente pra registrar.", false
			}
			id = p.ID
		}
		a, _ := strconv.Atoi(m[2])
		bb, _ := strconv.Atoi(m[3])
		j, err := t.RegistrarPlacar(id, a, bb)
		if err != nil {
			return "⚠️ " + err.Error(), false
		}
		antesMataMata := t.MataMata
		t.Atualizar()

		var out strings.Builder
		fmt.Fprintf(&out, "📝 *J%d registrado*\n%s %d x %d %s\n🏅 %s leva.",
			j.ID, t.NomeDupla(j.A), j.PlacarA, j.PlacarB, t.NomeDupla(j.B), t.NomeDupla(j.Vencedor()))
		out.WriteString("\n\nManda a foto da dupla comemorando que eu grudo nesse jogo. 📸")

		if !antesMataMata && t.MataMata {
			out.WriteString("\n\n🔔 *FASE DE GRUPOS ENCERRADA!*\n\n")
			out.WriteString(t.RenderChaves())
		} else if t.Campea == 0 {
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
		return fmt.Sprintf("↩️ Desfiz o resultado do J%d. Manda de novo: `!placar %d 2x1`", j.ID, j.ID), true

	case "zerar", "reset":
		if strings.ToUpper(strings.TrimSpace(resto)) != "CONFIRMA" {
			return "Isso apaga duplas, grupos e resultados. Se for isso mesmo: `!zerar CONFIRMA`", false
		}
		*t = *NovoTorneio(t.Chat)
		return "🧹 Zerei tudo. Comece cadastrando: `!dupla João & Maria`", true
	}

	return "", false
}

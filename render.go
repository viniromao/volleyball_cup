package main

import (
	"fmt"
	"sort"
	"strings"
)

func corta(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func (t *Torneio) LinhaJogo(j *Jogo) string {
	a := t.NomeTime(j.A)
	b := t.NomeTime(j.B)
	if !j.Jogado {
		return fmt.Sprintf("J%d  %s  x  %s", j.ID, a, b)
	}
	foto := ""
	if j.Foto != "" {
		foto = " 📸"
	}
	return fmt.Sprintf("J%d  %s *%d x %d* %s  ✅ %s%s", j.ID, a, j.PlacarA, j.PlacarB, b, t.NomeTime(j.Vencedor()), foto)
}

func RenderAjuda() string {
	return strings.Join([]string{
		"🏐 *COPA DE VÔLEI — PONTOS CORRIDOS*",
		"",
		"*Antes de começar*",
		"`!cup time João & Maria & Pedro` — cadastra um time, com quantas pessoas quiser (separa com `&`, `,` ou `e`; sem separador, cada palavra é uma pessoa)",
		"`!cup times` — lista os times inscritos",
		"`!cup remover João` — tira o time do João",
		"",
		"`!cup ate 17` — até quantos pontos vai cada jogo",
		"",
		"*Gerar a tabela*",
		"`!cup sortear` — todos contra todos, ida e volta",
		"",
		"*Durante a copa*",
		"`!cup jogos` — o que falta jogar",
		"`!cup placar 7 21x18` — placar do jogo J7 (sempre pelo ID)",
		"`!cup placar 7 21x15` — manda de novo pra corrigir o J7",
		"`!cup tabela` — a classificação",
		"`!cup campeonato` — tabela + todas as rodadas",
		"`!cup desfazer` — apaga o último resultado",
		"",
		"*Foto*",
		"manda a foto/sticker do time comemorando logo depois do placar (ou com `!cup placar 7 21x18` na legenda) que eu grudo no jogo.",
		"",
		"`!cup zerar CONFIRMA` — recomeça tudo do zero",
	}, "\n")
}

func (t *Torneio) RenderTimes() string {
	if len(t.Times) == 0 {
		return "Nenhum time inscrito ainda. Manda `!cup time Fulano & Ciclano & Beltrano`."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "🏐 *TIMES INSCRITOS* (%d)\n\n", len(t.Times))
	for i, e := range t.Times {
		fmt.Fprintf(&b, "%d. %s\n", i+1, e.Nome())
	}
	if !t.Sorteado {
		n := len(t.Times)
		fmt.Fprintf(&b, "\nDá %d jogos (cada time enfrenta os outros 2 vezes). Manda `!cup sortear` quando fechar a lista.", n*(n-1))
	}
	return b.String()
}

func (t *Torneio) RenderTabela() string {
	if !t.Sorteado {
		return "Ainda não gerei a tabela. Manda `!cup sortear`."
	}
	var b strings.Builder
	if t.Alvo > 0 {
		fmt.Fprintf(&b, "📊 *CLASSIFICAÇÃO* _(jogos até %d)_\n```\n", t.Alvo)
	} else {
		b.WriteString("📊 *CLASSIFICAÇÃO*\n```\n")
	}
	fmt.Fprintf(&b, "%-2s %-15s %2s %2s %2s %5s %3s\n", "#", "TIME", "J", "V", "D", "SALDO", "P")
	for i, l := range t.Classificacao() {
		fmt.Fprintf(&b, "%-2d %-15s %2d %2d %2d %+5d %3d\n",
			i+1, corta(t.NomeTime(l.Time), 15), l.Jogos, l.Vitoria, l.Derrota, l.Saldo(), l.Pontos)
	}
	b.WriteString("```")
	b.WriteString(t.RenderSituacao())
	return b.String()
}

func (t *Torneio) RenderSituacao() string {
	var b strings.Builder
	if t.Campea != 0 {
		fmt.Fprintf(&b, "\n\n🏆 *CAMPEÃO: %s*", t.NomeTime(t.Campea))
		return b.String()
	}
	faltam := len(t.JogosPendentes())
	if lider := t.Lider(); lider != 0 && faltam < len(t.Jogos) {
		fmt.Fprintf(&b, "\n🥇 Líder: *%s*", t.NomeTime(lider))
	}
	if sem := t.SemChance(); len(sem) > 0 {
		var nomes []string
		for _, id := range sem {
			nomes = append(nomes, t.NomeTime(id))
		}
		fmt.Fprintf(&b, "\n❌ Sem chance de título: %s", strings.Join(nomes, ", "))
	}
	if faltam > 0 {
		fmt.Fprintf(&b, "\n⏳ Faltam %d jogos", faltam)
	}
	return b.String()
}

func (t *Torneio) porRodada(jogos []*Jogo) ([]int, map[int][]*Jogo) {
	m := map[int][]*Jogo{}
	var ordem []int
	for _, j := range jogos {
		if _, ok := m[j.Rodada]; !ok {
			ordem = append(ordem, j.Rodada)
		}
		m[j.Rodada] = append(m[j.Rodada], j)
	}
	sort.Ints(ordem)
	return ordem, m
}

func (t *Torneio) RenderJogos() string {
	if !t.Sorteado {
		return "Ainda não gerei a tabela. Manda `!cup sortear`."
	}
	pend := t.JogosPendentes()
	if len(pend) == 0 {
		if t.Campea != 0 {
			return fmt.Sprintf("Acabou! 🏆 *%s* é o campeão.\n\nManda `!cup tabela` pra ver como terminou.", t.NomeTime(t.Campea))
		}
		return "Todos os jogos já foram registrados."
	}
	var b strings.Builder
	b.WriteString("⏭️ *JOGOS PENDENTES*\n")
	ordem, m := t.porRodada(pend)
	primeiro := pend[0].ID
	for _, r := range ordem {
		fmt.Fprintf(&b, "\n_Rodada %d_\n", r)
		for _, j := range m[r] {
			prefixo := "  "
			if j.ID == primeiro {
				prefixo = "▶ "
			}
			fmt.Fprintf(&b, "%s%s\n", prefixo, t.LinhaJogo(j))
		}
		b.WriteString(t.linhaFolga(r))
	}
	b.WriteString("\nPra registrar: `!cup placar <id> 21x18` — ex: `!cup placar " + fmt.Sprint(primeiro) + " 21x18`")
	return b.String()
}

func (t *Torneio) RenderCampeonato() string {
	if !t.Sorteado {
		return t.RenderTimes()
	}
	var b strings.Builder
	b.WriteString(t.RenderTabela())
	b.WriteString("\n\n🗓️ *RODADAS*\n")
	ordem, m := t.porRodada(t.Jogos)
	for _, r := range ordem {
		fmt.Fprintf(&b, "\n_Rodada %d_\n", r)
		for _, j := range m[r] {
			fmt.Fprintf(&b, "%s\n", t.LinhaJogo(j))
		}
		b.WriteString(t.linhaFolga(r))
	}
	return b.String()
}

func (t *Torneio) linhaFolga(rodada int) string {
	folga := t.Folga(rodada)
	if len(folga) == 0 {
		return ""
	}
	var nomes []string
	for _, id := range folga {
		nomes = append(nomes, t.NomeTime(id))
	}
	return "😴 folga: " + strings.Join(nomes, ", ") + "\n"
}

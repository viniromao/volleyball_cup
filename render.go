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
	a := t.NomeDupla(j.A)
	b := t.NomeDupla(j.B)
	if !j.Jogado {
		return fmt.Sprintf("J%d  %s  x  %s", j.ID, a, b)
	}
	foto := ""
	if j.Foto != "" {
		foto = " 📸"
	}
	return fmt.Sprintf("J%d  %s *%d x %d* %s  ✅ %s%s", j.ID, a, j.PlacarA, j.PlacarB, b, t.NomeDupla(j.Vencedor()), foto)
}

func RenderAjuda() string {
	return strings.Join([]string{
		"🏐 *COPA DE VÔLEI — PONTOS CORRIDOS*",
		"",
		"*Antes de começar*",
		"`!dupla João & Maria` — cadastra uma dupla",
		"`!duplas` — lista as duplas inscritas",
		"`!remover João` — tira a dupla do João",
		"`!formato sets` ou `!formato pontos`",
		"",
		"*Gerar a tabela*",
		"`!sortear` — todos contra todos, turno único",
		"`!sortear 2` — turno e returno",
		"",
		"*Durante a copa*",
		"`!jogos` — o que falta jogar",
		"`!placar 7 2x1` — placar do jogo J7 (sempre pelo ID)",
		"`!placar 7 2x0` — manda de novo pra corrigir o J7",
		"`!tabela` — a classificação",
		"`!campeonato` — tabela + todas as rodadas",
		"`!desfazer` — apaga o último resultado",
		"",
		"*Foto*",
		"manda a foto/sticker da dupla comemorando logo depois do placar (ou com `!placar 7 2x1` na legenda) que eu grudo no jogo.",
		"",
		"`!zerar CONFIRMA` — recomeça tudo do zero",
	}, "\n")
}

func (t *Torneio) RenderDuplas() string {
	if len(t.Duplas) == 0 {
		return "Nenhuma dupla inscrita ainda. Manda `!dupla Fulano & Ciclano`."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "🏐 *DUPLAS INSCRITAS* (%d)\n\n", len(t.Duplas))
	for i, d := range t.Duplas {
		fmt.Fprintf(&b, "%d. %s\n", i+1, d.Nome())
	}
	if !t.Sorteado {
		n := len(t.Duplas)
		fmt.Fprintf(&b, "\nDá %d jogos no turno único. Manda `!sortear` quando fechar a lista.", n*(n-1)/2)
	}
	return b.String()
}

func (t *Torneio) RenderTabela() string {
	if !t.Sorteado {
		return "Ainda não gerei a tabela. Manda `!sortear`."
	}
	unidade := "SD"
	if t.Formato == "pontos" {
		unidade = "SP"
	}
	var b strings.Builder
	b.WriteString("📊 *CLASSIFICAÇÃO*\n```\n")
	fmt.Fprintf(&b, "%-2s %-15s %2s %2s %2s %4s %3s\n", "#", "DUPLA", "J", "V", "D", unidade, "P")
	for i, l := range t.Classificacao() {
		fmt.Fprintf(&b, "%-2d %-15s %2d %2d %2d %+4d %3d\n",
			i+1, corta(t.NomeDupla(l.Dupla), 15), l.Jogos, l.Vitoria, l.Derrota, l.Saldo(), l.Pontos)
	}
	b.WriteString("```")
	b.WriteString(t.RenderSituacao())
	return b.String()
}

func (t *Torneio) RenderSituacao() string {
	var b strings.Builder
	if t.Campea != 0 {
		fmt.Fprintf(&b, "\n\n🏆 *CAMPEÃ: %s*", t.NomeDupla(t.Campea))
		return b.String()
	}
	faltam := len(t.JogosPendentes())
	if lider := t.Lider(); lider != 0 && faltam < len(t.Jogos) {
		fmt.Fprintf(&b, "\n🥇 Líder: *%s*", t.NomeDupla(lider))
	}
	if sem := t.SemChance(); len(sem) > 0 {
		var nomes []string
		for _, id := range sem {
			nomes = append(nomes, t.NomeDupla(id))
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
		return "Ainda não gerei a tabela. Manda `!sortear`."
	}
	pend := t.JogosPendentes()
	if len(pend) == 0 {
		if t.Campea != 0 {
			return fmt.Sprintf("Acabou! 🏆 *%s* é a campeã.\n\nManda `!tabela` pra ver como terminou.", t.NomeDupla(t.Campea))
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
	}
	b.WriteString("\nPra registrar: `!placar <id> 2x1` — ex: `!placar " + fmt.Sprint(primeiro) + " 2x1`")
	return b.String()
}

func (t *Torneio) RenderCampeonato() string {
	if !t.Sorteado {
		return t.RenderDuplas()
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
	}
	return b.String()
}

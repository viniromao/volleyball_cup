package main

import (
	"fmt"
	"sort"
	"strings"
)

func letraGrupo(i int) string {
	return string(rune('A' + i))
}

func (t *Torneio) ladoNome(j *Jogo, lado int) string {
	id, de := j.A, j.DeA
	if lado == 1 {
		id, de = j.B, j.DeB
	}
	if id != 0 {
		return t.NomeDupla(id)
	}
	if de != 0 {
		return fmt.Sprintf("vencedor do J%d", de)
	}
	return "a definir"
}

func (t *Torneio) LinhaJogo(j *Jogo) string {
	a := t.ladoNome(j, 0)
	b := t.ladoNome(j, 1)
	if j.Jogado {
		marca := "✅"
		venc := t.NomeDupla(j.Vencedor())
		foto := ""
		if j.Foto != "" {
			foto = " 📸"
		}
		return fmt.Sprintf("J%d  %s *%d x %d* %s  %s %s%s", j.ID, a, j.PlacarA, j.PlacarB, b, marca, venc, foto)
	}
	if !j.Pronto() {
		return fmt.Sprintf("J%d  %s  x  %s  ⏳", j.ID, a, b)
	}
	return fmt.Sprintf("J%d  %s  x  %s", j.ID, a, b)
}

func RenderAjuda() string {
	return strings.Join([]string{
		"🏐 *COPA DE VÔLEI — COMANDOS*",
		"",
		"*Antes do sorteio*",
		"`!dupla João & Maria` — cadastra uma dupla",
		"`!duplas` — lista as duplas inscritas",
		"`!remover João` — tira a dupla do João",
		"`!formato sets` ou `!formato pontos`",
		"",
		"*Sorteio*",
		"`!sortear` — sorteia os grupos e monta os jogos",
		"`!sortear 2` — força 2 grupos",
		"",
		"*Durante a copa*",
		"`!jogos` — o que falta jogar",
		"`!placar 2x1` — resultado do jogo da vez",
		"`!placar 7 2x0` — resultado do jogo J7",
		"`!tabela` — classificação dos grupos",
		"`!chaves` — grupos, mata-mata e quem caiu",
		"`!desfazer` — apaga o último resultado",
		"",
		"*Foto*",
		"manda a foto/sticker da dupla comemorando logo depois do placar (ou com `!placar 2x1` na legenda) que eu grudo no jogo.",
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
		g := quantosGrupos(len(t.Duplas))
		fmt.Fprintf(&b, "\nDá pra fazer %d grupo(s). Manda `!sortear` quando fechar.", g)
	}
	return b.String()
}

func (t *Torneio) RenderTabela() string {
	if !t.Sorteado {
		return "Ainda não sorteei os grupos. Manda `!sortear`."
	}
	var b strings.Builder
	b.WriteString("📊 *CLASSIFICAÇÃO*\n")
	unidade := "sets"
	if t.Formato == "pontos" {
		unidade = "pts"
	}
	for g := range t.Grupos {
		fmt.Fprintf(&b, "\n*Grupo %s*\n```\n", letraGrupo(g))
		fmt.Fprintf(&b, "%-22s %2s %2s %2s %4s\n", "dupla", "P", "V", "D", unidade)
		for i, l := range t.Classificacao(g) {
			nome := t.NomeDupla(l.Dupla)
			if len(nome) > 22 {
				nome = nome[:21] + "…"
			}
			marca := " "
			if i < 2 {
				marca = "›"
			}
			fmt.Fprintf(&b, "%s%-21s %2d %2d %2d %+4d\n", marca, nome, l.Pontos, l.Vitoria, l.Derrota, l.Saldo())
		}
		b.WriteString("```")
	}
	return b.String()
}

func (t *Torneio) RenderJogos() string {
	if !t.Sorteado {
		return "Ainda não sorteei os grupos. Manda `!sortear`."
	}
	var pend []*Jogo
	for _, j := range t.Jogos {
		if !j.Jogado {
			pend = append(pend, j)
		}
	}
	if len(pend) == 0 {
		if t.Campea != 0 {
			return fmt.Sprintf("Acabou! 🏆 *%s* é a campeã.", t.NomeDupla(t.Campea))
		}
		return "Todos os jogos da fase atual já foram registrados."
	}
	var b strings.Builder
	b.WriteString("⏭️ *JOGOS PENDENTES*\n\n")
	for i, j := range pend {
		prefixo := "  "
		if i == 0 && j.Pronto() {
			prefixo = "▶ "
		}
		fmt.Fprintf(&b, "%s%s\n", prefixo, t.LinhaJogo(j))
	}
	if p := t.ProximoJogo(); p != nil {
		fmt.Fprintf(&b, "\nPra registrar o jogo da vez: `!placar 2x1`")
	}
	return b.String()
}

func (t *Torneio) RenderChaves() string {
	if !t.Sorteado {
		return t.RenderDuplas()
	}
	var b strings.Builder
	b.WriteString("🏐 *COPA DE VÔLEI — CHAVES*\n")

	for g := range t.Grupos {
		fmt.Fprintf(&b, "\n📋 *GRUPO %s*\n", letraGrupo(g))
		for _, j := range t.Jogos {
			if j.Fase == "grupo" && j.Grupo == g {
				fmt.Fprintf(&b, "%s\n", t.LinhaJogo(j))
			}
		}
	}

	if t.MataMata {
		b.WriteString("\n🔥 *MATA-MATA*\n")
		rodadas := map[int][]*Jogo{}
		var ordem []int
		for _, j := range t.Jogos {
			if j.Fase == "grupo" {
				continue
			}
			if _, ok := rodadas[j.Rodada]; !ok {
				ordem = append(ordem, j.Rodada)
			}
			rodadas[j.Rodada] = append(rodadas[j.Rodada], j)
		}
		sort.Ints(ordem)
		for _, r := range ordem {
			jogos := rodadas[r]
			fmt.Fprintf(&b, "\n_%s_\n", strings.ToUpper(jogos[0].Fase))
			for _, j := range jogos {
				fmt.Fprintf(&b, "%s\n", t.LinhaJogo(j))
			}
		}
	} else {
		restam := 0
		for _, j := range t.JogosDaFaseDeGrupos() {
			if !j.Jogado {
				restam++
			}
		}
		fmt.Fprintf(&b, "\n🔥 *MATA-MATA*\nsai quando fechar a fase de grupos (faltam %d jogos)\n", restam)
	}

	if elim := t.Eliminadas(); len(elim) > 0 {
		var nomes []string
		for _, id := range elim {
			nomes = append(nomes, t.NomeDupla(id))
		}
		fmt.Fprintf(&b, "\n❌ *Eliminadas:* %s", strings.Join(nomes, ", "))
	}
	if vivas := t.Vivas(); t.MataMata && len(vivas) > 0 && t.Campea == 0 {
		var nomes []string
		for _, id := range vivas {
			nomes = append(nomes, t.NomeDupla(id))
		}
		fmt.Fprintf(&b, "\n✅ *Ainda na disputa:* %s", strings.Join(nomes, ", "))
	}
	if t.Campea != 0 {
		fmt.Fprintf(&b, "\n\n🏆 *CAMPEÃ: %s*", t.NomeDupla(t.Campea))
	}
	return b.String()
}

func (t *Torneio) Vivas() []int {
	if !t.MataMata {
		return nil
	}
	elim := map[int]bool{}
	for _, id := range t.Eliminadas() {
		elim[id] = true
	}
	var out []int
	for _, d := range t.Duplas {
		if !elim[d.ID] {
			out = append(out, d.ID)
		}
	}
	return out
}

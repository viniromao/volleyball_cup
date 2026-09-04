package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type Dupla struct {
	ID    int    `json:"id"`
	A     string `json:"a"`
	B     string `json:"b"`
	Grupo int    `json:"grupo"`
}

func (d *Dupla) Nome() string {
	return d.A + " & " + d.B
}

type Jogo struct {
	ID      int       `json:"id"`
	Fase    string    `json:"fase"`
	Grupo   int       `json:"grupo"`
	Rodada  int       `json:"rodada"`
	Slot    int       `json:"slot"`
	A       int       `json:"a"`
	B       int       `json:"b"`
	DeA     int       `json:"de_a"`
	DeB     int       `json:"de_b"`
	PlacarA int       `json:"placar_a"`
	PlacarB int       `json:"placar_b"`
	Jogado  bool      `json:"jogado"`
	Foto    string    `json:"foto"`
	Quando  time.Time `json:"quando"`
}

func (j *Jogo) Pronto() bool {
	return j.A != 0 && j.B != 0
}

func (j *Jogo) Vencedor() int {
	if !j.Jogado {
		return 0
	}
	if j.PlacarA > j.PlacarB {
		return j.A
	}
	return j.B
}

func (j *Jogo) Perdedor() int {
	if !j.Jogado {
		return 0
	}
	if j.PlacarA > j.PlacarB {
		return j.B
	}
	return j.A
}

type Torneio struct {
	Chat           string    `json:"chat"`
	Formato        string    `json:"formato"`
	Duplas         []*Dupla  `json:"duplas"`
	Jogos          []*Jogo   `json:"jogos"`
	Grupos         [][]int   `json:"grupos"`
	Sorteado       bool      `json:"sorteado"`
	MataMata       bool      `json:"mata_mata"`
	Campea         int       `json:"campea"`
	ProxID         int       `json:"prox_id"`
	Historico      []int     `json:"historico"`
	UltimoJogo     int       `json:"ultimo_jogo"`
	UltimoEm       time.Time `json:"ultimo_em"`
	AnunciouCampea bool      `json:"anunciou_campea"`
}

func NovoTorneio(chat string) *Torneio {
	return &Torneio{Chat: chat, Formato: "sets", ProxID: 1}
}

func (t *Torneio) novoID() int {
	id := t.ProxID
	t.ProxID++
	return id
}

func (t *Torneio) Dupla(id int) *Dupla {
	for _, d := range t.Duplas {
		if d.ID == id {
			return d
		}
	}
	return nil
}

func (t *Torneio) NomeDupla(id int) string {
	if d := t.Dupla(id); d != nil {
		return d.Nome()
	}
	return "?"
}

func (t *Torneio) Jogo(id int) *Jogo {
	for _, j := range t.Jogos {
		if j.ID == id {
			return j
		}
	}
	return nil
}

func (t *Torneio) AddDupla(a, b string) (*Dupla, error) {
	if t.Sorteado {
		return nil, errors.New("o sorteio já foi feito, use !zerar CONFIRMA pra recomeçar")
	}
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return nil, errors.New("preciso dos dois nomes")
	}
	for _, d := range t.Duplas {
		if strings.EqualFold(d.A, a) || strings.EqualFold(d.B, a) || strings.EqualFold(d.A, b) || strings.EqualFold(d.B, b) {
			return nil, fmt.Errorf("%s já está na dupla %s", a, d.Nome())
		}
	}
	d := &Dupla{ID: t.novoID(), A: a, B: b, Grupo: -1}
	t.Duplas = append(t.Duplas, d)
	return d, nil
}

func (t *Torneio) RemoveDupla(alvo string) (*Dupla, error) {
	if t.Sorteado {
		return nil, errors.New("o sorteio já foi feito, use !zerar CONFIRMA pra recomeçar")
	}
	alvo = strings.ToLower(strings.TrimSpace(alvo))
	for i, d := range t.Duplas {
		nome := strings.ToLower(d.Nome())
		if strings.Contains(nome, alvo) {
			t.Duplas = append(t.Duplas[:i], t.Duplas[i+1:]...)
			return d, nil
		}
	}
	return nil, errors.New("não achei essa dupla")
}

func quantosGrupos(n int) int {
	if n < 4 {
		return 1
	}
	return (n + 3) / 4
}

func (t *Torneio) Sortear(grupos int) error {
	if len(t.Duplas) < 2 {
		return errors.New("preciso de pelo menos 2 duplas")
	}
	if grupos <= 0 {
		grupos = quantosGrupos(len(t.Duplas))
	}
	if grupos > len(t.Duplas)/2 {
		grupos = len(t.Duplas) / 2
	}
	if grupos < 1 {
		grupos = 1
	}

	t.Jogos = nil
	t.Historico = nil
	t.MataMata = false
	t.Campea = 0
	t.ProxID = 1
	for _, d := range t.Duplas {
		d.ID = t.novoID()
	}

	ordem := make([]*Dupla, len(t.Duplas))
	copy(ordem, t.Duplas)
	rand.Shuffle(len(ordem), func(i, j int) { ordem[i], ordem[j] = ordem[j], ordem[i] })

	t.Grupos = make([][]int, grupos)
	for i, d := range ordem {
		g := i % grupos
		if (i/grupos)%2 == 1 {
			g = grupos - 1 - g
		}
		d.Grupo = g
		t.Grupos[g] = append(t.Grupos[g], d.ID)
	}

	for g, ids := range t.Grupos {
		for i := 0; i < len(ids); i++ {
			for k := i + 1; k < len(ids); k++ {
				t.Jogos = append(t.Jogos, &Jogo{
					ID:    t.novoID(),
					Fase:  "grupo",
					Grupo: g,
					A:     ids[i],
					B:     ids[k],
				})
			}
		}
	}
	t.Sorteado = true
	return nil
}

func (t *Torneio) JogosDaFaseDeGrupos() []*Jogo {
	var out []*Jogo
	for _, j := range t.Jogos {
		if j.Fase == "grupo" {
			out = append(out, j)
		}
	}
	return out
}

func (t *Torneio) GruposEncerrados() bool {
	jogos := t.JogosDaFaseDeGrupos()
	if len(jogos) == 0 {
		return false
	}
	for _, j := range jogos {
		if !j.Jogado {
			return false
		}
	}
	return true
}

func (t *Torneio) ProximoJogo() *Jogo {
	for _, j := range t.Jogos {
		if !j.Jogado && j.Pronto() {
			return j
		}
	}
	return nil
}

func (t *Torneio) pontos(vencedorSets, perdedorSets int) (int, int) {
	if t.Formato == "pontos" {
		return 3, 0
	}
	if perdedorSets >= 1 {
		return 2, 1
	}
	return 3, 0
}

type Linha struct {
	Dupla   int
	Pontos  int
	Vitoria int
	Derrota int
	Pro     int
	Contra  int
}

func (l Linha) Saldo() int { return l.Pro - l.Contra }

func (t *Torneio) Classificacao(grupo int) []Linha {
	if grupo < 0 || grupo >= len(t.Grupos) {
		return nil
	}
	idx := map[int]*Linha{}
	var linhas []*Linha
	for _, id := range t.Grupos[grupo] {
		l := &Linha{Dupla: id}
		idx[id] = l
		linhas = append(linhas, l)
	}
	for _, j := range t.Jogos {
		if j.Fase != "grupo" || j.Grupo != grupo || !j.Jogado {
			continue
		}
		la, lb := idx[j.A], idx[j.B]
		if la == nil || lb == nil {
			continue
		}
		la.Pro += j.PlacarA
		la.Contra += j.PlacarB
		lb.Pro += j.PlacarB
		lb.Contra += j.PlacarA
		if j.PlacarA > j.PlacarB {
			p, q := t.pontos(j.PlacarA, j.PlacarB)
			la.Pontos += p
			lb.Pontos += q
			la.Vitoria++
			lb.Derrota++
		} else {
			p, q := t.pontos(j.PlacarB, j.PlacarA)
			lb.Pontos += p
			la.Pontos += q
			lb.Vitoria++
			la.Derrota++
		}
	}
	sort.SliceStable(linhas, func(i, k int) bool {
		a, b := linhas[i], linhas[k]
		if a.Pontos != b.Pontos {
			return a.Pontos > b.Pontos
		}
		if a.Saldo() != b.Saldo() {
			return a.Saldo() > b.Saldo()
		}
		if a.Pro != b.Pro {
			return a.Pro > b.Pro
		}
		if v := t.confrontoDireto(a.Dupla, b.Dupla); v != 0 {
			return v == a.Dupla
		}
		return a.Dupla < b.Dupla
	})
	out := make([]Linha, len(linhas))
	for i, l := range linhas {
		out[i] = *l
	}
	return out
}

func (t *Torneio) confrontoDireto(a, b int) int {
	for _, j := range t.Jogos {
		if !j.Jogado {
			continue
		}
		if (j.A == a && j.B == b) || (j.A == b && j.B == a) {
			return j.Vencedor()
		}
	}
	return 0
}

func (t *Torneio) classificados() []int {
	if len(t.Grupos) == 1 {
		cl := t.Classificacao(0)
		n := 4
		if len(cl) < 4 {
			n = 2
		}
		if len(cl) < n {
			n = len(cl)
		}
		out := make([]int, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, cl[i].Dupla)
		}
		return out
	}
	var primeiros, segundos []int
	for g := range t.Grupos {
		cl := t.Classificacao(g)
		if len(cl) > 0 {
			primeiros = append(primeiros, cl[0].Dupla)
		}
		if len(cl) > 1 {
			segundos = append(segundos, cl[1].Dupla)
		}
	}
	return append(primeiros, segundos...)
}

type slotv struct {
	dupla int
	jogo  int
}

func (s slotv) vazio() bool { return s.dupla == 0 && s.jogo == 0 }

func nomeFase(slots int) string {
	switch slots {
	case 2:
		return "final"
	case 4:
		return "semifinal"
	case 8:
		return "quartas"
	case 16:
		return "oitavas"
	}
	return fmt.Sprintf("fase de %d", slots)
}

func (t *Torneio) GerarMataMata() {
	seeds := t.classificados()
	if len(seeds) < 2 {
		return
	}
	n := 1
	for n < len(seeds) {
		n *= 2
	}
	atual := make([]slotv, n)
	for i, id := range seeds {
		atual[i] = slotv{dupla: id}
	}
	rodada := 0
	for len(atual) > 1 {
		m := len(atual)
		prox := make([]slotv, m/2)
		for i := 0; i < m/2; i++ {
			a, b := atual[i], atual[m-1-i]
			if b.vazio() {
				prox[i] = a
				continue
			}
			if a.vazio() {
				prox[i] = b
				continue
			}
			j := &Jogo{
				ID:     t.novoID(),
				Fase:   nomeFase(m),
				Grupo:  -1,
				Rodada: rodada,
				Slot:   i,
				A:      a.dupla,
				B:      b.dupla,
				DeA:    a.jogo,
				DeB:    b.jogo,
			}
			t.Jogos = append(t.Jogos, j)
			prox[i] = slotv{jogo: j.ID}
		}
		atual = prox
		rodada++
	}
	t.MataMata = true
}

func (t *Torneio) propagar() {
	for _, j := range t.Jogos {
		if !j.Jogado {
			continue
		}
		v := j.Vencedor()
		for _, k := range t.Jogos {
			if k.DeA == j.ID {
				k.A = v
			}
			if k.DeB == j.ID {
				k.B = v
			}
		}
	}
	t.Campea = 0
	for _, j := range t.Jogos {
		if j.Fase == "final" && j.Jogado {
			t.Campea = j.Vencedor()
		}
	}
}

func (t *Torneio) Eliminadas() []int {
	if !t.MataMata {
		return nil
	}
	visto := map[int]bool{}
	var out []int
	marcar := func(id int) {
		if id != 0 && !visto[id] {
			visto[id] = true
			out = append(out, id)
		}
	}
	for _, d := range t.Duplas {
		if !t.passouDaFaseDeGrupos(d.ID) {
			marcar(d.ID)
		}
	}
	for _, j := range t.Jogos {
		if j.Fase != "grupo" && j.Jogado {
			marcar(j.Perdedor())
		}
	}
	return out
}

func (t *Torneio) passouDaFaseDeGrupos(id int) bool {
	for _, c := range t.classificados() {
		if c == id {
			return true
		}
	}
	return false
}

func (t *Torneio) RegistrarPlacar(jogoID, a, b int) (*Jogo, error) {
	j := t.Jogo(jogoID)
	if j == nil {
		return nil, fmt.Errorf("não existe o jogo J%d", jogoID)
	}
	if !j.Pronto() {
		return nil, fmt.Errorf("o J%d ainda depende de outros resultados", jogoID)
	}
	if a == b {
		return nil, errors.New("no vôlei não tem empate")
	}
	if t.Formato == "sets" {
		if a > 2 || b > 2 || (a != 2 && b != 2) {
			return nil, errors.New("no formato sets o placar é 2x0, 2x1, 1x2 ou 0x2 (troque com !formato pontos)")
		}
	}
	j.PlacarA = a
	j.PlacarB = b
	j.Jogado = true
	j.Quando = time.Now()
	t.Historico = append(t.Historico, j.ID)
	t.UltimoJogo = j.ID
	t.UltimoEm = time.Now()
	return j, nil
}

func (t *Torneio) Desfazer() (*Jogo, error) {
	if len(t.Historico) == 0 {
		return nil, errors.New("não tem resultado pra desfazer")
	}
	id := t.Historico[len(t.Historico)-1]
	t.Historico = t.Historico[:len(t.Historico)-1]
	j := t.Jogo(id)
	if j == nil {
		return nil, errors.New("jogo sumiu")
	}
	j.Jogado = false
	j.PlacarA, j.PlacarB = 0, 0
	if t.MataMata && t.faseDeGruposFoiDesfeita() {
		t.removerMataMata()
	}
	t.limparPropagacao()
	t.propagar()
	t.UltimoJogo = 0
	return j, nil
}

func (t *Torneio) faseDeGruposFoiDesfeita() bool {
	return !t.GruposEncerrados()
}

func (t *Torneio) removerMataMata() {
	var restantes []*Jogo
	for _, j := range t.Jogos {
		if j.Fase == "grupo" {
			restantes = append(restantes, j)
		}
	}
	t.Jogos = restantes
	t.MataMata = false
	t.Campea = 0
}

func (t *Torneio) limparPropagacao() {
	for _, j := range t.Jogos {
		if j.DeA != 0 {
			j.A = 0
		}
		if j.DeB != 0 {
			j.B = 0
		}
	}
}

func (t *Torneio) Atualizar() {
	if t.Sorteado && !t.MataMata && t.GruposEncerrados() {
		t.GerarMataMata()
	}
	t.propagar()
}

func (t *Torneio) FotoDaDupla(id int) string {
	for i := len(t.Jogos) - 1; i >= 0; i-- {
		j := t.Jogos[i]
		if j.Jogado && j.Foto != "" && j.Vencedor() == id {
			return j.Foto
		}
	}
	return ""
}

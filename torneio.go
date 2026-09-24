package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// Equipe é um time da copa, com quantas pessoas quiser.
type Equipe struct {
	ID        int      `json:"id"`
	Jogadores []string `json:"jogadores"`
}

func (e *Equipe) Nome() string {
	n := len(e.Jogadores)
	if n <= 2 {
		return strings.Join(e.Jogadores, " & ")
	}
	return strings.Join(e.Jogadores[:n-1], ", ") + " & " + e.Jogadores[n-1]
}

// duplaAntiga é como os times eram salvos quando a copa era só de duplas.
type duplaAntiga struct {
	ID int    `json:"id"`
	A  string `json:"a"`
	B  string `json:"b"`
}

type Jogo struct {
	ID      int       `json:"id"`
	Rodada  int       `json:"rodada"`
	Turno   int       `json:"turno"`
	A       int       `json:"a"`
	B       int       `json:"b"`
	PlacarA int       `json:"placar_a"`
	PlacarB int       `json:"placar_b"`
	Jogado  bool      `json:"jogado"`
	Foto    string    `json:"foto"`
	Quando  time.Time `json:"quando"`
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

func (j *Jogo) Envolve(id int) bool {
	return j.A == id || j.B == id
}

type Torneio struct {
	Chat           string         `json:"chat"`
	Alvo           int            `json:"alvo"`
	Turnos         int            `json:"turnos"`
	Times          []*Equipe      `json:"times"`
	Duplas         []*duplaAntiga `json:"duplas,omitempty"`
	Jogos          []*Jogo        `json:"jogos"`
	Sorteado       bool           `json:"sorteado"`
	Campea         int            `json:"campea"`
	ProxID         int            `json:"prox_id"`
	Historico      []int          `json:"historico"`
	UltimoJogo     int            `json:"ultimo_jogo"`
	UltimoEm       time.Time      `json:"ultimo_em"`
	AnunciouCampea bool           `json:"anunciou_campea"`
}

func NovoTorneio(chat string) *Torneio {
	return &Torneio{Chat: chat, Turnos: 1, ProxID: 1}
}

// Migrar converte as duplas salvas no formato antigo em times de 2.
func (t *Torneio) Migrar() {
	for _, d := range t.Duplas {
		t.Times = append(t.Times, &Equipe{ID: d.ID, Jogadores: []string{d.A, d.B}})
	}
	t.Duplas = nil
}

func (t *Torneio) novoID() int {
	id := t.ProxID
	t.ProxID++
	return id
}

func (t *Torneio) Equipe(id int) *Equipe {
	for _, e := range t.Times {
		if e.ID == id {
			return e
		}
	}
	return nil
}

func (t *Torneio) NomeTime(id int) string {
	if e := t.Equipe(id); e != nil {
		return e.Nome()
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

func (t *Torneio) AddTime(nomes []string) (*Equipe, error) {
	if t.Sorteado {
		return nil, errors.New("a tabela já foi gerada, use `!cup zerar CONFIRMA` pra recomeçar")
	}
	var jogadores []string
	for _, n := range nomes {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		for _, j := range jogadores {
			if strings.EqualFold(j, n) {
				return nil, fmt.Errorf("%s apareceu duas vezes no mesmo time", n)
			}
		}
		for _, e := range t.Times {
			for _, j := range e.Jogadores {
				if strings.EqualFold(j, n) {
					return nil, fmt.Errorf("%s já está no time %s", n, e.Nome())
				}
			}
		}
		jogadores = append(jogadores, n)
	}
	if len(jogadores) == 0 {
		return nil, errors.New("preciso de pelo menos um nome")
	}
	e := &Equipe{ID: t.novoID(), Jogadores: jogadores}
	t.Times = append(t.Times, e)
	return e, nil
}

func (t *Torneio) RemoveTime(alvo string) (*Equipe, error) {
	if t.Sorteado {
		return nil, errors.New("a tabela já foi gerada, use `!cup zerar CONFIRMA` pra recomeçar")
	}
	alvo = strings.ToLower(strings.TrimSpace(alvo))
	for i, e := range t.Times {
		if strings.Contains(strings.ToLower(e.Nome()), alvo) {
			t.Times = append(t.Times[:i], t.Times[i+1:]...)
			return e, nil
		}
	}
	return nil, errors.New("não achei esse time")
}

func gerarRodadas(ids []int) [][][2]int {
	lista := append([]int{}, ids...)
	if len(lista)%2 == 1 {
		lista = append(lista, 0)
	}
	n := len(lista)
	var rodadas [][][2]int
	for r := 0; r < n-1; r++ {
		var jogos [][2]int
		for i := 0; i < n/2; i++ {
			a, b := lista[i], lista[n-1-i]
			if a == 0 || b == 0 {
				continue
			}
			if r%2 == 1 {
				a, b = b, a
			}
			jogos = append(jogos, [2]int{a, b})
		}
		rodadas = append(rodadas, jogos)
		resto := append([]int{}, lista[1:]...)
		resto = append([]int{resto[len(resto)-1]}, resto[:len(resto)-1]...)
		lista = append([]int{lista[0]}, resto...)
	}
	return rodadas
}

func (t *Torneio) Sortear() error {
	if len(t.Times) < 2 {
		return errors.New("preciso de pelo menos 2 times")
	}

	t.Jogos = nil
	t.Historico = nil
	t.Campea = 0
	t.AnunciouCampea = false
	t.Turnos = turnos
	t.ProxID = 1
	for _, e := range t.Times {
		e.ID = t.novoID()
	}

	ordem := make([]int, 0, len(t.Times))
	for _, e := range t.Times {
		ordem = append(ordem, e.ID)
	}
	rand.Shuffle(len(ordem), func(i, j int) { ordem[i], ordem[j] = ordem[j], ordem[i] })

	base := gerarRodadas(ordem)
	rodada := 0
	for turno := 1; turno <= turnos; turno++ {
		for _, jogos := range base {
			rodada++
			for _, par := range jogos {
				a, b := par[0], par[1]
				if turno%2 == 0 {
					a, b = b, a
				}
				t.Jogos = append(t.Jogos, &Jogo{
					ID:     t.novoID(),
					Rodada: rodada,
					Turno:  turno,
					A:      a,
					B:      b,
				})
			}
		}
	}
	t.Sorteado = true
	return nil
}

func (t *Torneio) Rodadas() int {
	max := 0
	for _, j := range t.Jogos {
		if j.Rodada > max {
			max = j.Rodada
		}
	}
	return max
}

func (t *Torneio) Folga(rodada int) []int {
	jogando := map[int]bool{}
	achou := false
	for _, j := range t.Jogos {
		if j.Rodada != rodada {
			continue
		}
		achou = true
		jogando[j.A] = true
		jogando[j.B] = true
	}
	if !achou {
		return nil
	}
	var out []int
	for _, e := range t.Times {
		if !jogando[e.ID] {
			out = append(out, e.ID)
		}
	}
	return out
}

func (t *Torneio) JogosPendentes() []*Jogo {
	var out []*Jogo
	for _, j := range t.Jogos {
		if !j.Jogado {
			out = append(out, j)
		}
	}
	return out
}

func (t *Torneio) ProximoJogo() *Jogo {
	for _, j := range t.Jogos {
		if !j.Jogado {
			return j
		}
	}
	return nil
}

func (t *Torneio) Encerrado() bool {
	return t.Sorteado && len(t.Jogos) > 0 && len(t.JogosPendentes()) == 0
}

const (
	pontosPorVitoria = 3
	turnos           = 2
)

type Linha struct {
	Time     int
	Jogos    int
	Pontos   int
	Vitoria  int
	Derrota  int
	Pro      int
	Contra   int
	Restam   int
	Maximo   int
	SemChanc bool
}

func (l Linha) Saldo() int { return l.Pro - l.Contra }

func (t *Torneio) Classificacao() []Linha {
	idx := map[int]*Linha{}
	var linhas []*Linha
	for _, e := range t.Times {
		l := &Linha{Time: e.ID}
		idx[e.ID] = l
		linhas = append(linhas, l)
	}
	for _, j := range t.Jogos {
		la, lb := idx[j.A], idx[j.B]
		if la == nil || lb == nil {
			continue
		}
		if !j.Jogado {
			la.Restam++
			lb.Restam++
			continue
		}
		la.Jogos++
		lb.Jogos++
		la.Pro += j.PlacarA
		la.Contra += j.PlacarB
		lb.Pro += j.PlacarB
		lb.Contra += j.PlacarA
		if j.PlacarA > j.PlacarB {
			la.Pontos += pontosPorVitoria
			la.Vitoria++
			lb.Derrota++
		} else {
			lb.Pontos += pontosPorVitoria
			lb.Vitoria++
			la.Derrota++
		}
	}

	melhorAtual := 0
	for _, l := range linhas {
		l.Maximo = l.Pontos + pontosPorVitoria*l.Restam
		if l.Pontos > melhorAtual {
			melhorAtual = l.Pontos
		}
	}
	for _, l := range linhas {
		l.SemChanc = l.Maximo < melhorAtual
	}

	sort.SliceStable(linhas, func(i, k int) bool {
		a, b := linhas[i], linhas[k]
		if a.Pontos != b.Pontos {
			return a.Pontos > b.Pontos
		}
		if a.Vitoria != b.Vitoria {
			return a.Vitoria > b.Vitoria
		}
		if a.Saldo() != b.Saldo() {
			return a.Saldo() > b.Saldo()
		}
		if a.Pro != b.Pro {
			return a.Pro > b.Pro
		}
		if v := t.confrontoDireto(a.Time, b.Time); v != 0 {
			return v == a.Time
		}
		return a.Time < b.Time
	})

	out := make([]Linha, len(linhas))
	for i, l := range linhas {
		out[i] = *l
	}
	return out
}

func (t *Torneio) confrontoDireto(a, b int) int {
	saldoA := 0
	for _, j := range t.Jogos {
		if !j.Jogado {
			continue
		}
		if j.A == a && j.B == b {
			saldoA += j.PlacarA - j.PlacarB
		} else if j.A == b && j.B == a {
			saldoA += j.PlacarB - j.PlacarA
		}
	}
	if saldoA > 0 {
		return a
	}
	if saldoA < 0 {
		return b
	}
	return 0
}

func (t *Torneio) Lider() int {
	cl := t.Classificacao()
	if len(cl) == 0 {
		return 0
	}
	return cl[0].Time
}

func (t *Torneio) SemChance() []int {
	var out []int
	for _, l := range t.Classificacao() {
		if l.SemChanc {
			out = append(out, l.Time)
		}
	}
	return out
}

func (t *Torneio) RegistrarPlacar(jogoID, a, b int) (*Jogo, bool, error) {
	j := t.Jogo(jogoID)
	if j == nil {
		return nil, false, fmt.Errorf("não existe o jogo J%d", jogoID)
	}
	if a < 0 || b < 0 {
		return nil, false, errors.New("placar não tem número negativo")
	}
	if a == b {
		return nil, false, errors.New("empate não decide jogo, confere o placar")
	}
	venc, perd := a, b
	if b > a {
		venc, perd = b, a
	}
	if t.Alvo == 0 {
		t.Alvo = venc
	} else if venc < t.Alvo {
		return nil, false, fmt.Errorf("os jogos vão até %d, então o vencedor tem que chegar lá (mude com `!cup ate %d`)", t.Alvo, venc)
	} else if venc > t.Alvo && perd < t.Alvo-1 {
		return nil, false, fmt.Errorf("%dx%d passa de %d sem ser vantagem — confere o placar (jogo até %d só passa disso empatando em %d)", venc, perd, t.Alvo, t.Alvo, t.Alvo-1)
	}
	corrigido := j.Jogado
	j.PlacarA = a
	j.PlacarB = b
	j.Jogado = true
	j.Quando = time.Now()
	if !corrigido {
		t.Historico = append(t.Historico, j.ID)
	}
	t.UltimoJogo = j.ID
	t.UltimoEm = time.Now()
	t.AnunciouCampea = false
	return j, corrigido, nil
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
	t.UltimoJogo = 0
	t.AnunciouCampea = false
	return j, nil
}

func (t *Torneio) DefinirAlvo(n int) error {
	if n < 0 {
		return errors.New("o alvo não pode ser negativo")
	}
	for _, j := range t.Jogos {
		if !j.Jogado {
			continue
		}
		venc := j.PlacarA
		if j.PlacarB > venc {
			venc = j.PlacarB
		}
		if n > 0 && venc < n {
			return fmt.Errorf("o J%d terminou em %dx%d, que não chega a %d — corrija o placar antes", j.ID, j.PlacarA, j.PlacarB, n)
		}
	}
	t.Alvo = n
	return nil
}

func (t *Torneio) Atualizar() {
	t.Campea = 0
	if t.Encerrado() {
		t.Campea = t.Lider()
	}
}

func (t *Torneio) FotoDoTime(id int) string {
	for i := len(t.Jogos) - 1; i >= 0; i-- {
		j := t.Jogos[i]
		if j.Jogado && j.Foto != "" && j.Vencedor() == id {
			return j.Foto
		}
	}
	return ""
}

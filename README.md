# 🏐 Bot da Copa de Vôlei

Bot de WhatsApp em Go que organiza uma copa de duplas em pontos corridos num
grupo de amigos: cadastra duplas, gera a tabela de todos contra todos,
registra placares por ID e mantém a classificação em dia.

Conecta como **aparelho conectado** (mesmo mecanismo do WhatsApp Web) usando
`whatsmeow` — é assim que dá pra ler e responder em **grupo**, coisa que a API
oficial (Cloud API) não faz.

## Rodar

```bash
go build -buildvcs=false -o copa-volei-bot .
./copa-volei-bot
```

Na primeira execução aparece um QR no terminal:
**WhatsApp > Aparelhos conectados > Conectar aparelho**.

A sessão fica em `dados/sessao.db` — depois disso não precisa escanear de novo.
O estado da copa fica em `dados/torneios.json` (um torneio por conversa) e as
fotos em `dados/midia/`.

Variáveis: `COPA_DATA` (pasta de dados, padrão `dados`) e `COPA_LOG`
(`DEBUG`/`INFO`/`WARN`, padrão `WARN`).

## Comandos

| Comando | O que faz |
|---|---|
| `!ajuda` | lista tudo |
| `!dupla João & Maria` | cadastra uma dupla |
| `!duplas` | lista as inscritas |
| `!remover João` | tira a dupla |
| `!formato sets\|pontos` | melhor de 3 sets (padrão) ou set único em pontos |
| `!sortear` | gera a tabela, turno único |
| `!sortear 2` | turno e returno |
| `!jogos` | o que falta jogar, com os IDs |
| `!placar 7 2x1` | placar do jogo J7 |
| `!tabela` | a classificação |
| `!campeonato` | tabela + todas as rodadas |
| `!desfazer` | apaga o último resultado |
| `!zerar CONFIRMA` | recomeça do zero |

### Placar é sempre pelo ID do jogo

Cada jogo tem um ID fixo (`J7`), que aparece em `!jogos` e `!campeonato`. O
placar vai sempre com esse ID — a ordem em que os jogos acontecem não importa,
então dá pra registrar fora de ordem, com duas quadras rodando ao mesmo tempo:

```
!placar 7 2x1
!placar 12 2x0
!placar 9 1x2
```

Mandar de novo no mesmo ID **corrige** o resultado anterior (ele responde
`J7 corrigido` e refaz a tabela). `!desfazer` apaga o último placar registrado.

### Fotos

Manda a foto (ou sticker) da dupla comemorando **logo depois do placar** — ou
com `!placar 7 2x1` na legenda da imagem — que ela fica grudada naquele jogo
(aparece um 📸 na lista de rodadas). No fim, a foto da campeã volta com a
mensagem do título. A janela pra anexar é de 20 minutos após o placar.

## Formato

Pontos corridos: **todas as duplas jogam contra todas**, e campeã é quem somar
mais pontos no fim. Não tem grupo nem mata-mata, ninguém é eliminado.

- A tabela é gerada pelo método do círculo, então dentro de cada rodada nenhuma
  dupla joga duas vezes seguidas — todo mundo descansa parecido.
- Com número ímpar de duplas, uma folga por rodada.
- `!sortear 2` faz turno e returno (o mando de quadra inverte no returno).
- Pontuação no formato `sets`: 2x0 = 3 pts, 2x1 = 2 pts, 1x2 = 1 pt, 0x2 = 0.
  No formato `pontos` (set único, ex: 21x18): vitória 3, derrota 0.
- Desempate: pontos → vitórias → saldo de sets → sets ganhos → confronto direto.
- Enquanto rola, a tabela mostra o líder e quem já está **sem chance de título**
  (quem não alcança mais o líder nem ganhando tudo que falta).

## Deploy na VPS

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o copa-volei-bot .
scp copa-volei-bot copa-volei-bot.service usuario@vps:/tmp/
```

Na VPS:

```bash
sudo useradd -r -m -d /opt/copa-volei-bot copa
sudo mv /tmp/copa-volei-bot /opt/copa-volei-bot/
sudo mv /tmp/copa-volei-bot.service /etc/systemd/system/
sudo chown -R copa:copa /opt/copa-volei-bot
sudo -u copa /opt/copa-volei-bot/copa-volei-bot   # escaneia o QR uma vez, ctrl+c
sudo systemctl enable --now copa-volei-bot
sudo journalctl -u copa-volei-bot -f
```

Binário estático, sem cgo e sem dependência de sistema — o SQLite é o driver
puro-Go `modernc.org/sqlite`.

## Aviso

Bibliotecas não-oficiais de WhatsApp ficam fora dos termos de uso. Pra uma
brincadeira em grupo pequeno o risco é baixo, mas se der pra usar um número
secundário (chip antigo, número virtual), melhor.

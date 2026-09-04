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

Quando o sorteio sai, o bot manda a bandeira da copa
(`assets/bandeira.png`) antes da tabela. Ele procura o arquivo em
`$COPA_BANDEIRA`, depois em `./assets/`, depois ao lado do executável — se não
achar, só avisa no log e segue sem a imagem.

A sessão fica em `dados/sessao.db` — depois disso não precisa escanear de novo.
O estado da copa fica em `dados/torneios.json` (um torneio por conversa) e as
fotos em `dados/midia/`.

Variáveis: `COPA_DATA` (pasta de dados, padrão `dados`), `COPA_LOG`
(`DEBUG`/`INFO`/`WARN`, padrão `WARN`) e `COPA_BANDEIRA` (caminho da imagem do
sorteio).

## Comandos

| Comando | O que faz |
|---|---|
| `!ajuda` | lista tudo |
| `!dupla João & Maria` | cadastra uma dupla |
| `!duplas` | lista as inscritas |
| `!remover João` | tira a dupla |
| `!ate 17` | até quantos pontos vai cada jogo |
| `!sortear` | gera a tabela, ida e volta |
| `!jogos` | o que falta jogar, com os IDs |
| `!placar 7 21x18` | placar do jogo J7 |
| `!tabela` | a classificação |
| `!campeonato` | tabela + todas as rodadas |
| `!desfazer` | apaga o último resultado |
| `!zerar CONFIRMA` | recomeça do zero |

### Placar é sempre pelo ID do jogo

Cada jogo tem um ID fixo (`J7`), que aparece em `!jogos` e `!campeonato`. O
placar vai sempre com esse ID — a ordem em que os jogos acontecem não importa,
então dá pra registrar fora de ordem, com duas quadras rodando ao mesmo tempo:

```
!placar 7 21x18
!placar 12 21x15
!placar 9 17x21
```

Mandar de novo no mesmo ID **corrige** o resultado anterior (ele responde
`J7 corrigido` e refaz a tabela). `!desfazer` apaga o último placar registrado.

### Fotos

Manda a foto (ou sticker) da dupla comemorando **logo depois do placar** — ou
com `!placar 7 21x18` na legenda da imagem — que ela fica grudada naquele jogo
(aparece um 📸 na lista de rodadas). No fim, a foto da campeã volta com a
mensagem do título. A janela pra anexar é de 20 minutos após o placar.

## Formato

Pontos corridos: **todas as duplas jogam contra todas, duas vezes**, e campeã é
quem somar mais pontos no fim. Não tem grupo nem mata-mata, ninguém é eliminado.

- O placar é a **quantidade de pontos do jogo** (`21x18`), não sets. Quem fizer
  mais pontos vence.
- Todos os jogos vão até o mesmo alvo. O bot anota o alvo sozinho no primeiro
  placar que você registrar (`17x6` → alvo 17) e confere os seguintes: recusa
  um placar cujo vencedor não chegou lá, e recusa um que passou do alvo sem ser
  vantagem (`20x5` com alvo 17). Vantagem passa normal: `19x17`, `18x16`.
  `!ate 21` muda o alvo, `!ate livre` desliga a conferência.
- Cada vitória vale **3 pontos** na tabela, independente da margem. A margem
  entra no saldo, que é o primeiro critério de desempate.
- A tabela é gerada pelo método do círculo, então dentro de cada rodada nenhuma
  dupla joga duas vezes seguidas — todo mundo descansa parecido.
- **Número ímpar de duplas**: uma dupla folga por rodada, e ao longo do turno
  cada uma folga exatamente uma vez (duas no total, com o returno). A folga
  aparece em `!jogos` e `!campeonato` como `😴 folga: Fulano & Ciclano`.
  Ninguém joga a mais nem a menos: todas fazem os mesmos `2×(n-1)` jogos.
- No returno o mando de quadra inverte: quem foi o lado A na ida é o lado B na
  volta.
- Desempate: pontos → vitórias → saldo de pontos → pontos marcados →
  confronto direto.
- Enquanto rola, a tabela mostra o líder e quem já está **sem chance de título**
  (quem não alcança mais o líder nem ganhando tudo que falta). Quem está
  empatado em pontos com o líder segue vivo, porque ainda pode levar no
  desempate.

## Deploy na VPS

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o copa-volei-bot .
scp -r copa-volei-bot copa-volei-bot.service assets usuario@vps:/tmp/
```

Na VPS:

```bash
sudo useradd -r -m -d /opt/copa-volei-bot copa
sudo mv /tmp/copa-volei-bot /tmp/assets /opt/copa-volei-bot/
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

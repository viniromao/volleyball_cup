# 🏐 Bot da Copa de Vôlei

Bot de WhatsApp em Go que organiza uma copa de duplas num grupo de amigos:
cadastra duplas, sorteia grupos, monta a fase de grupos e o mata-mata,
registra placares e mostra as chaves com quem caiu e quem continua.

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
| `!sortear` | sorteia os grupos e monta os jogos |
| `!sortear 2` | força a quantidade de grupos |
| `!jogos` | o que falta jogar |
| `!placar 2x1` | resultado do jogo da vez |
| `!placar 7 2x1` | resultado do jogo J7 |
| `!tabela` | classificação dos grupos |
| `!chaves` | grupos, mata-mata, eliminadas e quem segue |
| `!desfazer` | apaga o último resultado |
| `!zerar CONFIRMA` | recomeça do zero |

### Fotos

Manda a foto (ou sticker) da dupla comemorando **logo depois do placar** — ou
com `!placar 2x1` na legenda da imagem — que ela fica grudada naquele jogo
(aparece um 📸 na chave). No fim, a foto da campeã volta com a mensagem do
título. A janela pra anexar é de 20 minutos após o placar.

## Formato

- Grupos sorteados na hora: ~4 duplas por grupo (`!sortear N` muda isso).
- Fase de grupos: todos contra todos dentro do grupo.
- Pontuação (formato `sets`): 2x0 = 3 pts, 2x1 = 2 pts, 1x2 = 1 pt, 0x2 = 0.
- Desempate: pontos → saldo de sets → sets ganhos → confronto direto.
- Classificam-se as 2 melhores de cada grupo (ou as 4 melhores, se houver
  um grupo só). Quem não passa, cai já na chave como eliminada.
- Mata-mata com chaveamento por posição (1º de um grupo pega 2º de outro) e
  bye pros melhores quando o número de classificadas não é potência de 2.

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

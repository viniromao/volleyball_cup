# 🏐 Bot da Copa de Vôlei

Bot de WhatsApp em Go que organiza uma copa de times em pontos corridos num
grupo de amigos: cadastra times (de qualquer tamanho), gera a tabela de todos contra todos,
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

Todo comando começa com `!cup` — o resto das mensagens do grupo (inclusive
outros bots com `!`) é ignorado. `!cup` sozinho mostra a ajuda.

| Comando | O que faz |
|---|---|
| `!cup` / `!cup ajuda` | lista tudo |
| `!cup time João & Maria & Pedro` | cadastra um time |
| `!cup times` | lista os inscritos |
| `!cup remover João` | tira o time |
| `!cup ate 17` | até quantos pontos vai cada jogo |
| `!cup sortear` | gera a tabela, ida e volta |
| `!cup jogos` | o que falta jogar, com os IDs |
| `!cup placar 7 21x18` | placar do jogo J7 |
| `!cup tabela` | a classificação |
| `!cup campeonato` | tabela + todas as rodadas |
| `!cup desfazer` | apaga o último resultado |
| `!cup zerar CONFIRMA` | recomeça do zero |

### Times

Cada time tem quantas pessoas quiser, e o tamanho não precisa bater: um time de
3 enfrenta um de 2 normalmente. Os nomes se separam com `&`, `,`, `+` ou `e`
(`!cup time Ana, Bia e Caio`); sem separador, cada palavra é uma pessoa
(`!cup time Ana Bia Caio`) — pra nome com sobrenome, use o separador
(`!cup time João Silva & Maria`). A mesma pessoa não entra em dois times.

Torneios salvos do tempo das duplas continuam valendo: cada dupla vira um time
de 2 quando o bot carrega os dados.

### Placar é sempre pelo ID do jogo

Cada jogo tem um ID fixo (`J7`), que aparece em `!cup jogos` e `!cup campeonato`. O
placar vai sempre com esse ID — a ordem em que os jogos acontecem não importa,
então dá pra registrar fora de ordem, com duas quadras rodando ao mesmo tempo:

```
!cup placar 7 21x18
!cup placar 12 21x15
!cup placar 9 17x21
```

Mandar de novo no mesmo ID **corrige** o resultado anterior (ele responde
`J7 corrigido` e refaz a tabela). `!cup desfazer` apaga o último placar registrado.

### Fotos

Manda a foto (ou sticker) do time comemorando **logo depois do placar** — ou
com `!cup placar 7 21x18` na legenda da imagem — que ela fica grudada naquele jogo
(aparece um 📸 na lista de rodadas). No fim, a foto do campeão volta com a
mensagem do título. A janela pra anexar é de 20 minutos após o placar.

## Formato

Pontos corridos: **todos os times jogam contra todos, duas vezes**, e campeão é
quem somar mais pontos no fim. Não tem grupo nem mata-mata, ninguém é eliminado.

- O placar é a **quantidade de pontos do jogo** (`21x18`), não sets. Quem fizer
  mais pontos vence.
- Todos os jogos vão até o mesmo alvo. O bot anota o alvo sozinho no primeiro
  placar com folga de 2 ou mais (`17x6` → alvo 17). Placar apertado (`8x7`)
  não diz o alvo — pode ser jogo até 7 que foi pra vantagem — então o bot
  pergunta e espera um `!cup ate 7`. Com o alvo, confere os seguintes: recusa
  um placar cujo vencedor não chegou lá, e recusa um que passou do alvo sem ser
  vantagem (`20x5` com alvo 17). Vantagem passa normal: `19x17`, `18x16`.
  `!cup ate 21` muda o alvo, `!cup ate livre` desliga a conferência.
- Cada vitória vale **3 pontos** na tabela, independente da margem. A margem
  entra no saldo, que é o primeiro critério de desempate.
- A tabela é gerada pelo método do círculo, então dentro de cada rodada nenhum
  time joga duas vezes seguidas — todo mundo descansa parecido.
- **Número ímpar de times**: um time folga por rodada, e ao longo do turno
  cada um folga exatamente uma vez (duas no total, com o returno). A folga
  aparece em `!cup jogos` e `!cup campeonato` como `😴 folga: Fulano & Ciclano`.
  Ninguém joga a mais nem a menos: todos fazem os mesmos `2×(n-1)` jogos.
- No returno o mando de quadra inverte: quem foi o lado A na ida é o lado B na
  volta.
- Desempate: pontos → vitórias → saldo de pontos → pontos marcados →
  confronto direto.
- Enquanto rola, a tabela mostra o líder e quem já está **sem chance de título**
  (quem não alcança mais o líder nem ganhando tudo que falta). Quem está
  empatado em pontos com o líder segue vivo, porque ainda pode levar no
  desempate.

## Deploy na VPS

O bot roda 24/7 na VPS como serviço systemd (`copa-volei-bot.service`, usuário
`copa`, pasta `/opt/copa-volei-bot`) com `Restart=always`, então volta sozinho
se cair ou se a máquina reiniciar.

Todo push na `main` dispara o workflow `.github/workflows/deploy.yml`, que
compila o binário para linux/amd64, manda ele e a pasta `assets/` por `scp` e
reinicia o serviço. Leva cerca de um minuto.

Secrets usados pelo workflow (em *Settings > Secrets and variables > Actions*):

| Secret | O que é |
|---|---|
| `VPS_HOST` | endereço da VPS |
| `VPS_USER` | usuário do deploy (`copa`) |
| `VPS_SSH_KEY` | chave privada ed25519 só desse usuário |
| `VPS_KNOWN_HOSTS` | linhas do `known_hosts` da VPS |

O usuário `copa` não tem senha e só pode usar `sudo` para
`systemctl restart|status|is-active copa-volei-bot` (`/etc/sudoers.d/copa`).

A pasta `dados/` (sessão do WhatsApp, torneios e fotos) vive só na VPS e nunca
vai pro git. Primeira instalação numa máquina nova: copiar `dados/` de uma
máquina já pareada ou rodar o binário uma vez na mão pra escanear o QR.

Só uma instância pode usar a sessão por vez — rodar o bot localmente enquanto o
serviço está de pé derruba um dos dois.

Logs:

```bash
ssh usuario@vps 'journalctl -u copa-volei-bot -f'
```

Binário estático, sem cgo e sem dependência de sistema — o SQLite é o driver
puro-Go `modernc.org/sqlite`.

## Aviso

Bibliotecas não-oficiais de WhatsApp ficam fora dos termos de uso. Pra uma
brincadeira em grupo pequeno o risco é baixo, mas se der pra usar um número
secundário (chip antigo, número virtual), melhor.

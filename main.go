package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	qrterminal "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"
)

const janelaFoto = 20 * time.Minute

type Bot struct {
	cli      *whatsmeow.Client
	mu       sync.Mutex
	torneios map[string]*Torneio
	dir      string
	inicio   time.Time
}

func main() {
	dir := os.Getenv("COPA_DATA")
	if dir == "" {
		dir = "dados"
	}
	if err := os.MkdirAll(filepath.Join(dir, "midia"), 0o755); err != nil {
		fatal(err)
	}

	ctx := context.Background()
	dsn := "file:" + filepath.Join(dir, "sessao.db") + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		fatal(err)
	}
	db.SetMaxOpenConns(1)

	nivel := os.Getenv("COPA_LOG")
	if nivel == "" {
		nivel = "WARN"
	}
	container := sqlstore.NewWithDB(db, "sqlite", waLog.Stdout("DB", nivel, true))
	if err := container.Upgrade(ctx); err != nil {
		fatal(err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		fatal(err)
	}

	bot := &Bot{
		torneios: map[string]*Torneio{},
		dir:      dir,
		inicio:   time.Now(),
	}
	if err := bot.Carregar(); err != nil {
		fatal(err)
	}

	bot.cli = whatsmeow.NewClient(device, waLog.Stdout("WA", nivel, true))
	bot.cli.AddEventHandler(bot.handler)

	if bot.cli.Store.ID == nil {
		qrChan, err := bot.cli.GetQRChannel(ctx)
		if err != nil {
			fatal(err)
		}
		if err := bot.cli.Connect(); err != nil {
			fatal(err)
		}
		for evt := range qrChan {
			switch evt.Event {
			case "code":
				fmt.Println("\n📱 WhatsApp > Aparelhos conectados > Conectar aparelho\nEscaneia esse QR:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			case "success":
				fmt.Println("✅ Conectado!")
			default:
				fmt.Println("QR:", evt.Event)
			}
		}
	} else if err := bot.cli.Connect(); err != nil {
		fatal(err)
	}

	fmt.Println("🏐 bot da copa no ar. manda !ajuda no grupo. ctrl+c pra sair.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	bot.cli.Disconnect()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "erro:", err)
	os.Exit(1)
}

func (b *Bot) handler(raw any) {
	switch evt := raw.(type) {
	case *events.Message:
		b.aoReceber(evt)
	case *events.Connected:
		fmt.Println("conectado ao whatsapp")
	case *events.LoggedOut:
		fmt.Println("sessão encerrada no celular. apague dados/sessao.db e escaneie o QR de novo.")
		os.Exit(1)
	}
}

func desembrulha(msg *waE2E.Message) *waE2E.Message {
	for i := 0; i < 4 && msg != nil; i++ {
		switch {
		case msg.GetEphemeralMessage().GetMessage() != nil:
			msg = msg.GetEphemeralMessage().GetMessage()
		case msg.GetViewOnceMessage().GetMessage() != nil:
			msg = msg.GetViewOnceMessage().GetMessage()
		case msg.GetViewOnceMessageV2().GetMessage() != nil:
			msg = msg.GetViewOnceMessageV2().GetMessage()
		case msg.GetDocumentWithCaptionMessage().GetMessage() != nil:
			msg = msg.GetDocumentWithCaptionMessage().GetMessage()
		default:
			return msg
		}
	}
	return msg
}

func textoDe(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if c := msg.GetConversation(); c != "" {
		return c
	}
	if e := msg.GetExtendedTextMessage(); e != nil {
		return e.GetText()
	}
	if i := msg.GetImageMessage(); i != nil {
		return i.GetCaption()
	}
	if v := msg.GetVideoMessage(); v != nil {
		return v.GetCaption()
	}
	return ""
}

func (b *Bot) aoReceber(evt *events.Message) {
	if evt.Info.Timestamp.Before(b.inicio) {
		return
	}
	msg := desembrulha(evt.Message)
	texto := strings.TrimSpace(textoDe(msg))
	temMidia := msg.GetImageMessage() != nil || msg.GetStickerMessage() != nil

	if !strings.HasPrefix(texto, "!") && !temMidia {
		return
	}

	chat := evt.Info.Chat.String()

	b.mu.Lock()
	t, ok := b.torneios[chat]
	if !ok {
		t = NovoTorneio(chat)
		b.torneios[chat] = t
	}

	var resposta string
	var salvar bool
	sorteadoAntes := t.Sorteado
	if strings.HasPrefix(texto, "!") {
		resposta, salvar = b.Executar(t, texto)
	}
	sorteouAgora := !sorteadoAntes && t.Sorteado
	alvoFoto := t.UltimoJogo
	dentroDaJanela := time.Since(t.UltimoEm) < janelaFoto
	b.mu.Unlock()

	if sorteouAgora {
		b.hastearBandeira(evt.Info.Chat)
	}
	if resposta != "" {
		b.responder(evt.Info.Chat, resposta)
	}

	if temMidia && alvoFoto != 0 && dentroDaJanela {
		if b.anexarFoto(evt, chat, alvoFoto) {
			salvar = true
		}
	}

	b.mu.Lock()
	campea := t.Campea
	anunciar := campea != 0 && !t.AnunciouCampea
	if anunciar {
		t.AnunciouCampea = true
		salvar = true
	}
	if salvar {
		if err := b.Salvar(); err != nil {
			fmt.Fprintln(os.Stderr, "erro ao salvar:", err)
		}
	}
	foto := ""
	nome := ""
	if anunciar {
		foto = t.FotoDaDupla(campea)
		nome = t.NomeDupla(campea)
	}
	b.mu.Unlock()

	if anunciar {
		b.comemorar(evt.Info.Chat, nome, foto)
	}
}

var naoAlfa = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func (b *Bot) anexarFoto(evt *events.Message, chat string, jogoID int) bool {
	msg := desembrulha(evt.Message)
	var dados []byte
	var err error
	ext := ".jpg"
	if img := msg.GetImageMessage(); img != nil {
		dados, err = b.cli.Download(context.Background(), img)
		if strings.Contains(img.GetMimetype(), "png") {
			ext = ".png"
		}
	} else if st := msg.GetStickerMessage(); st != nil {
		dados, err = b.cli.Download(context.Background(), st)
		ext = ".webp"
	} else {
		return false
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro no download:", err)
		return false
	}

	pasta := filepath.Join(b.dir, "midia", naoAlfa.ReplaceAllString(chat, "_"))
	if err := os.MkdirAll(pasta, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "erro na pasta:", err)
		return false
	}
	caminho := filepath.Join(pasta, fmt.Sprintf("J%d%s", jogoID, ext))
	if err := os.WriteFile(caminho, dados, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "erro ao gravar:", err)
		return false
	}

	b.mu.Lock()
	t := b.torneios[chat]
	if t == nil {
		b.mu.Unlock()
		return false
	}
	j := t.Jogo(jogoID)
	if j == nil {
		b.mu.Unlock()
		return false
	}
	novo := j.Foto == ""
	j.Foto = caminho
	venc := t.NomeDupla(j.Vencedor())
	b.mu.Unlock()

	if novo {
		b.responder(evt.Info.Chat, fmt.Sprintf("📸 Guardei essa como a foto do J%d — *%s*.", jogoID, venc))
	}
	return true
}

func (b *Bot) responder(chat types.JID, texto string) {
	_, err := b.cli.SendMessage(context.Background(), chat, &waE2E.Message{
		Conversation: proto.String(texto),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro ao enviar:", err)
	}
}

func (b *Bot) enviarImagem(chat types.JID, caminho, legenda string) error {
	ext := strings.ToLower(filepath.Ext(caminho))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return fmt.Errorf("formato não enviável: %s", ext)
	}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return err
	}
	up, err := b.cli.Upload(context.Background(), dados, whatsmeow.MediaImage)
	if err != nil {
		return err
	}
	mime := "image/jpeg"
	if ext == ".png" {
		mime = "image/png"
	}
	msg := &waE2E.ImageMessage{
		Mimetype:      proto.String(mime),
		URL:           &up.URL,
		DirectPath:    &up.DirectPath,
		MediaKey:      up.MediaKey,
		FileEncSHA256: up.FileEncSHA256,
		FileSHA256:    up.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(dados))),
	}
	if legenda != "" {
		msg.Caption = proto.String(legenda)
	}
	_, err = b.cli.SendMessage(context.Background(), chat, &waE2E.Message{ImageMessage: msg})
	return err
}

func (b *Bot) hastearBandeira(chat types.JID) {
	caminho := b.bandeira()
	if caminho == "" {
		return
	}
	if err := b.enviarImagem(chat, caminho, ""); err != nil {
		fmt.Fprintln(os.Stderr, "erro ao enviar a bandeira:", err)
	}
}

func (b *Bot) bandeira() string {
	var candidatos []string
	if p := os.Getenv("COPA_BANDEIRA"); p != "" {
		candidatos = append(candidatos, p)
	}
	candidatos = append(candidatos, filepath.Join("assets", "bandeira.png"))
	if exe, err := os.Executable(); err == nil {
		candidatos = append(candidatos, filepath.Join(filepath.Dir(exe), "assets", "bandeira.png"))
	}
	for _, c := range candidatos {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	fmt.Fprintln(os.Stderr, "aviso: não achei assets/bandeira.png, sigo sem a bandeira")
	return ""
}

func (b *Bot) comemorar(chat types.JID, nome, foto string) {
	legenda := fmt.Sprintf("🏆🏐 *CAMPEÃS DA COPA: %s* 🏐🏆\n\nAcabou! Parabéns, duplas. `!tabela` mostra como terminou.", nome)
	if foto == "" {
		b.responder(chat, legenda)
		return
	}
	if err := b.enviarImagem(chat, foto, legenda); err != nil {
		b.responder(chat, legenda)
	}
}

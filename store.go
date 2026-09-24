package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func (b *Bot) arquivo() string {
	return filepath.Join(b.dir, "torneios.json")
}

func (b *Bot) Carregar() error {
	data, err := os.ReadFile(b.arquivo())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := json.Unmarshal(data, &b.torneios); err != nil {
		return err
	}
	for _, t := range b.torneios {
		t.Migrar()
	}
	return nil
}

func (b *Bot) Salvar() error {
	data, err := json.MarshalIndent(b.torneios, "", "  ")
	if err != nil {
		return err
	}
	tmp := b.arquivo() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, b.arquivo())
}

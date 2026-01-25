// internal/adapters/inbound/http/web/embed.go
package web

import (
	"embed"
	"io/fs"
	"os"
)

// Embeda TUDO que precisa em produção.
// Ajuste os paths conforme seu repo.
var (
	//go:embed public/** templates/**
	embedded embed.FS
)

type FSBundle struct {
	Public    fs.FS
	Templates fs.FS
}

func LoadFS(env string) (FSBundle, error) {
	// DEV: lê do disco (feedback loop rápido)
	if env == "dev" {
		// assume que o working dir é a raiz do repo
		return FSBundle{
			Public:    os.DirFS("internal/adapters/inbound/http/web/public"),
			Templates: os.DirFS("internal/adapters/inbound/http/web/templates"),
		}, nil
	}

	// PROD: usa embed (binário self-contained)
	pub, err := fs.Sub(embedded, "public")
	if err != nil {
		return FSBundle{}, err
	}
	tpl, err := fs.Sub(embedded, "templates")
	if err != nil {
		return FSBundle{}, err
	}

	return FSBundle{Public: pub, Templates: tpl}, nil
}

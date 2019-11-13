package main

import (
	"leobcn/packages/cfgmngr"
	"log"

	"github.com/davecgh/go-spew/spew"
)

type Config struct {
	Path      string `toml:"path" long:"path" description:"ruta de trabajo"`
	Modo      string `toml:"modo" long:"modo" description:"modo de trabajo"`
	MaxConcur int    `toml:"max_conc" long:"maxconc" description:"cantidad máxima de procesos concurrentes"`
	Mode      string `toml:"-" long:"mode" def:"PRODUCTION"`
	Product   string `toml:"product" long:"product" def:"MLM-DEFAULT" env:"STK_PRODUCT"`
	Test      bool   `long:"test" env:"STK_TEST"`

	Version bool `short:"v" long:"version"`
}

func main() {
	cfg := Config{}
	if err := cfgmngr.Parse(&cfg, "config.toml"); err != nil {
		log.Fatal(err)
	}

	spew.Dump(cfg)
}

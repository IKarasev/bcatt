package main

import (
	"github.com/IKarasev/bcatt/internal/blockchain"
	"github.com/IKarasev/bcatt/internal/emulator"
	"github.com/IKarasev/bcatt/internal/globals"

	"fmt"
	"os"
)

func main() {
	confPath := "config.yaml"
	if len(os.Args) > 1 {
		confPath = os.Args[1]
	}

	cf, _ := globals.ReadConfig(confPath)

	if err := blockchain.InitBcSettings(cf); err != nil {
		fmt.Println(err)
	}
	if err := emulator.InitSettings(cf); err != nil {
		fmt.Println(err)
	}
	wb := emulator.NewEmulatorWeb().DefaultNodeManager()
	wb.RcMngr.WithNetmap(cf.Blockchain.NetMap)
	if emulator.WITH_LOG {
		wb.StartWithLogger()
	} else {
		wb.Start()
	}
}

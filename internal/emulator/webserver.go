package emulator

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/IKarasev/bcatt/internal/agregator"
	nm "github.com/IKarasev/bcatt/internal/nodemanager"
)

type EmulatorWeb struct {
	RcMngr            *nm.NodeManager
	DataAgr           *agregator.DataAgregator
	E                 *echo.Echo
	rssChan           RssChan
	RssReadUpdateTime time.Duration
	ctx               context.Context
}

func NewEmulatorWeb() *EmulatorWeb {
	return &EmulatorWeb{
		RcMngr:            nm.NewNodeManager(),
		DataAgr:           agregator.NewDataAgreator(),
		E:                 echo.New(),
		rssChan:           make(RssChan),
		RssReadUpdateTime: RSS_READ_UPDATE_TIME,
		ctx:               context.Background(),
	}
}

func (wb *EmulatorWeb) DefaultNodeManager() *EmulatorWeb {
	wb.RcMngr = nm.DefaultNodeManager()
	if START_UTXO.Active {
		if START_UTXO.All {
			wb.RcMngr.GenerateUtxoForAll(START_UTXO.Nmin, START_UTXO.Nmax, START_UTXO.Vmin, START_UTXO.Vmax)
		} else {
			wb.RcMngr.GenerateUtxoForWallet(START_UTXO.Wallets, START_UTXO.Nmin, START_UTXO.Nmax, START_UTXO.Vmin, START_UTXO.Vmax)
		}
	}
	return wb
}

func (wb *EmulatorWeb) StartWithLogger() {
	wb.E.Use(middleware.Logger())
	wb.E.Logger.Fatal(wb.Start())
}

func (wb *EmulatorWeb) Start() error {
	fmt.Println(HTTP_ADDR + ":" + HTTP_PORT)
	ctx, ctxDone := context.WithCancel(wb.ctx)
	wb.ctx = ctx
	defer ctxDone()
	defer close(wb.rssChan)

	wb.initRoutes()
	err := wb.E.Start(HTTP_ADDR + ":" + HTTP_PORT)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (wb *EmulatorWeb) initRoutes() {
	wd, _ := os.Getwd()
	wd = wd + "/assets"

	wb.E.Static("/static", "assets")

	wb.E.GET("/", wb.HandleIndex)
	// wb.E.GET("/test", wb.HandleTest)
	wb.E.GET("/sse", wb.HandleSse)
	wb.E.GET("/nodelist", wb.HandleNodeList)
	wb.E.GET("/selectminer", wb.HandleMinerSelect)
	wb.E.GET("/setminer", wb.HandleMinerSet)
	wb.E.GET("/tick", wb.HandleTick)
	wb.E.GET("/settings", wb.HandleEmulationSettings)

	gNode := wb.E.Group("/node")
	gNode.GET("/slist", wb.HandleNodeSelectList)
	gNode.POST("/info", wb.HandleNodeInfo)
	gNode.POST("/block", wb.HandleBlockDetails)
	gNode.POST("/block/tr", wb.HandleBlockTransactions)

	gWallet := wb.E.Group("/wallet")
	gWallet.POST("/slist", wb.HandleWalletList)
	gWallet.POST("/utxotable", wb.HandleWalletUtxoTable)
	gWallet.POST("/addtr", wb.HandleAddTransaction)
	gWallet.POST("/blocktr", wb.HandleWalletBlockTr)

	gEvil := wb.E.Group("/evil")
	gEvil.GET("/load", wb.HandleEvilLoad)
	gEvil.GET("/steal", wb.HandleEvilSteal)
	gEvil.GET("/create", wb.HandleEvilCreate)
	gEvil.GET("/mine", wb.HandleEvilMine)
	gEvil.GET("/inject", wb.HandleEvilInject)
	gEvil.GET("/send", wb.HandleEvilSend)

	gEvilSet := gEvil.Group("/set")
	gEvilSet.POST("/height", wb.HandleEvilSetHeihgt)
	gEvilSet.POST("/time", wb.HandleEvilSetTime)
	gEvilSet.POST("/hash", wb.HandleEvilSetHash)
	gEvilSet.POST("/nonce", wb.HandleEvilSetInt)
	gEvilSet.POST("/coinbase", wb.HandleEvilSetInt)
	gEvilSet.POST("/tr/sign", wb.HandleEvilSetTrHashValue)
	gEvilSet.POST("/tr/pk", wb.HandleEvilSetTrHashValue)
	gEvilSet.POST("/tr/utxo", wb.HandleEvilSetTrUtxo)

	gEvilDel := gEvil.Group("/del")
	gEvilDel.POST("/utxo", wb.HandleEvilDelUtxo)
	gEvilDel.POST("/tr", wb.HandleEvilDelTr)

	gEvilAdd := gEvil.Group("/add")
	gEvilAdd.GET("/tr", wb.HandleEvilAddTr)
	gEvilAdd.POST("/utxo", wb.HandleEvilAddUtxo)

	gVisual := wb.E.Group("/visual")
	gVisual.GET("/net", wb.HandleGetNetMap)
	gVisual.GET("/fork", wb.HandleGetChainFork)
}

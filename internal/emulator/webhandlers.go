package emulator

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IKarasev/bcatt/internal/blockchain"
	"github.com/IKarasev/bcatt/views"

	"github.com/labstack/echo/v4"
)

func (wb *EmulatorWeb) HandleIndex(c echo.Context) error {
	return renderTempl(c, views.Index(strconv.Itoa(wb.RcMngr.Tick)))
}

func (wb *EmulatorWeb) HandleSse(c echo.Context) error {

	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-c.Request().Context().Done():
			log.Printf("WebServer: SSE client disconnected, ip: %v\n", c.RealIP())
			return nil
		case <-wb.ctx.Done():
			log.Println("WebServer: main ctx done!")
			return nil
		case msg := <-wb.rssChan:
			if err := msg.MarshalTo(w); err != nil {
				log.Println("WebServer: Failed to marshal event")
				return err
			}
			w.Flush()
		}
		time.Sleep(wb.RssReadUpdateTime)
	}
}

func (wb *EmulatorWeb) HandleNodeList(c echo.Context) error {
	ncList := make([]views.NodeCellInput, len(wb.RcMngr.Nodes))
	i := 0
	minerId := ""
	if wb.RcMngr.MainNode() != nil {
		minerId = wb.RcMngr.MainNode().Id
	}
	for _, node := range wb.RcMngr.Nodes {
		n := views.NodeCellInput{}
		n.Name = node.Name
		n.Id = node.Id
		n.Coinbase = strconv.Itoa(node.CoinbaseUtxoAmount())
		n.WName = node.Wallet.Name
		n.WAddress = node.Wallet.Addr
		n.WCoins = strconv.Itoa(node.Wallet.Balance())
		n.Miner = n.Id == minerId

		b := node.GetLastBlock()
		if b != nil {
			n.BHeight = strconv.Itoa(b.Body.Coinbase)
			n.BHash = b.HashString()
			n.BCoinbase = strconv.Itoa(b.Body.Coinbase)
			n.BNonce = strconv.Itoa(b.Header.Nonce)
			n.BRoot = b.RootString()
		}
		ncList[i] = n
		i++
	}
	if err := renderTempl(c, views.NodeCellList(ncList)); err != nil {
		log.Printf("ERROR: Node cell list: %s", err)
	}
	slices.SortFunc(ncList, func(a, b views.NodeCellInput) int {
		if a.Name > b.Name {
			return 1
		} else if a.Name < b.Name {
			return -1
		}
		return 0
	})
	return nil
}

func (wb *EmulatorWeb) HandleBlockDetails(ctx echo.Context) error {
	nId := ctx.FormValue("node")
	bh := ctx.FormValue("block")
	h, err := strconv.Atoi(bh)
	if err != nil {
		wb.RssLogErrorSend("Block detailes failed: invalid block height")
		return err
	}
	b, err := wb.RcMngr.GetNodeBlock(nId, h)
	if err != nil {
		wb.RssLogErrorSend(fmt.Sprintf("Block detailes failed: %s", err))
		return err
	}
	bi := blockToItem(b)
	bi.Height = bh
	return renderTempl(ctx, views.NodeBLocksBlockDetailed(nId, bi))
}

func (wb *EmulatorWeb) HandleBlockTransactions(ctx echo.Context) error {
	nId := ctx.FormValue("node")
	bh := ctx.FormValue("block")
	h, err := strconv.Atoi(bh)
	if err != nil {
		wb.RssLogErrorSend("Block Transactions failed: Invalid block height")
		return err
	}
	b, err := wb.RcMngr.GetNodeBlock(nId, h)
	if err != nil {
		wb.RssLogErrorSend(fmt.Sprintf("Block Transactions failed: %s", err))
		return err
	}
	trItems := make([]views.BlockTransactionItem, len(b.Body.Transactions))
	for i, tr := range b.Body.Transactions {
		t := views.BlockTransactionItem{
			Sign:       blockchain.BytesToString(tr.Sign),
			Pk:         blockchain.BytesToString(tr.Pk),
			InputUtxo:  make([]views.UtxoItem, len(tr.InputUtxo)),
			OutputUtxo: make([]views.UtxoItem, len(tr.OutputUtxo)),
		}
		j := 0
		for _, u := range tr.InputUtxo.SortedItems() {
			t.InputUtxo[j] = views.UtxoItem{
				Amount: strconv.Itoa(u.Amount),
				Addr:   u.Addr,
			}
			j++
		}
		j = 0
		for _, u := range tr.OutputUtxo.SortedItems() {
			t.OutputUtxo[j] = views.UtxoItem{
				Amount: strconv.Itoa(u.Amount),
				Addr:   u.Addr,
			}
			j++
		}
		trItems[i] = t
	}
	return renderTempl(ctx, views.NodeBlockTransactions(trItems))
}

func (wb *EmulatorWeb) HandleMinerSelect(ctx echo.Context) error {
	wb.RssLogInfoSend("Selecting new Miner")
	wb.RcMngr.SelectMainNode()
	wb.RssSendMinerSelect()
	return nil
}

func (wb *EmulatorWeb) HandleMinerSet(ctx echo.Context) error {
	id := ctx.FormValue("nodeid")
	if id == "" {
		return nil
	}
	wb.RssLogInfoSend(fmt.Sprintf("Setting new Miner %s", id))
	if _, err := wb.RcMngr.SetMainNode(id); err != nil {
		wb.RssLogErrorSend("ERROR: %s", err)
		return nil
	}
	wb.RssSendMinerSelect()
	return nil
}

func (wb *EmulatorWeb) HandleTick(ctx echo.Context) error {
	if len(wb.RcMngr.Nodes) == 0 {
		wb.RssLogErrorSend("No Nodes exists. Aborting TICK operation")
		return nil
	}
	if wb.RcMngr.Tick == 0 {
		wb.RssLogInfoSend("First tick: initiating GENESIS block")
		return wb.HandleTickGenesis(ctx)
	}
	return wb.HandleTickGeneral(ctx)
}

func (wb *EmulatorWeb) HandleTickGenesis(ctx echo.Context) error {
	logTitle := "Genesis Block: "
	if wb.RcMngr.MainNode() == nil {
		wb.RssLogInfoSend(logTitle + "no main node selected. Selecting Main node.")
		wb.SelectMainNode()
		time.Sleep(OP_PAUSE_MILISEC)
	}
	n := wb.RcMngr.MainNode()
	wb.RssLogInfoSend(fmt.Sprintf("Node %s: Creating genesis block candidate", n.Name))
	_, err := n.InitGenesisBlock()
	if err != nil {
		wb.RssLogErrorSend(logTitle + err.Error() + " : ABORTING")
		return err
	}

	time.Sleep(OP_PAUSE_MILISEC)

	wb.RssLogOKSend(logTitle + "genesis block candidate created")
	wb.RssLogInfoSend(logTitle + "Start block mining")
	t := time.Now()
	b, err := wb.RcMngr.Mine()
	d := time.Since(t)
	if err != nil {
		wb.RssLogErrorSend(fmt.Sprintf("%sMine fialed: %s", logTitle, err))
		return err
	}

	wb.aggregatorAddBlock(b, n.Name, false)

	wb.RcMngr.Tick++

	wb.RssLogOKSend(logTitle+"Genesis block mined succesfully in %2.f seconds", d.Seconds())

	wb.RssNodeAllUpdates(n.Id)

	wb.RssLogInfoSend("Sending genesis block to other nodes")
	wb.spreadBlock(n.Id, b)
	wb.RssLogOKSend(logTitle + "Genesis block add succesfully")

	time.Sleep(OP_PAUSE_MILISEC)
	wb.SelectMainNode()
	wb.RssTick()
	return nil
}

func (wb *EmulatorWeb) HandleTickGeneral(ctx echo.Context) error {
	logPrefix := fmt.Sprintf("New Tick (%d): ", wb.RcMngr.Tick+1)

	ferr := func(msg string) error {
		wb.RssLogErrorSend(msg)
		return fmt.Errorf(logPrefix+"%s", msg)
	}

	wb.RssLogInfoSend(logPrefix + "Starting")

	n := wb.RcMngr.MainNode()
	if n == nil {
		return ferr("No Miner node selected. Aborting new tick.")
	}
	wb.RssLogInfoSend(logPrefix+"Node [%s] start mining...", n.Name)
	t := time.Now()
	b, err := wb.RcMngr.Mine()
	if err != nil {
		wb.RcMngr.Tick++
		wb.RssTick()
		wb.SelectMainNode()
		return ferr(err.Error())
	}
	wb.aggregatorAddBlock(b, n.Name, false)
	wb.RssLogOKSend(logPrefix+"Node [%s] finished mining in %.2f seconds", n.Name, time.Since(t).Seconds())
	wb.RssNodeAllUpdates(n.Id)
	wb.RssLogInfoSend(logPrefix + "Sending block to other Nodes")

	wb.spreadBlock(n.Id, b)

	wb.RcMngr.Tick++
	wb.RssTick()

	wb.SelectMainNode()
	wb.RcMngr.TrGen.Clear()
	return nil
}

func (wb *EmulatorWeb) HandleAddTransaction(ctx echo.Context) error {
	mode := ctx.FormValue("mode")
	if mode == "gen" {
		return wb.HandleAddTransactionGen(ctx)
	} else if mode == "genfull" {
		return wb.HandleAddTransactionGenFull(ctx)
	}
	return wb.HandleAddTransactionNew(ctx)
}

func (wb *EmulatorWeb) HandleAddTransactionNew(ctx echo.Context) error {
	logTitle := "New transaction: "
	ferr := func(msg string) error {
		wb.RssLogErrorSend(logTitle + msg)
		return renderTempl(ctx, views.WalletTrResult(false, msg))
	}

	wb.RssLogInfoSend(logTitle + "START")

	widFrom := ctx.FormValue("WalletList")
	widTo := ctx.FormValue("sendTo")

	if widFrom == "" {
		return ferr("From address is empty")
	}
	if widTo == "" {
		return ferr("To address is empty")
	}
	if widTo == widFrom {
		return ferr("From address and To address nust not be equal")
	}

	inpForm, err := ctx.FormParams()
	if err != nil {
		fmt.Println("Error Form Params\n", err)
		return ferr("Server error")
	}

	delete(inpForm, "WalletList")
	delete(inpForm, "sendTo")
	delete(inpForm, "mode")

	inUtxo := make([]string, 0, len(inpForm))
	outAmount := make([]int, 0, len(inpForm))

	for id := range inpForm {
		am, err0 := strconv.Atoi(ctx.FormValue(id))
		if err0 != nil {
			return ferr(fmt.Sprintf("Input from [%s] is not integer", id))
		}
		if am < 0 {
			return ferr(fmt.Sprintf("Input from [%s] is negative ", id))
		}
		if am == 0 {
			continue
		}
		inUtxo = append(inUtxo, id)
		outAmount = append(outAmount, am)
	}

	if len(inUtxo) == 0 {
		return ferr("No amounts to spend given")
	}

	w, ok := wb.RcMngr.Wallets[widFrom]

	if !ok {
		return ferr(fmt.Sprintf("Wallet [%s] does not exist", widFrom))
	}
	if _, ok = wb.RcMngr.Wallets[widTo]; !ok {
		return ferr(fmt.Sprintf("Wallet [%s] does not exist", widTo))
	}

	t, err := w.NewTransaction(inUtxo, outAmount, widTo)
	if err != nil {
		return ferr(err.Error())
	}

	wb.RssLogInfoSend(logTitle + " Transaction ready. Sending to main node...")

	if err = wb.RcMngr.MainNode().AddVerifyTransaction(*t); err != nil {
		return ferr(err.Error())
	}

	wb.RssLogOKSend("Transaction added succesfully")
	wb.RcMngr.TrGen.AddSpentUtxoFromTr(t)

	return renderTempl(ctx, views.WalletTrResult(true, "Transaction added succesfully"))
}

func (wb *EmulatorWeb) HandleAddTransactionGen(ctx echo.Context) error {
	logTitle := "Gen transaction: "

	ferr := func(msg string) error {
		wb.RssLogErrorSend(logTitle + msg)
		return renderTempl(ctx, views.WalletTrResult(false, msg))
	}
	wb.RssLogInfoSend(logTitle + "START")

	widFrom := ctx.FormValue("WalletList")
	widTo := ctx.FormValue("sendTo")

	if widFrom == "" {
		return ferr("From address is empty")
	}
	if widTo == "" {
		return ferr("To address is empty")
	}
	if widTo == widFrom {
		return ferr("From address and To address nust not be equal")
	}
	fromWallet, ok := wb.RcMngr.Wallets[widFrom]
	if !ok {
		return ferr("wallet [" + widFrom + "] not found")
	}
	n := wb.RcMngr.MainNode()
	if n == nil {
		return ferr("Main node not selected")
	}

	_, totalSend, err := wb.RcMngr.TrGen.GenTransaction(n, fromWallet, widTo)

	if err != nil {
		return ferr(err.Error())
	}

	wb.RssLogOKSend("Transaction generated: %d from %s to %s", totalSend, fromWallet.Name, wb.RcMngr.Wallets[widTo].Name)
	return renderTempl(ctx, views.WalletTrResult(true, "Transaction added succesfully"))
}

func (wb *EmulatorWeb) HandleAddTransactionGenFull(ctx echo.Context) error {
	logTitle := "Transaction generation:"

	trNumStr := ctx.FormValue("tr")
	utxoNumStr := ctx.FormValue("utxo")

	if trNumStr == "" || utxoNumStr == "" {
		wb.RssLogErrorSend("%s Input parameters are empty", logTitle)
		return nil
	}

	trNum, err := strconv.Atoi(trNumStr)
	if err != nil || trNum < 1 {
		wb.RssLogErrorSend("%s Invalid transaction quantity", logTitle)
		return nil
	}
	utxoNum, err := strconv.Atoi(utxoNumStr)
	if err != nil || utxoNum < 1 {
		wb.RssLogErrorSend("%s UInvalid utxo quantity", logTitle)
		return nil
	}

	wb.RssLogInfoSend("%s start generation %d transactions, %d utxo each", logTitle, trNum, utxoNum)
	err = wb.RcMngr.GenerateTransactions(trNum, utxoNum)
	if err != nil {
		wb.RssLogErrorSend("%s %s", logTitle, err)
	} else {
		wb.RssLogOKSend("%s %d transaction with %d utxo generated", logTitle, trNum, utxoNum)
	}
	return nil
}

func (wb *EmulatorWeb) HandleNodeInfo(ctx echo.Context) error {
	nid := ctx.FormValue("nodeId")
	if nid == "" {
		return renderTempl(ctx, views.ItemNotFound("Нода", "Нода не выбрана"))
	}
	n, ok := wb.RcMngr.Nodes[nid]
	if !ok {
		return renderTempl(ctx, views.ItemNotFound("Нода", "Выбранная нода не найдена"))
	}
	nf := views.NodeInfoSm{
		Name:        n.Name,
		Id:          n.Id,
		Coinbase:    strconv.Itoa(n.CoinbaseUtxoAmount()),
		TotalUtxo:   strconv.Itoa(len(n.Utxo)),
		TotalBlocks: strconv.Itoa(len(n.BlockChain)),
		Miner:       true,
	}
	ul := make([]views.UtxoItem, len(n.Utxo))
	if len(ul) > 0 {
		cbu := n.Utxo[blockchain.COINBASE_ADDR]
		ul[0] = views.UtxoItem{
			Amount: strconv.Itoa(cbu.Amount),
			Addr:   cbu.Addr,
		}
		i := 1
		for _, u := range n.Utxo {
			if u.Addr != blockchain.COINBASE_ADDR {
				ul[i] = views.UtxoItem{
					Amount: strconv.Itoa(u.Amount),
					Addr:   u.Addr,
				}
				i++
			}
		}
	}

	bl := make([]views.BlockInfoSmallItem, len(n.BlockChain))
	for i, b := range n.BlockChain {
		bl[i] = blockToItem(b)
	}
	if wb.RcMngr.MainNode() == n {
		if wb.RcMngr.MainNode().BlockCandidate != nil {
			cb := blockToItem(wb.RcMngr.MainNode().BlockCandidate)
			cb.Hash = "Candidate"
			cb.Height = "-1"
			bl = append(bl, cb)
		}
	}
	return renderTempl(ctx, views.NodeInfoFull(nf, bl, ul))
}

func (wb *EmulatorWeb) HandleNodeSelectList(ctx echo.Context) error {
	l := make([]views.SelectListItem, len(wb.RcMngr.Nodes))
	i := 0
	for _, n := range wb.RcMngr.Nodes {
		l[i] = views.SelectListItem{
			Id:   n.Id,
			Name: n.Name,
		}
		i++
	}
	slices.SortFunc(l, func(a, b views.SelectListItem) int {
		if a.Name > b.Name {
			return 1
		} else if a.Name < b.Name {
			return -1
		}
		return 0
	})
	return renderTempl(ctx, views.NodeSelectList(l))
}

func (wb *EmulatorWeb) HandleWalletPage(ctx echo.Context) error {
	return fmt.Errorf("HandleWalletPage not implemented")
}

func (wb *EmulatorWeb) HandleWalletBlockTr(ctx echo.Context) error {
	wid := ctx.FormValue("WalletList")
	bhIn := ctx.FormValue("BlockHeight")
	if wid == "" {
		wb.RssLogErrorSend("Empty wallet id")
		return renderTempl(ctx, views.ItemNotFound("Wallet", "Wallet is not selected"))
	}
	if bhIn == "" {
		wb.RssLogErrorSend("Embty block height")
		return renderTempl(ctx, views.ItemNotFound("Block", "Block height is not not given"))
	}
	wb.RssLogInfoSend(fmt.Sprintf("Searching wallet %s transactions in block %s", wid, bhIn))
	bh, err := strconv.Atoi(bhIn)
	if err != nil {
		wb.RssLogErrorSend("Block height value is invalid")
		return renderTempl(ctx, views.ItemNotFound("Block", "Block height value is invalid"))
	}
	n := wb.RcMngr.GetSetMainNode()
	if bh < 0 || bh >= len(n.BlockChain) {
		wb.RssLogErrorSend("No Block with height " + bhIn)
		return renderTempl(ctx, views.ItemNotFound("Block "+bhIn, "Block not found"))
	}
	if _, ok := wb.RcMngr.Wallets[wid]; !ok {
		wb.RssLogErrorSend("Wallet [" + wid + "] not found")
		return renderTempl(ctx, views.ItemNotFound("Wallet ["+wid+"]", "Wallet not found"))
	}
	b := n.BlockChain[bh]
	trs := []views.WalletBlockTrItem{}
	for _, t := range b.Body.Transactions {
		inu, outu := t.FilterUtxoByWallet(wid)
		if len(inu) == 0 && len(outu) == 0 {
			continue
		}
		item := views.WalletBlockTrItem{
			Sign:       t.SignString(),
			Pk:         t.PkString(),
			InputUtxo:  make([]string, len(inu)),
			OutputUtxo: make([]string, len(outu)),
		}
		i := 0
		for _, u := range inu {
			item.InputUtxo[i] = strconv.Itoa(u.Amount)
			i++
		}
		i = 0
		for _, u := range outu {
			item.OutputUtxo[i] = strconv.Itoa(u.Amount)
			i++
		}
		trs = append(trs, item)
	}
	return renderTempl(
		ctx,
		views.WalletBlockInfo(b.Header.Time.Format("2006-01-02 15:04:05"), b.HashString(), trs),
	)
}

func (wb *EmulatorWeb) HandleWalletList(ctx echo.Context) error {
	wid := ctx.FormValue("wid")
	var wlist []views.SelectListItem
	i := 0
	if wid == "" {
		wlist = make([]views.SelectListItem, len(wb.RcMngr.Wallets))
		for _, w := range wb.RcMngr.Wallets {
			wlist[i] = walletToSelectListItems(w)
			i++
		}
	} else {
		wlist = []views.SelectListItem{}
		for _, w := range wb.RcMngr.Wallets {
			if strings.HasPrefix(w.Addr, wid) {
				wlist = append(wlist, walletToSelectListItems(w))
			}
		}
	}
	slices.SortFunc(wlist, func(a, b views.SelectListItem) int {
		if a.Name > b.Name {
			return 1
		} else if a.Name < b.Name {
			return -1
		}
		return 0
	})
	return renderTempl(ctx, views.WalletSelectList(wlist))
}

func (wb *EmulatorWeb) HandleEmulationSettings(ctx echo.Context) error {
	s := views.EmulationSettingsItem{
		NodeNum:       strconv.Itoa(len(wb.RcMngr.Nodes)),
		WalletNum:     strconv.Itoa(len(wb.RcMngr.Wallets)),
		CoinbaseStart: strconv.Itoa(blockchain.COINBASE_START_AMOUNT),
		RewardAmount:  strconv.Itoa(blockchain.REWARD_AMOUNT),
		Diff:          blockchain.MINE_DIFF,
		OpPause:       strconv.Itoa(int(OP_PAUSE_MILISEC.Milliseconds())),
	}
	return renderTempl(ctx, views.EmulationSettings(s))
}

func (wb *EmulatorWeb) HandleWalletUtxoTable(ctx echo.Context) error {
	wid := ctx.FormValue("WalletList")
	if wid == "" {
		return nil
	}
	w, ok := wb.RcMngr.Wallets[wid]
	if !ok {
		return nil
	}
	wSpent := wb.RcMngr.TrGen.GetWallentSpentUtxo(w.Addr)
	ul := make([]views.UtxoItem, len(w.Utxo))
	i := 0
	for id, u := range w.Utxo {
		ul[i] = views.UtxoItem{
			Id:     id,
			Amount: strconv.Itoa(u.Amount),
			Addr:   u.Addr,
		}
		if wSpent != nil {
			ul[i].Spent = wSpent.Check(id)
		}
		i++
	}

	return renderTempl(ctx, views.WalletUtxoTable(ul))
}

func (wb *EmulatorWeb) SelectMainNode() {
	wb.RcMngr.SelectMainNode()
	wb.RssSendMinerSelect()
}

// Evil Handlers

func (wb *EmulatorWeb) HandleEvilLoad(ctx echo.Context) error {
	if wb.RcMngr.EvilBlock == nil {
		return renderTempl(ctx, views.ItemNotFound("Block", "No current evil block is set. Steal the block first"))
	}
	bi := blockToItem(wb.RcMngr.EvilBlock)
	tri := blockTransitionToItems(wb.RcMngr.EvilBlock)
	return renderTempl(ctx, views.EvilBlock(bi, tri))
}

func (wb *EmulatorWeb) HandleEvilSteal(ctx echo.Context) error {
	wb.RssLogEvilSend("Stealing block candidate")
	n := wb.RcMngr.MainNode()
	if n == nil {
		return renderTempl(ctx, views.ItemNotFound("Main Node", "main node not set"))
	}
	b := n.BlockCandidate
	if b == nil {
		return renderTempl(ctx, views.ItemNotFound("Block", "block candidate not set"))
	}
	wb.RcMngr.EvilBlock = b.Clone()
	bItem := blockToItem(wb.RcMngr.EvilBlock)
	trItems := blockTransitionToItems(b)
	return renderTempl(ctx, views.EvilBlock(bItem, trItems))
}

func (wb *EmulatorWeb) HandleEvilCreate(ctx echo.Context) error {
	wb.RssLogEvilSend("Creating evilt block")
	wb.RcMngr.EvilBlock = blockchain.NewBlock()
	bItem := blockToItem(wb.RcMngr.EvilBlock)
	trItems := blockTransitionToItems(wb.RcMngr.EvilBlock)
	return renderTempl(ctx, views.EvilBlock(bItem, trItems))
}

func (wb *EmulatorWeb) HandleEvilSetHeihgt(ctx echo.Context) error {
	wb.RssLogEvilSend("Setting new height")

	if wb.RcMngr.EvilBlock == nil {
		return wb.evilBlockSetFail(ctx, "Evil: evil block not set. Steal new block.")
	}

	fh := ctx.FormValue("height")
	if fh == "" {
		return wb.evilBlockSetFail(ctx, "Evil: new height not provided")
	}
	h, err := strconv.Atoi(fh)
	if err != nil {
		return wb.evilBlockSetFail(ctx, "Evil: new height is not integer")
	}

	wb.RcMngr.EvilBlock.Header.Height = h

	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) HandleEvilSetTime(ctx echo.Context) error {
	wb.RssLogEvilSend("Setting new time")

	if wb.RcMngr.EvilBlock == nil {
		return wb.evilBlockSetFail(ctx, "Evil: no evil block set")
	}
	d := ctx.FormValue("date")
	t := ctx.FormValue("time")

	if d == "" || t == "" {
		return wb.evilBlockSetFail(ctx, "Date or time is empty")
	}

	datetime, err := time.Parse("2006-01-02 15:04:05", d+" "+t)

	if err != nil {
		return wb.evilBlockSetFail(ctx, "Evil: wrong date time format. Must be YYYY-mm-dd hh:mm:ss")
	}

	wb.RcMngr.EvilBlock.Header.Time = datetime

	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) HandleEvilSetHash(ctx echo.Context) error {
	wb.RssLogEvilSend("Setting header value...")
	if wb.RcMngr.EvilBlock == nil {
		return wb.evilBlockSetFail(ctx, "Evil: no evil block set")
	}

	hf := ctx.FormValue("hash")

	var h []byte
	var err error

	if hf == "" {
		h = []byte{}
	} else if h, err = blockchain.StringToBytes(hf); err != nil {
		return wb.evilBlockSetFail(ctx, "Hash value is not hex string")
	}

	t := ctx.FormValue("type")
	switch t {
	case "root":
		wb.RssLogEvilSend("Setting root")
		wb.RcMngr.EvilBlock.Header.Root = h
	case "prev":
		wb.RssLogEvilSend("Setting prev")
		wb.RcMngr.EvilBlock.Header.Prev = h
	case "hash":
		wb.RssLogEvilSend("Setting hash")
		wb.RcMngr.EvilBlock.Header.Hash = h
	default:
		return wb.evilBlockSetFail(ctx, "Evil: Unknown hash type %s", t)
	}
	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) HandleEvilSetNonce(ctx echo.Context) error {
	wb.RssLogEvilSend("Setting nonce")
	if wb.RcMngr.EvilBlock == nil {
		return wb.evilBlockSetFail(ctx, "Evil: no evil block set")
	}
	nf := ctx.FormValue("nonce")
	if nf == "" {
		return wb.evilBlockSetFail(ctx, "Evil: nonce is empty")
	}
	if n, err := strconv.Atoi(nf); err == nil {
		wb.RcMngr.EvilBlock.Header.Nonce = n
		return renderTempl(ctx, views.EvilActionResult(true))
	}
	return wb.evilBlockSetFail(ctx, "Evil: nonce value not integer")
}

func (wb *EmulatorWeb) HandleEvilSetInt(ctx echo.Context) error {
	if wb.RcMngr.EvilBlock == nil {
		return wb.evilBlockSetFail(ctx, "Evil: no evil block set")
	}
	vf := ctx.FormValue("value")
	if vf == "" {
		return wb.evilBlockSetFail(ctx, "Evil: value is empty")
	}

	v, err := strconv.Atoi(vf)
	if err != nil {
		return wb.evilBlockSetFail(ctx, "Evil: nonce value not integer")

	}

	t := ctx.FormValue("field")
	switch t {
	case "nonce":
		wb.RssLogEvilSend("Setting nonce")
		wb.RcMngr.EvilBlock.Header.Nonce = v
	case "coinbase":
		wb.RssLogEvilSend("Setting coinbase")
		wb.RcMngr.EvilBlock.Body.Coinbase = v
	default:
		return wb.evilBlockSetFail(ctx, "Evil: unknown field %s", t)
	}
	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) HandleEvilSetTrHashValue(ctx echo.Context) error {
	fTid := ctx.FormValue("tid")

	t, err := wb.evilGetTransaction(fTid)
	if err != nil {
		return wb.evilBlockSetFail(ctx, err.Error())
	}

	hType := ctx.FormValue("type")
	fVal := ctx.FormValue("value")

	var v []byte

	if fVal == "" {
		v = []byte{}
	} else if v, err = blockchain.StringToBytes(fVal); err != nil {
		return wb.evilBlockSetFail(ctx, "value is not hex string")
	}

	switch hType {
	case "sign":
		wb.RssLogEvilSend("Setting Transaction [%s] sign", fTid)
		t.Sign = v
	case "pk":
		wb.RssLogEvilSend("Setting Transaction [%s] pk", fTid)
		t.Pk = v
	default:
		return wb.evilBlockSetFail(ctx, "Uknown transaction field %s", hType)
	}

	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) HandleEvilSetTrUtxo(ctx echo.Context) error {
	fTid := ctx.FormValue("tid")
	t, err := wb.evilGetTransaction(fTid)
	if err != nil {
		return wb.evilBlockSetFail(ctx, err.Error())
	}

	wb.RssLogEvilSend("Setting Transaction [%s] utxo", fTid)

	fAmount := ctx.FormValue("amount")
	amount, err := strconv.Atoi(fAmount)

	if err != nil {
		return wb.evilBlockSetFail(ctx, "Amount is not integer")
	}

	addr := ctx.FormValue("addr")

	utype := ctx.FormValue("type")
	uid := ctx.FormValue("uid")

	if uid == "" {
		return wb.evilBlockSetFail(ctx, "Utxo id is empty")
	}

	switch utype {
	case "input":
		err := t.UpdateInputUtxo(uid, amount, addr)
		if err != nil {
			return wb.evilBlockSetFail(ctx, err.Error())
		}
	case "output":
		err := t.UpdateOutputUtxo(uid, amount, addr)
		if err != nil {
			return wb.evilBlockSetFail(ctx, err.Error())
		}
	default:
		return wb.evilBlockSetFail(ctx, "Unknown utxo type: %s", utype)
	}
	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) HandleEvilDelUtxo(ctx echo.Context) error {
	wb.RssLogEvilSend("Deleting utxo")
	fTid := ctx.FormValue("tid")
	t, err := wb.evilGetTransaction(fTid)
	if err != nil {
		wb.RssLogErrorSend("Evil: %s", err)
		return ctx.String(400, err.Error())
	}
	uid := ctx.FormValue("uid")
	utype := ctx.FormValue("type")

	switch utype {
	case "input":
		if err = t.DeleteInputUtxo(uid); err != nil {
			wb.RssLogErrorSend("Evil: %s", err)
			return ctx.String(400, err.Error())
		}
	case "output":
		if err = t.DeleteOutputUtxo(uid); err != nil {
			wb.RssLogErrorSend("Evil: %s", err)
			return ctx.String(400, err.Error())
		}
	default:
		wb.RssLogErrorSend("Evil: uknown utxo type: %s", utype)
		return ctx.String(400, "")
	}

	return nil
}

func (wb *EmulatorWeb) HandleEvilDelTr(ctx echo.Context) error {
	wb.RssLogEvilSend("Removing transaction")
	fTid := ctx.FormValue("tid")
	tid, err := strconv.Atoi(fTid)
	if err != nil {
		wb.RssLogErrorSend("Evil: transaction id is empty")
		return ctx.String(400, "tid is empty")
	}
	if tid < 0 || tid >= len(wb.RcMngr.EvilBlock.Body.Transactions) {
		wb.RssLogErrorSend("Evil: transaction id is not integer")
		return ctx.String(400, "tid is empty")
	}
	wb.RcMngr.EvilBlock.Body.Transactions[tid] = blockchain.Transaction{}
	return nil
}

func (wb *EmulatorWeb) HandleEvilAddUtxo(ctx echo.Context) error {
	wb.RssLogEvilSend("Adding Utxo")

	tid := ctx.FormValue("tid")
	t, err := wb.evilGetTransaction(tid)
	if err != nil {
		wb.RssLogErrorSend("Evil: %s", err)
		return ctx.String(400, "invalid transaction id")
	}

	fAmount := ctx.FormValue("amount")

	amount, err := strconv.Atoi(fAmount)
	if err != nil {
		wb.RssLogErrorSend("Evil: amount is not integer")
		return ctx.String(400, "Amount is not integer")
	}

	utype := ctx.FormValue("type")
	addr := ctx.FormValue("addr")

	uid := ""

	switch utype {
	case "input":
		uid = t.InputUtxo.NewRecord(addr, amount)
	case "output":
		uid = t.OutputUtxo.NewRecord(addr, amount)
	default:
		wb.RssLogErrorSend("Evil: unknown utxo type: %s", utype)
		return ctx.String(400, "unknown utxo type")
	}

	wb.RssLogEvilSend("Added new %s utxo", utype)
	uItem := views.UtxoItem{
		Id:     uid,
		Amount: fAmount,
		Addr:   addr,
	}

	return renderTempl(ctx, views.EvilTrNewUtxo(tid, utype, uItem))
}

func (wb *EmulatorWeb) HandleEvilAddTr(ctx echo.Context) error {
	wb.RssLogEvilSend("Adding new transaction")
	if wb.RcMngr.EvilBlock == nil {
		wb.RssLogErrorSend("Evil: no evil block. Steal the block first.")
		return ctx.String(400, "No evil block")
	}
	t := blockchain.NewTransaction()
	t.Sign = []byte{}
	t.Pk = []byte{}
	tid := wb.RcMngr.EvilBlock.AddTransaction(*t)

	tItem := views.BlockTransactionItem{
		Id:         strconv.Itoa(tid),
		Sign:       "",
		Pk:         "",
		InputUtxo:  []views.UtxoItem{},
		OutputUtxo: []views.UtxoItem{},
	}
	return renderTempl(ctx, views.EvilNewTr(tItem))
}

func (wb *EmulatorWeb) HandleEvilMine(ctx echo.Context) error {
	wb.RssLogEvilSend("Start mining evil block")
	if wb.RcMngr.EvilBlock == nil {
		wb.RssLogErrorSend("Evil: evil block not set. Steal block first.")
		return ctx.String(400, "Evil: evil block not set. Steal block first.")
	}
	n := wb.RcMngr.MainNode()
	if n == nil {
		wb.RssLogErrorSend("Evil: no main node set")
		return ctx.String(400, "Evil: no main node set")
	}
	n.BlockCandidate = wb.RcMngr.EvilBlock
	t := time.Now()
	b, err := n.MineUnsafe()
	if err != nil {
		wb.RssLogErrorSend("Evil: Failed to mine block")
		return ctx.String(400, "Evil: Failed to mine block")
	}
	wb.RssLogEvilSend("Mined with node %s in %.2f sec", n.Name, time.Since(t).Seconds())
	wb.RssNodeAllUpdates(n.Id)
	wb.RcMngr.EvilBlock = b.Clone()
	return wb.HandleEvilLoad(ctx)
}

func (wb *EmulatorWeb) HandleEvilInject(ctx echo.Context) error {
	wb.RssLogEvilSend("Injecting evil block")
	if wb.RcMngr.EvilBlock == nil {
		return wb.evilBlockSetFail(ctx, "Evil: evil block not set. Steal block first.")
	}
	n := wb.RcMngr.MainNode()
	if n == nil {
		return wb.evilBlockSetFail(ctx, "Evil: no main node set")
	}
	b := wb.RcMngr.EvilBlock.Clone()
	n.AddBlock(b)
	wb.aggregatorAddBlock(b, n.Name, true)
	// n.NewBlockCandidate()
	wb.RssGrapthSetAttacked(n.Name)
	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) HandleEvilSend(ctx echo.Context) error {
	wb.RssLogEvilSend("Sending evil block")
	if wb.RcMngr.EvilBlock == nil {
		return wb.evilBlockSetFail(ctx, "Evil: evil block not set. Steal block first.")
	}
	wb.spreadBlock("", wb.RcMngr.EvilBlock)
	wb.RssAllNodesUpdates()
	return renderTempl(ctx, views.EvilActionResult(true))
}

func (wb *EmulatorWeb) evilGetTransaction(tid string) (*blockchain.Transaction, error) {
	if wb.RcMngr.EvilBlock == nil {
		return nil, fmt.Errorf("Evil: evil block not set")
	}

	id, err := strconv.Atoi(tid)
	if err != nil {
		return nil, fmt.Errorf("Evil: transaction id is not integer")
	}

	if id < 0 || id >= len(wb.RcMngr.EvilBlock.Body.Transactions) {
		return nil, fmt.Errorf("Evil: transaction id is not found")
	}

	return &wb.RcMngr.EvilBlock.Body.Transactions[id], nil
}

func (wb *EmulatorWeb) evilBlockSetFail(ctx echo.Context, msg string, a ...any) error {
	wb.RssLogErrorSend(fmt.Sprintf(msg, a...))
	return renderTempl(ctx, views.EvilActionResult(false))
}

// END Evil handlers

// Visual handlers

func (wb *EmulatorWeb) HandleGetNetMap(ctx echo.Context) error {
	bn := BlockchainNet{
		Nodes: make([]BlockchainNetNode, 0, len(wb.RcMngr.Nodes)+len(wb.RcMngr.Wallets)),
		// Links: make([]BlockChainNetLink, 0, len(wb.RcMngr.Nodes)*2+len(wb.RcMngr.Wallets)),
	}
	n := wb.RcMngr.MainNode()
	for id, nd := range wb.RcMngr.Nodes {
		node := BlockchainNetNode{
			Id: nd.Name,
		}
		if n != nil && id == n.Id {
			node.Group = V_NODE_GROUP_MINER
		} else {
			node.Group = V_NODE_GROUP_NODE
		}
		node.Attacked = wb.DataAgr.NodesAttacks.IsAttacked(nd.Name)
		bn.Nodes = append(bn.Nodes, node)
	}

	nodeLinks := wb.RcMngr.NetLinkListWithNames()
	bn.Links = make([]BlockChainNetLink, 0, len(nodeLinks)+len(wb.RcMngr.Wallets))
	for _, link := range nodeLinks {
		bn.Links = append(bn.Links, BlockChainNetLink{Source: link[0], Target: link[1], Latency: strconv.Itoa(rand.Intn(200) + 20)})
	}

	for _, wl := range wb.RcMngr.Wallets {
		if strings.Contains(wl.Name, "Node") {
			continue
		}
		bn.Nodes = append(bn.Nodes, BlockchainNetNode{Id: wl.Name, Group: V_NODE_GROUP_USER})
		trgt := ""
		for _, nd := range wb.RcMngr.Nodes {
			trgt = nd.Name
			break
		}
		bn.Links = append(bn.Links, BlockChainNetLink{Source: wl.Name, Target: trgt, Latency: strconv.Itoa(rand.Intn(100) + 5)})
	}
	return ctx.JSON(200, bn)
}

func (wb *EmulatorWeb) HandleGetChainFork(ctx echo.Context) error {
	blocks := wb.DataAgr.GetForkBlocks()
	return ctx.JSON(200, blocks)
}

// Helper functions

type blockRecieveResult struct {
	node    string
	success bool
	msg     string
}

func (wb *EmulatorWeb) spreadBlock(senderId string, b *blockchain.Block) {
	spreadNum := len(wb.RcMngr.Nodes) - 1
	statusChan := make(chan blockRecieveResult, spreadNum)
	chanLen := 0
	defer close(statusChan)
	var wg sync.WaitGroup
	recievedList := make(map[string]struct{})
	recievedList[senderId] = struct{}{}
	var sendFunc func(sender string)
	sendFunc = func(sender string) {
		toList, ok := wb.RcMngr.NetMap[sender]
		if !ok {
			return
		}
		seenCount := 0
		for toId := range toList {
			if _, ok := recievedList[toId]; ok {
				seenCount += 1
				continue
			}
			chanLen += 1
			recievedList[toId] = struct{}{}
			nd := wb.RcMngr.Nodes[toId]
			bCopy := b.Clone()
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := nd.AddVerifyBlock(bCopy)
				if err != nil {
					statusChan <- blockRecieveResult{node: nd.Name, success: false, msg: err.Error()}
					return
				}
				statusChan <- blockRecieveResult{node: nd.Name, success: true}
			}()
		}
		if seenCount >= len(toList) {
			return
		}
		for toId := range toList {
			sendFunc(toId)
		}
	}
	timeStart := time.Now()
	sendFunc(senderId)
	wg.Wait()
	timePassed := time.Since(timeStart).Seconds()
	wb.RssLogInfoSend("Block spread in %.4f seconds", timePassed)

	for i := 0; i < chanLen; i++ {
		s := <-statusChan
		if s.success {
			wb.RssLogOKSend("Spreading block: Node [%s]: block accepted", s.node)
		} else {
			wb.RssLogErrorSend("Spreading block: Node [%s]: %s", s.node, s.msg)
		}
	}
	wb.RssAllNodesUpdates()
}

func blockToItem(b *blockchain.Block) views.BlockInfoSmallItem {
	bi := views.BlockInfoSmallItem{
		Height:   strconv.Itoa(b.Header.Height),
		Coinbase: strconv.Itoa(b.Body.Coinbase),
		Nonce:    strconv.Itoa(b.Header.Nonce),
		Hash:     b.HashString(),
		Prev:     b.PrevString(),
		Root:     b.RootString(),
		Time:     b.Header.Time.Format("2006-01-02 15:04:05"),
		TotalTr:  strconv.Itoa(len(b.Body.Transactions)),
	}
	return bi
}

func walletToSelectListItems(w *blockchain.Wallet) views.SelectListItem {
	return views.SelectListItem{
		Id:   w.Addr,
		Name: w.Name,
	}
}

func blockTransitionToItems(b *blockchain.Block) []views.BlockTransactionItem {
	trItems := make([]views.BlockTransactionItem, 0, len(b.Body.Transactions))
	for i, tr := range b.Body.Transactions {
		if tr.Sign == nil || tr.Pk == nil {
			continue
		}
		t := views.BlockTransactionItem{
			Id:         strconv.Itoa(i),
			Sign:       blockchain.BytesToString(tr.Sign),
			Pk:         blockchain.BytesToString(tr.Pk),
			InputUtxo:  make([]views.UtxoItem, len(tr.InputUtxo)),
			OutputUtxo: make([]views.UtxoItem, len(tr.OutputUtxo)),
		}
		j := 0
		for id, u := range tr.InputUtxo.SortedItems() {
			t.InputUtxo[j] = views.UtxoItem{
				Id:     id,
				Amount: strconv.Itoa(u.Amount),
				Addr:   u.Addr,
			}
			j++
		}
		j = 0
		for id, u := range tr.OutputUtxo.SortedItems() {
			t.OutputUtxo[j] = views.UtxoItem{
				Id:     id,
				Amount: strconv.Itoa(u.Amount),
				Addr:   u.Addr,
			}
			j++
		}
		trItems = append(trItems, t)
	}
	return trItems
}

func (wb *EmulatorWeb) aggregatorAddBlock(b *blockchain.Block, nodeName string, evil bool) {
	if cf, err := wb.DataAgr.ChainFork.AddBlock(b, nodeName, wb.RcMngr.Tick, evil); err == nil {
		if newBlockData, err := json.Marshal(cf.Block); err == nil {
			wb.RssForkNewBlock(newBlockData)
		}
	}
	if evil {
		wb.DataAgr.NodesAttacks.AddAttack(nodeName, wb.RcMngr.Tick)
	}
}

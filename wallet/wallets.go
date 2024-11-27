package wallet

import (
	"encoding/json"
	"log"
	"os"
)

const walletFile = "./tmp/wallet.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFile()
	return &wallets, err
}

func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := string(wallet.Address())
	ws.Wallets[address] = wallet
	return address
}

func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, ws)
	if err != nil {
		log.Panic(err)
	}

	return nil
}

func (ws *Wallets) SaveFile() {
	data, err := json.Marshal(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, data, 0644)
	if err != nil {
		log.Panic(err)
	}
}

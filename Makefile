GO=go
GOFLAGS=-v
MAIN_FILE=main.go

ADDRESS?=
FROM?=
TO?=
AMOUNT?=

.PHONY: all run clean run getbalance createblockchain printchain send

all: run

run:
	$(GO) run $(MAIN_FILE)

getbalance:
ifeq ($(ADDRESS),)
	@echo "usage: make getbalance ADDRESS=<wallet-address>"
	@exit 1
else
	$(GO) run $(MAIN_FILE) getbalance -address $(ADDRESS)
endif

createblockchain:
ifeq ($(ADDRESS),)
	@echo "usage: make createblockchain ADDRESS=<wallet-address>"
	@exit 1
else
	$(GO) run $(MAIN_FILE) createblockchain -address $(ADDRESS)
endif

printchain:
	$(GO) run $(MAIN_FILE) printchain

send:
ifeq ($(FROM),)
	@echo "usage: make send FROM=<from-address> TO=<to-address> AMOUNT=<amount>"
	@exit 1
endif
ifeq ($(TO),)
	@echo "usage: make send FROM=<from-address> TO=<to-address> AMOUNT=<amount>"
	@exit 1
endif
ifeq ($(AMOUNT),)
	@echo "usage: make send FROM=<from-address> TO=<to-address> AMOUNT=<amount>"
	@exit 1
endif
	$(GO) run $(MAIN_FILE) send -from $(FROM) -to $(TO) -amount $(AMOUNT)

createwallet:
	$(GO) run $(MAIN_FILE) createwallet

listaddresses:
	$(GO) run $(MAIN_FILE) listaddresses

help:
	@echo "使用可能なコマンド:"
	@echo "  make getbalance ADDRESS=<address>  - アドレスの残高を確認"
	@echo "  make createblockchain ADDRESS=<address> - ブロックチェーンを作成"
	@echo "  make printchain - ブロックチェーンの内容を表示"
	@echo "  make send FROM=<from> TO=<to> AMOUNT=<amount> - コインを送金"

clean:
	rm -rf ./tmp/blocks
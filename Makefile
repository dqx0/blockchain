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
    @if [ -z "$(ADDRESS)" ]; then \
        echo "使用法: make getbalance ADDRESS=<wallet-address>"; \
        exit 1; \
    fi
    $(GO) run $(MAIN_FILE) getbalance -address $(ADDRESS)

createblockchain:
    @if [ -z "$(ADDRESS)" ]; then \
        echo "使用法: make createblockchain ADDRESS=<wallet-address>"; \
        exit 1; \
    fi
    $(GO) run $(MAIN_FILE) createblockchain -address $(ADDRESS)

printchain:
    $(GO) run $(MAIN_FILE) printchain

send:
    @if [ -z "$(FROM)" ] || [ -z "$(TO)" ] || [ -z "$(AMOUNT)" ]; then \
        echo "使用法: make send FROM=<from-address> TO=<to-address> AMOUNT=<amount>"; \
        exit 1; \
    fi
    $(GO) run $(MAIN_FILE) send -from $(FROM) -to $(TO) -amount $(AMOUNT)

help:
    @echo "使用可能なコマンド:"
    @echo "  make getbalance ADDRESS=<address>  - アドレスの残高を確認"
    @echo "  make createblockchain ADDRESS=<address> - ブロックチェーンを作成"
    @echo "  make printchain - ブロックチェーンの内容を表示"
    @echo "  make send FROM=<from> TO=<to> AMOUNT=<amount> - コインを送金"

clean:
    rm -rf ./tmp/blocks
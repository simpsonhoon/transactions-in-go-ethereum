package main

import (
	"context"
	conf "go-daemon/config"
	"go-daemon/model"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// config 초기화
	cf := conf.GetConfig("./config/config.toml")

	// model 초기화
	md, err := model.NewModel(cf.DB.Host)
	if err != nil {
		log.Fatal(err)
	}

	// ethclint 초기화
	client, err := ethclient.Dial(cf.Network.URL)
	if err != nil {
		log.Fatal(err)
	}

	// subscribe, 블록 감지
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			block, err := client.BlockByNumber(context.Background(), header.Number)
			if err != nil {
				log.Fatal(err)
			}

			// TODO: 블록 구조체 생성
			b := model.Block{
				BlockHash:    block.Hash().Hex(),
				BlockNumber:  block.Number().Uint64(),
				GasLimit:     block.GasLimit(),
				GasUsed:      block.GasUsed(),
				Time:         block.Time(),
				Nonce:        block.Nonce(),
				Transactions: make([]model.Transaction, 0),
			}

			// TODO: 트랜잭션 추출
			txs := block.Transactions()
			if len(txs) > 0 {
				for _, tx := range txs {
					msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), block.BaseFee())
					if err != nil {
						log.Fatal(err)
					}

					myAddress := "0x50f2Ca639b8F2819F977b73807E0e52e05e6bd70"
					SSHTokenAddress := "0xb58E525a38bb9Dc9Fe4fb3C2b957f7A9863093bF"
					//특정 주소에서 발생한 트랜잭션만 추출
					if msg.From().Hex() == myAddress || msg.To().Hex() == SSHTokenAddress {

						// TODO: 트랜잭션 구조체 생성
						t := model.Transaction{
							TxHash:      tx.Hash().Hex(),
							To:          "", // 디폴트 값 처리
							From:        msg.From().Hex(),
							Nonce:       tx.Nonce(),
							GasPrice:    tx.GasPrice().Uint64(),
							GasLimit:    tx.Gas(),
							Amount:      tx.Value().Uint64(),
							BlockHash:   block.Hash().Hex(),
							BlockNumber: block.Number().Uint64(),
						}

						if tx.To() != nil {
							t.To = tx.To().Hex()
						}
						b.Transactions = append(b.Transactions, t)
					}
				}
			}
			//트랜잭션이 있을때만 DB에 저장하자
			if len(b.Transactions) > 0 {
				err = md.SaveBlock(&b)
				if err != nil {
					log.Fatal(err)
				}
			}

		}
	}
}

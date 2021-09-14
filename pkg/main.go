package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethsana/sana/pkg/logging"
	"github.com/ethsana/sana/pkg/node"
	"github.com/ethsana/sana/pkg/storage"
)

const (
	chequebookKey           = "swap_chequebook"
	chequebookDeploymentKey = "swap_chequebook_transaction_deployment"
	deployedTopic           = `0xc0ffc525a1c7689549d7f79b49eca900e61ac49b43d977f680bcc3b36224c004`
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf(`recober-chequbook <data-dir> <txhash> <xdai rpc>`)
		return
	}

	logger := logging.New(ioutil.Discard, 0)

	stateStore, err := node.InitStateStore(logger, os.Args[1])
	if err != nil {
		fmt.Printf(`init statestore fail: %v`, err)
		return
	}
	defer stateStore.Close()

	var chequebook common.Address
	err = stateStore.Get(chequebookKey, &chequebook)
	if err != nil && err != storage.ErrNotFound {
		fmt.Printf(`get chequebook fail: %v`, err)
		return
	}

	if err == storage.ErrNotFound {
		client, err := ethclient.Dial(os.Args[3])
		if err != nil {
			fmt.Printf(`ethclient dail fail: %v`, err)
			return
		}

		receipt, err := client.TransactionReceipt(context.TODO(), common.HexToHash(os.Args[2]))
		if err != nil {
			fmt.Printf(`get transaction receipt fail: %v`, err)
		}

		for _, l := range receipt.Logs {
			if l.Topics[0].Hex() == deployedTopic {
				chequebook = common.BytesToAddress(l.Data)
				break
			}
		}

		if (chequebook == common.Address{}) {
			fmt.Printf(`not found chequebook with transaction %v`, os.Args[2])
			return
		}

		err = stateStore.Put(chequebookKey, chequebook)
		if err != nil {
			fmt.Printf(`put chequebook fail: %v`, err)
			return
		}

		err = stateStore.Put(chequebookDeploymentKey, common.HexToHash(os.Args[2]))
		if err != nil {
			fmt.Printf(`put chequebook deploy transaction fail: %v`, err)
			return
		}
	}

	fmt.Printf("chequebook: %v\n", chequebook.Hex())
}

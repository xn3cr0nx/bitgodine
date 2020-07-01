package bitcoin

import (
	"io/ioutil"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// ClientConfig returns bitcoin client config object
func ClientConfig() *rpcclient.ConnConfig {
	certs, err := ioutil.ReadFile(viper.GetString("bitcoin.client.certs"))
	if err != nil {
		logger.Error("Bitcoin client", err, logger.Params{})
		certs = nil
	}

	return &rpcclient.ConnConfig{
		Host:         viper.GetString("bitcoin.client.host"),
		Endpoint:     viper.GetString("bitcoin.client.endpoint"),
		User:         viper.GetString("bitcoin.client.user"),
		Pass:         viper.GetString("bitcoin.client.pass"),
		Certificates: certs,
	}
}

func NewClient() (*rpcclient.Client, error) {
	ntfnHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txns []*btcutil.Tx) {
			log.Printf("Block connected: %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
			log.Printf("Block disconnected: %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
		},
	}
	client, err := rpcclient.New(ClientConfig(), &ntfnHandlers)
	if err != nil {
		return nil, err
	}
	return client, nil
}

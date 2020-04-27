package watcher

import (
	"context"
	"net/url"

	types "github.com/mapofzones/cosmos-watcher/types"
	rabbitmq "github.com/mapofzones/cosmos-watcher/x/tendermint-rabbit/RabbitMQ"
	"github.com/mapofzones/cosmos-watcher/x/tendermint-rabbit/block"
)

// Watcher implements the watcher interface described at the project root
// this particular implementation is used to listen on tendermint websocket and
// send results to RabbitMQ
type Watcher struct {
	tendermintAddr url.URL
	rabbitMQAddr   url.URL

	chainID string
}

// NewWatcher returns instanciated Watcher
func NewWatcher(tendermintRPCAddr, rabbitmqAddr string) (*Watcher, error) {
	//we checked if urls are valid in GetConfig already
	nodeURL, _ := url.Parse(tendermintRPCAddr)
	rabbitURL, _ := url.Parse(rabbitmqAddr)

	//figure out name of the blockchain to which we are connecting
	name, err := getNetworkName(*nodeURL)
	if err != nil {
		return nil, err
	}

	return &Watcher{tendermintAddr: *nodeURL, rabbitMQAddr: *rabbitURL, chainID: name}, nil
}

// serve buffers and sends txs to rabbitmq, returns errors if something is wrong
// Accepts input and output channels, also an error channel if something goes wrong
func (w *Watcher) serve(ctx context.Context, blockInput <-chan block.WithTxs, blockOutput chan<- types.Block, errors <-chan error) error {
	for {
		select {
		case block, ok := <-blockInput:
			if ok {
				select {
				case blockOutput <- types.Normalize(block):
				case err := <-errors:
					return err
				}
			}

		case err := <-errors:
			return err

		case <-ctx.Done():
			return nil
		}
	}
}

// Watch implements watcher interface
// Collects txs from tendermint websocket and sends them to rabbitMQ
func (w *Watcher) Watch(ctx context.Context) error {
	queue, errors, err := rabbitmq.BlockQueue(ctx, w.rabbitMQAddr)
	if err != nil {
		return err
	}

	blocks, err := block.GetBlockStream(ctx, w.tendermintAddr.String())
	if err != nil {
		return err
	}

	return w.serve(ctx, blocks, queue, errors)
}

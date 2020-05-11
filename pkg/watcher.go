package watcher

import (
	"context"
	"net/url"
	"strconv"

	rabbitmq "github.com/mapofzones/cosmos-watcher/pkg/RabbitMQ"
	"github.com/mapofzones/cosmos-watcher/pkg/block"
	types "github.com/mapofzones/cosmos-watcher/types"
	"github.com/tendermint/tendermint/rpc/client/http"
)

// Watcher implements the watcher interface described at the project root
// this particular implementation is used to listen on tendermint websocket and
// send results to RabbitMQ
type Watcher struct {
	tendermintAddr url.URL
	rabbitMQAddr   url.URL
	client         *http.HTTP
	chainID        string
	startingHeight int64
}

// NewWatcher returns instanciated Watcher
func NewWatcher(tendermintRPCAddr, rabbitmqAddr string, startingHeight string) (*Watcher, error) {
	//we checked if urls are valid in GetConfig already
	nodeURL, err := url.Parse(tendermintRPCAddr)
	if err != nil {
		return nil, err
	}
	rabbitURL, err := url.Parse(rabbitmqAddr)
	if err != nil {
		return nil, err
	}

	height, err := strconv.ParseInt(startingHeight, 10, 64)
	if err != nil {
		return nil, err
	}

	client, err := http.New("tcp"+"://"+nodeURL.Host, "/websocket")
	if err != nil {
		return nil, err
	}
	err = client.Start()
	if err != nil {
		return nil, err
	}
	//figure out name of the blockchain to which we are connecting
	info, err := client.Status()
	if err != nil {
		return nil, err
	}

	return &Watcher{tendermintAddr: *nodeURL, rabbitMQAddr: *rabbitURL, chainID: info.NodeInfo.Network, client: client, startingHeight: height}, nil
}

// serve buffers and sends txs to rabbitmq, returns errors if something is wrong
// Accepts input and output channels, also an error channel if something goes wrong
func (w *Watcher) serve(ctx context.Context, blockInput <-chan types.Block, blockOutput chan<- types.Block, errors <-chan error) error {
	for {
		select {
		case block, ok := <-blockInput:
			if ok {
				select {
				case blockOutput <- block:
				case err := <-errors:
					return err
				}
			}

		case err := <-errors:
			w.client.Stop()
			w.client.Wait()
			return err

		case <-ctx.Done():
			err := w.client.Stop()
			w.client.Wait()
			return err
		}
	}
}

// Watch implements watcher interface
// Collects txs from tendermint websocket and sends them to rabbitMQ
func (w *Watcher) Watch(ctx context.Context) error {
	queue, errors, err := rabbitmq.BlockQueue(ctx, w.rabbitMQAddr, w.chainID)
	if err != nil {
		return err
	}

	blocks := block.BlockStream(ctx, w.client, w.startingHeight)

	return w.serve(ctx, blocks, queue, errors)
}

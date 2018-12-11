package graphql

import (
	context "context"
)

func (r *subscriptionResolver) EthereumTransaction(ctx context.Context, txID string) (<-chan string, error) {

	receipt, err := r.ethService.TxListener.StartListener(txID)
	go func() {
		<-ctx.Done()
		// TODO(dankins): is there any cleanup necessary? What happens if the websocket closes?
		// probably should remove the observable
	}()

	return receipt.Updates, err

}

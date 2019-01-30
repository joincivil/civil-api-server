package airswap

import (
	"encoding/json"
	"math/big"
	"math/rand"
	"net/http"
	"time"

	log "github.com/golang/glog"
)

// Handlers implement http routes for the airswap order server
type Handlers struct {
	Pricing    *PricingManager
	Conversion PairPricing
}

// GetOrderRequest is an incoming POST request from the client server
type GetOrderRequest struct {
	MakerAddress string `json:"makerAddress"`
	TakerAddress string `json:"takerAddress"`
	MakerToken   string `json:"makerToken"`
	TakerToken   string `json:"takerToken"`
	TakerAmount  string `json:"takerAmount"`
	MakerAmount  string `json:"makerAmount"`
}

// GetOrderResponse is an outgoing order to be signed and sent by the client server.
type GetOrderResponse struct {
	MakerToken  string `json:"makerToken"`
	TakerToken  string `json:"takerToken"`
	MakerAmount string `json:"makerAmount"`
	TakerAmount string `json:"takerAmount"`
	Expiration  int64  `json:"expiration"`
	Nonce       uint32 `json:"nonce"`
}

// GetOrder is called when someone submits a request to BUY tokens
func (h *Handlers) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Errorf("bad request method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request GetOrderRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Errorf("bad request: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// MakerAmount is the number of tokens (in wei) that the person wants to receive in the trade (buy or sell)
	if request.MakerAmount != "" {
		// parse the amount of CVL they want to receive from the request
		requestedAmount, ok := new(big.Float).SetString(request.MakerAmount)
		if !ok {
			log.Errorf("bad makerAmount format: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// set how long the offer is valid for
		expiration := time.Now().Add(5 * time.Minute).Unix()
		// a random nonce to prevent replay attacks
		nonce := uint32(rand.Intn(99999))

		// requests are in wei, not the full unit
		wei, ok := new(big.Float).SetString("0.0000000000000000001")
		if !ok {
			log.Errorf("could not make wei: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var makerFloat float64
		makerFloat, _ = requestedAmount.Mul(requestedAmount, wei).Float64()
		priceInUSD := h.Pricing.GetQuote(makerFloat)
		price := priceInUSD * h.Conversion.USDToETH()

		// takerAmount is the # of tokens (in wei) the maker will receive for the order
		// this is just a fancy way of saying what the price is
		takerAmount := new(big.Float).SetFloat64(price)
		// price is in ETH, but we need it in wei
		takerAmount.Quo(takerAmount, wei)
		// return a big.Int instead of float
		takerInt, _ := takerAmount.Int(nil)

		order := GetOrderResponse{
			MakerToken:  request.MakerToken,
			TakerToken:  request.TakerToken,
			MakerAmount: request.MakerAmount,
			TakerAmount: takerInt.String(),
			Expiration:  expiration,
			Nonce:       nonce,
		}
		response, err := json.Marshal(order)
		if err != nil {
			log.Errorf("error encoding JSON response: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Infof("sending response to airswap request: %s", response)
		w.Header().Add("content-type", "application/json")
		_, err = w.Write(response)
		if err != nil {
			log.Errorf("error writing response: %s", err)
		}
	} else if request.TakerAmount != "" {
		log.Infof("not handling taker amount: %v", request.TakerAmount)
	} else {
		log.Errorf("bad request: no maker or taker amount supplied")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

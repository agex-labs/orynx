package currencypair

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	orynxtypes "github.com/agex-labs/orynx/pkg/types"
	oracletypes "github.com/agex-labs/orynx/x/oracle/types"
)

// OracleKeeper is an interface for interacting with the x/oracle state.
//
//go:generate mockery --name OracleKeeper --filename mock_oracle_keeper.go
type OracleKeeper interface {
	GetCurrencyPairFromID(ctx sdk.Context, id uint64) (cp orynxtypes.CurrencyPair, found bool)
	GetIDForCurrencyPair(ctx sdk.Context, cp orynxtypes.CurrencyPair) (uint64, bool)
	GetPriceForCurrencyPair(ctx sdk.Context, cp orynxtypes.CurrencyPair) (oracletypes.QuotePrice, error)
	GetNumCurrencyPairs(ctx sdk.Context) (uint64, error)
	GetNumRemovedCurrencyPairs(ctx sdk.Context) (uint64, error)
	GetAllCurrencyPairs(ctx sdk.Context) []orynxtypes.CurrencyPair
}

// CurrencyPairStrategy is a strategy for generating a unique ID and price representation for a given currency pair.
//
//go:generate mockery --name CurrencyPairStrategy --filename mock_currency_pair_strategy.go
type CurrencyPairStrategy interface { //nolint
	// ID returns the on-chain ID of the given currency pair. This method returns an error if the given currency
	// pair is not found in the x/oracle state.
	ID(ctx sdk.Context, cp orynxtypes.CurrencyPair) (uint64, error)

	// FromID returns the currency pair with the given ID. This method returns an error if the given ID is not
	// currently present for an existing currency pair.
	FromID(ctx sdk.Context, id uint64) (orynxtypes.CurrencyPair, error)

	// GetEncodedPrice returns the encoded price for the given currency pair. This method returns an error if the
	// given currency pair is not found in the x/oracle state or if the price cannot be encoded.
	GetEncodedPrice(
		ctx sdk.Context,
		cp orynxtypes.CurrencyPair,
		price *big.Int,
	) ([]byte, error)

	// GetDecodedPrice returns the decoded price for the given currency pair. This method returns an error if the
	// given currency pair is not found in the x/oracle state or if the price cannot be decoded.
	GetDecodedPrice(
		ctx sdk.Context,
		cp orynxtypes.CurrencyPair,
		priceBytes []byte,
	) (*big.Int, error)

	// GetMaxNumCP returns the number of pairs that the VEs should include.  This method returns an error if the size cannot
	// be queried from the x/oracle state.
	GetMaxNumCP(
		ctx sdk.Context,
	) (uint64, error)
}

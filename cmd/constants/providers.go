package constants

import (
	"github.com/agex-labs/orynx/oracle/config"
	"github.com/agex-labs/orynx/oracle/constants"
	"github.com/agex-labs/orynx/oracle/types"
	"github.com/agex-labs/orynx/providers/apis/agex"
	binanceapi "github.com/agex-labs/orynx/providers/apis/binance"
	bitstampapi "github.com/agex-labs/orynx/providers/apis/bitstamp"
	coinbaseapi "github.com/agex-labs/orynx/providers/apis/coinbase"
	"github.com/agex-labs/orynx/providers/apis/coingecko"
	"github.com/agex-labs/orynx/providers/apis/coinmarketcap"
	"github.com/agex-labs/orynx/providers/apis/defi/osmosis"
	"github.com/agex-labs/orynx/providers/apis/defi/raydium"
	"github.com/agex-labs/orynx/providers/apis/defi/uniswapv3"
	krakenapi "github.com/agex-labs/orynx/providers/apis/kraken"
	"github.com/agex-labs/orynx/providers/apis/marketmap"
	"github.com/agex-labs/orynx/providers/apis/polymarket"
	"github.com/agex-labs/orynx/providers/volatile"
	binancews "github.com/agex-labs/orynx/providers/websockets/binance"
	"github.com/agex-labs/orynx/providers/websockets/bitfinex"
	"github.com/agex-labs/orynx/providers/websockets/bitstamp"
	"github.com/agex-labs/orynx/providers/websockets/bybit"
	"github.com/agex-labs/orynx/providers/websockets/coinbase"
	"github.com/agex-labs/orynx/providers/websockets/cryptodotcom"
	"github.com/agex-labs/orynx/providers/websockets/gate"
	"github.com/agex-labs/orynx/providers/websockets/huobi"
	"github.com/agex-labs/orynx/providers/websockets/kraken"
	"github.com/agex-labs/orynx/providers/websockets/kucoin"
	"github.com/agex-labs/orynx/providers/websockets/mexc"
	"github.com/agex-labs/orynx/providers/websockets/okx"
	mmtypes "github.com/agex-labs/orynx/service/clients/marketmap/types"
)

var (
	Providers = []config.ProviderConfig{
		// DEFI providers
		{
			Name: raydium.Name,
			API:  raydium.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: uniswapv3.ProviderNames[constants.ETHEREUM],
			API:  uniswapv3.DefaultETHAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: uniswapv3.ProviderNames[constants.BASE],
			API:  uniswapv3.DefaultBaseAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: osmosis.Name,
			API:  osmosis.DefaultAPIConfig,
			Type: types.ConfigType,
		},

		// Exchange API providers
		{
			Name: binanceapi.Name,
			API:  binanceapi.DefaultNonUSAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: bitstampapi.Name,
			API:  bitstampapi.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: coinbaseapi.Name,
			API:  coinbaseapi.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: coingecko.Name,
			API:  coingecko.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: coinmarketcap.Name,
			API:  coinmarketcap.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: krakenapi.Name,
			API:  krakenapi.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: volatile.Name,
			API:  volatile.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		// Exchange WebSocket providers
		{
			Name:      binancews.Name,
			WebSocket: binancews.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      bitfinex.Name,
			WebSocket: bitfinex.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      bitstamp.Name,
			WebSocket: bitstamp.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      bybit.Name,
			WebSocket: bybit.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      coinbase.Name,
			WebSocket: coinbase.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      cryptodotcom.Name,
			WebSocket: cryptodotcom.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      gate.Name,
			WebSocket: gate.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      huobi.Name,
			WebSocket: huobi.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      kraken.Name,
			WebSocket: kraken.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      kucoin.Name,
			WebSocket: kucoin.DefaultWebSocketConfig,
			API:       kucoin.DefaultAPIConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      mexc.Name,
			WebSocket: mexc.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      okx.Name,
			WebSocket: okx.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},

		// Polymarket provider
		{
			Name: polymarket.Name,
			API:  polymarket.DefaultAPIConfig,
			Type: types.ConfigType,
		},

		// MarketMap provider
		{
			Name: marketmap.Name,
			API:  marketmap.DefaultAPIConfig,
			Type: mmtypes.ConfigType,
		},
	}

	AlternativeMarketMapProviders = []config.ProviderConfig{
		{
			Name: agex.Name,
			API:  agex.DefaultAPIConfig,
			Type: mmtypes.ConfigType,
		},
		{
			Name: agex.SwitchOverAPIHandlerName,
			API:  agex.DefaultSwitchOverAPIConfig,
			Type: mmtypes.ConfigType,
		},
		{
			Name: agex.ResearchAPIHandlerName,
			API:  agex.DefaultResearchAPIConfig,
			Type: mmtypes.ConfigType,
		},
		{
			Name: agex.ResearchCMCAPIHandlerName,
			API:  agex.DefaultResearchCMCAPIConfig,
			Type: mmtypes.ConfigType,
		},
	}

	MarketMapProviderNames = map[string]struct{}{
		agex.Name:                      {},
		agex.SwitchOverAPIHandlerName:  {},
		agex.ResearchAPIHandlerName:    {},
		agex.ResearchCMCAPIHandlerName: {},
		marketmap.Name:                 {},
	}
)

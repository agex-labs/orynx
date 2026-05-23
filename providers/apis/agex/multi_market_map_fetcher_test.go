package agex_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	orynxtypes "github.com/agex-labs/orynx/pkg/types"
	"github.com/agex-labs/orynx/providers/apis/agex"
	apihandlermocks "github.com/agex-labs/orynx/providers/base/api/handlers/mocks"
	providertypes "github.com/agex-labs/orynx/providers/types"
	mmclient "github.com/agex-labs/orynx/service/clients/marketmap/types"
	mmtypes "github.com/agex-labs/orynx/x/marketmap/types"
)

func TestAGEXMultiMarketMapFetcher(t *testing.T) {
	agexMainnetMMFetcher := apihandlermocks.NewAPIFetcher[mmclient.Chain, *mmtypes.MarketMapResponse](t)
	agexResearchMMFetcher := apihandlermocks.NewAPIFetcher[mmclient.Chain, *mmtypes.MarketMapResponse](t)

	fetcher := agex.NewAGEXResearchMarketMapFetcher(agexMainnetMMFetcher, agexResearchMMFetcher, zap.NewExample(), false)

	t.Run("test that if the mainnet api-price fetcher response is unresolved, we return it", func(t *testing.T) {
		ctx := context.Background()
		agexMainnetMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{
			UnResolved: map[mmclient.Chain]providertypes.UnresolvedResult{
				agex.AGEXChain: {
					ErrorWithCode: providertypes.NewErrorWithCode(fmt.Errorf("error"), providertypes.ErrorAPIGeneral),
				},
			},
		}, nil).Once()
		agexResearchMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{}, nil).Once()

		response := fetcher.Fetch(ctx, []mmclient.Chain{agex.AGEXChain})
		require.Len(t, response.UnResolved, 1)
	})

	t.Run("test that if the agex-research response is unresolved, we return that", func(t *testing.T) {
		ctx := context.Background()
		agexMainnetMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{
			Resolved: map[mmclient.Chain]providertypes.ResolvedResult[*mmtypes.MarketMapResponse]{
				agex.AGEXChain: providertypes.NewResult(&mmtypes.MarketMapResponse{}, time.Now()),
			},
		}, nil).Once()
		agexResearchMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{
			UnResolved: map[mmclient.Chain]providertypes.UnresolvedResult{
				agex.AGEXChain: {},
			},
		}, nil).Once()

		response := fetcher.Fetch(ctx, []mmclient.Chain{agex.AGEXChain})
		require.Len(t, response.UnResolved, 1)
	})

	t.Run("test if both responses are resolved, the tickers are appended to each other + validation fails", func(t *testing.T) {
		ctx := context.Background()
		agexMainnetMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{
			Resolved: map[mmclient.Chain]providertypes.ResolvedResult[*mmtypes.MarketMapResponse]{
				agex.AGEXChain: providertypes.NewResult(&mmtypes.MarketMapResponse{
					MarketMap: mmtypes.MarketMap{
						Markets: map[string]mmtypes.Market{
							"BTC/USD": {},
						},
					},
				}, time.Now()),
			},
		}, nil).Once()
		agexResearchMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{
			Resolved: map[mmclient.Chain]providertypes.ResolvedResult[*mmtypes.MarketMapResponse]{
				agex.AGEXChain: providertypes.NewResult(&mmtypes.MarketMapResponse{
					MarketMap: mmtypes.MarketMap{
						Markets: map[string]mmtypes.Market{
							"ETH/USD": {},
						},
					},
				}, time.Now()),
			},
		}, nil).Once()

		response := fetcher.Fetch(ctx, []mmclient.Chain{agex.AGEXChain})
		require.Len(t, response.UnResolved, 1)
	})

	t.Run("test that if both responses are resolved, the responses are aggregated + validation passes", func(t *testing.T) {
		ctx := context.Background()
		agexMainnetMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{
			Resolved: map[mmclient.Chain]providertypes.ResolvedResult[*mmtypes.MarketMapResponse]{
				agex.AGEXChain: providertypes.NewResult(&mmtypes.MarketMapResponse{
					MarketMap: mmtypes.MarketMap{
						Markets: map[string]mmtypes.Market{
							"BTC/USD": {
								Ticker: mmtypes.Ticker{
									CurrencyPair:     orynxtypes.NewCurrencyPair("BTC", "USD"),
									Decimals:         8,
									MinProviderCount: 1,
									Enabled:          true,
								},
								ProviderConfigs: []mmtypes.ProviderConfig{
									{
										Name:           "agex",
										OffChainTicker: "BTC/USD",
									},
								},
							},
						},
					},
				}, time.Now()),
			},
		}, nil).Once()
		agexResearchMMFetcher.On("Fetch", ctx, []mmclient.Chain{agex.AGEXChain}).Return(mmclient.MarketMapResponse{
			Resolved: map[mmclient.Chain]providertypes.ResolvedResult[*mmtypes.MarketMapResponse]{
				agex.AGEXChain: providertypes.NewResult(&mmtypes.MarketMapResponse{
					MarketMap: mmtypes.MarketMap{
						Markets: map[string]mmtypes.Market{
							"ETH/USD": {
								Ticker: mmtypes.Ticker{
									CurrencyPair:     orynxtypes.NewCurrencyPair("ETH", "USD"),
									Decimals:         8,
									MinProviderCount: 1,
								},
								ProviderConfigs: []mmtypes.ProviderConfig{
									{
										Name:           "agex",
										OffChainTicker: "BTC/USD",
									},
								},
							},
						},
					},
				}, time.Now()),
			},
		}, nil).Once()

		response := fetcher.Fetch(ctx, []mmclient.Chain{agex.AGEXChain})
		require.Len(t, response.Resolved, 1)

		marketMap := response.Resolved[agex.AGEXChain].Value.MarketMap

		require.Len(t, marketMap.Markets, 2)
		require.Contains(t, marketMap.Markets, "BTC/USD")
		require.Contains(t, marketMap.Markets, "ETH/USD")
	})
}

package agex

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/agex-labs/orynx/cmd/constants/marketmaps"
	"github.com/agex-labs/orynx/oracle/config"
	"github.com/agex-labs/orynx/providers/apis/coinmarketcap"
	apihandlers "github.com/agex-labs/orynx/providers/base/api/handlers"
	"github.com/agex-labs/orynx/providers/base/api/metrics"
	providertypes "github.com/agex-labs/orynx/providers/types"
	mmclient "github.com/agex-labs/orynx/service/clients/marketmap/types"
	mmtypes "github.com/agex-labs/orynx/x/marketmap/types"
)

var (
	_         mmclient.MarketMapFetcher = &MultiMarketMapRestAPIFetcher{}
	AGEXChain                           = mmclient.Chain{
		ChainID: ChainID,
	}
)

// NewAGEXResearchMarketMapFetcher returns a MultiMarketMapFetcher composed of agex mainnet + research
// apiDataHandlers.
func DefaultAGEXResearchMarketMapFetcher(
	rh apihandlers.RequestHandler,
	metrics metrics.APIMetrics,
	api config.APIConfig,
	logger *zap.Logger,
) (*MultiMarketMapRestAPIFetcher, error) {
	if rh == nil {
		return nil, fmt.Errorf("request handler is nil")
	}

	if metrics == nil {
		return nil, fmt.Errorf("metrics is nil")
	}

	if !api.Enabled {
		return nil, fmt.Errorf("api is not enabled")
	}

	if err := api.ValidateBasic(); err != nil {
		return nil, err
	}

	if len(api.Endpoints) != 2 {
		return nil, fmt.Errorf("expected two endpoint, got %d", len(api.Endpoints))
	}

	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	// make a agex research api-handler
	researchAPIDataHandler, err := NewResearchAPIHandler(logger, api)
	if err != nil {
		return nil, err
	}

	mainnetAPIDataHandler := &APIHandler{
		logger: logger,
		api:    api,
	}

	mainnetFetcher, err := apihandlers.NewRestAPIFetcher(
		rh,
		mainnetAPIDataHandler,
		metrics,
		api,
		logger,
	)
	if err != nil {
		return nil, err
	}

	researchFetcher, err := apihandlers.NewRestAPIFetcher(
		rh,
		researchAPIDataHandler,
		metrics,
		api,
		logger,
	)
	if err != nil {
		return nil, err
	}

	return NewAGEXResearchMarketMapFetcher(
		mainnetFetcher,
		researchFetcher,
		logger,
		api.Name == ResearchCMCAPIHandlerName,
	), nil
}

// MultiMarketMapRestAPIFetcher is an implementation of a RestAPIFetcher that wraps
// two underlying Fetchers for fetching the market-map according to agex mainnet and
// the additional markets that can be added according to the agex research json.
type MultiMarketMapRestAPIFetcher struct {
	// agex mainnet fetcher is the api-fetcher for the agex mainnet market-map
	agexMainnetFetcher mmclient.MarketMapFetcher

	// agex research fetcher is the api-fetcher for the agex research market-map
	agexResearchFetcher mmclient.MarketMapFetcher

	// logger is the logger for the fetcher
	logger *zap.Logger

	// isCMCOnly is a flag that indicates whether the fetcher should only return CoinMarketCap markets.
	isCMCOnly bool
}

// NewAGEXResearchMarketMapFetcher returns an aggregated market-map among the agex mainnet and the agex research json.
func NewAGEXResearchMarketMapFetcher(
	mainnetFetcher, researchFetcher mmclient.MarketMapFetcher,
	logger *zap.Logger,
	isCMCOnly bool,
) *MultiMarketMapRestAPIFetcher {
	return &MultiMarketMapRestAPIFetcher{
		agexMainnetFetcher:  mainnetFetcher,
		agexResearchFetcher: researchFetcher,
		logger:              logger.With(zap.String("module", "agex-research-market-map-fetcher")),
		isCMCOnly:           isCMCOnly,
	}
}

// Fetch fetches the market map from the underlying fetchers and combines the results. If any of the underlying
// fetchers fetch for a chain that is different from the chain that the fetcher is initialized with, those responses
// will be ignored.
func (f *MultiMarketMapRestAPIFetcher) Fetch(ctx context.Context, chains []mmclient.Chain) mmclient.MarketMapResponse {
	// call the underlying fetchers + await their responses
	// channel to aggregate responses
	agexMainnetResponseChan := make(chan mmclient.MarketMapResponse, 1) // buffer so that sends / receives are non-blocking
	agexResearchResponseChan := make(chan mmclient.MarketMapResponse, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	// fetch agex mainnet
	go func() {
		defer wg.Done()
		agexMainnetResponseChan <- f.agexMainnetFetcher.Fetch(ctx, chains)
		f.logger.Debug("fetched valid market-map from agex mainnet")
	}()

	// fetch agex research
	go func() {
		defer wg.Done()
		agexResearchResponseChan <- f.agexResearchFetcher.Fetch(ctx, chains)
		f.logger.Debug("fetched valid market-map from agex research")
	}()

	// wait for both fetchers to finish
	wg.Wait()

	agexMainnetMarketMapResponse := <-agexMainnetResponseChan
	agexResearchMarketMapResponse := <-agexResearchResponseChan

	// if the agex mainnet market-map response failed, return the agex mainnet failed response
	if _, ok := agexMainnetMarketMapResponse.UnResolved[AGEXChain]; ok {
		f.logger.Error("agex mainnet market-map fetch failed", zap.Any("response", agexMainnetMarketMapResponse))
		return agexMainnetMarketMapResponse
	}

	// if the agex research market-map response failed, return the agex research failed response
	if _, ok := agexResearchMarketMapResponse.UnResolved[AGEXChain]; ok {
		f.logger.Error("agex research market-map fetch failed", zap.Any("response", agexResearchMarketMapResponse))
		return agexResearchMarketMapResponse
	}

	// otherwise, add all markets from agex research
	agexMainnetMarketMap := agexMainnetMarketMapResponse.Resolved[AGEXChain].Value.MarketMap

	resolved, ok := agexResearchMarketMapResponse.Resolved[AGEXChain]
	if ok {
		for ticker, market := range resolved.Value.MarketMap.Markets {
			// if the market is not already in the agex mainnet market-map, add it
			if _, ok := agexMainnetMarketMap.Markets[ticker]; !ok {
				f.logger.Debug("adding market from agex research", zap.String("ticker", ticker))
				agexMainnetMarketMap.Markets[ticker] = market
			}
		}
	}

	// if the fetcher is only for CoinMarketCap markets, filter out all non-CMC markets
	if f.isCMCOnly {
		for ticker, market := range agexMainnetMarketMap.Markets {
			market.Ticker.MinProviderCount = 1
			agexMainnetMarketMap.Markets[ticker] = market

			var (
				seenCMC     = false
				cmcProvider mmtypes.ProviderConfig
			)

			for _, provider := range market.ProviderConfigs {
				if provider.Name == coinmarketcap.Name {
					seenCMC = true
					cmcProvider = provider
				}
			}

			// if we saw a CMC provider, add it to the market
			if seenCMC {
				market.ProviderConfigs = []mmtypes.ProviderConfig{cmcProvider}
				agexMainnetMarketMap.Markets[ticker] = market
				continue
			}

			// If we did not see a CMC provider, we can attempt to add it using the CMC marketmap
			cmcMarket, ok := marketmaps.CoinMarketCapMarketMap.Markets[ticker]
			if !ok {
				f.logger.Info("did not find CMC market for ticker", zap.String("ticker", ticker))
				delete(agexMainnetMarketMap.Markets, ticker)
				continue
			}

			// add the CMC provider to the market
			market.ProviderConfigs = cmcMarket.ProviderConfigs
			agexMainnetMarketMap.Markets[ticker] = market
		}
	}

	// validate the combined market-map
	if err := agexMainnetMarketMap.ValidateBasic(); err != nil {
		f.logger.Error("combined market-map failed validation", zap.Error(err))

		return mmclient.NewMarketMapResponseWithErr(
			chains,
			providertypes.NewErrorWithCode(
				fmt.Errorf("combined market-map failed validation: %w", err),
				providertypes.ErrorUnknown,
			),
		)
	}

	agexMainnetMarketMapResponse.Resolved[AGEXChain].Value.MarketMap = agexMainnetMarketMap

	return agexMainnetMarketMapResponse
}

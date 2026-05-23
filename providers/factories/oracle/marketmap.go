package oracle

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/agex-labs/orynx/oracle/config"
	"github.com/agex-labs/orynx/providers/apis/agex"
	"github.com/agex-labs/orynx/providers/apis/marketmap"
	"github.com/agex-labs/orynx/providers/base"
	apihandlers "github.com/agex-labs/orynx/providers/base/api/handlers"
	apimetrics "github.com/agex-labs/orynx/providers/base/api/metrics"
	providermetrics "github.com/agex-labs/orynx/providers/base/metrics"
	"github.com/agex-labs/orynx/service/clients/marketmap/types"
	mmtypes "github.com/agex-labs/orynx/x/marketmap/types"
)

// MarketMapProviderFactory returns a sample implementation of the market map provider. This provider
// is responsible for fetching updates to the canonical market map on the given chain.
func MarketMapProviderFactory(
	logger *zap.Logger,
	providerMetrics providermetrics.ProviderMetrics,
	apiMetrics apimetrics.APIMetrics,
	cfg config.ProviderConfig,
) (*types.MarketMapProvider, error) {
	// Validate the provider config.
	err := cfg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: cfg.API.MaxQueries,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: cfg.API.Timeout,
	}

	var (
		apiDataHandler   types.MarketMapAPIDataHandler
		ids              []types.Chain
		marketMapFetcher types.MarketMapFetcher
	)

	requestHandler, err := apihandlers.NewRequestHandlerImpl(client)
	if err != nil {
		return nil, err
	}

	switch cfg.Name {
	case agex.Name:
		apiDataHandler, err = agex.NewAPIHandler(logger, cfg.API)
		ids = []types.Chain{{ChainID: agex.ChainID}}
	case agex.SwitchOverAPIHandlerName:
		marketMapFetcher, err = agex.NewDefaultSwitchOverMarketMapFetcher(
			logger,
			cfg.API,
			requestHandler,
			apiMetrics,
		)
		ids = []types.Chain{{ChainID: agex.ChainID}}
	case agex.ResearchAPIHandlerName, agex.ResearchCMCAPIHandlerName:
		marketMapFetcher, err = agex.DefaultAGEXResearchMarketMapFetcher(
			requestHandler,
			apiMetrics,
			cfg.API,
			logger,
		)
		ids = []types.Chain{{ChainID: agex.ChainID}}
	default:
		marketMapFetcher, err = marketmap.NewMarketMapFetcher(
			logger,
			cfg.API,
			apiMetrics,
		)
		ids = []types.Chain{{ChainID: "local-node"}}
	}
	if err != nil {
		return nil, err
	}

	if marketMapFetcher == nil {
		marketMapFetcher, err = apihandlers.NewRestAPIFetcher(
			requestHandler,
			apiDataHandler,
			apiMetrics,
			cfg.API,
			logger,
		)
		if err != nil {
			return nil, err
		}
	}

	queryHandler, err := types.NewMarketMapAPIQueryHandlerWithMarketMapFetcher(
		logger,
		cfg.API,
		marketMapFetcher,
		apiMetrics,
	)
	if err != nil {
		return nil, err
	}

	return types.NewMarketMapProvider(
		base.WithName[types.Chain, *mmtypes.MarketMapResponse](cfg.Name),
		base.WithLogger[types.Chain, *mmtypes.MarketMapResponse](logger),
		base.WithAPIQueryHandler(queryHandler),
		base.WithAPIConfig[types.Chain, *mmtypes.MarketMapResponse](cfg.API),
		base.WithMetrics[types.Chain, *mmtypes.MarketMapResponse](providerMetrics),
		base.WithIDs[types.Chain, *mmtypes.MarketMapResponse](ids),
	)
}

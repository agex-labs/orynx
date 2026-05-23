package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	orynxtypes "github.com/agex-labs/orynx/pkg/types"

	"github.com/agex-labs/orynx/x/oracle/types"
)

func TestGenesisValidation(t *testing.T) {
	tcs := []struct {
		name       string
		cpgs       []types.CurrencyPairGenesis
		nextID     uint64
		expectPass bool
	}{
		{
			"if any of the currency-pair geneses are invalid - fail",
			[]types.CurrencyPairGenesis{
				{
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "AA",
						Quote: "BB",
					},
				},
				{
					// invalid CurrencyPairGenesis
					CurrencyPair: orynxtypes.CurrencyPair{
						Base: "BB",
					},
				},
			},
			0,
			false,
		},
		{
			"if the CurrencyPairPrice is nil, but the nonce is non-zero - fail",
			[]types.CurrencyPairGenesis{
				{
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "AA",
						Quote: "BB",
					},
					Nonce: 10,
				},
			},
			0,
			false,
		},
		{
			"if all of the currency-pair geneses are valid - pass",
			[]types.CurrencyPairGenesis{
				{
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "AA",
						Quote: "BB",
					},
					Id: 0,
				},
				{
					// invalid CurrencyPairGenesis
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "BB",
						Quote: "CC",
					},
					Id: 1,
				},
			},
			2,
			true,
		},
		{
			"if any of the CurrencyPairGenesis ID's are duplicated - fail",
			[]types.CurrencyPairGenesis{
				{
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "AA",
						Quote: "BB",
					},
					Id: 1,
				},
				{
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "BB",
						Quote: "CC",
					},
					Id: 1,
				},
			},
			3,
			false,
		},
		{
			"if any of the CurrencyPairs are repeated - fail",
			[]types.CurrencyPairGenesis{
				{
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "AA",
						Quote: "BB",
					},
					Id: 1,
				},
				{
					CurrencyPair: orynxtypes.CurrencyPair{
						Base:  "AA",
						Quote: "BB",
					},
					Id: 2,
				},
			},
			3,
			false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			gs := types.NewGenesisState(tc.cpgs, tc.nextID)
			err := gs.Validate()

			if tc.expectPass {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
			}
		})
	}
}

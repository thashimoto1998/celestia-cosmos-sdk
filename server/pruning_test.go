package server

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
)

func TestGetPruningOptionsFromFlags(t *testing.T) {
	tests := []struct {
		name            string
		initParams      func() *viper.Viper
		expectedOptions pruningtypes.PruningOptions
		wantErr         bool
	}{
		{
			name: FlagPruning,
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, pruningtypes.PruningOptionNothing)
				return v
			},
			expectedOptions: pruningtypes.PruneNothing,
		},
		{
			name: "custom pruning options",
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, pruningtypes.PruningOptionCustom)
				v.Set(FlagPruningKeepRecent, 1234)
				v.Set(FlagPruningKeepEvery, 4321)
				v.Set(FlagPruningInterval, 10)

				return v
			},
			expectedOptions: pruningtypes.PruningOptions{
				KeepRecent: 1234,
				KeepEvery:  4321,
				Interval:   10,
			},
		},
		{
			name: pruningtypes.PruningOptionDefault,
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, pruningtypes.PruningOptionDefault)
				return v
			},
			expectedOptions: pruningtypes.PruneDefault,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(j *testing.T) {
			viper.Reset()
			viper.SetDefault(FlagPruning, pruningtypes.PruningOptionDefault)
			v := tt.initParams()

			opts, err := GetPruningOptionsFromFlags(v)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tt.expectedOptions, opts)
		})
	}
}

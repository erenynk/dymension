package cli_test

import (
	"fmt"
	"strconv"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/testutil/network"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp/client/cli"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func networkWithLatestFinalizedStateIndexObjects(t *testing.T, n int) (*network.Network, []types.StateInfoIndex) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {

		stateInfo := types.StateInfo{
			StateInfoIndex: types.StateInfoIndex{
				RollappId: strconv.Itoa(i),
				Index:     uint64(i)},
			Status: types.STATE_STATUS_FINALIZED,
		}
		nullify.Fill(&stateInfo)
		state.StateInfoList = append(state.StateInfoList, stateInfo)
		state.LatestFinalizedStateIndexList = append(state.LatestFinalizedStateIndexList, stateInfo.StateInfoIndex)

	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.LatestFinalizedStateIndexList
}

func TestShowLatestFinalizedStateIndex(t *testing.T) {
	net, objs := networkWithLatestFinalizedStateIndexObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc        string
		idRollappId string

		args []string
		err  error
		obj  types.StateInfoIndex
	}{
		{
			desc:        "found",
			idRollappId: objs[0].RollappId,

			args: common,
			obj:  objs[0],
		},
		{
			desc:        "not found",
			idRollappId: strconv.Itoa(100000),

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idRollappId,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowLatestFinalizedStateInfo(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetLatestFinalizedStateInfoResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.StateInfo)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.StateInfo.StateInfoIndex),
				)
			}
		})
	}
}

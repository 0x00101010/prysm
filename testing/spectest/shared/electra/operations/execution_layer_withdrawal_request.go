package operations

import (
	"context"
	"path"
	"testing"

	"github.com/golang/snappy"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/electra"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/interfaces"
	enginev1 "github.com/prysmaticlabs/prysm/v5/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/testing/require"
	"github.com/prysmaticlabs/prysm/v5/testing/spectest/utils"
	"github.com/prysmaticlabs/prysm/v5/testing/util"
)

func RunExecutionLayerWithdrawalRequestTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "electra", "operations/execution_layer_withdrawal_request/pyspec_tests")
	if len(testFolders) == 0 {
		t.Fatalf("No test folders found for %s/%s/%s", config, "electra", "operations/execution_layer_withdrawal_request/pyspec_tests")
	}
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			executionLayerWithdrawalRequestFile, err := util.BazelFileBytes(folderPath, "execution_layer_withdrawal_request.ssz_snappy")
			require.NoError(t, err)
			executionLayerWithdrawalRequestSSZ, err := snappy.Decode(nil /* dst */, executionLayerWithdrawalRequestFile)
			require.NoError(t, err, "Failed to decompress")
			withdrawalRequest := &enginev1.ExecutionLayerWithdrawalRequest{}
			require.NoError(t, withdrawalRequest.UnmarshalSSZ(executionLayerWithdrawalRequestSSZ), "Failed to unmarshal")
			body := &ethpb.BeaconBlockBodyElectra{ExecutionPayload: &enginev1.ExecutionPayloadElectra{
				WithdrawalRequests: []*enginev1.ExecutionLayerWithdrawalRequest{
					withdrawalRequest,
				},
			}}
			RunBlockOperationTest(t, folderPath, body, func(ctx context.Context, s state.BeaconState, b interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error) {
				bod, ok := b.Block().Body().(interfaces.ROBlockBodyElectra)
				require.Equal(t, true, ok)
				e, err := bod.Execution()
				require.NoError(t, err)
				exe, ok := e.(interfaces.ExecutionDataElectra)
				require.Equal(t, true, ok)
				return electra.ProcessExecutionLayerWithdrawalRequests(ctx, s, exe.WithdrawalRequests())
			})
		})
	}
}

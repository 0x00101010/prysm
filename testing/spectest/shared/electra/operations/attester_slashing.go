package operations

import (
	"context"
	"path"
	"testing"

	"github.com/golang/snappy"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/blocks"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/validators"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/interfaces"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/testing/require"
	"github.com/prysmaticlabs/prysm/v5/testing/spectest/utils"
	"github.com/prysmaticlabs/prysm/v5/testing/util"
)

func RunAttesterSlashingTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "electra", "operations/attester_slashing/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			attSlashingFile, err := util.BazelFileBytes(folderPath, "attester_slashing.ssz_snappy")
			require.NoError(t, err)
			attSlashingSSZ, err := snappy.Decode(nil /* dst */, attSlashingFile)
			require.NoError(t, err, "Failed to decompress")
			attSlashing := &ethpb.AttesterSlashingElectra{}
			require.NoError(t, attSlashing.UnmarshalSSZ(attSlashingSSZ), "Failed to unmarshal")

			body := &ethpb.BeaconBlockBodyElectra{AttesterSlashings: []*ethpb.AttesterSlashingElectra{attSlashing}}
			RunBlockOperationTest(t, folderPath, body, func(ctx context.Context, s state.BeaconState, b interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error) {
				return blocks.ProcessAttesterSlashings(ctx, s, b.Block().Body().AttesterSlashings(), validators.SlashValidator)
			})
		})
	}
}

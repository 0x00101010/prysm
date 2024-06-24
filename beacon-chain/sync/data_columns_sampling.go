package sync

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/prysmaticlabs/prysm/v5/async"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/v5/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/peerdas"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/p2p"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/p2p/types"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/startup"
	fieldparams "github.com/prysmaticlabs/prysm/v5/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v5/config/params"
	"github.com/prysmaticlabs/prysm/v5/crypto/rand"
	eth "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/runtime/version"
)

const PeerRefreshInterval = 1 * time.Minute

type roundSummary struct {
	RequestedColumns []uint64
	MissingColumns   map[uint64]bool
}

// DataColumnSampler defines the interface for sampling data columns from peers for requested block root and samples count.
type DataColumnSampler interface {
	// Run starts the data column sampling service.
	Run(ctx context.Context)
}

var _ DataColumnSampler = (*DataColumnSampler1D)(nil)

type DataColumnSampler1D struct {
	sync.RWMutex

	p2p           p2p.P2P
	clock         *startup.Clock
	ctxMap        ContextByteVersions
	stateNotifier statefeed.Notifier

	// custodyColumns is a set of columns that the node is responsible for custody.
	custodyColumns map[uint64]bool
	// columnFromPeer maps a peer to the columns it is responsible for custody.
	columnFromPeer map[peer.ID]map[uint64]bool
	// peerFromColumn maps a column to the peer responsible for custody.
	peerFromColumn map[uint64]map[peer.ID]bool
}

// NewDataColumnSampler1D creates a new 1D data column sampler.
func NewDataColumnSampler1D(
	p2p p2p.P2P,
	clock *startup.Clock,
	ctxMap ContextByteVersions,
	stateNotifier statefeed.Notifier,
) *DataColumnSampler1D {
	columnToPeerMap := make(map[uint64]map[peer.ID]bool, params.BeaconConfig().NumberOfColumns)
	for i := uint64(0); i < params.BeaconConfig().NumberOfColumns; i++ {
		columnToPeerMap[i] = make(map[peer.ID]bool)
	}

	return &DataColumnSampler1D{
		p2p:            p2p,
		clock:          clock,
		ctxMap:         ctxMap,
		stateNotifier:  stateNotifier,
		columnFromPeer: make(map[peer.ID]map[uint64]bool),
		peerFromColumn: columnToPeerMap,
	}
}

// Run implements DataColumnSampler.
func (d *DataColumnSampler1D) Run(ctx context.Context) {
	// verify if we need to run sampling or not, if not, return directly
	csc := peerdas.CustodySubnetCount()
	columns, err := peerdas.CustodyColumns(d.p2p.NodeID(), csc)
	if err != nil {
		log.WithError(err).Error("Failed to determine local custody columns")
		return
	}
	d.custodyColumns = columns

	custodyColumnsCount := uint64(len(columns))
	if uint64(len(columns)) > params.BeaconConfig().NumberOfColumns/2 {
		log.WithFields(logrus.Fields{
			"custodyColumnsCount": custodyColumnsCount,
			"totalColumns":        params.BeaconConfig().NumberOfColumns,
		}).Debug("The node custodies at least the half the data columns, no need to sample")
		return
	}

	// initialize peer info first.
	d.refreshPeerInfo()

	// periodically refresh peer info to keep peer <-> column mapping up to date.
	async.RunEvery(ctx, PeerRefreshInterval, d.refreshPeerInfo)

	// start the sampling loop.
	d.samplingLoop(ctx)
}

func (d *DataColumnSampler1D) samplingLoop(ctx context.Context) {
	// listen to block arrival
	// verify
	// randomize columns
	// sample data columns with incremental das
	stateCh := make(chan *feed.Event, 1)
	stateSub := d.stateNotifier.StateFeed().Subscribe(stateCh)
	defer stateSub.Unsubscribe()

	for {
		select {
		case evt := <-stateCh:
			d.handleStateNotification(ctx, evt)
		case err := <-stateSub.Err():
			log.WithError(err).Error("DataColumnSampler1D subscription to state feed failed")
		case <-ctx.Done():
			log.Debug("Context canceled, exiting data column sampling loop.")
			return
		}
	}
}

// Refresh peer information.
func (d *DataColumnSampler1D) refreshPeerInfo() {
	d.Lock()
	defer d.Unlock()

	activePeers := d.p2p.Peers().Active()
	d.prunePeerInfo(activePeers)

	for _, pid := range d.p2p.Peers().Active() {
		if _, ok := d.columnFromPeer[pid]; ok {
			continue
		}

		csc := d.p2p.CustodyCountFromRemotePeer(pid)
		nid, err := p2p.ConvertPeerIDToNodeID(pid)
		if err != nil {
			log.WithError(err).WithField("peerID", pid).Error("Failed to convert peer ID to node ID")
			continue
		}

		columns, err := peerdas.CustodyColumns(nid, csc)
		if err != nil {
			log.WithError(err).WithField("peerID", pid).Error("Failed to determine peer custody columns")
			continue
		}

		d.columnFromPeer[pid] = columns
		for column := range columns {
			d.peerFromColumn[column][pid] = true
		}
	}
}

// prunePeerInfo prunes inactive peers from peerFromColumn and columnFromPeer.
// this should not be called outside of refreshPeerInfo without being locked.
func (d *DataColumnSampler1D) prunePeerInfo(activePeers []peer.ID) {
	active := make(map[peer.ID]bool)
	for _, pid := range activePeers {
		active[pid] = true
	}

	for pid := range d.columnFromPeer {
		if !active[pid] {
			d.prunePeer(pid)
		}
	}
}

// prunePeer removes a peer from stored peer info map, it should be called with lock held.
func (d *DataColumnSampler1D) prunePeer(pid peer.ID) {
	delete(d.columnFromPeer, pid)
	for _, peers := range d.peerFromColumn {
		delete(peers, pid)
	}
}

func (d *DataColumnSampler1D) handleStateNotification(ctx context.Context, event *feed.Event) {
	if event.Type != statefeed.BlockProcessed {
		return
	}

	data, ok := event.Data.(*statefeed.BlockProcessedData)
	if !ok {
		log.Error("Event feed data is not of type *statefeed.BlockProcessedData")
		return
	}

	if !data.Verified {
		// We only process blocks that have been verified
		log.Error("Data is not verified")
		return
	}

	if data.SignedBlock.Version() < version.Deneb {
		log.Debug("Pre Deneb block, skipping data column sampling")
		return
	}

	// Get the commitments for this block.
	commitments, err := data.SignedBlock.Block().Body().BlobKzgCommitments()
	if err != nil {
		log.WithError(err).Error("Failed to get blob KZG commitments")
		return
	}

	// Skip if there are no commitments.
	if len(commitments) == 0 {
		log.Debug("No commitments in block, skipping data column sampling")
		return
	}

	// Randomize columns for sample selection.
	randomizedColumns := randomizeColumns(d.custodyColumns)

	ok, _, err = d.incrementalDAS(ctx, data.BlockRoot, randomizedColumns)
	if err != nil {
		log.WithError(err).Error("Failed to run incremental DAS")
	}

	if ok {
		log.WithFields(logrus.Fields{
			"root":    fmt.Sprintf("%#x", data.BlockRoot),
			"columns": randomizedColumns,
		}).Debug("Data column sampling successful")
	} else {
		log.WithFields(logrus.Fields{
			"root":    fmt.Sprintf("%#x", data.BlockRoot),
			"columns": randomizedColumns,
		}).Warning("Data column sampling failed")
	}
}

// incrementalDAS samples data columns from active peers using incremental DAS.
// https://ethresear.ch/t/lossydas-lossy-incremental-and-diagonal-sampling-for-data-availability/18963#incrementaldas-dynamically-increase-the-sample-size-10
func (d *DataColumnSampler1D) incrementalDAS(
	ctx context.Context,
	root [fieldparams.RootLength]byte,
	columns []uint64,
) (bool, []roundSummary, error) {
	samplesPerSlot, missingColumnsCount := params.BeaconConfig().SamplesPerSlot, uint64(0)
	lo, hi := uint64(0), peerdas.ExtendedSampleCount(samplesPerSlot, missingColumnsCount)
	roundSummaries := make([]roundSummary, 0, 1) // We optimistically allocate only one round summary.

	for round := 1; ; /*No exit condition */ round++ {
		if hi > params.BeaconConfig().NumberOfColumns {
			// We already tried to sample all possible columns, this is the unhappy path.
			log.WithField("root", fmt.Sprintf("%#x", root)).Warning("Some columns are still missing after sampling all possible columns")
			return false, roundSummaries, nil
		}

		// Get the columns to sample for this round.
		columnsToSample := columns[lo:hi]
		columnsToSampleCount := hi - lo

		// Sample data columns from peers in parallel.
		retrieved, err := d.sampleDataColumns(ctx, root, columnsToSample)
		if err != nil {
			return false, roundSummaries, nil
		}

		missingSamples := make(map[uint64]bool)
		for _, column := range columnsToSample {
			if !retrieved[column] {
				missingSamples[column] = true
			}
		}

		roundSummaries = append(roundSummaries, roundSummary{
			RequestedColumns: columnsToSample,
			MissingColumns:   missingSamples,
		})

		retrievedCount := uint64(len(retrieved))
		if retrievedCount == columnsToSampleCount {
			// All columns were correctly sampled, this is the happy path.
			log.WithFields(logrus.Fields{
				"root":         fmt.Sprintf("%#x", root),
				"roundsNeeded": round,
			}).Debug("All columns were successfully sampled")
			return true, roundSummaries, nil
		}

		if retrievedCount > columnsToSampleCount {
			// This should never happen.
			return false, roundSummaries, errors.New("retrieved more columns than requested")
		}

		// missing columns over the allowed failures, extend the samples.
		missingColumnsCount += columnsToSampleCount - retrievedCount
		oldHi := hi
		lo = hi
		// which should be the input? samplesPerSlot?
		hi = peerdas.ExtendedSampleCount(samplesPerSlot, missingColumnsCount)

		log.WithFields(logrus.Fields{
			"root":                fmt.Sprintf("%#x", root),
			"round":               round,
			"missingColumnsCount": missingColumnsCount,
			"currentSampleCount":  oldHi,
			"nextSampleCount":     hi,
		}).Debug("Some columns are still missing after sampling this round.")
	}
}

func (d *DataColumnSampler1D) sampleDataColumns(
	ctx context.Context,
	root [fieldparams.RootLength]byte,
	columns []uint64,
) (map[uint64]bool, error) {
	// distribute samples to peer
	peerToColumns, err := d.distributeSamplesToPeer(columns)
	if err != nil {
		log.WithError(err).Error("Failed to distribute samples to peers")
		return nil, errors.Wrap(err, "failed distributing samples to peer")
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	res := make(map[uint64]bool)
	sampleFromPeer := func(pid peer.ID, cols map[uint64]bool) {
		defer wg.Done()
		retrieved := d.sampleDataColumnsFromPeer(ctx, pid, root, cols)

		mu.Lock()
		for col := range retrieved {
			res[col] = true
		}
		mu.Unlock()
	}

	// sample from peers in parallel
	for pid, cols := range peerToColumns {
		wg.Add(1)
		go sampleFromPeer(pid, cols)
	}

	wg.Wait()
	return res, nil
}

// distributeSamplesToPeer distributes samples to peers based on the columns they are responsible for.
// Currently it randomizes peer selection for a column and did not take into account whole peer distribution balance. It could be improved if needed.
func (d *DataColumnSampler1D) distributeSamplesToPeer(
	columns []uint64,
) (map[peer.ID]map[uint64]bool, error) {
	dist := make(map[peer.ID]map[uint64]bool)

	for _, col := range columns {
		peers := d.peerFromColumn[col]
		if len(peers) == 0 {
			return nil, errors.Errorf("no peers responsible for column %d", col)
		}

		pid := selectRandomPeer(peers)
		if _, ok := dist[pid]; !ok {
			dist[pid] = make(map[uint64]bool)
		}
		dist[pid][col] = true
	}

	return dist, nil
}

func (d *DataColumnSampler1D) sampleDataColumnsFromPeer(
	ctx context.Context,
	pid peer.ID,
	root [fieldparams.RootLength]byte,
	columns map[uint64]bool,
) map[uint64]bool {
	retrieved := make(map[uint64]bool)

	req := make(types.DataColumnSidecarsByRootReq, 0)
	for col := range columns {
		// Skip querying peers for columns we already have.
		if d.custodyColumns[col] {
			retrieved[col] = true
			continue
		}

		req = append(req, &eth.DataColumnIdentifier{
			BlockRoot:   root[:],
			ColumnIndex: col,
		})
	}

	// No columns to request.
	if len(req) == 0 {
		return retrieved
	}

	// Send the request to the peer.
	roDataColumns, err := SendDataColumnSidecarByRoot(ctx, d.clock, d.p2p, pid, d.ctxMap, &req)
	if err != nil {
		log.WithError(err).Error("Failed to send data column sidecar by root")
		return nil
	}

	for _, roDataColumn := range roDataColumns {
		actualRoot := roDataColumn.BlockRoot()
		if actualRoot != root {
			// TODO: Should we decrease the peer score here?
			log.WithFields(logrus.Fields{
				"peerID":        pid,
				"requestedRoot": fmt.Sprintf("%#x", root),
				"actualRoot":    fmt.Sprintf("%#x", actualRoot),
			}).Warning("Actual root does not match requested root")
			continue
		}

		retrieved[roDataColumn.ColumnIndex] = true
	}

	missingColumns := make(map[uint64]bool)
	for col := range columns {
		if !retrieved[col] {
			missingColumns[col] = true
			// TODO: Should we decrease the peer score here?
		}
	}

	log.WithFields(logrus.Fields{
		"peerID":           pid,
		"blockRoot":        fmt.Sprintf("%#x", root),
		"custodiedColumns": d.columnFromPeer[pid],
		"requestedColumns": sortedListFromMap(columns),
		"retrievedColumns": sortedListFromMap(retrieved),
	}).Debug("Peer data column sampling summary")
	return retrieved
}

// randomizeColumns returns a slice containing all columns in a random order.
func randomizeColumns(columns map[uint64]bool) []uint64 {
	// Create a slice from columns.
	randomized := make([]uint64, 0, len(columns))
	for column := range columns {
		randomized = append(randomized, column)
	}

	// Shuffle the slice.
	rand.NewGenerator().Shuffle(len(randomized), func(i, j int) {
		randomized[i], randomized[j] = randomized[j], randomized[i]
	})

	return randomized
}

// sortedListFromMap returns a sorted list of keys from a map.
func sortedListFromMap(m map[uint64]bool) []uint64 {
	result := make([]uint64, 0, len(m))
	for k := range m {
		result = append(result, k)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

// selectRandomPeer returns a random peer from the given list of peers.
func selectRandomPeer(peers map[peer.ID]bool) peer.ID {
	pick := rand.NewGenerator().Uint64() % uint64(len(peers))
	for k := range peers {
		if pick == 0 {
			return k
		}
		pick--
	}

	// This should never be reached.
	return peer.ID("")
}

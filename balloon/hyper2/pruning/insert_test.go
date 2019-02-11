package pruning

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneToInsert(t *testing.T) {

	testCases := []struct {
		index, value  []byte
		cachedBatches map[string][]byte
		storedBatches map[string][]byte
		expectedOp    Operation
	}{
		{
			// insert index = 0 on empty tree
			index:         []byte{0},
			value:         []byte{0},
			cachedBatches: map[string][]byte{},
			storedBatches: map[string][]byte{},
			expectedOp: putBatch(
				inner(pos(0, 8), 0, []byte{0x00, 0x00, 0x00, 0x00},
					inner(pos(0, 7), 1, []byte{0x00, 0x00, 0x00, 0x00},
						inner(pos(0, 6), 3, []byte{0x00, 0x00, 0x00, 0x00},
							inner(pos(0, 5), 7, []byte{0x00, 0x00, 0x00, 0x00},
								leaf(pos(0, 4), 15, []byte{0x00, 0x00, 0x00, 0x00},
									mutate(
										shortcut(pos(0, 4), 0, []byte{0x00, 0x00, 0x00, 0x00},
											[]byte{0}, []byte{0},
										),
										[]byte{0x00, 0x00, 0x00, 0x00},
									),
								),
								getDefault(pos(16, 4)),
							),
							getDefault(pos(32, 5)),
						),
						getDefault(pos(64, 6)),
					),
					getDefault(pos(128, 7)),
				),
				[]byte{0x00, 0x00, 0x00, 0x00},
			),
		},
		{
			// update index = 0 on tree with only one leaf
			index: []byte{0},
			value: []byte{0},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{0xe0, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x02},
			},
			expectedOp: putBatch(
				inner(pos(0, 8), 0, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					inner(pos(0, 7), 1, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						inner(pos(0, 6), 3, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
							inner(pos(0, 5), 7, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
								leaf(pos(0, 4), 15, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
									mutate(
										shortcut(pos(0, 4), 0, []byte{0x00, 0x00, 0x00, 0x00},
											[]byte{0}, []byte{0},
										),
										[]byte{0x00, 0x00, 0x00, 0x00}, // <-- the batch has been reset
									),
								),
								getDefault(pos(16, 4)),
							),
							getDefault(pos(32, 5)),
						),
						getDefault(pos(64, 6)),
					),
					getDefault(pos(128, 7)),
				),
				[]byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			),
		},
		{
			// insert key = 1 on tree with 1 leaf (index: 0, value: 0)
			// it should push down the previous leaf to the last level
			index: []byte{1},
			value: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{0xe0, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x02},
			},
			expectedOp: putBatch(
				inner(pos(0, 8), 0, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					inner(pos(0, 7), 1, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						inner(pos(0, 6), 3, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
							inner(pos(0, 5), 7, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
								leaf(pos(0, 4), 15, []byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
									mutate(
										inner(pos(0, 4), 0, []byte{0x00, 0x00, 0x00, 0x00},
											inner(pos(0, 3), 1, []byte{0x00, 0x00, 0x00, 0x00},
												inner(pos(0, 2), 3, []byte{0x00, 0x00, 0x00, 0x00},
													inner(pos(0, 1), 7, []byte{0x00, 0x00, 0x00, 0x00},
														leaf(pos(0, 0), 15, []byte{0x00, 0x00, 0x00, 0x00},
															mutate(
																shortcut(pos(0, 0), 0, []byte{0x00, 0x00, 0x00, 0x00},
																	[]byte{0}, []byte{0},
																),
																[]byte{0x00, 0x00, 0x00, 0x00}, // new batch
															),
														),
														leaf(pos(1, 0), 16, []byte{0x00, 0x00, 0x00, 0x00},
															mutate(
																shortcut(pos(1, 0), 0, []byte{0x00, 0x00, 0x00, 0x00},
																	[]byte{1}, []byte{1},
																),
																[]byte{0x00, 0x00, 0x00, 0x00}, // new batch
															),
														),
													),
													getDefault(pos(2, 1)),
												),
												getDefault(pos(4, 2)),
											),
											getDefault(pos(8, 3)),
										),
										[]byte{0x00, 0x00, 0x00, 0x00}, // reset previous shortcut
									),
								),
								getDefault(pos(16, 4)),
							),
							getDefault(pos(32, 5)),
						),
						getDefault(pos(64, 6)),
					),
					getDefault(pos(128, 7)),
				),
				[]byte{0xd1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // does not change
			),
		},
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4

	for i, c := range testCases {
		loader := NewFakeBatchLoader(c.cachedBatches, c.storedBatches, cacheHeightLimit)
		prunedOp, err := PruneToInsert(c.index, c.value, cacheHeightLimit, loader)
		require.NoError(t, err)
		assert.Equalf(t, c.expectedOp, prunedOp, "The pruned operation should match for test case %d", i)
	}
}

// // insert Key = 1 after 0

// PutBatchOp(
// 	InnerHashOp(pos(0, 8), 0,
// 		InnerHashOp(pos(0, 7), 1,
// 			InnerHashOp(pos(0, 6), 3
// 				InnerHashOp(pos(0, 5), 7,
// 					LeafOp(pos(0, 4), 15,
// 						MutateBatchOp(
// 							InnerHashOp(pos(0, 4), 0),
// 								InnerHashOp(pos(0, 3), 1,
// 									InnerHashOp(pos(0, 2), 3,
// 										InnerHashOp(pos(0, 1), 7,
// 											LeafOp(pos(0, 0), 15,
// 												MutateBatchOp(
// 													ShortcutLeafOp(pos(0, 4), 0, Key[0], Value),
// 													batch,
// 												),
// 											),
// 											LeafOp(pos(1, 0), 15,
// 												MutateBatchOp(
// 													ShortcutLeafOp(pos(0, 4), 0, Key[1], Value),
// 													batch,
// 												),
// 											),
// 										),
// 										GetDefaultOp(pos(2, 1)),
// 									),
// 									GetDefaultOp(pos(4, 2)),
// 								),
// 								GetDefaultOp(pos(8, 3)),
// 							),
// 							batch,
// 						),
// 					),
// 					GetDefaultOp(pos(16, 4)),
// 				),
// 				GetDefaultOp(pos(32, 5)),
// 			),
// 			GetDefaultOp(pos(64, 6)),
// 		),
// 		GetDefaultOp(pos(128, 7)),
// 	),
// 	batch,
// )

// // insert Key = 0 after nil

// PutBatchOp(
// 	InnerHashOp(pos(0, 8), 0,
// 		InnerHashOp(pos(0, 7), 1,
// 			InnerHashOp(pos(0, 6), 3
// 				InnerHashOp(pos(0, 5), 7,
// 					LeafOp(pos(0, 4), 15,
// 						MutateBatchOp(
// 							ShortcutLeafOp(pos(0, 4), 0, Key, Value),
// 							batch,
// 						),
// 					),
// 					GetDefaultOp(pos(16, 4)),
// 				),
// 				GetDefaultOp(pos(32, 5)),
// 			),
// 			GetDefaultOp(pos(64, 6)),
// 		),
// 		GetDefaultOp(pos(128, 7)),
// 	),
// 	batch,
// )

// // insert Key = 8 after 0

// PutBatchOp(
// 	InnerHashOp(pos(0, 8), 0,
// 		InnerHashOp(pos(0, 7), 1,
// 			InnerHashOp(pos(0, 6), 3
// 				InnerHashOp(pos(0, 5), 7,
// 					LeafOp(pos(0, 4), 15,
// 						MutateBatchOp(
// 							InnerHashOp(pos(0, 4), 0),
// 								ShortcutLeafOp(pos(0, 3), 1, Key[0], Value), // pushed down
// 								ShortcutLeafOp(pos(8, 3), 2, Key[8], Value), // new
// 							),
// 							batch,
// 						),
// 					),
// 					GetDefaultOp(pos(16, 4)),
// 				),
// 				GetDefaultOp(pos(32, 5)),
// 			),
// 			GetDefaultOp(pos(64, 6)),
// 		),
// 		GetDefaultOp(pos(128, 7)),
// 	),
// 	batch,
// )

// // insert Key = 12 after 0 and 8

// PutBatchOp(
// 	InnerHashOp(pos(0, 8), 0,
// 		InnerHashOp(pos(0, 7), 1,
// 			InnerHashOp(pos(0, 6), 3
// 				InnerHashOp(pos(0, 5), 7,
// 					LeafOp(pos(0, 4), 15,
// 						MutateBatchOp(
// 							InnerHashOp(pos(0, 4), 0),
// 								UseProvidedOp(pos(0, 3), 1), // provided
// 								InnerHashOp(pos(8, 3), 2,
// 									ShortcutLeafOp(pos(8, 2), 5, Key[8], Value), // pushed down
// 									ShortcutLeafOp(pos(12, 2), 6, Key[12], Value),
// 								),
// 							),
// 							batch,
// 						),
// 					),
// 					GetDefaultOp(pos(16, 4)),
// 				),
// 				GetDefaultOp(pos(32, 5)),
// 			),
// 			GetDefaultOp(pos(64, 6)),
// 		),
// 		GetDefaultOp(pos(128, 7)),
// 	),
// 	batch,
// )

// // insert Key = 1 after 0

// PutBatchOp(
// 	InnerHashOp(pos(0, 8), 0,
// 		InnerHashOp(pos(0, 7), 1,
// 			InnerHashOp(pos(0, 6), 3
// 				InnerHashOp(pos(0, 5), 7,
// 					LeafOp(pos(0, 4), 15,
// 						MutateBatchOp(
// 							InnerHashOp(pos(0, 4), 0),
// 								InnerHashOp(pos(0, 3), 1,
// 									InnerHashOp(pos(0, 2), 3,
// 										InnerHashOp(pos(0, 1), 7,
// 											LeafOp(pos(0, 0), 15,
// 												MutateBatchOp(
// 													ShortcutLeafOp(pos(0, 4), 0, Key[0], Value),
// 													batch,
// 												),
// 											),
// 											LeafOp(pos(1, 0), 15,
// 												MutateBatchOp(
// 													ShortcutLeafOp(pos(0, 4), 0, Key[1], Value),
// 													batch,
// 												),
// 											),
// 										),
// 										GetDefaultOp(pos(2, 1)),
// 									),
// 									GetDefaultOp(pos(4, 2)),
// 								),
// 								GetDefaultOp(pos(8, 3)),
// 							),
// 							batch,
// 						),
// 					),
// 					GetDefaultOp(pos(16, 4)),
// 				),
// 				GetDefaultOp(pos(32, 5)),
// 			),
// 			GetDefaultOp(pos(64, 6)),
// 		),
// 		GetDefaultOp(pos(128, 7)),
// 	),
// 	batch,
// )

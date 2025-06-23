package params_test

import (
	"math"
	"testing"
	"time"

	"github.com/OffchainLabs/prysm/v6/config/params"
	"github.com/OffchainLabs/prysm/v6/consensus-types/primitives"
	"github.com/OffchainLabs/prysm/v6/testing/require"
	"github.com/OffchainLabs/prysm/v6/time/slots"
)

func TestSlotTimeSchedule_CurrentSlot(t *testing.T) {
	slotsPerEpoch := params.BeaconConfig().SlotsPerEpoch
	tests := []struct {
		name string
		// Inputs
		sch     params.SlotTimeSchedule
		genesis time.Time
		// Want
		slot primitives.Slot
	}{
		{
			name: "single entry",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				},
			},
			genesis: time.Now().Add(-1 * 33 * time.Duration(slotsPerEpoch) * 12 * time.Second), // Genesis was 33 epochs ago.
			slot:    slots.UnsafeEpochStart(33),
		},
		{
			name: "multiple entries",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        32,
					SlotDuration: 10 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			genesis: func() time.Time {
				tm := time.Now()
				firstEpochDuration := 32 * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 12 * time.Second
				tm = tm.Add(-1 * firstEpochDuration)
				oneEpochDuration := time.Duration(params.BeaconConfig().SlotsPerEpoch) * 10 * time.Second
				tm = tm.Add(-1 * oneEpochDuration)

				return tm
			}(),
			slot: slots.UnsafeEpochStart(33),
		},
		{
			name: "multiple entries, last entry",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        32,
					SlotDuration: 10 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			genesis: func() time.Time {
				tm := time.Now()
				firstEpochDuration := 32 * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 12 * time.Second
				tm = tm.Add(-1 * firstEpochDuration)
				secondEpochDuration := 32 * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 10 * time.Second
				tm = tm.Add(-1 * secondEpochDuration)
				remaining := (100 - 64) * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 1 * time.Second
				tm = tm.Add(-1 * remaining)

				return tm
			}(),
			slot: slots.UnsafeEpochStart(100),
		},
		// TODO: Unsorted.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.slot, tt.sch.CurrentSlot(tt.genesis))
		})
	}
}

func TestSlotTimeSchedule_SinceGenesis(t *testing.T) {
	tests := []struct {
		name string
		// Inputs
		sch  params.SlotTimeSchedule
		slot primitives.Slot
		// Want
		since time.Duration
		error bool
	}{
		{
			name: "single entry",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				},
			},
			slot:  slots.UnsafeEpochStart(33),
			since: 12 * time.Second * 33 * time.Duration(params.BeaconConfig().SlotsPerEpoch),
		},
		{
			name: "single entry 1s",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: time.Second,
				},
			},
			slot:  16,
			since: 16 * time.Second,
		},
		{
			name: "multiple entries",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        32,
					SlotDuration: 10 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			slot: slots.UnsafeEpochStart(33),
			since: func() time.Duration {
				firstEpochDuration := 32 * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 12 * time.Second
				oneEpochDuration := time.Duration(params.BeaconConfig().SlotsPerEpoch) * 10 * time.Second
				return firstEpochDuration + oneEpochDuration
			}(),
		},
		{
			name: "multiple entries, last entry",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        32,
					SlotDuration: 10 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			slot: slots.UnsafeEpochStart(100),
			since: func() time.Duration {
				firstEntryDuration := 32 * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 12 * time.Second
				secondEntryDuration := 32 * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 10 * time.Second
				remaining := (100 - 64) * time.Duration(params.BeaconConfig().SlotsPerEpoch) * 1 * time.Second
				return firstEntryDuration + secondEntryDuration + remaining
			}(),
		},
		{
			name: "overflow",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			slot:  math.MaxUint64,
			error: true,
		},
		// TODO: Unsorted.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sch.SinceGenesis(tt.slot)
			if !tt.error {
				require.NoError(t, err)
			} else {
				require.Equal(t, true, err != nil, "did not get any error when one was expected")
			}
			require.Equal(t, tt.since, got)
		})
	}
}

func TestSlotTimeSchedule_SlotDuration(t *testing.T) {
	tests := []struct {
		name string
		// Inputs
		sch  params.SlotTimeSchedule
		slot primitives.Slot
		// Want
		duration time.Duration
	}{
		{
			name: "single entry",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				},
			},
			slot:     slots.UnsafeEpochStart(33),
			duration: 12 * time.Second,
		},
		{
			name: "multiple entries",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        32,
					SlotDuration: 10 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			slot:     slots.UnsafeEpochStart(33),
			duration: 10 * time.Second,
		},
		{
			name: "multiple entries, last entry",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        32,
					SlotDuration: 10 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			slot:     slots.UnsafeEpochStart(100),
			duration: 1 * time.Second,
		},
		{
			name: "multiple entries, first entry",
			sch: params.SlotTimeSchedule{
				{
					Epoch:        0,
					SlotDuration: 12 * time.Second,
				}, {
					Epoch:        32,
					SlotDuration: 10 * time.Second,
				}, {
					Epoch:        64,
					SlotDuration: 1 * time.Second,
				},
			},
			slot:     slots.UnsafeEpochStart(3),
			duration: 12 * time.Second,
		},
		// TODO: Unsorted.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sch.SlotDuration(tt.slot)
			require.Equal(t, tt.duration, got)
		})
	}
}

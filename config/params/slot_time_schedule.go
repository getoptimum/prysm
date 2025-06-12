package params

import (
	"errors"
	"fmt"
	"time"

	"github.com/OffchainLabs/prysm/v6/consensus-types/primitives"
)

var ErrInvalidSlotTimeScheduleNoGenesis = errors.New("invalid slot time schedule, missing an entry for epoch 0")

type SlotTimeSchedule []SlotTimeScheduleEntry

// SlotTimeScheduleEntry defines a schedule entry which adjusts the seconds per slot starting at
// the given epoch. This works similarly to the blob schedule.
type SlotTimeScheduleEntry struct {
	Epoch        primitives.Epoch
	SlotDuration time.Duration
}

// IsValid ensures that there is at least one entry with epoch 0 and that all entries have an epoch
// with a value less than MaxSafeEpoch. It also ensures that every duration is at least 1 second.
func (s SlotTimeSchedule) IsValid() error {
	return errors.New("not implemented")
}

func (s SlotTimeSchedule) CurrentSlot(genesis time.Time) primitives.Slot {
	s.sort()

	now := time.Now()
	// This is the non-optimized routine. It could be better by precomputing the start time of each epoch entry.
	d := s[0].SlotDuration
	if len(s) == 1 {
		return primitives.Slot(now.Sub(genesis) / d)
	}

	remaining := now.Sub(genesis)
	for i, e := range s {
		// TODO: iterate through the slot times to find the solution.

		// Is this the last bucket? If so, return the result.
		if i == len(s)-1 {
			return epochStart(e.Epoch) + primitives.Slot(remaining/e.SlotDuration)
		}
		// Does remaining fit in the current bucket?
		// fits = s[i+1].Epoch.Sub(uint64(e.Epoch)) * BeaconChain().SlotsPerEpoch * e.SlotDuration < remaining
		wholeEntryDuration := time.Duration(s[i+1].Epoch.Sub(uint64(e.Epoch))) * time.Duration(BeaconConfig().SlotsPerEpoch) * e.SlotDuration
		// Yes -> return StartSlot(e.Epoch) + remaining / e.SlotDuration.
		if remaining < wholeEntryDuration {
			return epochStart(e.Epoch) + primitives.Slot(remaining/e.SlotDuration)
		}
		// No -> remove the full bucket period from remaining.
		remaining -= wholeEntryDuration
	}

	return 0 // This should never happen. Maybe even panic? It's ensured safe by IsValid().
}

// CurrentSlotDuration returns the slot duration given the current slot on the schedule.
func (s SlotTimeSchedule) CurrentSlotDuration(genesis time.Time) time.Duration {
	return s.SlotDuration(s.CurrentSlot(genesis))
}

// SinceGenesis will return the amount of time since genesis for a given slot. May return an error
// when the slot value would cause an overflow or underflow.
func (s SlotTimeSchedule) SinceGenesis(slot primitives.Slot) (time.Duration, error) {
	s.sort()

	var tm time.Duration
	for i, e := range s {
		if i == len(s)-1 || epochStart(s[i+1].Epoch) > slot {
			delta, err := slot.SafeSub(uint64(epochStart(e.Epoch)))
			if err != nil {
				return 0, fmt.Errorf("failed to compute the number of slots into the epoch: %w", err)
			}
			return tm + (time.Duration(delta) * e.SlotDuration), nil
		}
		delta, err := s[i+1].Epoch.SafeSub(uint64(e.Epoch))
		if err != nil {
			return 0, fmt.Errorf("failed to compute the number of slots in a SlotTimeSchedule entry: %w", err)
		}
		tm += (time.Duration(primitives.Slot(delta)*BeaconConfig().SlotsPerEpoch) * e.SlotDuration)
	}

	return 0, errors.New("not implemented")
}

// This is a copy from slots.EpochStart, but avoids the circular dependency.
func epochStart(e primitives.Epoch) primitives.Slot {
	return primitives.Slot(e) * BeaconConfig().SlotsPerEpoch
}

func (s SlotTimeSchedule) sort() {
	// TODO: How to ensure the list is sorted at least once and remains sorted?
}

// SlotDuration returns the amount of time in a given slot. For example, 12 seconds per slot for
// Ethereum's original slot duration.
func (s SlotTimeSchedule) SlotDuration(slot primitives.Slot) time.Duration {
	s.sort()

	for i := len(s) - 1; i >= 0; i-- {
		if BeaconConfig().SlotsPerEpoch.Mul(uint64(s[i].Epoch)) >= slot {
			return s[i].SlotDuration
		}
	}
	return 0 // Maybe this should be an error, but handling an error on this would be really annoying.
}

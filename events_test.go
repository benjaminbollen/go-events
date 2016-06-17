package events

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestStartEventSwitch(t *testing.T) {
	evsw := NewEventSwitch()
	// calls over tendermint/go-common.BaseService
	started, err := evsw.Start()
	if started != true {
		t.Fatalf("Failed to start EventSwitch, error: %v", err)
	}
	if err != nil {
		t.Fatalf("Started with error: %v", err)
	}
}

func TestAddListenerForEvent(t *testing.T) {
	evsw := NewEventSwitch()
	started, err := evsw.Start()
	if started == false || err != nil {
		t.Errorf("Failed to start EventSwitch, error: %v", err)
	}
	messages := make(chan EventData)
	evsw.AddListenerForEvent("listener", "event",
		func (data EventData) {
			messages <- data
		})
	go evsw.FireEvent("event", "data")
	received := <-messages
	if received != "data" {
		t.Errorf("Message received does not match: %v", received)
	}
}

func TestAddListenerForMultipleEvents(t *testing.T) {
	evsw := NewEventSwitch()
	started, err := evsw.Start()
	if started == false || err != nil {
		t.Errorf("Failed to start EventSwitch, error: %v", err)
	}
	doneSum := make(chan uint64)
	doneSending := make(chan uint64)
	numbers := make(chan uint64, 4)
	// subscribe one listener for one event
	evsw.AddListenerForEvent("listener", "event",
		func (data EventData) {
			numbers <- data.(uint64)
		})
	// collect received events
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers
			sum += j
			if !more {
				doneSum <- sum
				close(doneSum)
				return
			}
		}
	}()
	// go fire events
	go func() {
		var sentsum uint64 = 0
		for i := uint64(1); i<= uint64(1000); i++ {
			sentsum += i
			evsw.FireEvent("event", i)
		}
		close(numbers)
		doneSending <- sentsum
		close(doneSending)
	}()
	checksum := <-doneSending
	eventsum := <-doneSum
  if checksum != eventsum {
		t.Errorf("Not all messages sent were received.\n")
	}
}

func TestAddListenerForDifferentEvents(t *testing.T) {
	evsw := NewEventSwitch()
	started, err := evsw.Start()
	if started == false || err != nil {
		t.Errorf("Failed to start EventSwitch, error: %v", err)
	}
	doneSum := make(chan uint64)
	doneSending1 := make(chan uint64)
	doneSending2 := make(chan uint64)
	doneSending3 := make(chan uint64)
	numbers := make(chan uint64, 4)
	// subscribe one listener to three events
	evsw.AddListenerForEvent("listener", "event1",
		func (data EventData) {
			numbers <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener", "event2",
		func (data EventData) {
			numbers <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener", "event3",
		func (data EventData) {
			numbers <- data.(uint64)
		})
	// collect received events
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers
			sum += j
			if !more {
				doneSum <- sum
				close(doneSum)
				return
			}
		}
	}()
	// go fire events
	sendEvents := func(event string, doneChan chan uint64) {
		var sentsum uint64 = 0
		for i := uint64(1); i<= uint64(1000) ; i++ {
			sentsum += i
			evsw.FireEvent(event, i)
		}
		doneChan <- sentsum
		close(doneChan)
		return
	}
	go sendEvents("event1", doneSending1)
	go sendEvents("event2", doneSending2)
	go sendEvents("event3", doneSending3)
	var checksum uint64 = 0
	checksum += <-doneSending1
	checksum += <-doneSending2
	checksum += <-doneSending3
	close(numbers)
	eventsum := <-doneSum
  if checksum != eventsum {
		t.Errorf("Not all messages sent were received.\n")
	}
}

func TestAddDifferentListenerForDifferentEvents(t *testing.T) {
	evsw := NewEventSwitch()
	started, err := evsw.Start()
	if started == false || err != nil {
		t.Errorf("Failed to start EventSwitch, error: %v", err)
	}
	doneSum1 := make(chan uint64)
	doneSum2 := make(chan uint64)
	doneSending1 := make(chan uint64)
	doneSending2 := make(chan uint64)
	doneSending3 := make(chan uint64)
	numbers1 := make(chan uint64, 4)
	numbers2 := make(chan uint64, 4)
	// subscribe two listener to three events
	evsw.AddListenerForEvent("listener1", "event1",
		func (data EventData) {
			numbers1 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener1", "event2",
		func (data EventData) {
			numbers1 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener1", "event3",
		func (data EventData) {
			numbers1 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener2", "event2",
		func (data EventData) {
			numbers2 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener2", "event3",
		func (data EventData) {
			numbers2 <- data.(uint64)
		})
	// collect received events for listener1
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers1
			sum += j
			if !more {
				doneSum1 <- sum
				close(doneSum1)
				return
			}
		}
	}()
	// collect received events for listener2
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers2
			sum += j
			if !more {
				doneSum2 <- sum
				close(doneSum2)
				return
			}
		}
	}()

	// go fire events
	sendEvents := func(event string, doneChan chan uint64, offset uint64) {
		var sentsum uint64 = 0
		for i := offset; i<= offset + uint64(999) ; i++ {
			sentsum += i
			evsw.FireEvent(event, i)
		}
		doneChan <- sentsum
		close(doneChan)
		return
	}
	go sendEvents("event1", doneSending1, uint64(1))
	go sendEvents("event2", doneSending2, uint64(1001))
	go sendEvents("event3", doneSending3, uint64(2001))
	checksumevent1 := <-doneSending1
	checksumevent2 := <-doneSending2
	checksumevent3 := <-doneSending3
	checksum1 := checksumevent1 + checksumevent2 + checksumevent3
	checksum2 := checksumevent2 + checksumevent3
	close(numbers1)
	close(numbers2)
	eventsum1 := <-doneSum1
	eventsum2 := <-doneSum2
  if checksum1 != eventsum1 ||
	 	checksum2 != eventsum2 {
		t.Errorf("Not all messages sent were received for different listeners to different events.\n")
	}
}

func TestAddAndRemoveListenerForEvents(t *testing.T) {
	evsw := NewEventSwitch()
	started, err := evsw.Start()
	if started == false || err != nil {
		t.Errorf("Failed to start EventSwitch, error: %v", err)
	}
	doneSum1 := make(chan uint64)
	doneSum2 := make(chan uint64)
	doneSending1 := make(chan uint64)
	doneSending2 := make(chan uint64)
	numbers1 := make(chan uint64, 4)
	numbers2 := make(chan uint64, 4)
	// subscribe two listener to three events
	evsw.AddListenerForEvent("listener", "event1",
		func (data EventData) {
			numbers1 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener", "event2",
		func (data EventData) {
			numbers2 <- data.(uint64)
		})
	// collect received events for event1
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers1
			sum += j
			if !more {
				doneSum1 <- sum
				close(doneSum1)
				return
			}
		}
	}()
	// collect received events for event2
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers2
			sum += j
			if !more {
				doneSum2 <- sum
				close(doneSum2)
				return
			}
		}
	}()
	// go fire events
	sendEvents := func(event string, doneChan chan uint64, offset uint64) {
		var sentsum uint64 = 0
		for i := offset; i<= offset + uint64(999) ; i++ {
			sentsum += i
			evsw.FireEvent(event, i)
		}
		doneChan <- sentsum
		close(doneChan)
		return
	}
	go sendEvents("event1", doneSending1, uint64(1))
	checksumevent1 := <-doneSending1
	// after sending all event1, unsubscribe for all events
	evsw.RemoveListener("listener")
	go sendEvents("event2", doneSending2, uint64(1001))
	checksumevent2 := <-doneSending2
	close(numbers1)
	close(numbers2)
	eventsum1 := <-doneSum1
	eventsum2 := <-doneSum2
  if checksumevent1 != eventsum1 ||
	 	checksumevent2 != uint64(1500500) ||
		eventsum2 != uint64(0) {
		t.Errorf("Not all messages sent were received or unsubscription did not register.\n")
	}
}

func TestRemoveListenersAsync(t *testing.T) {
	evsw := NewEventSwitch()
	started, err := evsw.Start()
	if started == false || err != nil {
		t.Errorf("Failed to start EventSwitch, error: %v", err)
	}
	doneSum1 := make(chan uint64)
	doneSum2 := make(chan uint64)
	doneSending1 := make(chan uint64)
	doneSending2 := make(chan uint64)
	doneSending3 := make(chan uint64)
	numbers1 := make(chan uint64, 4)
	numbers2 := make(chan uint64, 4)
	// subscribe two listener to three events
	evsw.AddListenerForEvent("listener1", "event1",
		func (data EventData) {
			numbers1 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener1", "event2",
		func (data EventData) {
			numbers1 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener1", "event3",
		func (data EventData) {
			numbers1 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener2", "event1",
		func (data EventData) {
			numbers2 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener2", "event2",
		func (data EventData) {
			numbers2 <- data.(uint64)
		})
	evsw.AddListenerForEvent("listener2", "event3",
		func (data EventData) {
			numbers2 <- data.(uint64)
		})
	// collect received events for event1
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers1
			sum += j
			if !more {
				doneSum1 <- sum
				close(doneSum1)
				return
			}
		}
	}()
	// collect received events for event2
	go func() {
		var sum uint64 = 0
		for {
			j, more := <-numbers2
			sum += j
			if !more {
				doneSum2 <- sum
				close(doneSum2)
				return
			}
		}
	}()
	// go fire events
	sendEvents := func(event string, doneChan chan uint64, offset uint64) {
		var sentsum uint64 = 0
		for i := offset; i<= offset + uint64(999) ; i++ {
			sentsum += i
			evsw.FireEvent(event, i)
		}
		doneChan <- sentsum
		close(doneChan)
		return
	}
	addListenersStress := func() {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		for k := uint16(0); k < 400; k++ {
			listenerNumber := r1.Intn(100) + 3
			eventNumber := r1.Intn(3) + 1
			go evsw.AddListenerForEvent(fmt.Sprintf("listener%v", listenerNumber),
				fmt.Sprintf("event%v", eventNumber),
				func (_ EventData) {})
		}
	}
	removeListenersStress := func() {
		s2 := rand.NewSource(time.Now().UnixNano())
		r2 := rand.New(s2)
		for k := uint16(0); k < 80; k++ {
			listenerNumber := r2.Intn(100) + 3
			go evsw.RemoveListener(fmt.Sprintf("listener%v", listenerNumber))
		}
	}
	go addListenersStress()
	go sendEvents("event1", doneSending1, uint64(1))
	go removeListenersStress()
	go sendEvents("event2", doneSending2, uint64(1001))
	go sendEvents("event3", doneSending3, uint64(2001))
	checksumevent1 := <-doneSending1
	checksumevent2 := <-doneSending2
	checksumevent3 := <-doneSending3
	checksum1 := checksumevent1 + checksumevent2 + checksumevent3
	checksum2 := checksumevent1 + checksumevent2 + checksumevent3
	close(numbers1)
	close(numbers2)
	eventsum1 := <-doneSum1
	eventsum2 := <-doneSum2
  if checksum1 != eventsum1 ||
	 	checksum2 != eventsum2 {
		t.Errorf("Not all messages sent were received.\n")
	}
}

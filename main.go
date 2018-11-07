package main

import (
	"fmt"
	"go-event-sourcing-sample/pkg/eventsourcing"
)

func main() {

	aggregateID := "123"

	history := []eventsourcing.Event{
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: "FrequentFlierAccount", Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 2, Reason: "StatusMatched", AggregateType: "FrequentFlierAccount", Data: StatusMatched{NewStatus: StatusSilver}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 3, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 4, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 5, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 6, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}
	fmt.Println(history)
	aggregate := NewFrequentFlierAccountFromHistory(history)
	fmt.Println("Before RecordFlightTaken")
	fmt.Println(aggregate)

	aggregate.RecordFlightTaken(1000, 3)
	fmt.Println("After RecordFlightTaken")
	fmt.Println(aggregate)

	newAggregate := CreateFrequentFlierAccount("666")
	newAggregate.RecordFlightTaken(20, 1233)
	fmt.Println(newAggregate)

	fmt.Println(newAggregate.aggregateRoot.Changes)
	fmt.Println("-----")

	copyAggregate := NewFrequentFlierAccountFromHistory(newAggregate.aggregateRoot.Changes)
	fmt.Println(copyAggregate.status)
	fmt.Println(copyAggregate.aggregateRoot.Changes)

}

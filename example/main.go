package main

import (
	"eventsourcing"
	"eventsourcing/eventstore/memory"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

func main() {

	aggregateID := eventsourcing.AggregateRootID("123")

	history := []eventsourcing.Event{
		{AggregateRootID: aggregateID, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: "FrequentFlierAccount", Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}},
		{AggregateRootID: aggregateID, Version: 2, Reason: "StatusMatched", AggregateType: "FrequentFlierAccount", Data: StatusMatched{NewStatus: StatusSilver}},
		{AggregateRootID: aggregateID, Version: 3, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}},
		{AggregateRootID: aggregateID, Version: 4, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}},
		{AggregateRootID: aggregateID, Version: 5, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		{AggregateRootID: aggregateID, Version: 6, Reason: "FlightTaken", AggregateType: "FrequentFlierAccount", Data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}
	fmt.Println(history)
	aggregate := NewFrequentFlierAccountFromHistory(history)
	fmt.Println("Before RecordFlightTaken")
	fmt.Println(aggregate)

	aggregate = CreateFrequentFlierAccount("morgan")
	aggregate.RecordFlightTaken(10, 5)

	eventStore := memory.Create()
	repo := eventsourcing.NewRepository(eventStore)
	repo.Save(aggregate)

	a := FrequentFlierAccountAggregate{}
	repo.Get(aggregate.ID(), &a)
	spew.Dump(a)

	aggregate.RecordFlightTaken(1000, 3)
	fmt.Println("After RecordFlightTaken")
	fmt.Println(aggregate)

	newAggregate := CreateFrequentFlierAccount("666")
	newAggregate.RecordFlightTaken(20, 1233)
	fmt.Println(newAggregate)

	fmt.Println(newAggregate.Changes())
	fmt.Println("-----")

	copyAggregate := NewFrequentFlierAccountFromHistory(newAggregate.Changes())
	fmt.Println(copyAggregate.status)
	fmt.Println(copyAggregate.Changes())

}

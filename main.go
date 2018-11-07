package main

import (
	"fmt"
)

func main() {

	aggregateID := "123"

	history := []Event{
		Event{aggregateRootID: aggregateID, version: 1, reason: "FrequentFlierAccountCreated", aggregateType: "FrequentFlierAccount", data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}},
		Event{aggregateRootID: aggregateID, version: 2, reason: "StatusMatched", aggregateType: "FrequentFlierAccount", data: StatusMatched{NewStatus: StatusSilver}},
		Event{aggregateRootID: aggregateID, version: 3, reason: "FlightTaken", aggregateType: "FrequentFlierAccount", data: FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}},
		Event{aggregateRootID: aggregateID, version: 4, reason: "FlightTaken", aggregateType: "FrequentFlierAccount", data: FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}},
		Event{aggregateRootID: aggregateID, version: 5, reason: "FlightTaken", aggregateType: "FrequentFlierAccount", data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		Event{aggregateRootID: aggregateID, version: 6, reason: "FlightTaken", aggregateType: "FrequentFlierAccount", data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
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

	fmt.Println(newAggregate.aggregateRoot.changes)
	fmt.Println("-----")

	copyAggregate := NewFrequentFlierAccountFromHistory(newAggregate.aggregateRoot.changes)
	fmt.Println(copyAggregate.status)
	fmt.Println(copyAggregate.aggregateRoot.changes)

}

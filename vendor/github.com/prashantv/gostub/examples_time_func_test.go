package gostub_test

import (
	"fmt"
	"time"

	"github.com/prashantv/gostub"
)

// Production code
var timeNow = time.Now

func GetDay() int {
	return timeNow().Day()
}

// Test code
func Example_stubTimeWithFunction() {
	var day = 2
	stubs := gostub.Stub(&timeNow, func() time.Time {
		return time.Date(2015, 07, day, 0, 0, 0, 0, time.UTC)
	})
	defer stubs.Reset()

	firstDay := GetDay()

	day = 3
	secondDay := GetDay()

	fmt.Printf("First day: %v, second day: %v\n", firstDay, secondDay)
	// Output:
	// First day: 2, second day: 3
}

// Test code
func Example_stubTimeWithConstant() {
	stubs := gostub.StubFunc(&timeNow, time.Date(2015, 07, 2, 0, 0, 0, 0, time.UTC))
	defer stubs.Reset()

	fmt.Println("Day:", GetDay())
	// Output:
	// Day: 2
}

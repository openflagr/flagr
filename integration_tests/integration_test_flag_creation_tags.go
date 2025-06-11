package main

import (
	"fmt"
	"log"

	"github.com/Allen-Career-Institute/flagr/pkg/entity"
	"github.com/Allen-Career-Institute/flagr/pkg/handler"
	"github.com/Allen-Career-Institute/flagr/pkg/util"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/models"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/restapi/operations/flag"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/restapi/operations/latch"
)

// Integration test to verify that flag creation correctly includes tags in the response
// This test demonstrates the fix for the issue where AB experiments and latches
// were not returning their associated tags in the creation response.
func main() {
	fmt.Println("🧪 Integration Test: Flag Creation with Tags")
	fmt.Println("============================================================")

	// Initialize test environment
	setupTestEnvironment()

	// Run test scenarios
	testABExperimentCreation()
	testABExperimentWithTemplate()
	testLatchCreation()

	fmt.Println("\n🎉 All integration tests passed!")
	fmt.Println("Flag creation now correctly includes tags in the response.")
}

func setupTestEnvironment() {
	fmt.Println("\n📋 Setting up test environment...")

	// Create a test database
	_ = entity.NewTestDB()
	fmt.Println("✅ Test database initialized")

	fmt.Println("✅ Test environment ready")
}

func testABExperimentCreation() {
	fmt.Println("\n🧪 Test 1: AB Experiment Flag Creation")
	fmt.Println("----------------------------------------")

	// Create CRUD handler
	c := handler.NewCRUD()

	// Create AB Experiment Flag
	fmt.Println("Creating AB experiment flag...")
	res := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("Integration Test AB Experiment"),
			Key:         "integration_test_ab_experiment",
		},
	})

	if flagOK, ok := res.(*flag.CreateFlagOK); ok {
		fmt.Printf("✅ AB Flag created successfully\n")
		fmt.Printf("   ID: %d\n", flagOK.Payload.ID)
		fmt.Printf("   Description: %s\n", *flagOK.Payload.Description)
		fmt.Printf("   Key: %s\n", flagOK.Payload.Key)

		// Verify tags are included
		if flagOK.Payload.Tags != nil && len(flagOK.Payload.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", *flagOK.Payload.Tags[0].Value)
			if *flagOK.Payload.Tags[0].Value == "AB" {
				fmt.Println("   ✅ AB tag correctly included in response!")
			} else {
				log.Fatalf("   ❌ Expected 'AB' tag, got '%s'", *flagOK.Payload.Tags[0].Value)
			}
		} else {
			log.Fatalf("   ❌ No tags found in response!")
		}

		// Verify basic flag properties
		if flagOK.Payload.Enabled != nil && *flagOK.Payload.Enabled == false {
			fmt.Println("   ✅ Flag created with default enabled=false")
		}

		if flagOK.Payload.DataRecordsEnabled != nil && *flagOK.Payload.DataRecordsEnabled == false {
			fmt.Println("   ✅ Flag created with default dataRecordsEnabled=false")
		}

	} else {
		log.Fatalf("❌ Failed to create AB flag: %+v", res)
	}
}

func testABExperimentWithTemplate() {
	fmt.Println("\n🧪 Test 2: AB Experiment Flag with Template")
	fmt.Println("----------------------------------------")

	// Create CRUD handler
	c := handler.NewCRUD()

	// Create AB Experiment Flag with Template
	fmt.Println("Creating AB experiment flag with simple_boolean_flag template...")
	res := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("Integration Test AB Experiment with Template"),
			Key:         "integration_test_ab_experiment_template",
			Template:    "simple_boolean_flag",
		},
	})

	if flagOK, ok := res.(*flag.CreateFlagOK); ok {
		fmt.Printf("✅ AB Flag with template created successfully\n")
		fmt.Printf("   ID: %d\n", flagOK.Payload.ID)
		fmt.Printf("   Description: %s\n", *flagOK.Payload.Description)
		fmt.Printf("   Key: %s\n", flagOK.Payload.Key)
		fmt.Printf("   Template: simple_boolean_flag\n")

		// Verify tags are included
		if flagOK.Payload.Tags != nil && len(flagOK.Payload.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", *flagOK.Payload.Tags[0].Value)
			if *flagOK.Payload.Tags[0].Value == "AB" {
				fmt.Println("   ✅ AB tag correctly included in response!")
			} else {
				log.Fatalf("   ❌ Expected 'AB' tag, got '%s'", *flagOK.Payload.Tags[0].Value)
			}
		} else {
			log.Fatalf("   ❌ No tags found in response!")
		}

		// Verify template-created segments
		if flagOK.Payload.Segments != nil && len(flagOK.Payload.Segments) > 0 {
			fmt.Printf("   Segments: %d\n", len(flagOK.Payload.Segments))
			segment := flagOK.Payload.Segments[0]
			if segment.RolloutPercent != nil && *segment.RolloutPercent == 100 {
				fmt.Println("   ✅ Template segment created with 100% rollout")
			}
		} else {
			log.Fatalf("   ❌ No segments found in response!")
		}

		// Verify template-created variants
		if flagOK.Payload.Variants != nil && len(flagOK.Payload.Variants) > 0 {
			fmt.Printf("   Variants: %d\n", len(flagOK.Payload.Variants))
			variant := flagOK.Payload.Variants[0]
			if variant.Key != nil && *variant.Key == "on" {
				fmt.Printf("   ✅ Template variant created with key: %s\n", *variant.Key)
			}
		} else {
			log.Fatalf("   ❌ No variants found in response!")
		}

	} else {
		log.Fatalf("❌ Failed to create AB flag with template: %+v", res)
	}
}

func testLatchCreation() {
	fmt.Println("\n🧪 Test 3: Latch Creation")
	fmt.Println("----------------------------------------")

	// Create CRUD handler
	c := handler.NewCRUD()

	// Create Latch
	fmt.Println("Creating latch...")
	res := c.CreateLatch(latch.CreateLatchParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("Integration Test Latch"),
			Key:         "integration_test_latch",
		},
	})

	if flagOK, ok := res.(*flag.CreateFlagOK); ok {
		fmt.Printf("✅ Latch created successfully\n")
		fmt.Printf("   ID: %d\n", flagOK.Payload.ID)
		fmt.Printf("   Description: %s\n", *flagOK.Payload.Description)
		fmt.Printf("   Key: %s\n", flagOK.Payload.Key)

		// Verify tags are included
		if flagOK.Payload.Tags != nil && len(flagOK.Payload.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", *flagOK.Payload.Tags[0].Value)
			if *flagOK.Payload.Tags[0].Value == "latch" {
				fmt.Println("   ✅ Latch tag correctly included in response!")
			} else {
				log.Fatalf("   ❌ Expected 'latch' tag, got '%s'", *flagOK.Payload.Tags[0].Value)
			}
		} else {
			log.Fatalf("   ❌ No tags found in response!")
		}

		// Verify latch-created segments
		if flagOK.Payload.Segments != nil && len(flagOK.Payload.Segments) > 0 {
			fmt.Printf("   Segments: %d\n", len(flagOK.Payload.Segments))
			segment := flagOK.Payload.Segments[0]
			if segment.RolloutPercent != nil && *segment.RolloutPercent == 100 {
				fmt.Println("   ✅ Latch segment created with 100% rollout")
			}
		} else {
			log.Fatalf("   ❌ No segments found in response!")
		}

		// Verify latch-created variants
		if flagOK.Payload.Variants != nil && len(flagOK.Payload.Variants) > 0 {
			fmt.Printf("   Variants: %d\n", len(flagOK.Payload.Variants))
			variant := flagOK.Payload.Variants[0]
			if variant.Key != nil && *variant.Key == "APPLICABLE" {
				fmt.Printf("   ✅ Latch variant created with key: %s\n", *variant.Key)
			}
		} else {
			log.Fatalf("   ❌ No variants found in response!")
		}

	} else {
		log.Fatalf("❌ Failed to create latch: %+v", res)
	}
}

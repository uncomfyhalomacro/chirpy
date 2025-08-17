package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
)

func TestJWT(t *testing.T) {
	testSigningKey := "Foo"
	testUUIDs := make([]uuid.UUID, 100)
	for i := 0; i < 100; i++ {
		testUUIDs = append(testUUIDs, uuid.New())
	}
	duration, err := time.ParseDuration("1h")
	if err != nil {
		t.Errorf("%v\n", err)
	}
	for _, testUUID := range testUUIDs {
		tokenString, err := MakeJWT(testUUID, testSigningKey, duration)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		id, err := ValidateJWT(tokenString, testSigningKey)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		if id == testUUID {
			t.Logf("Jwt is valid\n")
		} else {
			t.Errorf("id does not match: %v vs %v\n", id, testUUID)
		}
	}

}

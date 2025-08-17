package auth

import (
	"testing"
)

func TestPasswordCreation(t *testing.T) {
	testCases := []struct{
		input string
		expected string
	} {
		{
			input: "a password!",
			expected: "$2a$10$5ZySHgVNSm/ZfCDGD56uTuRgfbHTljAyg.pozO0AYeccADAWH6We.",
		},
		{
			input: "a password!",
			expected: "$2a$10$5ZySHgVNSm/ZfCDGD56uTuRgfbHTljAyg.pozO0AYeccADAWH6We.",
		},
		{
			input: "+hE6IU8sw0x0wA==",
			expected: "$2a$10$lRs0DvNOypcZ7RwbxhYu4uH122mQ6gqsTIklqkhpCqFtcbyqUUpmi",
		},
		{
			input: "4BfCfyk9AhN+iDaG",
			expected: "2a$10$MgZlb8muj9DWdKazfGK8xecTtLjo6H6YLXLLyGGZyBD2SkyQpplzC",
		},
	}

	for _, testCase := range testCases {
		hash, err := HashPassword(testCase.input)

		if err != nil {
			t.Errorf("Failed to generate hash for password: %v\n", err)
		}
		if hash != testCase.expected {
			t.Logf("As it should be, expected hash does not match against expected for password `%s`: %v vs %v\n", testCase.input, hash, testCase.expected)
		}
		if err = CheckPasswordHash(testCase.input, hash); err != nil {
			t.Errorf("Password hash check function failed: %v\n", err)
		}
	}

}

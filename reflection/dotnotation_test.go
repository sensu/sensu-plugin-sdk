package reflection

import (
	"testing"
)

type testServiceNowResult struct {
	Array1  []string
	Array2  []int
	Array3  []testServiceNowSysId
	Field1  string
	Field2  int
	Field3  float32
	Result1 *testServiceNowSysId
	Result2 *testServiceNowSysId
}

type testServiceNowSysId struct {
	SysId string
}

func TestDotNotation(t *testing.T) {
	sn := &testServiceNowResult{
		Array1: []string{"arraystring1", "arraystring2"},
		Array2: []int{9, 8},
		Array3: []testServiceNowSysId{{SysId: "b"}, {SysId: "c"}, {SysId: "d"}},
		Field1: "field1Value",
		Field2: 12345,
		Field3: 678.09,
		Result1: &testServiceNowSysId{
			SysId: "sysIdValue",
		},
		Result2: nil,
	}

	expectedResults := make(map[string]string)
	expectedResults["Field1"] = "field1Value"
	expectedResults["Field2"] = "12345"
	expectedResults["Field3"] = "678.09"
	expectedResults["Result1.SysId"] = "sysIdValue"
	expectedResults["Result2"] = "nil"
	expectedResults["Array1[0]"] = "arraystring1"
	expectedResults["Array1[1]"] = "arraystring2"
	expectedResults["Array2[0]"] = "9"
	expectedResults["Array2[1]"] = "8"
	expectedResults["Array3[0].SysId"] = "b"
	expectedResults["Array3[1].SysId"] = "c"
	expectedResults["Array3[2].SysId"] = "d"

	dotNotations := DotNotation(sn)
	if len(dotNotations) != 12 {
		t.Error("Entry should return 11 key/value pairs")
	}

	for i := 0; i < len(dotNotations); i++ {
		pair := dotNotations[i]
		expectedResult := expectedResults[pair.key]
		if len(expectedResult) > 0 {
			if expectedResult != pair.value {
				t.Errorf("Wrong result for key %s: actual: %s; expected: %s", pair.key, pair.value, expectedResult)
			}
		} else {
			t.Errorf("Unexpected key: %s", pair.key)
		}
		delete(expectedResults, pair.key)
	}

	// Check for the expected keys that were not found
	if len(expectedResults) > 0 {
		t.Errorf("Expected results not found: %v", expectedResults)
	}
}

func TestDotNotationString(t *testing.T) {
	pairs := []nameValuePair{
		{"key1", "value1"},
		{"key2", "value2"},
	}
	expectedString := "key1: value1\nkey2: value2\n"
	resultString := DotNotationToString(pairs, ": ")

	if resultString != expectedString {
		t.Errorf("Invalid encoding. Expected: <<<\n%s\n>>>\nResult: <<<\n%s\n>>>", expectedString, resultString)
	}

}

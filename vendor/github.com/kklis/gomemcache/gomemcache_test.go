/**
 * Run two following commands on the background before this test.
 * $ memcached
 * $ memcached -s /tmp/memcached.sock -a 0755
 */
package gomemcache

import (
	"bufio"
	"strconv"
	"strings"
	"testing"
)

const (
	KEY_1   string = "test-key-1"
	KEY_2   string = "test-key-2"
	VALUE_1 string = "test-value-1"
	VALUE_2 string = "test-value-2"
	FLAGS   int    = 1
)

var memc *Memcache

func TestDial_TCP(t *testing.T) {
	c, err := Dial("tcp", "127.0.0.1:11211")
	assertNoError(t, err)
	err = c.Close()
	assertNoError(t, err)
}

func TestDial_UNIX(t *testing.T) {
	c, err := Dial("unix", "/tmp/memcached.sock")
	assertNoError(t, err)
	err = c.Close()
	assertNoError(t, err)
}

func testConnect_TCP(t *testing.T) {
	c, err := Connect("127.0.0.1", 11211)
	assertNoError(t, err)
	err = c.Close()
	assertNoError(t, err)
}

func testConnect_UNIX(t *testing.T) {
	c, err := Connect("/tmp/memcached.sock", 0)
	assertNoError(t, err)
	err = c.Close()
	assertNoError(t, err)
}

func TestSet(t *testing.T) {
	connect(t)
	err := memc.Set(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	assertGet(t, KEY_1, VALUE_1)
	cleanUp()
}

func TestAdd(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	assertGet(t, KEY_1, VALUE_1)
	cleanUp()
}

func TestGetMulti(t *testing.T) {
	// given
	connect(t)
	memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	memc.Add(KEY_2, []uint8(VALUE_2), FLAGS, 0)
	// when
	receivedValues, err := memc.GetMulti(KEY_1, KEY_2)
	// then
	assertNoError(t, err)
	assertEqual(t, VALUE_1, string(receivedValues[KEY_1].Value))
	assertEqual(t, VALUE_2, string(receivedValues[KEY_2].Value))
	assertEqualInt(t, FLAGS, receivedValues[KEY_1].Flags)
	assertEqualInt(t, FLAGS, receivedValues[KEY_2].Flags)
	cleanUp()
}

func TestGetMultiForMissingKeys(t *testing.T) {
	// given
	connect(t)
	// when
	receivedValues, err := memc.GetMulti(KEY_1, KEY_2)
	// then
	assertNoError(t, err)
	if len(receivedValues) != 0 {
		t.Error("Expected emtpy result")
	}
	cleanUp()
}

func TestGetMultiForMissingOneKey(t *testing.T) {
	// given
	connect(t)
	memc.Add(KEY_2, []uint8(VALUE_2), FLAGS, 0)
	// when
	receivedValues, err := memc.GetMulti(KEY_1, KEY_2)
	// then
	assertNoError(t, err)
	assertEqual(t, "", string(receivedValues[KEY_1].Value))
	assertEqual(t, VALUE_2, string(receivedValues[KEY_2].Value))
	assertEqualInt(t, 0, receivedValues[KEY_1].Flags)
	assertEqualInt(t, FLAGS, receivedValues[KEY_2].Flags)
	cleanUp()
}

func TestGetForMissingKey(t *testing.T) {
	// given
	connect(t)
	// when
	receivedValue, _, err := memc.Get(KEY_1)
	// then
	if err != NotFoundError {
		t.Error("Expected NotFoundError. receivedValue:", string(receivedValue))
	}
	cleanUp()
}

func TestAddingPresentKEY_1(t *testing.T) {
	connect(t)
	err := memc.Set(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	assertGet(t, KEY_1, VALUE_1)
	err = memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	if err == nil {
		t.Error("Adding should fail because KEY_1 is already present")
	}
	cleanUp()
}

func TestReplace(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	newValue := "new value"
	err = memc.Replace(KEY_1, []uint8(newValue), FLAGS, 0)
	assertNoError(t, err)
	assertGet(t, KEY_1, newValue)
	cleanUp()
}

func TestPrepend(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	prefix := "prefix"
	err = memc.Prepend(KEY_1, []uint8(prefix), FLAGS, 0)
	assertNoError(t, err)
	assertGet(t, KEY_1, prefix+VALUE_1)
	cleanUp()
}

func TestAppend(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	suffix := "suffix"
	err = memc.Append(KEY_1, []uint8(suffix), FLAGS, 0)
	assertNoError(t, err)
	assertGet(t, KEY_1, VALUE_1+suffix)
	cleanUp()
}

func TestDelete(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	err = memc.Delete(KEY_1)
	assertNoError(t, err)
	_, _, err = memc.Get(KEY_1)
	if err == nil {
		t.Error("Data not removed from memcache")
	}
	cleanUp()
}

func TestFlushAll(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8(VALUE_1), FLAGS, 0)
	assertNoError(t, err)
	err = memc.FlushAll()
	assertNoError(t, err)
	_, _, err = memc.Get(KEY_1)
	if err == nil {
		t.Error("Data not removed from memcache")
	}
	cleanUp()
}

func TestIncr(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8("1234"), FLAGS, 0)
	assertNoError(t, err)
	i, err := memc.Incr(KEY_1, 9)
	assertNoError(t, err)
	if i != 1243 {
		t.Error("Value expected: 1243\nValue received: " + strconv.FormatUint(i, 10))
	}
	cleanUp()
}

func TestDecr(t *testing.T) {
	connect(t)
	err := memc.Add(KEY_1, []uint8("1243"), FLAGS, 0)
	assertNoError(t, err)
	i, err := memc.Decr(KEY_1, 9)
	assertNoError(t, err)
	if i != 1234 {
		t.Error("Value expected: 1234\nValue received: " + strconv.FormatUint(i, 10))
	}
	cleanUp()
}

func TestParseResponseWithOneValue(t *testing.T) {
	// given
	responseString := "VALUE " + KEY_1 + " " + strconv.Itoa(FLAGS) + " " + strconv.Itoa(len(VALUE_1)) + "\r\n" +
		VALUE_1 + "\r\n" +
		"END\r\n"
	stringReader := strings.NewReader(responseString)
	reader := bufio.NewReader(stringReader)
	// when
	value, receivedFlags, err := memc.readValue(reader, KEY_1)
	// then
	assertNoError(t, err)
	assertEqual(t, VALUE_1, string(value))
	assertEqualInt(t, FLAGS, receivedFlags)
}

func TestParseResponseWithTwoValues(t *testing.T) {
	// given
	responseString := "VALUE " + KEY_1 + " " + strconv.Itoa(FLAGS) + " " + strconv.Itoa(len(VALUE_1)) + "\r\n" +
		VALUE_1 + "\r\n" +
		"VALUE " + KEY_2 + " " + strconv.Itoa(FLAGS) + " " + strconv.Itoa(len(VALUE_2)) + "\r\n" +
		VALUE_2 + "\r\n" +
		"END\r\n"
	stringReader := strings.NewReader(responseString)
	reader := bufio.NewReader(stringReader)
	// when called first time
	value, receivedFlags, err := memc.readValue(reader, KEY_1)
	// then
	assertNoError(t, err)
	assertEqual(t, VALUE_1, string(value))
	assertEqualInt(t, FLAGS, receivedFlags)
	// when called second time
	value, receivedFlags, err = memc.readValue(reader, KEY_2)
	// then
	assertNoError(t, err)
	assertEqual(t, VALUE_2, string(value))
	assertEqualInt(t, FLAGS, receivedFlags)
}

func assertGet(t *testing.T, key string, expectedValue string) {
	receivedValue, receivedFlags, err := memc.Get(key)
	assertNoError(t, err)
	assertEqual(t, expectedValue, string(receivedValue))
	assertEqualInt(t, FLAGS, receivedFlags)
}

func connect(t *testing.T) {
	connection, err := Connect("127.0.0.1", 11211)
	assertNoError(t, err)
	memc = connection
}

func cleanUp() {
	memc.Delete(KEY_1)
	memc.Delete(KEY_2)
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func assertEqualInt(t *testing.T, expectedValue int, receivedValue int) {
	if receivedValue != expectedValue {
		t.Error("Value expected: " + strconv.Itoa(expectedValue) + "\nValue received: " + strconv.Itoa(receivedValue))
	}
}

func assertEqual(t *testing.T, expectedValue string, receivedValue string) {
	if receivedValue != expectedValue {
		t.Error("Value expected: " + expectedValue + "\nValue received: " + receivedValue)
	}
}

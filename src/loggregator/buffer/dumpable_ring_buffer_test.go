package buffer

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	messagetesthelpers "github.com/cloudfoundry/loggregatorlib/logmessage/testhelpers"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setup(bufferSize int) (chan *logmessage.Message, *DumpableRingBuffer) {
	channel := make(chan *logmessage.Message, 10)
	return channel, NewDumpableRingBuffer(channel, bufferSize)
}

func TestInputChannelAddToBuffer(t *testing.T) {
	channel, buffer := setup(10)
	message1 := messagetesthelpers.NewMessage(t, "message 1", "appId")
	channel <- message1
	message2 := messagetesthelpers.NewMessage(t, "message 2", "appId")
	channel <- message2

	close(channel)
	buffer.waitForClose()

	dump := buffer.Dump()
	assert.Equal(t, 2, len(dump))
	assert.Equal(t, message1, dump[0])
	assert.Equal(t, message2, dump[1])
}

func TestBufferDropsOldMessages(t *testing.T) {
	channel, buffer := setup(1)
	message1 := messagetesthelpers.NewMessage(t, "message 1", "appId")
	channel <- message1
	message2 := messagetesthelpers.NewMessage(t, "message 2", "appId")
	channel <- message2

	close(channel)
	buffer.waitForClose()

	dump := buffer.Dump()
	assert.Equal(t, 1, len(dump))
	assert.Equal(t, message2, dump[0])
}

func TestBufferHasChannelListenerWithLimitOne(t *testing.T) {
	inChannel, buffer := setup(1)
	message1 := messagetesthelpers.NewMessage(t, "message 1", "appId")
	inChannel <- message1
	message2 := messagetesthelpers.NewMessage(t, "message 2", "appId")
	inChannel <- message2
	message3 := messagetesthelpers.NewMessage(t, "message 3", "appId")
	inChannel <- message3
	close(inChannel)
	assert.Equal(t, message3, <-buffer.OutputChannel())
	select {
	case <-buffer.OutputChannel():
	case <-time.After(2 * time.Millisecond):
		t.Error("OutputChannel should be closed")
	}
}

func TestBufferHasChannelListenerWithLimitMany(t *testing.T) {
	inChannel, buffer := setup(3)
	message1 := messagetesthelpers.NewMessage(t, "message 1", "appId")
	inChannel <- message1
	message2 := messagetesthelpers.NewMessage(t, "message 2", "appId")
	inChannel <- message2
	message3 := messagetesthelpers.NewMessage(t, "message 3", "appId")
	inChannel <- message3
	close(inChannel)
	assert.Equal(t, message1, <-buffer.OutputChannel())
	assert.Equal(t, message2, <-buffer.OutputChannel())
	assert.Equal(t, message3, <-buffer.OutputChannel())
	select {
	case <-buffer.OutputChannel():
	case <-time.After(2 * time.Millisecond):
		t.Error("OutputChannel should be closed")
	}
}

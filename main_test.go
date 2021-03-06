package jobinator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	globalClient  *client
	fakeDataStore = make(map[string]int)
)

func TestNewClient(t *testing.T) {
	c, err := NewClient("sqlite3", "test.db", clientConfig{
		WorkerSleepTime: time.Second / 10,
	})
	assert.Nil(t, err)
	assert.NotNil(t, c)
	globalClient = c
}

func TestRegisterWorker(t *testing.T) {
	type testData struct {
		NumToSet int
		KeyToSet string
	}
	wf := func(j *jobRef) error {
		td := &testData{}
		err := j.ScanArgs(td)
		if err != nil {
			return err
		}
		fakeDataStore[td.KeyToSet] = td.NumToSet
		return nil
	}
	globalClient.RegisterWorker("numsetter", wf)
}

func TestQueueJob(t *testing.T) {
	type testData struct {
		NumToSet int
		KeyToSet string
	}
	td := &testData{
		NumToSet: 9,
		KeyToSet: "myKey",
	}
	err := globalClient.EnqueueJob("numsetter", td)
	assert.Nil(t, err)
}

func TestExecuteOne(t *testing.T) {
	err := globalClient.ExecuteOneJob()
	assert.Nil(t, err)
	assert.Equal(t, 9, fakeDataStore["myKey"])
}

func TestStartWorker(t *testing.T) {
	type thing struct {
		number int
	}
	np := &thing{
		number: 0,
	}
	wfIncNum := func(j *jobRef) error {
		np.number++
		return nil
	}
	globalClient.RegisterWorker("numthing", wfIncNum)
	globalClient.EnqueueJob("numthing", nil)
	globalClient.EnqueueJob("numthing", nil)
	globalClient.EnqueueJob("numthing", nil)
	globalClient.startWorker()
	time.Sleep(time.Second * 5)
	globalClient.stopWorker()
	assert.Equal(t, np.number, 3)
}

func TestPendingJobs(t *testing.T) {
	globalClient.EnqueueJob("willNeverExist", nil)
	count, err := globalClient.pendingJobs()
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

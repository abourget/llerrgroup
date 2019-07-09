package llerrgroup

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func Test_ParallelRuns(t *testing.T) {
	var cnt atomic.Uint32
	g := New(8)
	for i := 0; i < 12; i++ {
		if stop := g.Stop(); stop != false {
			t.Fail()
		}

		g.Go(func() error {
			cnt.Add(1)
			time.Sleep(time.Millisecond * 10)
			return nil
		})

	}
	errCh := make(chan error, 1)
	go func() {
		err := g.Wait()
		errCh <- err
	}()
	select {
	case <-time.After(time.Millisecond * 30):
		t.Fail()
	case err := <-errCh:
		require.NoError(t, err)
	}

	assert.Equal(t, uint32(12), cnt.Load())
	assert.Equal(t, 12, g.CallsCount())

}

func Test_ParallelLimits(t *testing.T) {
	var cnt atomic.Uint32
	g := New(1)
	for i := 0; i < 100; i++ {
		if stop := g.Stop(); stop != false {
			fmt.Println("failing on stop")
			t.Fail()
		}
		assert.Equal(t, i, int(cnt.Load()))
		g.Go(func() error {
			cnt.Add(1)
			return nil
		})
	}
	err := g.Wait()
	assert.NoError(t, err)

}

func Test_DynamicParallelReduction(t *testing.T) {
	var cnt atomic.Uint32
	parallelRunningIndication := false
	g := New(10)
	for i := 0; i < 100; i++ {
		if stop := g.Stop(); stop != false {
			fmt.Println("failing on stop")
			t.Fail()
		}
		if i == 19 {
			g.SetSize(1)
		}
		if i < 20 {
			if i != int(cnt.Load()) {
				parallelRunningIndication = true
			}
		} else {
			assert.Equal(t, i, int(cnt.Load()), "routines after #20 should be run one at a time")
		}
		g.Go(func() error {
			cnt.Add(1)
			return nil
		})
	}
	err := g.Wait()
	assert.NoError(t, err)

	assert.True(t, parallelRunningIndication, "cannot validate that first 20 routines run in parallel")
}

func Test_DynamicParallelIncrease(t *testing.T) {
	var cnt atomic.Uint32
	parallelRunningIndication := false
	g := New(1)
	for i := 0; i < 100; i++ {
		if stop := g.Stop(); stop != false {
			fmt.Println("failing on stop")
			t.Fail()
		}
		if i == 20 {
			g.SetSize(10)
		}
		if i > 20 {
			if i != int(cnt.Load()) {
				parallelRunningIndication = true
			}
		} else {
			assert.Equal(t, i, int(cnt.Load()), "routines before #20 should be run one at a time")
		}
		g.Go(func() error {
			cnt.Add(1)
			return nil
		})
	}
	err := g.Wait()
	assert.NoError(t, err)

	assert.True(t, parallelRunningIndication, "cannot validate that routines after #20 run in parallel")

}

func ExampleGroupParallelAndQuitSoon() {
	url := []string{"http://1", "http://2"}
	grp := New(15)
	for _, u := range url {
		if grp.Stop() {
			continue
		}

		u := u
		grp.Go(func() error {
			log.Println("Fetching", u)
			_, err := http.Get(u)
			return err
		})
	}
	err := grp.Wait()
	if err != nil {
		log.Println("Errored:", err)
	}
	log.Println("All done!")
}

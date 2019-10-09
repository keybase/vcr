package vcr

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

func TestDo(t *testing.T) {
	v := New("testdata").Record()
	req, err := http.NewRequest("GET", "https://keybase.io", nil)
	if err != nil {
		t.Fatal(err)
	}
	respRec, err := v.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	v.Play()
	respPlay, err := v.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	recordBody, err := ioutil.ReadAll(respRec.Body)
	if err != nil {
		t.Fatal(err)
	}
	playBody, err := ioutil.ReadAll(respPlay.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(recordBody, playBody) {
		t.Errorf("recordBody != playBody")
	}
}

func TestGet(t *testing.T) {
	v := New("testdata").Record()
	url := "https://keybase.io"
	respRec, err := v.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	v.Play()
	respPlay, err := v.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	recordBody, err := ioutil.ReadAll(respRec.Body)
	if err != nil {
		t.Fatal(err)
	}
	playBody, err := ioutil.ReadAll(respPlay.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(recordBody, playBody) {
		t.Errorf("recordBody != playBody")
	}
}

func TestPostForm(t *testing.T) {
	v := New("testdata").Record()
	endpoint := "https://keybase.io"
	vals := url.Values{}
	vals.Set("q", "keybase")
	respRec, err := v.PostForm(endpoint, vals)
	if err != nil {
		t.Fatal(err)
	}

	v.Play()
	respPlay, err := v.PostForm(endpoint, vals)
	if err != nil {
		t.Fatal(err)
	}

	recordBody, err := ioutil.ReadAll(respRec.Body)
	if err != nil {
		t.Fatal(err)
	}
	playBody, err := ioutil.ReadAll(respPlay.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(recordBody, playBody) {
		t.Errorf("recordBody != playBody")
	}
}

func TestConcurrency(t *testing.T) {
	// record and playback multiple requests in different orders (it's concurrent
	// so this is probabilistic)
	v := New("testdata/concurrency").Record()
	url := "https://keybase.io/_/api/1.0/merkle/root.json"

	errChRecord := make(chan error, 100)
	var wgRecord sync.WaitGroup
	getRequest := func(wg *sync.WaitGroup, errCh chan error) {
		defer wg.Done()
		_, err := v.Get(url)
		errCh <- err
	}
	doRequest := func(wg *sync.WaitGroup, errCh chan error) {
		defer wg.Done()
		req, err := http.NewRequest("GET", url, nil)
		errCh <- err
		_, err = v.Do(req)
		errCh <- err
	}
	for i := 0; i < 5; i++ {
		wgRecord.Add(2)
		go getRequest(&wgRecord, errChRecord)
		go doRequest(&wgRecord, errChRecord)
	}
	wgRecord.Wait()
	close(errChRecord)
	for err := range errChRecord {
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Log("finished recording concurrent requests. start playback...")
	v.Play()

	var wgPlay sync.WaitGroup
	errChPlay := make(chan error, 100)
	for i := 0; i < 5; i++ {
		wgPlay.Add(2)
		go getRequest(&wgPlay, errChPlay)
		go doRequest(&wgPlay, errChPlay)
	}
	wgPlay.Wait()
	close(errChPlay)
	for err := range errChPlay {
		if err != nil {
			t.Fatal(err)
		}
	}
}

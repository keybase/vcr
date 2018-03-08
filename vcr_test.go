package vcr

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
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

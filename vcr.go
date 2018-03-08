package vcr

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

type Mode int

const (
	_         = iota
	Play Mode = iota
	Record
	Live
)

var (
	ErrInvalidMode = errors.New("invalid mode")
)

type VCR struct {
	dir   string
	mode  Mode
	seqno int
	Debug bool
}

func New(dir string) *VCR {
	return &VCR{
		dir:  dir,
		mode: Play,
	}
}

func (v *VCR) Play() *VCR {
	v.mode = Play
	v.seqno = 0
	return v
}

func (v *VCR) Record() *VCR {
	v.mode = Record
	return v
}

func (v *VCR) Live() *VCR {
	v.mode = Live
	return v
}

func (v *VCR) SetDir(dir string) {
	v.dir = dir
	v.seqno = 0
}

func (v *VCR) Do(req *http.Request) (resp *http.Response, err error) {
	filename, err := v.doFilename(req)
	if err != nil {
		return nil, err
	}

	defer v.incSeqno()

	switch v.mode {
	case Play:
		return v.play(filename)
	case Record:
		return v.recordDo(req, filename)
	case Live:
		return v.liveDo(req)
	}

	return nil, ErrInvalidMode
}

func (v *VCR) Get(url string) (resp *http.Response, err error) {
	filename, err := v.getFilename(url)
	if err != nil {
		return nil, err
	}

	defer v.incSeqno()

	switch v.mode {
	case Play:
		return v.play(filename)
	case Record:
		return v.recordGet(url, filename)
	case Live:
		return v.liveGet(url)
	}

	return nil, ErrInvalidMode
}

func (v *VCR) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	filename, err := v.postFormFilename(url, data)
	if err != nil {
		return nil, err
	}

	defer v.incSeqno()

	switch v.mode {
	case Play:
		return v.play(filename)
	case Record:
		return v.recordPostForm(url, data, filename)
	case Live:
		return v.livePostForm(url, data)
	}

	return nil, ErrInvalidMode
}

func (v *VCR) incSeqno() {
	v.seqno += 1
}

func (v *VCR) play(filename string) (*http.Response, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return v.decodeResponse(data)
}

func (v *VCR) recordDo(req *http.Request, filename string) (*http.Response, error) {
	resp, err := v.liveDo(req)
	return v.writeResponse(resp, filename, err)
}

func (v *VCR) liveDo(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func (v *VCR) recordGet(url, filename string) (*http.Response, error) {
	resp, err := v.liveGet(url)
	return v.writeResponse(resp, filename, err)
}

func (v *VCR) liveGet(url string) (*http.Response, error) {
	return http.DefaultClient.Get(url)
}

func (v *VCR) recordPostForm(url string, data url.Values, filename string) (*http.Response, error) {
	resp, err := v.livePostForm(url, data)
	return v.writeResponse(resp, filename, err)
}

func (v *VCR) livePostForm(url string, data url.Values) (*http.Response, error) {
	return http.DefaultClient.PostForm(url, data)
}

func (v *VCR) reqHash(req *http.Request) (string, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(dump)
	return hex.EncodeToString(sum[:]), nil
}

func (v *VCR) urlHash(url string) (string, error) {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:]), nil
}

func (v *VCR) doFilename(req *http.Request) (string, error) {
	hash, err := v.reqHash(req)
	if err != nil {
		return "", err
	}
	return v.filename("do", hash), nil
}

func (v *VCR) getFilename(url string) (string, error) {
	hash, err := v.urlHash(url)
	if err != nil {
		return "", err
	}
	return v.filename("get", hash), nil
}

func (v *VCR) postFormFilename(url string, data url.Values) (string, error) {
	// fmt.Printf("postFormFilename url: %s\n", url)
	// fmt.Printf("postFormFilename data: %s\n", data.Encode())
	h := sha256.New()
	h.Write([]byte(url))
	h.Write([]byte(data.Encode()))
	hash := hex.EncodeToString(h.Sum(nil))
	return v.filename("postform", hash), nil
}

func (v *VCR) filename(prefix, hash string) string {
	return filepath.Join(v.dir, fmt.Sprintf("%s_%s_%d.vcr", prefix, hash, v.seqno))
}

func (v *VCR) encodeResponse(resp *http.Response) ([]byte, error) {
	return httputil.DumpResponse(resp, true)
}

func (v *VCR) decodeResponse(data []byte) (*http.Response, error) {
	buf := bytes.NewBuffer(data)
	return http.ReadResponse(bufio.NewReader(buf), nil)
}

func (v *VCR) writeResponse(resp *http.Response, filename string, reqErr error) (*http.Response, error) {
	if resp == nil {
		return nil, reqErr
	}

	enc, err := v.encodeResponse(resp)
	if err != nil {
		if reqErr != nil {
			return nil, reqErr
		}
		return nil, err
	}

	if v.Debug {
		_, err = os.Stat(filename)
		if !os.IsNotExist(err) {
			fmt.Printf("warning: file %s exists and will be overwritten\n", filename)
			existing, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			fmt.Printf("existing content:\n%s\n", existing)
			fmt.Printf("new content:\n%s\n", enc)
		}
	}

	if err := ioutil.WriteFile(filename, enc, 0644); err != nil {
		if reqErr != nil {
			return nil, reqErr
		}
		return nil, err
	}

	return resp, reqErr
}

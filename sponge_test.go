package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func tempArgs(args []string) func() {
	hold := os.Args
	os.Args = args
	return func() {
		os.Args = hold
	}
}

func tempFile() *os.File {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	return f
}

func tempStdout(f *os.File, canRemove bool) func() {
	hold := os.Stdout
	os.Stdout = f
	return func() {
		os.Stdout = hold
		if canRemove {
			os.Remove(f.Name())
		}
	}
}

type fakeIO func() bool

func (f *fakeIO) proc(buf []byte) (int, error) {
	if (*f)() {
		return 0, io.EOF
	} else {
		return len(buf), nil
	}
}

func (f *fakeIO) Read(buf []byte) (int, error) {
	return f.proc(buf)
}

func (f *fakeIO) Write(buf []byte) (int, error) {
	return f.proc(buf)
}

func TestGetSink_whenNoArgs_returnsStdout(t *testing.T) {
	defer tempArgs([]string{"sponge"})()

	f := tempFile()
	defer tempStdout(f, true)()

	if out := GetSink(); out != f {
		t.Errorf("Expected %v; got %v", f, out)
	}
}

func TestGetSink_withFirstArg_isOutFile(t *testing.T) {
	n := func() string {
		f := tempFile()
		n := f.Name()
		f.Close()
		return n
	}()
	defer os.Remove(n)

	defer tempArgs([]string{"sponge", n})()

	out := GetSink()
	f := out.(*os.File)
	if f.Name() != n {
		t.Errorf("Expected file with name %q; got %q", n, f.Name())
	}
}

func TestSponge_readerConsumedBeforeWrite(t *testing.T) {
	didRead := false
	didWrite := false
	n := 10
	r := fakeIO(func() bool {
		if didWrite {
			t.Error("Did read after write.")
		}
		if n > 0 {
			n--
			return false
		}
		didRead = true
		return true
	})
	w := fakeIO(func() bool {
		didWrite = true
		if !didRead {
			t.Error("Did write before reader was exhausted.")
		}
		return false
	})

	if err := Sponge(&r, &w); err != nil {
		t.Error(err)
	}
}

func TestSponge_inputEqualsOutput(t *testing.T) {
	testStr := `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam vitae ligula
lectus. Phasellus a turpis dictum, gravida leo mattis, tempus turpis. Vivamus
nec libero vestibulum, imperdiet tellus a, sodales sapien. Curabitur ornare orci
odio, mattis faucibus risus pretium non. Ut posuere ex eu sem condimentum
aliquet. Proin dictum lacus vitae semper bibendum. Nam convallis nibh sed
pellentesque dignissim. Fusce dignissim, erat non aliquet volutpat, urna purus
ullamcorper nisi, eu sagittis nibh libero at urna. Sed ac tempor eros, at
elementum lectus. Phasellus non lorem gravida, congue massa et, scelerisque
ligula. Phasellus at laoreet sapien. Donec ac mi ac nisi fringilla lacinia.
Vestibulum aliquet quis lorem quis sodales. Praesent non magna nec nunc finibus
maximus ut a ante. Nam ornare imperdiet pretium. Vestibulum id quam ornare,
volutpat purus id, hendrerit lorem.`
	r := strings.NewReader(testStr)
	var w strings.Builder

	if err := Sponge(r, &w); err != nil {
		t.Error(err)
	}

	if w.String() != testStr {
		t.Error("Bad copy.")
	}
}

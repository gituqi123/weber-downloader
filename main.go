package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {

	old_hash := ""

	hasher := sha256.New()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//resp, err := http.Get("https://ostan-es.ir/")
	resp, err := http.Get("https://farsnews.ir/")
	if err != nil {
		return
	}
	resp.Header.Set("User-Agent", "aft")
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return
	}
	//fmt.Printf("client: status code: %d\n")
	if resp.StatusCode == 200 {
		fmt.Printf("url: %s \n", resp.Request.URL.String())
		hasher.Write([]byte(buf.Bytes()))
		bs := hasher.Sum(nil)
		old_hash = hex.EncodeToString(bs)

	}

	for {
		hasher := sha256.New()
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		//resp, err := http.Get("https://ostan-es.ir/")
		resp, err := http.Get("https://farsnews.ir/")
		if err != nil {
			return
		}
		resp.Header.Set("User-Agent", "aft")
		defer resp.Body.Close()

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			return
		}
		//fmt.Printf("client: status code: %d\n")
		if resp.StatusCode == 200 {
			fmt.Printf("url: %s \n", resp.Request.URL.String())
			hasher.Write([]byte(buf.Bytes()))
			bs := hasher.Sum(nil)
			if old_hash != hex.EncodeToString(bs) {
				for {

					f, err := os.Open("./peru-alert.mp3")
					if err != nil {
						log.Fatal(err)
					}

					streamer, format, err := mp3.Decode(f)
					if err != nil {
						log.Fatal(err)
					}
					defer streamer.Close()

					speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

					done := make(chan bool)
					speaker.Play(beep.Seq(streamer, beep.Callback(func() {
						done <- true
					})))

					<-done
				}
			}
			old_hash = hex.EncodeToString(bs)
			fmt.Printf("%s\n", old_hash)
		}
		time.Sleep(3 * time.Second)

	}

	//final = resp.Request.URL.String()
	//return final, resp.StatusCode, buf, nil
}

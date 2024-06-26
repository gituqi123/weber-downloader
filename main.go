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
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func get_page(link string) string {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(link)
	if err != nil {
		os.Exit(-1)
	}
	resp.Header.Set("User-Agent", "aft")
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		os.Exit(-1)
	}
	//fmt.Printf("client: status code: %d\n")
	// if resp.StatusCode == 200 {
	// 	fmt.Printf("url: %s \n", resp.Request.URL.String())
	// }
	return buf.String()
}

func alert_sound() {
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
		time.Sleep(100 * time.Millisecond)
	}
}

func remove_false_changes(web_page string) string {
	black_list_strings := []string{"weather", "azan", "views_count", "btn3-style-gradient", "timetoday", "data-lightbox", "su-image-carousel", "wp-aparat", "", "typography_", "_StyledDynamicTypographyComponent"}
	//black_list_strings := []string{"typography_", "_StyledDynamicTypographyComponent"}
	splitted_string := strings.Split(web_page, "\n")
	for i := 0; i < len(splitted_string); i++ {
		for j := 0; j < len(black_list_strings); j++ {
			if strings.Contains(splitted_string[i], black_list_strings[j]) {
				splitted_string[i] = ""
				if i+1 != len(splitted_string) {
					splitted_string[i+1] = ""
				}
			}
		}
	}
	return strings.Join(splitted_string, "\n")
}

func checker(domain string) {
	old_hash := ""
	page := get_page(domain)
	cleaned_string := remove_false_changes(page)
	hasher := sha256.New()
	hasher.Write([]byte(cleaned_string))
	initial_hash := hasher.Sum(nil)
	old_hash = hex.EncodeToString(initial_hash)
	// fmt.Printf("%s %s \n\n\n", domain, cleaned_string)
	for {
		hasher := sha256.New()
		page := get_page(domain)
		cleaned_string := remove_false_changes(page)
		hasher.Write([]byte(cleaned_string))
		new_hash := hasher.Sum(nil)
		new_hash_string := hex.EncodeToString(new_hash)
		if old_hash != new_hash_string {
			fmt.Printf("[alert] %s domain changed [alert]\n", domain)
			fmt.Printf("domain: %s  old = %s , new= %s \n", domain, old_hash, new_hash_string)
			// fmt.Printf("%s %s \n\n\n", domain, cleaned_string)
			alert_sound()

		}
		fmt.Printf("[+] %s domain analyzed \n", domain)
		old_hash = hex.EncodeToString(new_hash)
		// fmt.Printf("domain: %s hash: %s\n", domain, old_hash)
		time.Sleep(60 * time.Second)
	}
}

func main() {

	go checker("google.com")
	select {}
}

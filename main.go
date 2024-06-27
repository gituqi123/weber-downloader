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

func get_page(link string) (int, string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("[ERROR] can't connect to website %s\n", link)
		return -1, "error"
	}
	resp.Header.Set("User-Agent", "aft")
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		fmt.Printf("[ERROR] can't read from %s \n", link)
		return -1, "error"
	}
	return resp.StatusCode, buf.String()
}

func alert_sound() {
	count := 0
	for count < 1 {

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
	black_list_strings := []string{"weather", "azan", "views_count", "btn3-style-gradient", "timetoday", "data-lightbox", "su-image-carousel", "wp-aparat", "theme_token", "time", "form-actions form-wrapper", "view-dom-id-", "_blank", "data-nid"}
	//black_list_strings := []string{"typography_", "_StyledDynamicTypographyComponent"}
	splitted_string := strings.Split(web_page, "\n")
	for i := 0; i < len(splitted_string); i++ {
		for j := 0; j < len(black_list_strings); j++ {
			if strings.Contains(splitted_string[i], black_list_strings[j]) {
				splitted_string[i] = ""
				if i+1 != len(splitted_string) {
					splitted_string[i+1] = ""
				}
				if i+2 < len(splitted_string) {
					splitted_string[i+2] = ""
				}
			}
		}
	}
	// fmt.Println(splitted_string)
	return strings.Join(splitted_string, "\n")
}

func checker(domain string) {
	dt := time.Now()
	old_hash := ""
	status_code, page := get_page(domain)
	for status_code == -1 {
		status_code, page = get_page(domain)
		time.Sleep(5 * time.Second)
	}

	cleaned_string := remove_false_changes(page)
	for status_code != 200 {
		fmt.Printf("[alert] %s domain -> %s  , status_code: %d \n", dt.String(), domain, status_code)
		alert_sound()
		time.Sleep(3 * time.Second)
	}
	hasher := sha256.New()
	hasher.Write([]byte(cleaned_string))
	initial_hash := hasher.Sum(nil)
	old_hash = hex.EncodeToString(initial_hash)
	// fmt.Printf("%s %s \n\n\n", domain, cleaned_string) // debug change detection
	for {
		hasher = sha256.New()
		status_code, page = get_page(domain)
		for status_code == -1 {
			status_code, page = get_page(domain)
			time.Sleep(5 * time.Second)
		}
		for status_code != 200 {
			fmt.Printf("[alert] %s domain -> %s  , status_code: %d \n", dt.String(), domain, status_code)
			alert_sound()
			time.Sleep(5 * time.Second)
		}
		cleaned_string = remove_false_changes(page)
		hasher.Write([]byte(cleaned_string))
		new_hash := hasher.Sum(nil)
		new_hash_string := hex.EncodeToString(new_hash)
		if old_hash != new_hash_string {
			dt = time.Now()
			fmt.Printf("[alert] %s  domain ->  %s changed  status_code: %d [alert]\n", dt.String(), domain, status_code)
			// fmt.Printf("%s %s \n\n\n", domain, cleaned_string) // debug change detection
			alert_sound()

		}
		lowercase_page := strings.ToLower(cleaned_string)
		if strings.Contains(lowercase_page, "hack") {
			dt = time.Now()
			fmt.Printf("[alert] %s  domain ->  %s hack keyword detected: %s  status_code: %d [alert]\n", dt.String(), domain, "hack", status_code)
			alert_sound()
		} else if strings.Contains(lowercase_page, "هک شد") {
			dt = time.Now()
			fmt.Printf("[alert] %s  domain ->  %s hack keyword detected: %s  status_code: %d [alert]\n", dt.String(), domain, "هک", status_code)
			alert_sound()
		} else if strings.Contains(lowercase_page, "مرگ") {
			dt = time.Now()
			fmt.Printf("[alert] %s  domain ->  %s hack keyword detected: %s  status_code: %d [alert]\n", dt.String(), domain, "مرگ", status_code)
			alert_sound()
		} else if strings.Contains(lowercase_page, "deface") {
			dt = time.Now()
			fmt.Printf("[alert] %s  domain ->  %s hack keyword detected: %s  status_code: %d [alert]\n", dt.String(), domain, "deface", status_code)
			alert_sound()
		}

		dt = time.Now()
		fmt.Printf("[+] %s  domain -> %s analyzed \n", dt.String(), domain)
		old_hash = hex.EncodeToString(new_hash)
		// fmt.Printf("domain: %s hash: %s\n", domain, old_hash)
		time.Sleep(5 * time.Second)
	}
}

func main() {

	go checker("https://google.ir/")
	select {}
}

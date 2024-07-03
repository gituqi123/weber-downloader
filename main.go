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

const ErrorColor = "\033[1;31m"
const InfoColor = "\033[1;34m"
const WarningColor = "\033[1;33m"
const colorNone = "\033[0m"

func get_page(link string) (int, string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("[ERROR] can't connect to website %s\n", link)
		return -1, "error"
	}
	//resp.Header.Set("User-Agent", "fta15")
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
	// count := 0
	// for count < 1 {

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
	// }
}

func remove_false_changes(web_page string) string {
	black_list_strings := []string{"weather", "azan", "views_count", "btn3-style-gradient", "timetoday", "data-lightbox", "su-image-carousel", "wp-aparat", "theme_token", "time", "form-actions form-wrapper", "view-dom-id-", "_blank", "data-nid", "geohack", "captcha", "\"__VIEWSTATE\" id=\"__VIEWSTATE\"", "__EVENTVALIDATION", "dnngo_gomenu", "dnngo_megamenue", "__RequestVerificationToken", "dnngo_megamenu", "csrf", "contact__item", "count", "token", "w-100 px-3 mb-0", "w-100 text-xs", "glide__container border-0", "nav-link text-light", "nav-link", "glide__slide", "move-on-hover", "ctl09_MenuView1", "ctl96_lblSiteHit", "ctl96_lblPage", "آمار بازدید", "آخرین بروز رسانی", "text-color", "product-item", "product-type-simple", "article class=\"post-item\"", "statistics-value", "block-content clear", "views-field views-field-field-image", "views-field views-field-title", "Challenge", "\"sid\" type=\"hidden\"", "OnlineUserCount", "BDC_CaptchaImageDiv", "BDC_SoundLink", "BotDetectCaptcha", "BDC_VCID_LoginCaptcha", "BDC_Hs_LoginCaptcha", "CaptchaImage.axd", "link rel='stylesheet'", "initResponsivePagination", "row vc_row wpb_row vc_row-fluid", "row vc_row vc_inner", "btn-bs-pagination next", "better-slider", "bsb-have-heading-color", "wpb_column", "bs-shortcode", "bs-slider-controls", ".bs-pretty-tabs-container", "digits_login_remember", "digits_reg_lastname", "login-field", "lostpasswordform", "remember-checkbox", "news-mul-content", "fa fa-angle-left font-13", "latest_posts-details-category", "form-control-static", "ctl00_rssmStyleSheet_TSSM", "TSM_CombinedScripts", "ctl00_cphMiddleTabs_Sampa", "ctl01_ctl09_Title", "wd-info-box"}
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
	//counter := 0
	dt := time.Now()
	old_hash := ""
	status_code, page := get_page(domain)
	for status_code == -1 {
		status_code, page = get_page(domain)
		time.Sleep(30 * time.Second)
	}

	cleaned_string := remove_false_changes(page)
	for status_code != 200 {
		if status_code >= 500 {
			fmt.Fprintf(os.Stdout, "%s [alert] %s domain -> %s  , status_code: %d \n%s", ErrorColor, dt.String(), domain, status_code, colorNone)
		} else {
			fmt.Fprintf(os.Stdout, "%s[alert] %s domain -> %s  , status_code: %d \n%s", InfoColor, dt.String(), domain, status_code, colorNone)
		}
		// alert_sound()
		time.Sleep(30 * time.Second)
		status_code, page = get_page(domain)
		dt = time.Now()
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
			time.Sleep(30 * time.Second)
		}
		for status_code != 200 {
			if status_code >= 500 {
				fmt.Fprintf(os.Stdout, "%s [alert] %s domain -> %s  , status_code: %d \n%s", ErrorColor, dt.String(), domain, status_code, colorNone)
			} else {
				fmt.Fprintf(os.Stdout, "%s[alert] %s domain -> %s  , status_code: %d \n%s", InfoColor, dt.String(), domain, status_code, colorNone)
			}

			// alert_sound()
			time.Sleep(30 * time.Second)
			status_code, page = get_page(domain)
			dt = time.Now()
		}
		cleaned_string = remove_false_changes(page)
		hasher.Write([]byte(cleaned_string))
		new_hash := hasher.Sum(nil)
		new_hash_string := hex.EncodeToString(new_hash)
		if old_hash != new_hash_string {
			dt = time.Now()
			fmt.Fprintf(os.Stdout, "%s [alert] %s  domain ->  %s changed  status_code: %d [alert]\n%s", ErrorColor, dt.String(), domain, status_code, colorNone)

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
		} else if strings.Contains(lowercase_page, "مرگ بر") {
			dt = time.Now()
			fmt.Printf("[alert] %s  domain ->  %s hack keyword detected: %s  status_code: %d [alert]\n", dt.String(), domain, "مرگ", status_code)
			alert_sound()
		} else if strings.Contains(lowercase_page, "deface") {
			dt = time.Now()
			fmt.Printf("[alert] %s  domain ->  %s hack keyword detected: %s  status_code: %d [alert]\n", dt.String(), domain, "deface", status_code)
			alert_sound()
		}

		dt = time.Now()
		// fmt.Printf("[+] %s  domain -> %s analyzed status_code: %d \n", dt.String(), domain, status_code)
		old_hash = hex.EncodeToString(new_hash)
		// fmt.Printf("domain: %s hash: %s\n", domain, old_hash)
		time.Sleep(5 * time.Second)
	}
}

func main() {

	select {}
}

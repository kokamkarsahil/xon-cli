package checker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/crypto/sha3"
)

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

type ApiResponse struct {
	Breaches [][]string `json:"breaches"`
}

func sha3Hash(input string) string {
	hash := sha3.NewLegacyKeccak512()
	_, _ = hash.Write([]byte(input))
	sha3 := hash.Sum(nil)
	return fmt.Sprintf("%x", sha3)[:10]
}

func CheckEmailExposure(email string) (bool, [][]string, error) {
	if !valid(email) {
		log.Warn("Email Is Inavlid")
		return true, nil, nil
	}
	url := fmt.Sprintf("https://private-anon-a37cc03621-xposedornot.apiary-proxy.com/v1/check-email/%s", email)
	resp, err := http.Get(url)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, err
	}

	var result ApiResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, nil, err
	}

	if len(result.Breaches) > 0 && len(result.Breaches[0]) > 0 {
		// If Email is exposed in at least one data breach
		return true, result.Breaches, nil
	} else {
		// No breaches found for the provided email
		return false, nil, nil
	}
}

func IsPasswordSafe(password string) bool {
	pwdHashStr := sha3Hash(password)
	apiURL := "https://passwords.xposedornot.com/api/v1/pass/anon/" + url.QueryEscape(pwdHashStr)
	var resp *http.Response
	var err error
	retryCount := 0
	maxRetries := 3
	retryDelay := 5 * time.Second

	for {
		resp, err = http.Get(apiURL)
		if err != nil {
			fmt.Println("Request failed:", err)
			return false
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				delay, _ := strconv.Atoi(retryAfter)
				time.Sleep(time.Duration(delay) * time.Second)
			} else {
				time.Sleep(retryDelay)
			}
			retryCount++
			if retryCount > maxRetries {
				fmt.Println("Max retries reached.")
				break
			}
		} else if resp.StatusCode == http.StatusNotFound {
			return true
		} else {
			break
		}
	}
	defer resp.Body.Close()

	return false
}

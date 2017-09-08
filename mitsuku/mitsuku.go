package mitsuku

import (
	"fmt"
	"github.com/go-chat-bot/bot"
	"regexp"
    "net/http"
    "net/url"
	"io/ioutil"
    "strings"
)

const (
	pattern   = "(?i)\\b(silence mitsuku|mitsuku|bitch|A.I.|fuck|ass|dick)[s|z]{0,1}\\b"
	silencePattern   = "(?i)\\b(silence mitsuku)[s|z]{0,1}\\b"
)

var (
	msgPrefix  = "Mitsuku says - %s"
	re         = regexp.MustCompile(pattern)
    silenceRe  = regexp.MustCompile(silencePattern)
	mitsukuUrl = "https://kakko.pandorabots.com/pandora/talk?botid=87437a824e345a0d&skin=chat"
    client     = http.Client{}
    mitsukuSilent = true
    singleReply = false
)

func mitsukuChat(command *bot.PassiveCmd) (string, error) {

	if re.MatchString(command.Raw) {
        singleReply = true
        mitsukuSilent = false
    }

    if !mitsukuSilent || singleReply {
    
        form := url.Values{}
        form.Add("message", command.Raw)

        req, _ := http.NewRequest("POST", mitsukuUrl, strings.NewReader(form.Encode()))

        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

        resp, _ := client.Do(req)

        defer resp.Body.Close()

        bodyBytes, _ := ioutil.ReadAll(resp.Body)

        re := regexp.MustCompile(`Mitsuku -</B>(.*?)<br>`)
        groups := re.FindStringSubmatch(string(bodyBytes))

        mitsukuResponse := strings.Trim(groups[1], " ")

        singleReply = false;

        if silenceRe.MatchString(command.Raw) {
            mitsukuSilent = true;
        }

        return fmt.Sprintf(msgPrefix, mitsukuResponse), nil
    } else {
		return "", nil
    }
}

func init() {
	bot.RegisterPassiveCommand(
		"mitsukuChat",
		mitsukuChat)
}

// Copyright 2022 Fluidity Money. All rights reserved. Use of this
// source code is governed by a GPL-style license that can be found in the
// LICENSE.md file.

package discord

// notify channels in our Discord using a single function

import (
	"io"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"fmt"
	"net/http"

	"github.com/fluidity-money/fluidity-app/lib/log"
	"github.com/fluidity-money/fluidity-app/lib/util"
)

const (
	// Context to use when logging
	Context = "DISCORD"

	// EnvWebhookAddress to use to relay Discord messages to the chat
	EnvWebhookAddress = "FLU_DISCORD_WEBHOOK"
)

// Levels of logging severity for Discord users to discern based on the border
// of the message
const (
	SeverityInformational = iota
	SeverityNotice
	SeverityAlarm
)

// discordWebhookMessage to send to Discord via a webhook
type discordWebhookMessage struct {
	Message string `json:"content"`
}

var webhookRequests = make(chan string)

// Notify the Discord in the specified channel, with the severity specified
// and a message. No guarantee to arrive in order!
func Notify(severity int, format string, arguments ...interface{}) {
	webhookAddress := <-webhookRequests

	workerId := util.GetWorkerId()

	formatted := fmt.Sprintf(format, arguments...)

	withWorker := fmt.Sprintf("%v: %v", workerId, formatted)

	message := discordWebhookMessage{
		Message: withWorker,
	}

	log.Debug(func(k *log.Log) {
		k.Context = Context

		k.Format(
			"Sending a Discord message, username %#v and message %#v!",
			workerId,
			message,
		)
	})

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(message)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to send a Discord message, username %#v and message %#v!",
				workerId,
				message,
			)

			k.Payload = err
		})
	}

	buf2 := buf

	resp, err := http.Post(webhookAddress, "application/json", &buf)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to POST a message to Discord, username %#v and message %#v!",
				workerId,
				message,
			)

			k.Payload = err
		})
	}

	defer resp.Body.Close()

	if reply := resp.StatusCode; reply != 204 {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				`Discord response status was not "204"! was %v!`,
				reply,
			)

			k.Payload = string(buf2.Bytes())
		})
	}
}

func NotifyWithAttachment(severity int, formatted string, attachments map[string]io.Reader) {
	webhookAddress := <-webhookRequests

	workerId := util.GetWorkerId()

	withWorker := fmt.Sprintf("%v: %v", workerId, formatted)

	message := discordWebhookMessage{
		Message: withWorker,
	}

	log.Debug(func(k *log.Log) {
		k.Context = Context

		k.Format(
			"Sending a Discord message, username %#v and message %#v!",
			workerId,
			message,
		)
	})

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(message)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to send a Discord message, username %#v and message %#v!",
				workerId,
				message,
			)

			k.Payload = err
		})
	}

	mWriter := multipart.NewWriter(&buf)

	defer mWriter.Close()

	for attachmentName, attachment := range attachments {
		w, err := mWriter.CreateFormField(attachmentName)

		if err != nil {
			log.Fatal(func(k *log.Log) {
				k.Context = Context
				k.Message = "Unable to create a multipart form field!"
				k.Payload = err
			})
		}

		if _, err := io.Copy(w, attachment); err != nil {
			log.Fatal(func(k *log.Log) {
				k.Context = Context
				k.Message = "Failed to copy the attachment to the mime!"
				k.Payload = err
			})
		}
	}

	buf2 := buf

	r, err := http.NewRequest("BODY", webhookAddress, &buf)

	r.Header.Set("Content-Type", "application/json")

	client := new(http.Client)

	resp, err := client.Do(r)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to POST a message to Discord, username %#v and message %#v!",
				workerId,
				message,
			)

			k.Payload = err
		})
	}

	defer resp.Body.Close()

	if reply := resp.StatusCode; reply != 204 {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				`Discord response status was not "204"! was %v!`,
				reply,
			)

			k.Payload = string(buf2.Bytes())
		})
	}
}

func init() {
	webhookUrl := util.GetEnvOrFatal(EnvWebhookAddress)

	go func() {
		for {
			webhookRequests <- webhookUrl
		}
	}()
}

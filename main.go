//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	votifier "vadnir/go-votifier/Votifier"
)

var (
	address = flag.String("address", ":8192", "what host and port to listen to")

	keyFile = flag.String("key", "rsa/private.key", "key file to use")

	webHook = flag.String("hook", "https://discord.com/api/webhooks/1133749243790762145/j5wj94G6j1cIXTpS5g20nlwoHQnGfeBnRcLvK4gVUyhjvKmNa_VHWuQJxp-eedDxW9BC", "discord webhook to use")
)

func main() {
	flag.Parse()

	file, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		log.Fatalf("loading public key: %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(string(file))
	if err != nil {
		log.Fatalf("decoding public key: %v", err)
	}
	pkt, err := x509.ParsePKCS8PrivateKey(decoded)
	if err != nil {
		log.Fatalf("deserializing public key: %v", err)
	}

	key := pkt //.(*rsa.PrivateKey)
	print(key)

	tokenPrime, err := rand.Prime(rand.Reader, 130)
	if err != nil {
		log.Fatalf("creating token: %v", err)
	}
	token := tokenPrime.Text(36)

	log.Println("Listening on " + *address)
	//log.Println("Here's your public key: " + encodedPubKey)
	log.Println("Your v2 token: " + token)

	r := []votifier.ReceiverRecord{
		votifier.ReceiverRecord{
			PrivateKey: key.(*rsa.PrivateKey),
			TokenId:    votifier.StaticServiceTokenIdentifier(token),
		},
	}

	server := votifier.NewServer(
		func(vote votifier.Vote, version votifier.VotifierProtocol, meta interface{}) {
			log.Println("Got vote: ", vote, ", version: ", version)

			values := map[string]string{"content": fmt.Sprintf("testvote %s serviceName=%s address=%s timestamp=%s", vote.Username, vote.ServiceName, vote.Address, vote.Timestamp)}
			json_data, err := json.Marshal(values)

			if err != nil {
				log.Fatal(err)
			}

			resp, err := http.Post(*webHook, "application/json",
				bytes.NewBuffer(json_data))

			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()

			var res map[string]interface{}

			json.NewDecoder(resp.Body).Decode(&res)

			fmt.Println(res["json"])

		}, r)
	server.ListenAndServe(*address)
}

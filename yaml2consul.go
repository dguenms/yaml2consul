package main

import (
    "fmt"
    "os"
    "github.com/kylelemons/go-gypsy/yaml"
    "github.com/hashicorp/consul/api"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/kms"
    "github.com/aws/aws-sdk-go/aws"
)

type Client struct {
    kv *api.KV
    kms *kms.KMS
    key string
}

func newClient(profile string, region string, key string) *Client {
    c := new(Client)

    client, err := api.NewClient(api.DefaultConfig())
    if err != nil {
        panic(err)
    }

    c.kv = client.KV()

    sess, err := session.NewSessionWithOptions(session.Options {
        Config: aws.Config {
            Region: aws.String(region),
        },
        Profile: profile,
    })
    if err != nil {
        panic(err)
    }

    c.kms = kms.New(sess)

    c.key = key

    return c
}

func flatten(m map[string]string, node yaml.Node, key string) (map[string]string) {
    switch c := node.(type) {
        case yaml.Map:
            for k, v := range c {
                flatten(m, v, fmt.Sprintf("%s/%s", key, k))
            }
        case yaml.List:
            for i, e := range c {
                flatten(m, e, fmt.Sprintf("%s/%d", key, i))
            }
        case yaml.Scalar:
            m[key[1:]] = c.String()
    }

    return m
}

func parse(file* yaml.File) (map[string]string) {
    return flatten(make(map[string]string), file.Root, "")
}

func (client *Client) put(m map[string]string) {
    for k, v := range m {
        kvpair := &api.KVPair{Key: k, Value: client.encrypt(v)}
        _, err := client.kv.Put(kvpair, nil)
        if err != nil {
            panic(err)
        }
    }
}

func (client *Client) lookup(prefix string) {
    kvpairs, _, err := client.kv.List(prefix, nil)
    if err != nil {
        panic(err)
    }

    for _, kvpair := range kvpairs {
        fmt.Printf("%v -> %v\n", kvpair.Key, string(client.decrypt(kvpair.Value)))
    }
}

func (client *Client) encrypt(plaintext string) []byte {
    params := &kms.EncryptInput {
        KeyId:     aws.String(client.key),
        Plaintext: []byte(plaintext),
    }
    resp, err := client.kms.Encrypt(params)

    if err != nil {
        panic(err)
    }

    return resp.CiphertextBlob
}

func (client *Client) decrypt(ciphertext []byte) []byte {
    params := &kms.DecryptInput {
        CiphertextBlob: []byte(ciphertext),
    }
    resp, err := client.kms.Decrypt(params)

    if err != nil {
        panic(err)
    }

    return resp.Plaintext
}

func main() {
    if len(os.Args) < 2 {
        panic("Usage: yaml2consul <yaml file>")
    }

    filename := os.Args[1]
    file, err := yaml.ReadFile(filename)

    if err != nil {
        panic(err)
    }

    flattened_map := parse(file)

    for k, v := range flattened_map {
        fmt.Printf("%#v -> %#v\n", k, v)
    }

    client := newClient("learning", "us-west-2", "bd616dec-a26a-4200-a2fd-6898c7e5c0d5")

    client.put(flattened_map)

    client.lookup("")
}

package main

import (
    "fmt"
    "os"
    "github.com/kylelemons/go-gypsy/yaml"
    "github.com/hashicorp/consul/api"
)

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

func put(m map[string]string) {
    client, err := api.NewClient(api.DefaultConfig())
    if err != nil {
        panic(err)
    }

    kv := client.KV()

    for k, v := range m {
        kvpair := &api.KVPair{Key: k, Value: []byte(v)}
        _, err = kv.Put(kvpair, nil)
        if err != nil {
            panic(err)
        }
    }
}

func lookup(prefix string) {
    client, err := api.NewClient(api.DefaultConfig())
    if err != nil {
        panic(err)
    }

    kv := client.KV()

    kvpairs, _, err := kv.List(prefix, nil)
    if err != nil {
        panic(err)
    }

    for _, kvpair := range kvpairs {
        fmt.Printf("%v -> %v\n", kvpair.Key, string(kvpair.Value))
    }
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

    put(flattened_map)

    lookup("")
}

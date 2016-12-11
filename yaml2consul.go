package main

import (
    "fmt"
    "os"
    "github.com/kylelemons/go-gypsy/yaml"
)

func nodeToMap(node yaml.Node) (yaml.Map) {
    m, ok := node.(yaml.Map)
    if !ok {
        panic(fmt.Sprintf("%v is not of type map", node))
    }
    return m
}

func nodeToList(node yaml.Node) (yaml.List) {
    m, ok := node.(yaml.List)
    if !ok {
        panic(fmt.Sprintf("%v is not of type list", node))
    }
    return m
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

func main() {
    filename := os.Args[1]
    file, err := yaml.ReadFile(filename)

    if err != nil {
        panic(err)
    }

    basemap := file.Root.(yaml.Map)

    for k, v := range basemap {
        fmt.Printf("%#v -> %#v\n", k, v)
    }

    flattened_map := parse(file)

    for k, v := range flattened_map {
        fmt.Printf("%#v -> %#v\n", k, v)
    }
}

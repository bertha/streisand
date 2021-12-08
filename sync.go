package main

type PrefixQuery struct {
    Prefix Prefix
    Depth uint
}

type Query struct {
    Got []Hash
    Prefixes []PrefixRequest
}

type QueryResponse struct {
    Got []bool
    Prefixes []map[int]Hash
}


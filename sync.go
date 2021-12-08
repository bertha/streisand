package main

type PrefixQuery struct {
    Prefix Prefix
    Depth uint
}

type Query struct {
    Got []Hash
    Prefixes []PrefixQuery
}

type QueryResponse struct {
    Got []bool
    Prefixes []map[int]Hash
}


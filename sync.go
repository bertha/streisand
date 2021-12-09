package main

type Notification struct {
	XorsIGot           []XorTeabag
	HashesYouMightLack []Hash
}

type XorTeabag struct {
	prefix Prefix
	xors   []byte
}

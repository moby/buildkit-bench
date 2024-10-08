package testutil

var pins = map[string]map[string]string{
	// busybox 1.36
	"busybox:latest": {
		"amd64":   "sha256:023917ec6a886d0e8e15f28fb543515a5fcd8d938edb091e8147db4efed388ee",
		"arm64v8": "sha256:1fa89c01cd0473cedbd1a470abb8c139eeb80920edf1bc55de87851bfb63ea11",
		"library": "sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79",
	},
	// alpine 3.18
	"alpine:latest": {
		"amd64":   "sha256:25fad2a32ad1f6f510e528448ae1ec69a28ef81916a004d3629874104f8a7f70",
		"arm64v8": "sha256:e3bd82196e98898cae9fe7fbfd6e2436530485974dc4fb3b7ddb69134eda2407",
		"library": "sha256:82d1e9d7ed48a7523bdebc18cf6290bdb97b82302a8a9c27d4fe885949ea94d1",
	},
	// golang 1.22 alpine 3.20
	"golang:1.22-alpine": {
		"amd64":   "sha256:51c59ce1d82286f8c6498ad78f97528fc7896fcff59997bd02b4b76c7f4979ca",
		"arm64v8": "sha256:ffadbf655b022c09e1fe1a14d2026cab688978a43d580c6e971ea2790cfaf212",
		"library": "sha256:48eab5e3505d8c8b42a06fe5f1cf4c346c167cc6a89e772f31cb9e5c301dcf60",
	},
	// python 3.12.6-bookworm
	"python:latest": {
		"amd64":   "sha256:cd07fcc5721f0d1ae2097291a30315176a997d8819e278827be0e090ba187bd2",
		"arm64v8": "sha256:d171522f62f530a7044d501539b97ebf093807864d821cf1ca110ade3b1dcff0",
		"library": "sha256:fcad5ffb670a9f1edc5cc232b2b321e617aaaae1a22c54242964178e408e0057",
	},
}

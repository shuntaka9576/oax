module github.com/shuntaka9576/oax

go 1.18

require (
	github.com/alecthomas/kong v0.7.1
	github.com/pelletier/go-toml v1.9.5
	github.com/r3labs/sse/v2 v2.10.0
)

require (
	golang.org/x/net v0.0.0-20191116160921-f9c825593386 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
)

replace github.com/r3labs/sse/v2 => ./sse

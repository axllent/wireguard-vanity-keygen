# WireGuard vanity keygen

[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/wireguard-vanity-keygen)](https://goreportcard.com/report/github.com/axllent/wireguard-vanity-keygen)

A command-line vanity (public) key generator for [WireGuard](https://www.wireguard.com/). It only matches the prefix of generated public keys, and not whether the search matches anywhere in the public key. The concept is based on [wireguard-vanity-address](https://github.com/warner/wireguard-vanity-address), however I wanted something a little more streamlined.


## Features

- Generates compliant [curve25519](https://cr.yp.to/ecdh.html) private and public keys
- Configurable multi-core processing (defaults to all cores)
- Optional case sensitive searching
- Search multiple prefixes at once
- Exit after results limit reached (defaults to 1)
- Displays probability and estimated runtime based on quick benchmark


## Usage options

```
Usage: wireguard-vanity-keygen [OPTIONS] <SEARCH> [<SEARCH>...]

Options:
  -c, --case-sensitive   case sensitive match (default false)
  -t, --threads int      threads (default 4)
  -l, --limit int        limit results to n (exists after) (default 1)
```


## Example

```
$ wireguard-vanity-keygen -l 4 test pc1/ 
Calculating speed: 49,950 calculations per second using 4 CPU cores
Case-insensitive search, exiting after 4 results
Probability for "test": 1 in 2,085,136 (approx 41 seconds per match)
Probability for "pc1/": 1 in 5,914,624 (approx 1 minute per match)

Press Ctrl-c to cancel

private OFVUjUoTNQp94fNPB9GCLzxiJPTbN03rcDPrVd12uFc=   public tEstMXL/3ZzAd2TnVlr1BNs/+eOnKzSHpGUnjspk3kc=
private gInIEDmENYbyuaWR1W/KLfximExwbcCg45W2WOmEc0I=   public TestKmA/XVagDW/JsHBXk5mhYJ6E1N1lAWeIeCttgRs=
private yDQLNiQlfnMGhUBsbLQjoBbuNezyHug31Qa1Ht6cgkw=   public PC1/3oUId241TLYImJLUObR8NNxz4HXzG4z+EazfWxY=
private +CUqn4jcKoL8pw53pD4IzfMKW/IMceDWKcM2W5Dxtn4=   public teStmGXZwiJl9HmfnTSmk83girtiIH8oZEa6PFJ8F1Y=
private 2G0X+IvBLw3NRfRnHb8diIXp96NQ9wSu4gdqPidy3nw=   public tESt3DBU40Q/Zkp0d1aeb6HOgEOsEM3BxzNqLckKhhc=
private EMaUfQvAEABpQV/21ALJP5YtyGerRXAn8u67j2AQzVs=   public pC1/t2x5V99Y1SBqNgPZDPsa6r+L5y3BJ4XUCJMar3g=
private wNuHOKCfoH1emfvijXNBoc/7KjrEXUeof7tSdGWvRFo=   public PC1/jXQosaBad2HePOm/w1KjCZ82eT3qNbfzNDZiwTs=
private 8IdcNsman/ZRGvqWzw1e5cRfhhdtAAmk02X9TkQxhHI=   public pC1/N8coOcXmcwO09QXxLrF5/BoHQfvp/qsysGPXiw0=
```

## Installing

Download the [latest binary release](https://github.com/axllent/wireguard-vanity-keygen/releases/latest) for your system, 
or build from source `go get -u github.com/axllent/wireguard-vanity-keygen`(go >= 1.11 required)


## FAQ

### What characters can I search for?

Valid characters include `A-Z`, `a-z`, `0-9`, `/` and `+`. There are no other characters in a hash.


### Why does `test` & `tes1` show different probabilities despite having 4 characters each?

With case-insensitive searches (default), a-z have the chance of matching both uppercase and lowercase. A search for "cat" can match `Cat`, `cAT` etc.


### How accurate are the estimated times?

They are not (and cannot be) accurate. Keys are completely randomly generated, and the estimate is based on a law of averages. For instance, you could find a match for a one in a billion chance on the very first hit, or it could take you 5 billion attempts. It will however give you an indication based on your CPU speed, word count, case sensitivity, and use of numbers or characters.


### Why do I need this?

You don't. I wrote it because I run a WireGuard server, which does not provide any reference as to who the key belongs to (`wg` on the server). Using vanity keys, I can at least identify connections. I also wanted to learn more about multi-core processing in Golang.

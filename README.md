# WireGuard vanity keygen

[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/wireguard-vanity-keygen)](https://goreportcard.com/report/github.com/axllent/wireguard-vanity-keygen)

A command-line vanity (public) key generator for [WireGuard](https://www.wireguard.com/). It only matches the prefix of generated public keys, and not whether the search matches anywhere in the public key. The concept is based on [wireguard-vanity-address](https://github.com/warner/wireguard-vanity-address), however I wanted something a little more streamlined.


## Features

- Generates compliant [curve25519](https://cr.yp.to/ecdh.html) private and public keys
- Configurable multi-core processing (defaults to all cores)
- Optional case sensitive searching
- Optional regex searching
- Search multiple prefixes at once
- Exit after results limit reached (defaults to 1)
- Displays probability and estimated runtime based on quick benchmark


## Usage options

```
Usage: wireguard-vanity-keygen [OPTIONS] <SEARCH> [<SEARCH>...]

Options:
  -c, --case-sensitive   case sensitive match (default false)
  -t, --threads int      threads (defaults to all available cores minus 1)
  -l, --limit int        limit results to n (exists after) (default 1)
```


## Example

```
$ wireguard-vanity-keygen -l 3 test pc1/ "^pc7[+/]"
Calculating speed: 49,950 calculations per second using 4 CPU cores
Case-insensitive search, exiting after 4 results
Probability for "test": 1 in 2,085,136 (approx 41 seconds per match)
Probability for "pc1/": 1 in 5,914,624 (approx 1 minute per match)
Cannot calculate probability for the regular expression "^pc7[/+]"

Press Ctrl-c to cancel

private OFVUjUoTNQp94fNPB9GCLzxiJPTbN03rcDPrVd12uFc=   public tEstMXL/3ZzAd2TnVlr1BNs/+eOnKzSHpGUnjspk3kc=
private gInIEDmENYbyuaWR1W/KLfximExwbcCg45W2WOmEc0I=   public TestKmA/XVagDW/JsHBXk5mhYJ6E1N1lAWeIeCttgRs=
private yDQLNiQlfnMGhUBsbLQjoBbuNezyHug31Qa1Ht6cgkw=   public PC1/3oUId241TLYImJLUObR8NNxz4HXzG4z+EazfWxY=
private QIbJgxy83+F/1kdogcF+T04trs+1N9gAr1t5th2tLXM=   public Pc7+h172sx0TfIMikjgszM/B8i8/ghi7qJVOwWQtx0w=
private +CUqn4jcKoL8pw53pD4IzfMKW/IMceDWKcM2W5Dxtn4=   public teStmGXZwiJl9HmfnTSmk83girtiIH8oZEa6PFJ8F1Y=
private EMaUfQvAEABpQV/21ALJP5YtyGerRXAn8u67j2AQzVs=   public pC1/t2x5V99Y1SBqNgPZDPsa6r+L5y3BJ4XUCJMar3g=
private wNuHOKCfoH1emfvijXNBoc/7KjrEXUeof7tSdGWvRFo=   public PC1/jXQosaBad2HePOm/w1KjCZ82eT3qNbfzNDZiwTs=
private gJtn0woDChGvyN2eSdc7mTpAFA/nA6jykJeK5bYYfFA=   public Pc7+UEJSHiWsQ9zkO2q+guqDK4sc3VMDMgJu+h/bOFI=
private IMyPmYm/v0SPmB62hC8l6kfxT3/Lfp7dMioo+SM6T2c=   public Pc7/uVfD/ZftxWBHwYbaudEywUS61biBcpj5Tw830Q4=
```

## Timings

To give you a rough idea of how long it will take to generate keys, the following table lists
estimated timings for each match on a system that reported  "`Calculating speed: 230,000 calculations per second using 19 CPU cores`" when it started:

| Length  | Case-insensitive | Case-sensitive |
| :------ | :--------------- | :------------- |
| 3 chars | 0 seconds        | 1 second       |
| 4 chars | 9 seconds        | 1 minute       |
| 5 chars | 5 minutes        | 1.25 hours     |
| 6 chars | 4 hours          | 3.5 days       |
| 7 chars | 6 days           | 7 months       |
| 8 chars | 7 months         | 38 years       |
| 9 chars | 22 years         | 175 years      |

Note that the above timings are for finding a result for any search term.
Passing multiple search terms will not substantially increase the time,
but increasing the limit to two (`--limit 2`) will double the estimated time, three will triple the time, etc.

If any search term contains numbers, the timings would fall somewhere between the case-insensitive and case-sensitive columns.

Of course, your mileage will differ, depending on the number, and speed, of your CPU cores.

## Regular Expressions

Since each additional letter in a search term increasing the search time exponentially, searching by regular expression may
reduce the time considerably. Here are some examples:

1. `.*word.*` - find word anywhere in the key (`word.*` and `.*word` will also work)
2. `^.{0,10}word` - find word anywhere in the first 10 letters of the key
3. `word1.*word2` - find two words, anywhere in the key
4. `^[s5][o0][ll]ar` - find 'solar' or the visually similar 's01ar`, at the beginning of the key
5. `^(best|next)[/+]` - find 'best' or the 'next' best, at the beginning of the key, with `/` or `+` as a delimiter

A good guide on Go's regular expression syntax is at https://pkg.go.dev/regexp/syntax.

NOTE: If your search term contains shell metacharacters, such as `|`, or `^`, you will need to quote the search time.
On Windows, you must use double quotes (`"`), and not single quotes (`'`) when quoting a search term.

NOTE: It is possible to create regular expressions that will never match a key.
To guard against this, shorten your search term to use just one character in each section of your regular expression.
If you don't get a hit after a few minutes, assume the regular expression may never match.

## Installing

Download the [latest binary release](https://github.com/axllent/wireguard-vanity-keygen/releases/latest) for your system, 
or build from source `go install github.com/axllent/wireguard-vanity-keygen@latest`.


## FAQ

### What characters can I search for?

Valid characters include `A-Z`, `a-z`, `0-9`, `/` and `+`. There are no other characters in a hash.

You can also use regex expressions to search.

### Why does `test` & `tes1` show different probabilities despite having 4 characters each?

With case-insensitive searches (default), a-z have the chance of matching both uppercase and lowercase. A search for "cat" can match `Cat`, `cAT` etc.


### How accurate are the estimated times?

They are not (and cannot be) accurate. Keys are completely randomly generated, and the estimate is based on a law of averages. For instance, you could find a match for a one in a billion chance on the very first hit, or it could take you 5 billion attempts. It will however give you an indication based on your CPU speed, word count, case sensitivity, and use of numbers or characters.


### Why do I need this?

You don't. I wrote it because I run a WireGuard server, which does not provide any reference as to who the key belongs to (`wg` on the server). Using vanity keys, I can at least identify connections. I also wanted to learn more about multi-core processing in Golang.

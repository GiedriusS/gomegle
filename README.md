# Gomegle
A library written in Go for interacting with Omegle (omegle.com)

# Purpose
Create an easy to use library that fully implements the whole Omegle
protocol. Also, it must include an example client for basic
interaction with Omegle's chat

# Feature matrix
|                            | Implemented | Error checking |
|----------------------------|-------------|----------------|
| /stoplookingforcommonlikes | no          | no             |
| /generate                  | no          | no             |
| /send                      | yes         | yes            |
| /status                    | yes         | yes            |
| /typing                    | yes         | yes            |
| /stoppedtyping             | yes         | yes            |
| /disconnect                | yes         | yes            |
| /events                    | not fully   | yes            |

Some more features are missing from the table but they are kind of obscure so they will be implemented later.

# Credits
Part of this library is based on this awesome [document](https://gist.github.com/nucular/e19264af8d7fc8a26ece)

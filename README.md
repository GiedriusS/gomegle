# Gomegle
A library written in Go for interacting with Omegle (omegle.com) chat

# Purpose
Create an easy to use library that fully implements the whole Omegle chat
protocol. Also, it must include an example client for basic interaction

# Feature matrix
|                            | Implemented | Error checking |
|----------------------------|-------------|----------------|
| /stoplookingforcommonlikes | yes         | yes            |
| /generate                  | no          | no             |
| /send                      | yes         | yes            |
| /status                    | yes         | yes            |
| /typing                    | yes         | yes            |
| /stoppedtyping             | yes         | yes            |
| /disconnect                | yes         | yes            |
| /events                    | not fully   | yes            |
| /start                     | yes         | yes            |

Some more features are missing from the table but they are kind of obscure so they will be implemented later.

# Credits
Part of this library is based on this awesome [document](https://gist.github.com/nucular/e19264af8d7fc8a26ece)

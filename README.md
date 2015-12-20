# Gomegle
A library written in Go for interacting with Omegle (omegle.com) chat

# Purpose
Create an easy to use library that fully implements the whole Omegle chat
protocol. Also, it must include an example client for basic interaction

# Feature matrix
|                            | Implemented | Error checking |
|----------------------------|-------------|----------------|
| /stoplookingforcommonlikes | yes         | yes            |
| /generate                  | yes         | yes            |
| /send                      | yes         | yes            |
| /status                    | yes         | yes            |
| /typing                    | yes         | yes            |
| /stoppedtyping             | yes         | yes            |
| /disconnect                | yes         | yes            |
| /events                    | yes         | yes            |
| /start                     | yes         | yes            |
| /recaptcha                 | yes         | yes            |

# Credits
Part of this library is based on this awesome [document](https://gist.github.com/nucular/e19264af8d7fc8a26ece)

# FAQ
* How can I get "statusInfo" event from UpdateEvents()

You can't because a simple string slice isn't expressive enough to properly store info returned in that object. Thus, we ask you to explicitly ask for status information via GetStatus().

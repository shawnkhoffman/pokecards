# Pokecards CLI

- [How to run this program](#how-to-run-this-program)
  - [From binary](#from-binary)
  - [From source code](#from-source-code)
  - [From Docker](#from-docker)
- [Usage](#usage)
  - [Getting help](#getting-help)
  - [Drawing fewer cards](#drawing-fewer-cards)
  - [Setting the log file](#setting-the-log-file)
- [Telemetry](#telemetry)

This is a CLI program built using Go that generates a list of Pokémon cards that match a specific criteria. The program uses the [Pokémon TCG API](https://pokemontcg.io/) to retrieve the cards and OpenTelemetry to trace the HTTP requests.

## How to run this program

You have three options to use this program: from [binary](#from-binary), from [source code](#from-source-code), or from [Docker](#from-docker).

### From binary

This solution offers the best user experience. Simply download the compatible [binary](https://github.com/shawnkhoffman/pokecards/releases/latest) for your system and run it from the command line. Depending on your system, you may need to make the binary executable or configure your security settings (this is the case with [MacOS](https://support.apple.com/lv-lv/guide/mac-help/mh40616/mac)).

### From source code

You need to have Go installed on your system. You can download and install Go from the official website: [https://golang.org/dl/](https://golang.org/dl/).

```sh
go run main.go -limit=<number_of_cards> -log=<log_file_name>
```

### From Docker

This solution was written with the fewest test cases and is only being provided as a proof of concept. To use this solution, you need to have Docker installed on your system. You can download and install Docker from the official website: [https://www.docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop).

```sh
docker run -v "$(pwd):/app/logs" --rm shawnkhoffman/pokecards [OPTIONS...] | jq
```

## Usage

It is recommended that you have [jq](https://stedolan.github.io/jq/) installed on your system (for beautifying output). The rest of this README will assume as such.

### Getting help

```sh
$ ./pokecards --help

Usage of ./pokecards:
  -limit int
        the number of cards to retrieve (default 10)
  -log string
        the file to write logs to (default "./logs/pokecards.log")
```

### Drawing fewer cards

The `-limit` flag sets the number of cards to retrieve (default value is 10).

```sh
$ ./pokecards -limit 1 | jq
{
  "Cards": [
    {
      "ID": "base1-17",
      "Name": "Beedrill",
      "Types": [
        "Grass"
      ],
      "HP": "80",
      "Rarity": "Rare"
    }
  ]
}
```

### Setting the log file

The `-log` flag sets the name of the file to write logs to (default value is `./logs/pokecards.log`).

```sh
./pokecards -limit 1 -log foo.log | jq
{
  "Cards": [
    {
      "ID": "base1-17",
      "Name": "Beedrill",
      "Types": [
        "Grass"
      ],
      "HP": "80",
      "Rarity": "Rare"
    }
  ]
}
```

## Telemetry

This program uses [OpenTelemetry](https://opentelemetry.io/) to trace HTTP requests and writes telemetry data to a log file. The telemetry data includes information about the number of cards retrieved, the HTTP status code, and any errors that occurred during the request.

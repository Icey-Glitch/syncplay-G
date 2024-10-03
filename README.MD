# Syncplay Proxy & Server

This project implements a proxy server for the Syncplay protocol. It intercepts and processes messages between Syncplay clients and the Syncplay server, allowing for custom handling and logging of messages.

also a work in progres go syncplay server 

***very wip***
## Features

- Intercepts and logs messages between Syncplay clients and the server.
- Adds spaces after colons in JSON messages for better readability.
- Supports forwarding data between clients and the server.

## Project Structure

- `main.go`: Implements the Syncplay server and handles client connections.
- `tester.go`: Implements a proxy server that forwards data between clients and the Syncplay server.
- `flake.nix`: Nix flake configuration for setting up the development environment.

## Setup

### Prerequisites

- Go (Golang) installed on your system.
- Nix package manager installed on your system.

### Using Nix Flakes

1. **Install Nix**: Follow the instructions on the [Nix website](https://nixos.org/download.html) to install Nix.
2. **Enable Flakes**: Add the following lines to your `~/.config/nix/nix.conf` file:
    ```plaintext
    experimental-features = nix-command flakes
    ```

3. **Enter the Development Shell**:
    ```sh
    nix develop
    or 
    edit .envrc
    use the flake
    ```

## Usage

### Running the Syncplay Server

To run the Syncplay server, execute the following command:

```sh
go run ./src/main.go
```


### Debuging
you can also proxy the offical server or this server using the tester by seting the server to port 12345 and the client to 8080

```sh
go run ./src/tester.go
```
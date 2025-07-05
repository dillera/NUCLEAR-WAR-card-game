# Nuclear War Game Server

## Overview

This project is a Go-based server for the classic "Nuclear War" card game. It provides a RESTful API to manage game state, handle player actions, and enforce game rules. The server supports a full gameplay loop, including player elimination, a "Final Strike" retaliation mechanic for eliminated players, and victory conditions.

This server is designed to be a backend for any potential game client (web, mobile, or desktop). A simple Python test client is included to demonstrate and validate the server's functionality.

## Prerequisites

Before you begin, ensure you have the following installed:
- [Go](https://golang.org/doc/install) (version 1.18 or higher)
- [Python](https://www.python.org/downloads/) (version 3.8 or higher)

## Setup & Installation

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd nuclear-war-game-server
    ```

2.  **Set up the Go environment:**
    Download the necessary Go modules.
    ```bash
    go mod tidy
    ```

3.  **Set up the Python virtual environment:**
    Create and activate a virtual environment to manage Python dependencies.
    ```bash
    python3 -m venv venv
    source venv/bin/activate
    ```

4.  **Install Python dependencies:**
    Install the required `requests` library for the test client.
    ```bash
    pip install requests
    ```

## Running the Server

To start the game server, run the following command from the project's root directory:

```bash
go run main.go
```

The server will start and listen for requests on `http://localhost:8080`.

## Running the Test Client

The test client simulates a full two-player game, from creation to completion, and verifies the core game logic, including attacks, retaliation, and winner declaration.

**Important:** Make sure the server is running in a separate terminal before executing the client.

To run the test client, execute the following command from the project's root directory:

```bash
source venv/bin/activate  # If the virtual environment is not already active
python3 client/test_client.py
```

## API Endpoints

The server exposes the following REST API endpoints:

| Method | Path                               | Description                                                                                             |
| :----- | :--------------------------------- | :------------------------------------------------------------------------------------------------------ |
| `POST` | `/games`                           | Creates a new game instance and returns its initial state.                                              |
| `GET`  | `/games/{gameID}`                  | Retrieves the general game state. Add `?playerID=<id>` to get a player-specific view including their hand. |
| `POST` | `/games/{gameID}/join`             | Adds a new player to the game. Requires a JSON body: `{"playerName": "YourName"}`.                      |
| `POST` | `/games/{gameID}/start`            | Starts the game, deals cards to players, and sets the game state to `in_progress`.                      |
| `POST` | `/games/{gameID}/play`             | Allows a player to play a card from their hand. Requires a JSON body with `playerID`, `cardID`, `location`. |
| `POST` | `/games/{gameID}/attack`           | Executes an attack from one player to another. Requires a JSON body with `attackerID` and `targetID`.     |
| `POST` | `/games/{gameID}/pass`             | Allows the current player to pass their turn. Requires a JSON body: `{"playerID": "YourID"}`.             |

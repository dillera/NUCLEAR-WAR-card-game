# Nuclear War Game Server

This project is a Go-based server for the Nuclear War card game, with a Python-based ASCII client for interactive gameplay.

## How to Play

### 1. Start the Server

Navigate to the project's root directory and run the following command. This will start the game server on its default port (`:8080`).

```sh
go run .
```

The server will log when it starts and when players join or perform actions.

### 2. Play the Game

The game is played using the Python client. You will need at least two terminal windows to simulate a two-player game.

#### Dependencies

The client requires the `requests` library. Make sure it's installed:

```sh
pip install requests
```

#### Launching the Client

In each terminal window, navigate to the `client` directory and run the Python script:

```sh
cd client
python game_client.py
```

#### Creating and Joining a Game

1.  **First Player (Create Game):**
    *   When the client launches, choose `1` to create a new game.
    *   Enter a name for your player.
    *   The client will display a **Game ID**. Make a note of this ID.

2.  **Second Player (Join Game):**
    *   In the second terminal, launch the client and choose `2` to join an existing game.
    *   Enter the **Game ID** provided by the first player.
    *   Enter a name for your player.

3.  **Starting the Game:**
    *   Once two or more players have joined, the option to `start` the game will become available.
    *   Any player can type `start` (or the corresponding number) and press Enter to begin the game.

## Gameplay

The client provides a real-time view of the game state:
*   A list of all players and their status.
*   Whose turn it is.
*   Your hand of cards, including their IDs.
*   A list of available commands.

Follow the on-screen prompts to play cards, pass your turn, and lead your nation to victory!

## Testing

To run the full suite of unit and integration tests, run the following command from the project root:

```sh
go test ./...
```

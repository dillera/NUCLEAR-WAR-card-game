import curses
import requests
import json
import time
import textwrap

BASE_URL = "http://localhost:8080"

# --- Curses Helper Functions ---

def draw_menu(stdscr, title, options, selected_idx):
    h, w = stdscr.getmaxyx()
    stdscr.clear()
    stdscr.addstr(2, w//2 - len(title)//2, title)
    for i, option in enumerate(options):
        x = w//2 - len(option)//2
        y = h//2 - len(options)//2 + i
        if i == selected_idx:
            stdscr.attron(curses.color_pair(1))
            stdscr.addstr(y, x, option)
            stdscr.attroff(curses.color_pair(1))
        else:
            stdscr.addstr(y, x, option)
    stdscr.refresh()

def get_input(stdscr, prompt):
    h, w = stdscr.getmaxyx()
    stdscr.addstr(h-2, 2, prompt)
    stdscr.clrtoeol()
    curses.echo()
    s = stdscr.getstr(h-2, 2 + len(prompt)).decode('utf-8').strip()
    curses.noecho()
    return s

def draw_card(win, y, x, card):
    card_height, card_width = 7, 22
    card_win = win.derwin(card_height, card_width, y, x)
    card_win.box()
    card_type = card.get('type', 'Unknown')
    card_win.addstr(1, 2, f"[{card_type}]", curses.A_BOLD)
    
    name = card.get('name', 'N/A')
    wrapped_name = textwrap.wrap(name, card_width - 4)
    for i, line in enumerate(wrapped_name):
        if i < 2: # Max 2 lines for name
            card_win.addstr(2 + i, 2, line)

    if card_type == 'Warhead':
        yield_val = card.get('yield', 0)
        card_win.addstr(5, 2, f"Yield: {yield_val} MT")
    elif card_type == 'Delivery System':
        pass # No extra info needed
    elif card_type == 'Secret':
        card_win.addstr(4, 2, "(Secret)")

    win.refresh()

def draw_game_state(stdscr, game_state, player_id):
    stdscr.clear()
    h, w = stdscr.getmaxyx()
    
    # Game Info
    stdscr.addstr(1, 2, f"Game ID: {game_state.get('id', 'N/A')} | Your Player ID: {player_id}")
    stdscr.addstr(2, 2, f"Game State: {game_state.get('state', 'N/A')}")
    
    # Players
    players = game_state.get('players', {})
    stdscr.addstr(4, 2, "Players:", curses.A_BOLD)
    for i, p in enumerate(players.values()):
        elim_status = " (Eliminated)" if p.get('is_eliminated') else ""
        pop_str = f"{p.get('population', 0):,}"
        player_line = f"- {p.get('name', 'N/A')} (Pop: {pop_str}){elim_status}"
        if p.get('id') == game_state.get('current_turn_player_id'):
            player_line += " <-- CURRENT TURN"
        stdscr.addstr(5 + i, 4, player_line)

    # Player Hand
    hand = game_state.get('player_hand', [])
    stdscr.addstr(h - 12, 2, "Your Hand:", curses.A_BOLD)
    if not hand:
        stdscr.addstr(h - 10, 4, "(No cards in hand)")
    else:
        for i, card in enumerate(hand):
            draw_card(stdscr, h - 10, 2 + (i * 24), card)

    stdscr.addstr(h - 3, 2, "'q' to quit | Enter command:")
    stdscr.refresh()

# --- API Functions ---

def create_game():
    res = requests.post(f"{BASE_URL}/games")
    res.raise_for_status()
    return res.json()

def join_game(game_id, player_name):
    res = requests.post(f"{BASE_URL}/games/{game_id}/join", json={"playerName": player_name})
    res.raise_for_status()
    return res.json()

def get_game_state(game_id, player_id):
    res = requests.get(f"{BASE_URL}/games/{game_id}?playerID={player_id}")
    res.raise_for_status()
    return res.json()

def start_game(game_id):
    res = requests.post(f"{BASE_URL}/games/{game_id}/start")
    res.raise_for_status()
    return res.json()

def post_command(game_id, player_id, command, args):
    payload = {"playerID": player_id, **args}
    res = requests.post(f"{BASE_URL}/games/{game_id}/{command}", json=payload)
    # Don't raise for status, as we want to handle game-specific errors
    return res

# --- Main Application Logic ---

def main(stdscr):
    curses.curs_set(0)
    stdscr.nodelay(1) # non-blocking input
    curses.start_color()
    curses.init_pair(1, curses.COLOR_BLACK, curses.COLOR_WHITE)

    # Main Menu
    menu_options = ["Create New Game", "Join Existing Game"]
    current_option = 0
    while True:
        draw_menu(stdscr, "NUCLEAR WAR", menu_options, current_option)
        key = stdscr.getch()
        if key == curses.KEY_UP and current_option > 0:
            current_option -= 1
        elif key == curses.KEY_DOWN and current_option < len(menu_options) - 1:
            current_option += 1
        elif key == curses.KEY_ENTER or key in [10, 13]:
            break

    stdscr.clear()
    try:
        if current_option == 0: # Create Game
            game_info = create_game()
            game_id = game_info['id']
            player_name = ""
            while not player_name:
                player_name = get_input(stdscr, "Enter your name: ")
            player_info = join_game(game_id, player_name)
            player_id = player_info['id']
        else: # Join Game
            game_id = ""
            while not game_id:
                game_id = get_input(stdscr, "Enter Game ID: ")
            player_name = ""
            while not player_name:
                player_name = get_input(stdscr, "Enter your name: ")
            player_info = join_game(game_id, player_name)
            player_id = player_info['id']
    except requests.exceptions.RequestException as e:
        stdscr.addstr(10, 2, f"Error connecting to server: {e}")
        stdscr.addstr(12, 2, "Press any key to exit.")
        stdscr.nodelay(0)
        stdscr.getch()
        return

    # Game Loop
    while True:
        try:
            game_state = get_game_state(game_id, player_id)
            draw_game_state(stdscr, game_state, player_id)
        except requests.exceptions.RequestException:
            # Handle server disconnect gracefully
            pass

        key = stdscr.getch()
        if key == ord('q'):
            break
        elif key == curses.KEY_ENTER or key in [10, 13]:
            cmd_line = get_input(stdscr, "> ")
            parts = cmd_line.split()
            if not parts:
                continue
            
            command = parts[0].lower()
            args = {}
            if command == 'play':
                if len(parts) == 3:
                    args = {'cardID': parts[1], 'location': parts[2]}
                    post_command(game_id, player_id, 'play', args)
            elif command == 'attack':
                if len(parts) == 2:
                    args = {'attackerID': player_id, 'targetID': parts[1]}
                    post_command(game_id, player_id, 'attack', args)
            elif command == 'pass':
                post_command(game_id, player_id, 'pass', {})
            elif command == 'start':
                start_game(game_id)

        time.sleep(0.5) # Refresh rate

if __name__ == "__main__":
    curses.wrapper(main)

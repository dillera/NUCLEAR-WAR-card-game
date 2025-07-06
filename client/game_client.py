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
    # Set terminal state for text input
    curses.curs_set(1)  # Make cursor visible
    stdscr.nodelay(0)   # Use blocking input
    curses.echo()       # Echo characters to the screen

    stdscr.addstr(h-2, 2, prompt)
    stdscr.clrtoeol()
    stdscr.refresh()

    s = stdscr.getstr(h-2, 2 + len(prompt)).decode('utf-8').strip()

    # Restore terminal state for game loop
    curses.noecho()
    stdscr.nodelay(1)
    curses.curs_set(0)
    return s

def draw_card(win, y, x, card):
    card_height, card_width = 7, 22
    card_win = win.derwin(card_height, card_width, y, x)
    card_win.box()
    card_type = card.get('type', 'Unknown')
    card_win.addstr(1, 2, f"[{card_type}]", curses.A_BOLD)
    
    name = card.get('name', 'N/A')
    card_win.addstr(2, 2, card.get('name', 'Unnamed Card')[:card_width-4])

    # Display Card ID
    card_id = card.get('id', '')
    if card_id:
        card_win.addstr(3, 2, f"ID: {card_id[:10]}", curses.A_DIM) # Show first 10 chars of ID
    
    # Add description or other details
    desc = card.get('description', '')
    if card.get('type') == 'Warhead':
        desc = f"Yield: {card.get('value', 0)} MT"
    
    if desc:
        card_win.addstr(5, 2, desc[:card_width-4])
    win.refresh()

def draw_game_state(stdscr, game_state, player_id):
    stdscr.clear()
    h, w = stdscr.getmaxyx()

    # --- Section 1: Game and Player Info ---
    stdscr.addstr(1, 2, f"Game ID: {game_state.get('gameID', 'N/A')} | Your Player ID: {player_id[:8]}...", curses.A_BOLD)
    stdscr.addstr(2, 2, f"Game State: {game_state.get('state', 'N/A')}")

    # --- Section 2: Player List ---
    stdscr.addstr(4, 2, "Players:", curses.A_BOLD)
    player_y = 5
    all_players = []
    if game_state.get('playerName'):
        all_players.append({
            'id': player_id,
            'name': f"{game_state.get('playerName')} (You)",
            'population': game_state.get('playerPopulation'),
            'isEliminated': game_state.get('isEliminated', False)
        })
    all_players.extend(game_state.get('opponents', []))

    for p in all_players:
        pop = p.get('population', 0)
        elim_status = "(ELIMINATED)" if p.get('isEliminated') else f"(Pop: {pop:,})"
        player_info = f"- {p.get('name', 'Unknown')} {elim_status}"

        attr = curses.A_NORMAL
        if p.get('id') == game_state.get('currentTurnPlayerId'):
            attr = curses.A_BOLD
        if p.get('isEliminated'):
            attr |= curses.A_DIM

        stdscr.addstr(player_y, 4, player_info[:w-5], attr)
        player_y += 1

    # --- Section 3: Turn Indicator / Game Ready Message ---
    y_offset = player_y + 2 # Dynamic positioning with extra space
    if game_state.get('state') == 'waiting_for_players' and len(all_players) >= 2:
        start_msg = "*** GAME READY TO START! Type 'start' to begin ***"
        stdscr.attron(curses.A_BOLD | curses.color_pair(1))
        stdscr.addstr(y_offset, max(2, (w - len(start_msg)) // 2), start_msg)
        stdscr.attroff(curses.A_BOLD | curses.color_pair(1))
    elif game_state.get('state') not in ['waiting_for_players', 'game_over']:
        current_turn_player_id = game_state.get('currentTurnPlayerId')
        if current_turn_player_id == player_id:
            turn_msg = "It's YOUR turn!"
            stdscr.attron(curses.A_BOLD | curses.color_pair(1))
            stdscr.addstr(y_offset, max(2, (w - len(turn_msg)) // 2), turn_msg)
            stdscr.attroff(curses.A_BOLD | curses.color_pair(1))
        else:
            current_player_name = game_state.get('currentTurnPlayer', 'Unknown')
            wait_msg = f"Waiting for {current_player_name} to play..."
            stdscr.addstr(y_offset, max(2, (w - len(wait_msg)) // 2), wait_msg)

    # --- Section 4: Player's Hand ---
    hand_y = h - 10
    stdscr.addstr(hand_y - 2, 2, "Your Hand:", curses.A_BOLD)
    hand = game_state.get('playerHand', [])
    if not hand:
        stdscr.addstr(hand_y, 4, "(No cards in hand)")
    else:
        card_display_width = 24
        max_cards = (w - 4) // card_display_width
        cards_to_draw = hand[:max_cards]
        for i, card in enumerate(cards_to_draw):
            draw_card(stdscr, hand_y, 2 + (i * card_display_width), card)
        if len(hand) > max_cards:
            more_cards_x = 2 + (max_cards * card_display_width)
            if more_cards_x < w - 15:
                stdscr.addstr(hand_y + 8, more_cards_x, f"(+ {len(hand) - max_cards} more)")

    # --- Section 5: Available Commands ---
    cmd_y_start = h - 7
    stdscr.addstr(cmd_y_start - 2, 2, "Available Commands:", curses.A_BOLD)
    commands = game_state.get('availableCommands', [])
    if commands:
        cmd_idx = 0
        for cmd_data in commands:
            cmd_str = f"{cmd_idx+1}. {cmd_data['name']} - {cmd_data['description']}"
            stdscr.addstr(cmd_y_start -1 + cmd_idx, 4, cmd_str[:w-6])
            cmd_idx += 1
        
        if any(cmd['name'] == 'play' for cmd in commands):
            hint = "Play locations: active, deterrent_1, deterrent_2"
            stdscr.addstr(cmd_y_start -1 + cmd_idx + 1, 4, hint, curses.A_DIM)
    else:
        stdscr.addstr(cmd_y_start -1, 4, "(No commands available right now)")

    # --- Section 6: Final Prompt ---
    stdscr.addstr(h - 1, 2, "Press 'q' to quit | Press Enter to input a command")
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

            # Display the Game ID so it can be shared
            stdscr.clear()
            h, w = stdscr.getmaxyx()
            msg1 = f"Game Created! Your Game ID is: {game_id}"
            msg2 = "Share this ID with other players so they can join."
            msg3 = "Press any key to continue..."
            stdscr.addstr(h//2 - 2, w//2 - len(msg1)//2, msg1)
            stdscr.addstr(h//2, w//2 - len(msg2)//2, msg2)
            stdscr.addstr(h//2 + 2, w//2 - len(msg3)//2, msg3)
            stdscr.refresh()
            stdscr.nodelay(0)  # Wait for user to press a key
            stdscr.getch()
            stdscr.nodelay(1)  # Restore non-blocking mode

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

    game_loop(stdscr, game_id, player_id)

def game_loop(stdscr, game_id, player_id):
    """Main loop for handling game state updates and user input."""
    while True:
        try:
            game_state = get_game_state(game_id, player_id)
            draw_game_state(stdscr, game_state, player_id)
        except requests.exceptions.RequestException as e:
            # Display a non-blocking error message
            h, w = stdscr.getmaxyx()
            error_msg = f"Connection error: {e}"
            stdscr.addstr(h - 1, 2, error_msg[:w-3]) # Truncate to fit
            stdscr.clrtoeol()

        key = stdscr.getch()
        if key == ord('q'):
            break
        elif key == curses.KEY_ENTER or key in [10, 13]:
            # Get latest command list from server
            game_state = get_game_state(game_id, player_id)
            commands = game_state.get('availableCommands', [])

            # Show command prompt
            cmd_line = get_input(stdscr, "> ")

            # Check if input is a number (shortcut)
            if cmd_line.isdigit():
                cmd_num = int(cmd_line)
                if 1 <= cmd_num <= len(commands):
                    # Convert number to command name from server's list
                    command = commands[cmd_num-1]['name']
                    # Reconstruct cmd_line to handle commands with args later
                    cmd_line = command
            
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
                    # Note: The API expects 'targetID', not 'target_id'
                    args = {'targetID': parts[1]}
                    post_command(game_id, player_id, 'attack', args)
            elif command == 'pass':
                post_command(game_id, player_id, 'pass', {})
            elif command == 'start':
                start_game(game_id)

        time.sleep(0.5) # Refresh rate

if __name__ == "__main__":
    curses.wrapper(main)

import requests
import json
import time

BASE_URL = "http://localhost:8080"

def print_response(res):
    print(f"Status Code: {res.status_code}")
    try:
        print("Response JSON:")
        print(json.dumps(res.json(), indent=2))
    except json.JSONDecodeError:
        print("Response Text:")
        print(res.text)
    print("-"*20)

def main():
    # 1. Create a new game with retries
    print("Creating a new game...")
    game_id = None
    for i in range(5): # Retry up to 5 times
        try:
            res = requests.post(f"{BASE_URL}/games")
            if res.status_code == 201:
                print_response(res)
                game_id = res.json()['id']
                break
            else:
                print(f"Attempt {i+1}: Failed to create game, status code: {res.status_code}")
        except requests.exceptions.ConnectionError as e:
            print(f"Attempt {i+1}: Could not connect to server. Retrying in 2 seconds...")
        time.sleep(2)
    
    if not game_id:
        print("Could not create game after multiple attempts.")
        return

    # 2. Add two players
    print(f"Joining game {game_id} as 'Player 1'...")
    res = requests.post(f"{BASE_URL}/games/{game_id}/join", json={"playerName": "Player 1"})
    print_response(res)
    player_1_id = res.json()['id']

    print(f"Joining game {game_id} as 'Player 2'...")
    res = requests.post(f"{BASE_URL}/games/{game_id}/join", json={"playerName": "Player 2"})
    print_response(res)
    player_2_id = res.json()['id']

    # 3. Get game state to verify players joined
    print(f"Getting game state for {game_id}...")
    res = requests.get(f"{BASE_URL}/games/{game_id}")
    print_response(res)

    # 4. Start the game
    print(f"Starting game {game_id}...")
    res = requests.post(f"{BASE_URL}/games/{game_id}/start")
    print_response(res)

    # 5. Opening Round: Each player plays a secret card
    player_ids = [player_1_id, player_2_id]
    for p_id in player_ids:
        print(f"--- Player {p_id}'s turn (Opening Round) ---")
        # Get player-specific game state to see their hand
        res = requests.get(f"{BASE_URL}/games/{game_id}?playerID={p_id}")
        game_state = res.json()
        player_hand = game_state.get('player_hand')

        # Find a secret card in hand
        secret_card = None
        for card in player_hand:
            if card['type'] == 'Secret':
                secret_card = card
                break
        
        if not secret_card:
            print(f"Player {p_id} has no secret card to play.")
            continue

        # Play the secret card
        print(f"Player {p_id} is playing secret card '{secret_card['name']}' to face_down_1...")
        play_card_data = {
            "playerID": p_id,
            "cardID": secret_card['id'],
            "location": "face_down_1"
        }
        res = requests.post(f"{BASE_URL}/games/{game_id}/play", json=play_card_data)
        print_response(res)

    # 5. Main Game Loop: Players attack each other until one is eliminated.
    print("\n--- Main Game Loop --- ")
    attacker_id, target_id = player_1_id, player_2_id
    
    while True:
        # Get game state to check for winner
        res = requests.get(f"{BASE_URL}/games/{game_id}")
        game_state = res.json()
        if game_state.get('winner'):
            print(f"\nGame over! Winner is {game_state['winner']['name']}")
            break

        print(f"\n--- It's Player {attacker_id}'s turn. --- ")

        # Get player's hand
        res = requests.get(f"{BASE_URL}/games/{game_id}?playerID={attacker_id}")
        hand = res.json().get('player_hand', [])
        warhead = next((c for c in hand if c['type'] == 'Warhead'), None)
        delivery = next((c for c in hand if c['type'] == 'Delivery System'), None)

        if warhead and delivery:
            # Play cards needed for an attack
            print(f"Player {attacker_id} is playing '{delivery['name']}' and '{warhead['name']}'.")
            requests.post(f"{BASE_URL}/games/{game_id}/play", json={'playerID': attacker_id, 'cardID': delivery['id'], 'location': 'face_up'})
            requests.post(f"{BASE_URL}/games/{game_id}/play", json={'playerID': attacker_id, 'cardID': warhead['id'], 'location': 'face_up'})

            # Launch the attack
            print(f"Player {attacker_id} is attacking Player {target_id}.")
            attack_res = requests.post(f"{BASE_URL}/games/{game_id}/attack", json={'attackerID': attacker_id, 'targetID': target_id})
            print_response(attack_res)

            # Check if the target was eliminated and gets a Final Strike
            game_state_after_attack = requests.get(f"{BASE_URL}/games/{game_id}").json()
            if game_state_after_attack['state'] == 'final_strike':
                print(f"\n--- Player {target_id} was eliminated and gets a Final Strike! ---")
                # The eliminated player (target) now attacks back (at the original attacker)
                final_strike_attacker_id = target_id
                final_strike_target_id = attacker_id

                res = requests.get(f"{BASE_URL}/games/{game_id}?playerID={final_strike_attacker_id}")
                hand = res.json().get('player_hand', [])
                fs_warhead = next((c for c in hand if c['type'] == 'Warhead'), None)
                fs_delivery = next((c for c in hand if c['type'] == 'Delivery System'), None)

                if fs_warhead and fs_delivery:
                    print(f"Player {final_strike_attacker_id} is playing '{fs_delivery['name']}' and '{fs_warhead['name']}' for their Final Strike.")
                    requests.post(f"{BASE_URL}/games/{game_id}/play", json={'playerID': final_strike_attacker_id, 'cardID': fs_delivery['id'], 'location': 'face_up'})
                    requests.post(f"{BASE_URL}/games/{game_id}/play", json={'playerID': final_strike_attacker_id, 'cardID': fs_warhead['id'], 'location': 'face_up'})
                    
                    print(f"Player {final_strike_attacker_id} is launching their Final Strike against Player {final_strike_target_id}.")
                    fs_res = requests.post(f"{BASE_URL}/games/{game_id}/attack", json={'attackerID': final_strike_attacker_id, 'targetID': final_strike_target_id})
                    print_response(fs_res)
                else:
                    print(f"Player {final_strike_attacker_id} has no cards for a Final Strike. Passing turn.")
                    # If they can't attack, they must effectively pass. The server should handle this state transition.
                    # This part of the test assumes the server will advance the turn after a final strike attempt.

        else:
            # If no attack cards, pass the turn
            print(f"Player {attacker_id} has no attack cards. Passing turn.")
            pass_res = requests.post(f"{BASE_URL}/games/{game_id}/pass", json={'playerID': attacker_id})
            print_response(pass_res)

        # Swap attacker and target for the next turn
        attacker_id, target_id = target_id, attacker_id

    # Final verification
    print("\n--- Verifying Final Game State ---")
    res = requests.get(f"{BASE_URL}/games/{game_id}")
    final_state = res.json()
    print_response(res)

    assert final_state['state'] == 'finished', f"FAIL: Game state should be 'finished', but is '{final_state['state']}'."
    assert final_state['winner'] is not None, "FAIL: A winner should be declared."
    
    winner = final_state['winner']
    eliminated_player_id = player_1_id if winner['id'] == player_2_id else player_2_id
    eliminated_player = final_state['players'][eliminated_player_id]

    assert eliminated_player['is_eliminated'] == True, f"FAIL: Player {eliminated_player['name']} should be marked as eliminated."
    assert winner['population'] > 0, f"FAIL: Winner {winner['name']} should have population > 0."

    print("\nSUCCESS: Game finished correctly, a winner was declared, and player elimination was verified.")
    print("Test script finished.")

if __name__ == "__main__":
    main()

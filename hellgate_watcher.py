import json
import os
import requests
from PIL import Image, ImageDraw, ImageFont, ImageEnhance
from datetime import datetime,timedelta
import time
import aiohttp,asyncio

BASE_URL_EUROPE = "https://gameinfo-ams.albiononline.com/api/gameinfo"
RENDER_API_URL = "https://render.albiononline.com/v1/item/"


SERVER_URL = BASE_URL_EUROPE
RATE_LIMIT_DELAY_SECONDS = 0.5

IMAGE_FOLDER = "./images"
COVEREDBATTLESJSON="./covered_battle_reports.json"

TIMEOUT=30

LAYOUT = {
    "Bag": (0, 0),
    "Head": (1, 0),
    "Cape": (2, 0),
    
    "MainHand": (0, 1),
    "Armor": (1, 1),
    "OffHand": (2, 1),
    
    "Potion": (0, 2),
    "Shoes": (1, 2),
    "Food": (2, 2)
}
IMAGE_SIZE = 217
CANVAS_SIZE = (3*IMAGE_SIZE, 3*IMAGE_SIZE)

VERBOSELOGGING = False

HEALING_WEAPONS = [
    "MAIN_HOLYSTAFF",
    "2H_HOLYSTAFF",
    "2H_DIVINESTAFF",
    "MAIN_HOLYSTAFF_MORGANA",
    "2H_HOLYSTAFF_HELL",
    "2H_HOLYSTAFF_UNDEAD",
    "MAIN_HOLYSTAFF_AVALON",
    "2H_HOLYSTAFF_CRYSTAL",
    "MAIN_NATURESTAFF",
    "2H_NATURESTAFF",
    "2H_WILDSTAFF",
    "MAIN_NATURESTAFF_KEEPER",
    "2H_NATURESTAFF_HELL",
    "2H_NATURESTAFF_KEEPER",
    "MAIN_NATURESTAFF_AVALON",
    "MAIN_NATURESTAFF_CRYSTAL",
]

async def get_recent_battles(server_url, limit=50,page=1):
    battles = []
    request = f"{server_url}/battles?limit={limit}&sort=recent&offset={page*limit}"

    battle_data = await fetch_response_from_request_url(request, return_json=True)
    if battle_data:
        battles = battle_data
    return battles


def find_10_man_battles (battles):
    sorted_battles = []
    for battle in battles:
        nb_players = len(battle["players"])
        if nb_players == 10:
            sorted_battles.append(battle)
    return sorted_battles

async def find_5v5_battles(battles):
    hellgate_battles = []
    for battle in battles:
        if await is_five_vs_five_battle(battle['id']):
            hellgate_battles.append(battle)
    return hellgate_battles

async def is_five_vs_five_battle(id):
    request = f"{SERVER_URL}/events/battle/{id}"
    battle_events = await fetch_response_from_request_url(request, return_json=True)
    if not battle_events: return False # Handle case where request fails
    for kill in battle_events:
        if kill["groupMemberCount"] != 5:
            return False
    return True

def is_ip_capped():
    pass

def printfile(list_of_battles):
    with open("output.json",'w+') as f:
        f.write(str(list_of_battles))

def get_battle_data(battle):
    if not battle:
        return {"TeamA": [], "TeamB": [], "killers": [], "victims": []}

    team_a_players = {}
    team_b_players = {}
    team_a_ids = set()
    team_b_ids = set()

    # Helper to add a player to a team's dict and id set
    def add_player(player_data, team_dict, team_ids):
        player_id = player_data['Id']
        if player_id not in team_ids:
            team_ids.add(player_id)
            team_dict[player_id] = {
                "id": player_id,
                "name": player_data['Name'],
                "equipment": player_data['Equipment']
            }

    # Initialize with the first event
    first_event = battle[0]
    
    add_player(first_event['Killer'], team_a_players, team_a_ids)
    for participant in first_event['Participants']:
        if participant['Id'] != first_event['Killer']['Id']:
            add_player(participant, team_a_players, team_a_ids)
    add_player(first_event['Victim'], team_b_players, team_b_ids)

    # Process all events to build the full teams
    for event in battle[1:]:
        print_team_names(list(team_a_players.values()))
        print_team_names(list(team_b_players.values()))
        print(event['Killer']['Name'])
        print(event['Victim']['Name'])
        
        killer_id = event['Killer']['Id']
        victim_id = event['Victim']['Id']

        # If we know the killer is on team A, the victim must be on team B
        if killer_id in team_a_ids:
            add_player(event['Victim'], team_b_players, team_b_ids)

        # If we know the killer is on team B, the victim must be on team A
        elif killer_id in team_b_ids:
            add_player(event['Victim'], team_a_players, team_a_ids)

    
    TeamA = list(team_a_players.values())
    TeamB = list(team_b_players.values())

    if len(TeamA) != 5 or len(TeamB) != 5:
        return {"TeamA": [], "TeamB": [], "killers": [], "victims": []}

    killers=[]
    victims=[]
    for event in battle:
        killers.append(event["Killer"]["Name"])
        victims.append(event["Victim"]["Name"])

    TeamA = sort_teams_by_class(TeamA)
    TeamB = sort_teams_by_class(TeamB)

    return {
        "TeamA": TeamA,
        "TeamB": TeamB,
        "killers": killers,
        "victims": victims
    }

async def generate_item_image_from_json(item:dict):
    
    item_image_path = f"{IMAGE_FOLDER}/items/{item["Type"]}&{item["Quality"]}.png"
    
    if VERBOSELOGGING:
        print(f"fetching {item_image_path}")

    if os.path.exists(item_image_path):
        return item_image_path

    request = f"{RENDER_API_URL}{item['Type']}.png?quality={item['Quality']}"
    image = await fetch_response_from_request_url(request, return_json=False)
    if not image: return None # Or a path to a placeholder image

    with open(item_image_path,'wb') as f:
        f.write(image)
    return item_image_path

async def generate_equipment_image_from_json(equipment_json:dict):
    images = {}

    for item_slot,item in equipment_json.items():
        if item is None:
            continue

        image = await generate_item_image_from_json(item)
        if image is None:
            continue
        images[item_slot] = image

    equipment_image = Image.new('RGB',CANVAS_SIZE, (40,40,40,255))

    for item_slot,image in images.items():
        if item_slot in LAYOUT:
            item_image = Image.open(image).convert('RGBA')
            coords = (LAYOUT[item_slot][0]*IMAGE_SIZE, LAYOUT[item_slot][1]*IMAGE_SIZE)
            R, G, B, A = item_image.split()
            equipment_image.paste(item_image,coords,A)

    image_name = ""
    for item_slot,item in equipment_json.items():
        if item is None:
            continue
        image_name += item["Type"]

    equipment_image_path = f"{IMAGE_FOLDER}/equipments/{image_name}.png"

    if VERBOSELOGGING:
        print(f"generating {equipment_image_path}")
    equipment_image.save(equipment_image_path)
    return equipment_image_path
            
async def generate_battle_report_image(battle_events,id):
    
    data = get_battle_data(battle_events)

    if not data:
        return None
    
    # --- New Design Parameters ---
    EQUIPMENT_IMAGE_SIZE = 651
    SIDE_PADDING = 100
    TOP_BOTTOM_PADDING = 50
    SPACING = 30
    MIDDLE_GAP = 200
    PLAYER_NAME_AREA_HEIGHT = 60

    # Calculate canvas dimensions
    CANVAS_WIDTH = (2 * SIDE_PADDING) + (5 * EQUIPMENT_IMAGE_SIZE) + ((5 - 1) * SPACING)
    CANVAS_HEIGHT = (2 * TOP_BOTTOM_PADDING) + (2 * (EQUIPMENT_IMAGE_SIZE + PLAYER_NAME_AREA_HEIGHT)) + MIDDLE_GAP
    CANVAS_SIZE = (CANVAS_WIDTH, CANVAS_HEIGHT)

    battle_report_image = Image.new('RGB', CANVAS_SIZE, (40, 40, 40, 255))
    draw = ImageDraw.Draw(battle_report_image)
    
    try:
        # Using a common font, you might need to provide a path to a .ttf file
        # For bold, use a specific bold font file like 'arialbd.ttf'
        player_name_font = ImageFont.truetype("arialbd.ttf", 40) # Keep player name font size
        timestamp_font = ImageFont.truetype("arial.ttf", 60) # Increased font size for timestamp
    except IOError:
        print("Arial font not found. Using default font.")
        player_name_font = ImageFont.load_default()
        timestamp_font = ImageFont.load_default()

    dead_players = set(data["victims"])

    # --- Draw Team A ---
    y_pos = TOP_BOTTOM_PADDING
    for i, player in enumerate(data["TeamA"]):
        x_pos = SIDE_PADDING + i * (EQUIPMENT_IMAGE_SIZE + SPACING)
        
        # Draw player name
        try:
            # Center the name above the equipment image
            bbox = draw.textbbox((0, 0), player["name"], font=player_name_font)
            text_width = bbox[2] - bbox[0]
            draw.text((x_pos + (EQUIPMENT_IMAGE_SIZE - text_width) / 2, y_pos), player["name"], font=player_name_font, fill=(255, 255, 255))
        except AttributeError: # Fallback for older Pillow versions
            draw.text((x_pos, y_pos), player["name"], font=player_name_font, fill=(255, 255, 255))

        # Paste equipment image
        equipment_image_path = await generate_equipment_image_from_json(player["equipment"])
        equipment_image = Image.open(equipment_image_path).convert('RGBA')

        # Make dead players gray
        if player["name"] in dead_players:
            enhancer = ImageEnhance.Color(equipment_image)
            equipment_image = enhancer.enhance(0.3) 

        R, G, B, A = equipment_image.split()
        battle_report_image.paste(equipment_image, (x_pos, y_pos + PLAYER_NAME_AREA_HEIGHT), A)

    # --- Draw Team B ---
    y_pos = TOP_BOTTOM_PADDING + PLAYER_NAME_AREA_HEIGHT + EQUIPMENT_IMAGE_SIZE + MIDDLE_GAP
    for i, player in enumerate(data["TeamB"]):
        x_pos = SIDE_PADDING + i * (EQUIPMENT_IMAGE_SIZE + SPACING)

        # Draw player name
        try:
            # Center the name above the equipment image
            bbox = draw.textbbox((0, 0), player["name"], font=player_name_font)
            text_width = bbox[2] - bbox[0]
            draw.text((x_pos + (EQUIPMENT_IMAGE_SIZE - text_width) / 2, y_pos), player["name"], font=player_name_font, fill=(255, 255, 255))
        except AttributeError: # Fallback for older Pillow versions
            draw.text((x_pos, y_pos), player["name"], font=player_name_font, fill=(255, 255, 255))

        # Paste equipment image
        equipment_image_path = await generate_equipment_image_from_json(player["equipment"])
        equipment_image = Image.open(equipment_image_path).convert('RGBA')

        # Make dead players gray
        if player["name"] in dead_players:
            enhancer = ImageEnhance.Color(equipment_image)
            equipment_image = enhancer.enhance(0.3)


        R, G, B, A = equipment_image.split()
        battle_report_image.paste(equipment_image, (x_pos, y_pos + PLAYER_NAME_AREA_HEIGHT), A)

    # --- Draw Timestamp ---
    if battle_events:
        # Parse timestamps from the event data
        start_time = datetime.fromisoformat(battle_events[0]['TimeStamp'].replace('Z', '+00:00'))
        end_time = datetime.fromisoformat(battle_events[-1]['TimeStamp'].replace('Z', '+00:00'))
        
        # Calculate duration
        duration = end_time - start_time
        duration_minutes = int(duration.total_seconds() // 60)
        duration_seconds = int(duration.total_seconds() % 60)

        # Format the text strings
        start_time_text = f"Start Time: {start_time.strftime('%H:%M:%S')} UTC"
        duration_text = f"Duration: {duration_minutes:02d}m {duration_seconds:02d}s"

        line_spacing = 20 # Pixels between the two lines of text

        # Calculate text position for centering
        timestamp_y = TOP_BOTTOM_PADDING + PLAYER_NAME_AREA_HEIGHT + EQUIPMENT_IMAGE_SIZE + (MIDDLE_GAP // 2)
        
        # Using textbbox for better centering if available (Pillow >= 8.0.0)
        try:
            start_bbox = draw.textbbox((0, 0), start_time_text, font=timestamp_font)
            start_text_width = start_bbox[2] - start_bbox[0]
            text_height = start_bbox[3] - start_bbox[1] # Height of a single line of text

            # Calculate vertical positions for centered text with spacing
            start_text_y = timestamp_y - (text_height + line_spacing) / 2
            duration_text_y = start_text_y + text_height + line_spacing

            draw.text(((CANVAS_WIDTH - start_text_width) / 2, start_text_y), start_time_text, font=timestamp_font, fill=(255, 255, 255))

            duration_bbox = draw.textbbox((0, 0), duration_text, font=timestamp_font)
            duration_text_width = duration_bbox[2] - duration_bbox[0]
            draw.text(((CANVAS_WIDTH - duration_text_width) / 2, duration_text_y), duration_text, font=timestamp_font, fill=(255, 255, 255))
        except AttributeError: # Fallback for older Pillow versions
            draw.text((CANVAS_WIDTH / 2 - 400, timestamp_y - 50), start_time_text, font=timestamp_font, fill=(255, 255, 255))
            draw.text((CANVAS_WIDTH / 2 - 400, timestamp_y), duration_text, font=timestamp_font, fill=(255, 255, 255))
            
    battle_report_image_path = f"{IMAGE_FOLDER}/battle_reports/battle_report_{id}.png"
    
    print(f"generating {battle_report_image_path}")
    battle_report_image.save(battle_report_image_path)
    return battle_report_image_path

def sort_teams_by_class(team):

    sorted_team = []

    for player in team:
        if player not in sorted_team and "PLATE" in player["equipment"]["Armor"]["Type"] and not "PLATE_ROYAL" in player["equipment"]["Armor"]["Type"] :
            sorted_team.append(player)

    for player in team:
        if player not in sorted_team and "PLATE_ROYAL" in player["equipment"]["Armor"]["Type"]:
            sorted_team.append(player)

    for player in team:
        if player not in sorted_team and "LEATHER" in player["equipment"]["Armor"]["Type"]  and not is_healer(player):
            sorted_team.append(player)

    for player in team:
        if player not in sorted_team and not is_healer(player):
            sorted_team.append(player)

    for player in team:
        if player not in sorted_team and is_healer(player):
            sorted_team.append(player)

    for player in team:
        if player not in sorted_team:
            sorted_team.append(player)
            
    return sorted_team

def is_healer(player):
    weapon = player["equipment"]["MainHand"]["Type"][3:].split('@')[0]
    if weapon in HEALING_WEAPONS:
        return True
    return False

async def get_recent_battle_reports():
    battles = []
    battle_report_paths = []
    exceeded_minute = False
    page_number = 0
    while not exceeded_minute:
        exceeded_minute = battles_exceed_last_minute(battles)
        if exceeded_minute:
            break
        else:
            battles.extend(await get_recent_battles(SERVER_URL,limit=50,page=page_number))
            page_number += 1
            time.sleep(RATE_LIMIT_DELAY_SECONDS)
    
    print(f"Parsed {len(battles)} Battles")
    
    battles = find_10_man_battles(battles)
    print(f"Found {len(battles)} battles with 10 players")

    battles = await find_5v5_battles(battles)
    print(f"Found {len(battles)} 5v5 battles")

    covered_battle_reports = load_covered_battles()
    for battle in battles:
        id = battle["id"]
        if id not in covered_battle_reports:
            covered_battle_reports.append(id)
            battle_events = requests.get(f"{SERVER_URL}/events/battle/{id}").json()
            battle_report_image_path = await generate_battle_report_image(battle_events,id)
            if battle_report_image_path:
                battle_report_paths.append(battle_report_image_path)

    battle_report_paths.sort()

    id = 287202041
    if id not in covered_battle_reports:
        covered_battle_reports.append(id)
        battle_events = requests.get(f"{SERVER_URL}/events/battle/{id}").json()
        battle_report_path = await generate_battle_report_image(battle_events,id)
        if battle_report_path:
            battle_report_paths.append(battle_report_path)

    save_covered_battles(covered_battle_reports)
    return battle_report_paths

def battles_exceed_last_minute(battles):
    if not battles:
        return False
    
    duration = datetime.fromisoformat(battles[0]["startTime"])-datetime.fromisoformat(battles[-1]["startTime"])
    one_minute = timedelta(minutes=1)
    
    if duration < one_minute:
        return False
    else:
        return True
    
def load_covered_battles():
    try:
        with open(COVEREDBATTLESJSON, 'r') as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        return {} 

def save_covered_battles(covered_battles):
    with open(COVEREDBATTLESJSON, 'w') as f:
        json.dump(covered_battles, f, indent=4)

async def fetch_response_from_request_url(url, return_json=True):
    async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=TIMEOUT)) as session:
        try:
            # Initiate the request and manage the response object
            async with session.get(url) as response:
                
                # Check for bad HTTP status codes (e.g., 404, 500)
                # This raises a ClientResponseError if the status is >= 400
                response.raise_for_status()
                
                if return_json:
                    return await response.json()
                else:
                    return await response.read()
                
        except aiohttp.ClientResponseError as e:
            # Handles 4xx and 5xx responses
            print(f"HTTP Error {e.status} for URL {url}: {e.message}")
            return None
        except aiohttp.ClientConnectorError:
            # Handles issues like DNS failures or connection refused
            print(f"Connection Error: Could not connect to {url}")
            return None
        except asyncio.TimeoutError:
            # Handles timeout errors
            print(f"Request timed out after {TIMEOUT} seconds for URL {url}")
            return None
        except aiohttp.ContentTypeError:
            # Handles cases where the response is not valid JSON
            print(f"Failed to parse JSON response from {url}")
            return None
        except Exception as e:
            # Catch all other unexpected errors
            print(f"An unexpected error occurred while fetching {url}: {e}")
            return None

def print_team_names(team):
        print([player["name"]for player in team])



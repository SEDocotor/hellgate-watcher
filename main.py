from bot import bot
import dotenv,os
import hellgate_watcher
import asyncio

dotenv.load_dotenv()
DISCORDTOKEN = os.getenv('DISCORDTOKEN')

async def async_main():
    id = 287328728
    hellgate_watcher.clear_covered_battles()

    kill = await hellgate_watcher.fetch_response_from_request_url('https://gameinfo-ams.albiononline.com/api/gameinfo/events/287377014', return_json=True)
    player = kill["Killer"]
    max_ip = hellgate_watcher.get_max_ip_player(player)
    print(f"player: {player["Name"]} can have up to {max_ip} IP")

    is_capped = hellgate_watcher.is_ip_capped(await hellgate_watcher.get_battle_data_from_id(id))
    print(f"Is IP capped: {is_capped}")
    # await hellgate_watcher.generate_report_image_from_id(id)

def main():
    hellgate_watcher.clear_covered_battles()
    bot.run(DISCORDTOKEN)

if __name__ == "__main__":
    main()
    # asyncio.run(async_main())

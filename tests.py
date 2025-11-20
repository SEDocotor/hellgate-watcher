from src.hellgate_watcher import gen_battle_report_by_id,get_battle,get_max_ip_player
import asyncio
import dotenv
import os

dotenv.load_dotenv()
DISCORDTOKEN = os.getenv('DISCORDTOKEN')

async def main():
    await gen_battle_report_by_id('287823952')
    battle = await get_battle('287823952')
    
if __name__ == "__main__":
    asyncio.run(main())
    
